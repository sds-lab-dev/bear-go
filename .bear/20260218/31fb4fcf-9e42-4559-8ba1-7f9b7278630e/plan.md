# Claude Code 클라이언트 모듈 및 스펙 에이전트 명확화 질문 단계 구현 플랜

## 수정 이력 (Revision Delta)

- 최초 플랜 작성. 사용자 의사결정 3건 반영:
  - UUID v4 생성: `github.com/google/uuid` 사용
  - JSON Schema 검증: Claude Code CLI가 이미 검증하므로 클라이언트 측에서는 `encoding/json` 파싱만 수행
  - 스트리밍 중간 과정 TUI: 전용 BubbleTea 모델로 구현

---

## 1. Overview

### 목표

Bear AI Developer의 핵심 실행 플로우 4번 단계를 구현한다. 구체적으로:

1. Claude Code CLI를 외부 프로세스로 실행하여 스트리밍 통신하는 범용 클라이언트 모듈(`claudecode` 패키지)을 구현한다.
2. 기존 메인 플로우(배너 → 워크스페이스 → 요구사항 입력) 이후, 클라이언트 모듈을 통해 스펙 에이전트에게 명확화 질문을 요청하고, 그 진행 과정과 결과를 BubbleTea TUI로 표시한 뒤 프로그램을 종료한다.
3. `CLAUDE_CODE_API_KEY` 환경변수 검증을 프로그램 시작 시점에 추가한다.

### 비목표

- 명확화 질문에 대한 사용자 답변 수집 (플로우 5번)
- 명확화 질문-답변 반복 루프
- 스펙 드래프트 생성 (플로우 6번 이후)
- CLI 프로세스 타임아웃 처리
- QA_LOG_TEXT 동적 채움 (이번 작업에서는 항상 빈 문자열)

---

## 2. Assumptions & Open Questions

### 가정

- Claude Code CLI의 `stream-json` 출력은 줄바꿈 구분 JSON(NDJSON) 형식이다. 각 줄은 독립적인 JSON 오브젝트이다.
- `stream-json` 이벤트의 `type` 필드 값: `"system"`, `"assistant"`, `"user"`, `"result"`, `"stream_event"`.
- `type: "result"`인 이벤트에는 `subtype` 필드가 있으며, 성공 시 `"success"`, 에러 시 `"error_max_turns"`, `"error_during_execution"`, `"error_max_budget_usd"`, `"error_max_structured_output_retries"` 등의 값을 가진다.
- 성공적인 result 이벤트의 `structured_output` 필드에 JSON Schema에 맞게 검증된 최종 출력이 들어있다. Claude Code CLI가 이미 검증을 수행하므로 클라이언트에서 중복 검증하지 않는다.
- `type: "assistant"` 이벤트에는 `content` 배열이 있으며, 각 요소는 `type` 필드로 `"text"`, `"tool_use"` 등을 구분한다.
- `type: "user"` 이벤트에는 `content` 배열이 있으며, `tool_result` 등의 블록이 포함된다.
- Go 1.25 환경에서 빌드한다.
- `exec.LookPath`가 시스템 PATH에서 바이너리를 탐색하는 데 사용 가능하다 (Rust의 `which::which`에 대응).

### 오픈 질문

- 없음. 모든 의사결정 완료.

---

## 3. Proposed Design

### 아키텍처

```text
main.go
  └─ app.Run()
       ├─ validateAPIKey()          [신규]
       ├─ 배너, 워크스페이스, 요구사항 (기존)
       ├─ claudecode.NewClient()    [신규 패키지]
       │    └─ findClaudeBinary()
       ├─ client.Query()            [스트리밍 CLI 실행]
       │    ├─ stdout → NDJSON 파싱 → 콜백 호출
       │    └─ result → structured_output 반환
       ├─ SpecStreamModel           [신규 BubbleTea 모델]
       │    └─ 스트리밍 메시지 실시간 표시
       └─ 결과 표시 및 종료
```

### 패키지 구조

- `claudecode/` — Claude Code CLI 클라이언트 모듈 (범용, 재사용 가능)
  - `client.go` — Client 구조체, NewClient 생성자, Query 메소드
  - `binary.go` — findClaudeBinary 함수
  - `stream.go` — stream-json 파싱 로직, 이벤트 타입 정의
  - `client_test.go` — Client 테스트
  - `binary_test.go` — 바이너리 탐색 테스트
  - `stream_test.go` — 스트리밍 파싱 테스트
- `ui/` — 기존 패키지에 스트리밍 TUI 모델 추가
  - `spec_stream.go` — 스트리밍 진행 상황 및 결과 표시 BubbleTea 모델
  - `spec_stream_test.go` — 해당 모델 테스트
- `app/` — 기존 패키지에 환경변수 검증 및 플로우 통합
  - `app.go` — Run 함수 수정
  - `env.go` — 환경변수 검증 함수
  - `env_test.go` — 환경변수 검증 테스트

### 핵심 설계 결정

1. **Claude Code 클라이언트는 인터페이스가 아닌 구조체로 구현**: 현재 단일 구현만 존재하므로, 불필요한 추상화를 피하고 구조체로 직접 구현한다. 향후 필요 시 인터페이스를 추출할 수 있다.

2. **세션 관리는 Client 구조체 내부 상태로**: 첫 Query 호출 시 UUID v4를 생성하여 저장하고, 후속 호출에서 재사용한다. 세션 ID 존재 여부로 최초/후속을 구분한다.

3. **스트리밍 콜백은 함수 타입으로**: Query 메소드에 콜백 함수를 전달하여 호출자가 중간 메시지를 처리한다. 콜백 시그니처는 스트리밍 이벤트 구조체를 받는 형태이다.

4. **BubbleTea 모델에서 goroutine 기반 스트리밍 수신**: Query를 별도 goroutine에서 실행하고, channel을 통해 BubbleTea 모델로 이벤트를 전달한다. BubbleTea의 `tea.Cmd` 패턴을 사용하여 이벤트를 수신한다.

5. **시스템 프롬프트와 사용자 프롬프트 템플릿은 상수로**: `app` 패키지에 상수로 정의하되, 사용자 프롬프트의 템플릿 변수 치환은 호출 시 수행한다.

### 에러 처리

- 모든 에러는 Go 관용적 `error` 반환을 사용한다.
- `claudecode` 패키지 내부에서는 도메인 특화 에러 변수(sentinel error)를 정의한다.
- CLI 프로세스 비정상 종료, I/O 에러, JSON 파싱 실패는 각각 구분 가능한 에러로 반환한다.

### 인터페이스 / 계약

**claudecode.Client**:
- 생성자: API Key와 워킹 디렉토리를 받아 Client를 생성. 바이너리 탐색 실패 시 에러 반환.
- Query 메소드: 시스템 프롬프트, 사용자 프롬프트, JSON Schema 문자열, 콜백 함수를 받아 실행. 최종 결과를 `json.RawMessage`로 반환. 에러 시 error 반환.

**콜백 함수 타입**: 스트리밍 이벤트 구조체를 받는 함수. 이벤트 구조체에는 타입(assistant/user)과 원본 JSON 데이터가 포함된다.

**SpecStreamModel**: BubbleTea 모델로, Query 실행 중 중간 메시지를 표시하고, 완료 후 명확화 질문 결과를 표시한다.

---

## 4. Implementation

### TASK-00: Claude Code CLI 바이너리 탐색 및 클라이언트 모듈 기반 구현

**목적**: Claude Code CLI 바이너리 탐색 로직, 스트리밍 이벤트 타입 정의, Client 구조체와 Query 메소드를 구현하는 `claudecode` 패키지를 생성한다. 이 태스크는 전체 Claude Code 클라이언트 모듈의 핵심 기능을 포함한다.

**Dependencies**: none

#### 파일 1: `claudecode/binary.go`

**목적**: Claude Code CLI 바이너리의 시스템 내 위치를 탐색하는 함수를 제공한다.

**새로운 심볼**:
- `findClaudeBinary` (비공개 함수)
- `ErrBinaryNotFound` (공개 에러 변수)
- `absoluteFallbackPaths` (비공개 패키지 레벨 변수)
- `homeRelativeFallbackPaths` (비공개 패키지 레벨 변수)

**편집 의도**:

패키지 레벨 변수로 절대 경로 폴백 목록과 홈 디렉토리 상대 경로 폴백 목록을 정의한다.

절대 경로 폴백 목록 (순서대로):
- `/usr/local/bin/claude`
- `/usr/bin/claude`

홈 디렉토리 상대 경로 폴백 목록 (순서대로):
- `.local/bin/claude`
- `.npm-global/bin/claude`
- `node_modules/.bin/claude`
- `.yarn/bin/claude`
- `.claude/local/claude`

ErrBinaryNotFound는 sentinel 에러로, 바이너리를 찾지 못했을 때 반환한다.

findClaudeBinary 함수를 도입한다. 테스트 가능성을 위해 의존성을 주입받는다.

```pseudocode
FUNCTION findClaudeBinary:
INPUTS: lookPath - PATH에서 실행 파일을 찾는 함수, homeDir - 홈 디렉토리 경로를 반환하는 함수, fileExists - 파일 존재 여부를 확인하는 함수
OUTPUTS: found binary path as string
FAILURE: ErrBinaryNotFound

IF lookPath finds claude in system PATH THEN
    RETURN found path
ENDIF

Retrieve home directory via homeDir
IF home directory is available THEN
    LOOP over each entry in homeRelativeFallbackPaths:
        Build candidate by joining home directory and entry
        IF fileExists confirms candidate exists THEN
            RETURN candidate path
        ENDIF
    ENDLOOP
ENDIF

LOOP over each entry in absoluteFallbackPaths:
    IF fileExists confirms candidate exists THEN
        RETURN candidate path
    ENDIF
ENDLOOP

RETURN ErrBinaryNotFound
```

#### 파일 2: `claudecode/stream.go`

**목적**: stream-json 이벤트 타입 정의와 NDJSON 스트림 파싱 로직을 제공한다.

**새로운 심볼**:
- `StreamEventType` (공개 타입 — 문자열 기반)
- `StreamEventTypeAssistant`, `StreamEventTypeUser`, `StreamEventTypeResult`, `StreamEventTypeSystem`, `StreamEventTypeStreamEvent` (공개 상수)
- `StreamMessage` (공개 구조체)
- `ContentBlock` (공개 구조체)
- `ResultData` (공개 구조체)
- `StreamCallback` (공개 함수 타입)
- `ErrStreamParseFailed` (공개 에러 변수)
- `ErrResultError` (공개 에러 변수)
- `ErrNoResultReceived` (공개 에러 변수)
- `processStream` (비공개 함수)

**편집 의도**:

StreamEventType은 문자열 기반 타입으로, 이벤트 종류를 나타내는 상수들을 정의한다. 값: `"assistant"`, `"user"`, `"result"`, `"system"`, `"stream_event"`.

StreamMessage 구조체를 정의한다. JSON 태그를 사용하여 stream-json 이벤트의 공통 필드를 매핑한다:
- Type 필드: 이벤트 타입
- Subtype 필드: result 이벤트의 서브타입 (문자열)
- IsError 필드: result 이벤트의 에러 여부 (boolean)
- Content 필드: assistant/user 메시지의 콘텐츠 블록 배열
- StructuredOutput 필드: result 이벤트의 최종 출력 (`json.RawMessage`)
- RawJSON 필드: 원본 JSON 라인 (문자열, JSON 태그 `-`)

ContentBlock 구조체는 content 배열의 개별 요소를 나타낸다:
- Type 필드: 블록 타입 (text, tool_use, tool_result 등)
- Text 필드: text 블록의 텍스트 내용
- Name 필드: tool_use 블록의 도구 이름
- Input 필드: tool_use 블록의 입력 (`json.RawMessage`)
- Content 필드: tool_result 등의 중첩 콘텐츠 (문자열)

ResultData 구조체는 Query 메소드의 최종 반환 결과를 나타낸다:
- Output 필드: `json.RawMessage` — structured_output 원본

StreamCallback은 StreamMessage를 받는 함수 타입이다.

에러 변수: ErrStreamParseFailed (JSON 파싱 실패), ErrResultError (result 이벤트가 에러인 경우), ErrNoResultReceived (스트림이 result 없이 종료된 경우).

processStream 함수를 도입한다.

```pseudocode
FUNCTION processStream:
INPUTS: reader for stdout lines, callback of type StreamCallback
OUTPUTS: ResultData
FAILURE: ErrStreamParseFailed or ErrResultError or ErrNoResultReceived

LOOP over each line from reader using scanner:
    Trim whitespace from line
    IF line is empty THEN
        skip to next line
    ENDIF

    Parse line as JSON into StreamMessage
    IF parse fails THEN
        RETURN ErrStreamParseFailed wrapped with line content
    ENDIF

    Store raw JSON line on the message

    IF message type is assistant or user THEN
        Invoke callback with the message
    ENDIF

    IF message type is result THEN
        IF message is error or subtype indicates error THEN
            RETURN ErrResultError wrapped with subtype info
        ENDIF
        IF subtype is success THEN
            Build ResultData from structured output field
            RETURN ResultData
        ENDIF
    ENDIF
ENDLOOP

IF no result was received THEN
    RETURN ErrNoResultReceived
ENDIF
```

#### 파일 3: `claudecode/client.go`

**목적**: Claude Code CLI 클라이언트의 핵심 구조체와 Query 메소드를 제공한다.

**새로운 심볼**:
- `Client` (공개 구조체)
- `NewClient` (공개 생성자 함수)
- `Query` (공개 메소드)
- `ErrProcessStartFailed` (공개 에러 변수)
- `ErrProcessExitError` (공개 에러 변수)

**편집 의도**:

Client 구조체를 정의한다:
- apiKey 필드: 문자열
- workingDir 필드: 문자열
- binaryPath 필드: 문자열 (생성 시 탐색 결과)
- sessionID 필드: 문자열 (첫 Query 시 생성)

NewClient 생성자 함수를 도입한다. 의존성 주입을 위해 바이너리 탐색 함수를 옵션으로 받을 수 있도록 functional options 패턴 대신 간단한 접근을 사용한다: 내부적으로 findClaudeBinary를 호출하되, 테스트 시 binaryPath를 직접 설정할 수 있도록 비공개 필드로 둔다.

```pseudocode
FUNCTION NewClient:
INPUTS: apiKey, workingDir
OUTPUTS: pointer to Client
FAILURE: ErrBinaryNotFound if binary not found

Call findClaudeBinary with default system functions
IF binary not found THEN
    RETURN ErrBinaryNotFound
ENDIF

Build Client with apiKey and workingDir and found binary path
RETURN pointer to Client
```

Query 메소드를 도입한다.

```pseudocode
FUNCTION Query on Client:
INPUTS: systemPrompt, userPrompt, jsonSchema as strings, callback of type StreamCallback
OUTPUTS: ResultData
FAILURE: various errors for process start, IO, parsing, and result errors

Create temp file in OS temp dir for system prompt
Write systemPrompt content to temp file
Defer removal of temp file

Build CLI argument list:
    -p
    --model claude-opus-4-6
    --output-format stream-json
    --verbose
    --include-partial-messages
    --allow-dangerously-skip-permissions
    --permission-mode bypassPermissions
    --tools with the full tool list string
    --append-system-prompt-file with temp file path
    --json-schema with the jsonSchema string

IF sessionID is empty THEN
    Generate new UUID v4 using google/uuid
    Store as sessionID on Client
    Append --session-id with new UUID to arguments
ELSE
    Append --resume with existing sessionID to arguments
ENDIF

Create exec.Command with binary path and arguments
Set command Dir to workingDir
Set command Stdin to a reader from userPrompt string
Set command environment variables:
    Inherit current process env
    Set ANTHROPIC_API_KEY to apiKey
    Set CLAUDE_CODE_EFFORT_LEVEL to high
    Set CLAUDE_CODE_DISABLE_AUTO_MEMORY to 0
    Set CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY to 1

Obtain stdout pipe from command
Obtain stderr pipe from command

IF starting command fails THEN
    RETURN ErrProcessStartFailed
ENDIF

Start goroutine to drain stderr into a buffer

Call processStream with stdout pipe reader and callback
Store result or error

Wait for command to finish
IF command exit code is nonzero AND no stream result error THEN
    RETURN ErrProcessExitError with stderr content
ENDIF

RETURN stream processing result
```

#### 파일 4: `claudecode/binary_test.go`

**목적**: findClaudeBinary 함수의 탐색 로직을 테스트한다.

**새로운 심볼**:
- `TestFindClaudeBinary_FoundInPath`
- `TestFindClaudeBinary_FoundInHomeFallback`
- `TestFindClaudeBinary_FoundInAbsoluteFallback`
- `TestFindClaudeBinary_FallbackOrder`
- `TestFindClaudeBinary_NotFound`
- `TestFindClaudeBinary_NoHomeDir`

**편집 의도**:

findClaudeBinary에 주입 가능한 의존성을 모킹하여 다양한 시나리오를 테스트한다.

테스트 케이스:
- PATH에서 찾은 경우: lookPath가 성공하면 해당 경로 반환 확인
- PATH에 없고 홈 폴백에서 찾은 경우: lookPath 실패, homeDir 성공, 첫 번째 존재하는 홈 상대 경로 반환 확인
- PATH와 홈 폴백 모두 실패, 절대 경로에서 찾은 경우
- 폴백 순서 검증: 여러 경로가 존재할 때 스펙에 명시된 순서 중 첫 번째 경로가 반환되는지 확인
- 모든 경로에서 찾지 못한 경우: ErrBinaryNotFound 반환 확인
- HOME 환경변수가 없는 경우: 홈 폴백을 건너뛰고 절대 경로 폴백으로 넘어가는지 확인

#### 파일 5: `claudecode/stream_test.go`

**목적**: processStream 함수의 NDJSON 파싱 및 이벤트 라우팅 로직을 테스트한다.

**새로운 심볼**:
- `TestProcessStream_AssistantMessageCallsCallback`
- `TestProcessStream_UserMessageCallsCallback`
- `TestProcessStream_ResultMessageNotPassedToCallback`
- `TestProcessStream_SystemMessageIgnored`
- `TestProcessStream_StreamEventIgnored`
- `TestProcessStream_SuccessResultReturnsStructuredOutput`
- `TestProcessStream_ErrorResultReturnsError`
- `TestProcessStream_InvalidJSONReturnsError`
- `TestProcessStream_NoResultReturnsError`
- `TestProcessStream_EmptyLinesSkipped`

**편집 의도**:

strings.NewReader를 사용하여 미리 정의된 NDJSON 문자열을 processStream에 전달하고 결과를 검증한다.

테스트 케이스:
- assistant 메시지: 콜백이 호출되고 메시지 내용이 올바른지 확인
- user 메시지: 콜백이 호출되고 메시지 내용이 올바른지 확인
- result 메시지: 콜백에 전달되지 않는지 확인
- system/stream_event 메시지: 콜백에 전달되지 않는지 확인
- 성공 result: structured_output이 올바르게 파싱되어 ResultData로 반환되는지 확인
- 에러 result (is_error: true): ErrResultError 반환 확인
- 에러 result (subtype가 에러 계열): ErrResultError 반환 확인
- 잘못된 JSON: ErrStreamParseFailed 반환 확인
- result 없이 스트림 종료: ErrNoResultReceived 반환 확인
- 빈 줄: 무시되는지 확인

#### 파일 6: `claudecode/client_test.go`

**목적**: Client 구조체의 세션 관리 로직과 CLI 인자 조립 로직을 테스트한다.

**새로운 심볼**:
- `TestNewClient_BinaryNotFound`
- `TestQuery_SessionIDGeneratedOnFirstCall`
- `TestQuery_SessionIDReusedOnSubsequentCall`
- `TestQuery_EnvironmentVariablesSet`
- `TestQuery_CLIArgumentsCorrect`
- `TestQuery_SystemPromptTempFileCleanup`
- `TestQuery_StdinReceivesUserPrompt`

**편집 의도**:

Client의 내부 로직을 테스트한다. 실제 CLI 프로세스를 실행하지 않고 테스트하기 위해 다음 접근을 사용한다:

- NewClient 테스트: binaryPath를 직접 설정하거나, findClaudeBinary의 의존성을 모킹하여 바이너리 미발견 시 에러 반환 확인
- 세션 관리 테스트: Client에 테스트용 빌더를 제공하여, 빈 sessionID 상태에서 Query 호출 시 UUID가 생성되는지, 두 번째 호출 시 동일 UUID가 재사용되는지 확인. 실제 프로세스 실행 대신, 명령 조립 로직을 별도 비공개 함수로 분리하여 테스트 가능하게 한다.
- CLI 인자/환경변수 테스트: 명령 조립 함수가 올바른 인자와 환경변수를 생성하는지 확인
- 임시 파일 정리 테스트: Query 완료 후 시스템 프롬프트 임시 파일이 삭제되는지 확인

참고: CLI 인자 조립과 환경변수 설정 로직을 별도 비공개 함수 `buildCommand`로 분리하여 테스트 가능성을 높인다. 이 함수는 `client.go`에 위치하며 Query 메소드에서 호출된다.

---

### TASK-01: 환경변수 검증, 스트리밍 TUI 모델, 메인 플로우 통합

**목적**: `CLAUDE_CODE_API_KEY` 환경변수 검증, 스트리밍 진행 과정과 결과를 표시하는 BubbleTea 모델, 그리고 이 모든 것을 기존 메인 플로우에 통합하는 작업을 수행한다.

**Dependencies**: TASK-00

#### 파일 1: `app/env.go`

**목적**: 프로그램 시작 시 필수 환경변수를 검증하는 함수를 제공한다.

**새로운 심볼**:
- `ValidateAPIKeyEnv` (공개 함수)
- `ErrAPIKeyNotSet` (공개 에러 변수)

**편집 의도**:

ErrAPIKeyNotSet은 sentinel 에러로, `CLAUDE_CODE_API_KEY` 환경변수가 없거나 빈 문자열인 경우에 사용한다. 에러 메시지는 "CLAUDE_CODE_API_KEY environment variable is not set or empty"로 한다.

ValidateAPIKeyEnv 함수를 도입한다. 의존성 주입을 위해 환경변수 조회 함수를 파라미터로 받는다.

```pseudocode
FUNCTION ValidateAPIKeyEnv:
INPUTS: lookupEnv - environment variable lookup function
OUTPUTS: API key string value
FAILURE: ErrAPIKeyNotSet

Lookup CLAUDE_CODE_API_KEY via lookupEnv
IF not found or value is empty string THEN
    RETURN ErrAPIKeyNotSet
ENDIF

RETURN the API key value
```

#### 파일 2: `app/env_test.go`

**목적**: ValidateAPIKeyEnv 함수를 테스트한다.

**새로운 심볼**:
- `TestValidateAPIKeyEnv_ValidKey`
- `TestValidateAPIKeyEnv_NotSet`
- `TestValidateAPIKeyEnv_EmptyString`

**편집 의도**:

기존 테스트 패턴(editor_test.go 등)을 따라 의존성 주입으로 테스트한다:
- 유효한 키: 정상 반환 확인
- 환경변수 미설정: ErrAPIKeyNotSet 반환 확인
- 빈 문자열: ErrAPIKeyNotSet 반환 확인

#### 파일 3: `app/prompts.go`

**목적**: 스펙 에이전트의 시스템 프롬프트와 사용자 프롬프트 템플릿을 상수로 정의한다.

**새로운 심볼**:
- `SpecAgentSystemPrompt` (공개 상수)
- `SpecClarificationJSONSchema` (공개 상수)
- `BuildSpecClarificationUserPrompt` (공개 함수)

**편집 의도**:

SpecAgentSystemPrompt 상수에 스펙의 FR-SPEC-2에 명시된 시스템 프롬프트 전문을 저장한다. raw string literal을 사용한다.

SpecClarificationJSONSchema 상수에 FR-SPEC-4에 명시된 JSON Schema 문자열을 저장한다.

BuildSpecClarificationUserPrompt 함수를 도입한다. 이 함수는 사용자 요구사항 원문과 QA 로그 텍스트를 받아 FR-SPEC-3에 명시된 사용자 프롬프트 템플릿의 플레이스홀더를 치환한 최종 문자열을 반환한다.

```pseudocode
FUNCTION BuildSpecClarificationUserPrompt:
INPUTS: originalUserRequest, qaLogText
OUTPUTS: completed prompt string

Take the template string defined in FR-SPEC-3
Replace the placeholder for original user request with originalUserRequest
Replace the placeholder for QA log with qaLogText
RETURN the resulting string
```

#### 파일 4: `ui/spec_stream.go`

**목적**: Claude Code CLI의 스트리밍 출력을 실시간 표시하고, 최종 명확화 질문 결과를 표시하는 BubbleTea 모델을 제공한다.

**새로운 심볼**:
- `SpecStreamResult` (공개 구조체)
- `SpecStreamModel` (공개 구조체)
- `NewSpecStreamModel` (공개 생성자 함수)
- `streamEventMsg` (비공개 커스텀 메시지 타입)
- `streamDoneMsg` (비공개 커스텀 메시지 타입)
- `streamErrorMsg` (비공개 커스텀 메시지 타입)
- `maxDisplayLines` (비공개 상수 — 값 5)
- `AgentActivityStyle` (공개 스타일 변수 — `ui/styles.go`에 추가)
- `QuestionStyle` (공개 스타일 변수 — `ui/styles.go`에 추가)
- `InfoStyle` (공개 스타일 변수 — `ui/styles.go`에 추가)

**편집 의도**:

**ui/styles.go에 추가할 스타일:**

기존 스타일 변수 목록 끝(SuccessStyle 뒤)에 다음 스타일을 추가한다:
- AgentActivityStyle: lipgloss 색상 4(파란색) 전경색. 에이전트 활동 메시지 표시용.
- QuestionStyle: lipgloss 색상 5(마젠타) 전경색, 볼드. 명확화 질문 표시용.
- InfoStyle: lipgloss 색상 6(시안) 전경색. 안내 메시지 표시용.

**SpecStreamResult 구조체:**
- Questions 필드: 문자열 슬라이스 — 명확화 질문 목록
- Err 필드: error — 에러 발생 시
- Cancelled 필드: bool — 사용자 취소 여부

**커스텀 메시지 타입:**
- streamEventMsg: StreamMessage를 감싸는 구조체. 스트리밍 중간 이벤트를 BubbleTea 모델로 전달.
- streamDoneMsg: ResultData를 감싸는 구조체. 스트리밍 완료를 알림.
- streamErrorMsg: error를 감싸는 구조체. 에러 발생을 알림.

**SpecStreamModel 구조체:**
- messages 필드: 문자열 슬라이스 — 현재 표시 중인 메시지 라인들
- result 필드: SpecStreamResult 포인터 — 최종 결과
- done 필드: bool — 완료 여부
- width 필드: int
- ready 필드: bool
- queryFunc 필드: 함수 타입 — Query 실행 함수 (의존성 주입)

**NewSpecStreamModel 생성자:**

queryFunc를 파라미터로 받는다. queryFunc는 콜백과 함께 호출되어 Query를 실행하는 함수이다. 이를 통해 테스트 시 실제 CLI 호출 없이 모킹할 수 있다.

```pseudocode
FUNCTION NewSpecStreamModel:
INPUTS: queryFunc - function that runs query with a callback
OUTPUTS: SpecStreamModel

Build model with queryFunc stored
RETURN model
```

**Init 메소드:**

goroutine에서 Query를 시작하는 tea.Cmd를 반환한다.

```pseudocode
FUNCTION Init on SpecStreamModel:
OUTPUTS: tea.Cmd

RETURN a command that starts the query execution
    The command calls queryFunc with a callback
    The callback sends streamEventMsg to the BubbleTea program
    On success: sends streamDoneMsg with result
    On error: sends streamErrorMsg with error
```

실제로는 channel 기반 접근을 사용한다. Init에서 goroutine을 시작하고 channel에 메시지를 보내며, tea.Cmd가 channel에서 메시지를 하나씩 읽어 BubbleTea 루프로 전달한다. 구체적으로:

Init이 반환하는 tea.Cmd에서:
1. goroutine을 시작하여 queryFunc를 실행한다.
2. queryFunc의 콜백에서 streamEventMsg를 channel에 보낸다.
3. queryFunc 완료 시 streamDoneMsg 또는 streamErrorMsg를 channel에 보낸다.
4. tea.Cmd는 channel에서 첫 번째 메시지를 읽어 반환한다.

**Update 메소드:**

```pseudocode
FUNCTION Update on SpecStreamModel:
INPUTS: msg of type tea.Msg
OUTPUTS: updated model, tea.Cmd

IF msg is tea.WindowSizeMsg THEN
    Update width
    Set ready to true
    RETURN model with no command
ENDIF

IF msg is tea.KeyMsg THEN
    IF key is ctrl+c THEN
        Mark as cancelled in result
        RETURN model with tea.Quit
    ENDIF
ENDIF

IF msg is streamEventMsg THEN
    Extract display text from message content blocks
    Truncate to maxDisplayLines
    Append to messages list keeping only latest entries
    RETURN model with command to read next channel message
ENDIF

IF msg is streamDoneMsg THEN
    Parse structured output as questions JSON
    Store questions in result
    Mark done
    RETURN model with tea.Quit
ENDIF

IF msg is streamErrorMsg THEN
    Store error in result
    Mark done
    RETURN model with tea.Quit
ENDIF

RETURN model with no command
```

메시지 콘텐츠 블록에서 표시 텍스트 추출 로직 (비공개 헬퍼 함수 `formatStreamMessage`):
- assistant 메시지: content 배열을 순회하며, text 블록은 텍스트를, tool_use 블록은 도구 이름을 포함하는 형식 문자열을 생성
- user 메시지: content 배열을 순회하며, tool_result 등의 내용을 형식 문자열로 생성
- 생성된 문자열을 줄 단위로 분할하여 maxDisplayLines(5줄)로 제한

**View 메소드:**

```pseudocode
FUNCTION View on SpecStreamModel:
OUTPUTS: rendered string

IF not ready THEN
    RETURN initializing text
ENDIF

Build output string builder

IF not done THEN
    Render activity header with AgentActivityStyle
    LOOP over messages:
        Render each line with DescriptionStyle
    ENDLOOP
    RETURN built string
ENDIF

IF result has error THEN
    Render error message with ErrorStyle
    RETURN built string
ENDIF

IF result questions list is empty THEN
    Render info message with InfoStyle indicating no clarification needed
    RETURN built string
ENDIF

Render questions header with PromptLabelStyle
LOOP over questions with index:
    Render each question with QuestionStyle prefixed by number
ENDLOOP

RETURN built string
```

**Result 메소드:**

```pseudocode
FUNCTION Result on SpecStreamModel:
OUTPUTS: SpecStreamResult

RETURN stored result
```

#### 파일 5: `ui/spec_stream_test.go`

**목적**: SpecStreamModel의 동작을 테스트한다.

**새로운 심볼**:
- `TestSpecStreamModel_InitStartsQuery`
- `TestSpecStreamModel_StreamEventUpdatesMessages`
- `TestSpecStreamModel_MessagesTruncatedToMaxLines`
- `TestSpecStreamModel_DoneWithQuestionsShowsQuestions`
- `TestSpecStreamModel_DoneWithEmptyQuestionsShowsInfo`
- `TestSpecStreamModel_ErrorShowsErrorMessage`
- `TestSpecStreamModel_CtrlCCancels`
- `TestSpecStreamModel_ViewBeforeReady`

**편집 의도**:

기존 user_request_prompt_test.go의 패턴을 따른다. queryFunc를 모킹하여 실제 CLI 호출 없이 테스트한다.

테스트 케이스:
- Init이 tea.Cmd를 반환하는지 확인
- streamEventMsg 수신 시 messages가 업데이트되는지 확인
- 5줄 초과 메시지가 잘리는지 확인
- streamDoneMsg 수신 시 질문 목록이 결과에 저장되는지 확인
- 빈 질문 배열로 streamDoneMsg 수신 시 올바른 상태 확인
- streamErrorMsg 수신 시 에러가 결과에 저장되는지 확인
- Ctrl+C로 취소 시 Cancelled 플래그 확인
- ready 전 View가 초기화 텍스트를 반환하는지 확인

#### 파일 6: `app/app.go` (기존 파일 수정)

**목적**: 환경변수 검증과 클라이언트 모듈 호출, 스트리밍 TUI 표시를 기존 플로우에 통합한다.

**편집 의도**:

**import 추가**: 기존 import 블록에 `claudecode` 패키지와 `encoding/json` 추가.

**Run 함수 수정 — 환경변수 검증 추가**:

기존 Run 함수의 시작 부분(terminalWidth 조회 전)에 환경변수 검증을 추가한다. 삽입 위치: Run 함수의 첫 번째 줄 (현재 `terminalWidth, err := queryTerminalWidth(stdout)` 전).

```pseudocode
FUNCTION Run modified beginning:
INPUTS: stdout, stderr

Call ValidateAPIKeyEnv with os.LookupEnv
IF error THEN
    RETURN error wrapped with context message
ENDIF

Store returned API key for later use

... existing terminal width check, banner, workspace, request flow ...
```

**Run 함수 수정 — 클라이언트 생성 및 Query 실행 추가**:

기존 사용자 요구사항 결과 확인 (`if requestResult.Cancelled`) 이후, 현재 `return nil` 전에 새로운 로직을 추가한다.

```pseudocode
... after requestResult.Cancelled check ...

Create claudecode Client with apiKey and workspace path from result
IF client creation fails THEN
    RETURN error wrapped with context message
ENDIF

Build user prompt using BuildSpecClarificationUserPrompt with request text and empty QA log

Define queryFunc that captures client and prompts:
    queryFunc accepts a StreamCallback
    Calls client Query with system prompt and user prompt and JSON schema and callback
    RETURN the result

Create SpecStreamModel with queryFunc
Run BubbleTea program with the model
Extract SpecStreamResult

IF result cancelled THEN
    RETURN nil
ENDIF

IF result has error THEN
    RETURN error wrapped with context message
ENDIF

RETURN nil
```

#### 파일 7: `ui/styles.go` (기존 파일 수정)

**목적**: 스트리밍 TUI에서 사용할 새로운 스타일 변수를 추가한다.

**편집 위치**: 기존 SuccessStyle 변수 선언 뒤.

**편집 의도**: 위 `ui/spec_stream.go`에서 설명한 3개 스타일 변수(AgentActivityStyle, QuestionStyle, InfoStyle)를 추가한다.

---

## 5. Testing & Validation

### 단위 테스트

| 테스트 파일 | 테스트 대상 | 주요 검증 내용 |
|---|---|---|
| `claudecode/binary_test.go` | findClaudeBinary | PATH 탐색, 홈 폴백, 절대 폴백, 순서, 미발견 에러 |
| `claudecode/stream_test.go` | processStream | 이벤트 타입 라우팅, 콜백 호출, 결과 파싱, 에러 처리 |
| `claudecode/client_test.go` | Client, Query | 세션 관리, CLI 인자 조립, 환경변수, 임시 파일 정리 |
| `app/env_test.go` | ValidateAPIKeyEnv | 유효/미설정/빈 문자열 케이스 |
| `ui/spec_stream_test.go` | SpecStreamModel | 이벤트 표시, 줄 수 제한, 결과 표시, 에러 표시, 취소 |

### 수동 테스트 체크리스트

1. `CLAUDE_CODE_API_KEY`를 설정하지 않고 프로그램 실행 → 배너 없이 에러 메시지 출력 후 비정상 종료 코드 확인
2. `CLAUDE_CODE_API_KEY`를 빈 문자열로 설정하고 프로그램 실행 → 동일 결과 확인
3. `CLAUDE_CODE_API_KEY`를 유효한 값으로 설정하고 프로그램 실행 → 기존 플로우(배너 → 워크스페이스 → 요구사항) 후 스트리밍 진행 상황 TUI 표시 확인
4. 스트리밍 중 에이전트 활동(텍스트, 도구 사용) 메시지가 실시간으로 업데이트되는지 확인
5. 각 메시지가 최대 5줄로 제한되는지 확인
6. 명확화 질문이 반환되면 질문 목록이 표시되는지 확인
7. 빈 질문 배열이 반환되면 안내 메시지가 표시되는지 확인
8. 질문/안내 표시 후 프로그램이 정상 종료되는지 확인
9. Claude Code CLI 바이너리가 없는 환경에서 적절한 에러 메시지 출력 확인

### 성공/실패 판별

- 모든 단위 테스트: `go test ./...` 통과
- 빌드: `go build ./...` 성공
- 수동 테스트: 위 체크리스트 항목 모두 통과

---

## 6. Risk Analysis

### 기술 리스크

| 리스크 | 영향 | 완화 |
|---|---|---|
| stream-json 이벤트 형식이 문서와 다를 수 있음 | 파싱 실패 | StreamMessage 구조체에 `json.RawMessage` 필드를 두어 유연하게 처리. 미지의 필드는 무시. |
| Claude Code CLI 버전 업데이트로 이벤트 구조 변경 | 호환성 문제 | 최소한의 필드만 파싱하고, 알 수 없는 타입은 무시하는 방어적 파싱 |
| BubbleTea 모델에서 goroutine channel 통신 타이밍 이슈 | 메시지 누락 | buffered channel 사용, 최종 done/error 메시지는 반드시 전달되도록 보장 |
| 시스템 프롬프트 임시 파일 정리 실패 | 디스크 공간 누수 | defer로 정리 보장. OS 임시 디렉토리 사용으로 OS 자체 정리 메커니즘도 활용 |

### 보안 고려사항

- API 키는 환경변수로 전달하며, CLI 인자에 포함하지 않는다 (프로세스 목록 노출 방지).
- 시스템 프롬프트 임시 파일은 OS 기본 임시 디렉토리에 생성되며, Query 완료 후 즉시 삭제한다.

### 롤백 전략

- 이 변경은 기존 플로우의 끝에 새로운 단계를 추가하는 것이므로, 기존 기능에 대한 회귀 위험이 매우 낮다.
- 유일한 기존 파일 수정: `app/app.go`의 Run 함수에 환경변수 검증과 후속 단계 추가, `ui/styles.go`에 스타일 변수 추가.
- 롤백 시: `app/app.go`에서 추가된 코드 블록 제거, `ui/styles.go`에서 추가된 스타일 제거, 신규 파일/패키지 삭제.

---

## 7. Implementation Notes

### 빌드 및 테스트 명령

```
go mod tidy          # 새로운 의존성(google/uuid) 설치
go build ./...       # 전체 빌드
go test ./...        # 전체 테스트
go vet ./...         # 정적 분석
```

### 새로운 외부 의존성

- `github.com/google/uuid` — UUID v4 생성용 (사용자 승인 완료)

### 리포지토리 컨벤션

- 패키지명: 소문자, 단일 단어 또는 짧은 복합어 (`claudecode`, `app`, `ui`)
- 에러 변수: `Err` 접두사 사용, `errors.New`로 정의
- BubbleTea 모델 패턴: 구조체 + `New` 생성자 + `Init`/`Update`/`View` 메소드 + `Result` 메소드
- 커스텀 메시지: 비공개 구조체 (소문자)
- 의존성 주입: 함수 파라미터 또는 구조체 필드로 주입
- 테스트: `testing` 패키지 사용, 헬퍼 함수로 모델 초기화, 의존성 모킹
- View에서 `ready` 플래그 체크: WindowSizeMsg 수신 전까지 초기화 텍스트 표시
- 에러 래핑: `fmt.Errorf("context: %w", err)` 패턴

### 파일별 영향 범위 요약

| 파일 | 액션 | 설명 |
|---|---|---|
| `claudecode/binary.go` | 신규 | 바이너리 탐색 |
| `claudecode/stream.go` | 신규 | 스트리밍 이벤트 타입 및 파싱 |
| `claudecode/client.go` | 신규 | Client 구조체, Query 메소드 |
| `claudecode/binary_test.go` | 신규 | 바이너리 탐색 테스트 |
| `claudecode/stream_test.go` | 신규 | 스트리밍 파싱 테스트 |
| `claudecode/client_test.go` | 신규 | Client 테스트 |
| `app/env.go` | 신규 | 환경변수 검증 |
| `app/env_test.go` | 신규 | 환경변수 검증 테스트 |
| `app/prompts.go` | 신규 | 프롬프트 상수 및 빌더 |
| `ui/spec_stream.go` | 신규 | 스트리밍 TUI 모델 |
| `ui/spec_stream_test.go` | 신규 | 스트리밍 TUI 테스트 |
| `ui/styles.go` | 수정 | 3개 스타일 변수 추가 |
| `app/app.go` | 수정 | 환경변수 검증 + 클라이언트 + TUI 통합 |
| `go.mod` | 수정 | google/uuid 의존성 추가 |