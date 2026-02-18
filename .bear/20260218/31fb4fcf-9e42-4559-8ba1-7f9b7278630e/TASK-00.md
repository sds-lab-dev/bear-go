# Metadata
- Workspace: /workspace-bear-worktree-f712fdf8-375c-4b00-9933-b7400741f895
- Base Branch: bear/integration/31fb4fcf-9e42-4559-8ba1-7f9b7278630e-c4862f2d-786e-42a8-a8c6-2195c5981080
- Base Commit: 279668d (Merge branch 'bear/integration/92c4b298-766b-4756-b53d-1caafabd56fc-3cbbd921-83f9-484c-bdaa-ea929f1a95fe')

# Task Summary
TASK-00: Claude Code CLI 바이너리 탐색 및 클라이언트 모듈 기반 구현. `claudecode` 패키지를 새로 생성하여 바이너리 탐색 로직, 스트리밍 이벤트 타입 정의, Client 구조체와 Query 메소드를 구현했다.

# Scope Summary
- 이 태스크는 `claudecode` 패키지의 3개 소스 파일(binary.go, stream.go, client.go)과 3개 테스트 파일(binary_test.go, stream_test.go, client_test.go)을 구현하는 것이다.
- DONE 기준: `go test ./claudecode/...` 모든 22개 테스트 통과, `go build ./claudecode/...` 빌드 성공, `go vet ./claudecode/...` 통과, google/uuid 의존성 go.mod에 추가.
- OUT OF SCOPE: 환경변수 검증(TASK-01), TUI 스트리밍 모델(TASK-01), 메인 플로우 통합(TASK-01), app/prompts.go(TASK-01), ui/styles.go 수정(TASK-01).

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음 (이 태스크에서는 환경변수 검증을 구현하지 않음)
- Build/test commands used: `go build ./...`, `go test ./claudecode/... -v`, `go vet ./...`
- Runtime assumptions: Go 1.25, Linux. /bin/echo, /bin/sh 사용 가능한 환경에서 테스트 실행.
- Secrets/credentials: google/uuid 패키지는 외부 의존성이지만 자격증명 불필요.

# Key Decisions (and rationale)
- Decision: `Query`를 `Client` 구조체의 메소드로 구현
  - Rationale: 플랜에 "Query (공개 메소드)"로 명시. Client의 상태(sessionID)를 변경하므로 메소드가 자연스러움.
  - Alternatives: 패키지 레벨 함수 → 초기 구현 후 플랜에 맞춰 메소드로 변경함.

- Decision: `buildCommand`를 비공개 함수로 분리하여 테스트 용이성 확보
  - Rationale: 플랜에 명시적으로 "buildCommand 비공개 함수로 분리하여 테스트"라고 지정됨.
  - Alternatives: 없음.

- Decision: `findClaudeBinary`에 의존성 주입(lookPath, homeDir, fileExists) 적용
  - Rationale: 플랜에 명시. 테스트에서 모킹 가능.

- Decision: processStream에서 에러 subtype를 map으로 관리
  - Rationale: 4개의 에러 subtype을 깔끔하게 검사하기 위해. 확장 가능.

# Invariants (MUST HOLD)
- `findClaudeBinary`는 PATH → 홈 상대 폴백 → 절대 폴백 순서를 반드시 유지해야 한다.
- `processStream`은 assistant/user 타입만 콜백에 전달하고, result는 콜백에 전달하지 않아야 한다.
- 첫 번째 Query 호출 시 --session-id, 두 번째 이후는 --resume을 사용해야 한다.
- 시스템 프롬프트 임시 파일은 Query 완료 후 반드시 삭제되어야 한다.
- 환경변수 ANTHROPIC_API_KEY는 CLI 인자가 아닌 환경변수로 전달해야 한다 (보안).

# Prohibited Changes (DO NOT DO)
- `claudecode` 패키지의 공개 인터페이스(Client, Query, StreamMessage, StreamCallback, ResultData 등)를 임의로 변경하지 말 것. TASK-01에서 이 인터페이스를 사용한다.
- ErrBinaryNotFound, ErrStreamParseFailed, ErrResultError, ErrNoResultReceived, ErrProcessStartFailed, ErrProcessExitError의 이름이나 의미를 변경하지 말 것.
- toolList 상수의 도구 목록을 임의로 변경하지 말 것.
- homeRelativeFallbackPaths, absoluteFallbackPaths의 순서를 변경하지 말 것.

# What Changed in the current task
- New files:
  - `claudecode/binary.go`: findClaudeBinary 함수, ErrBinaryNotFound, 폴백 경로 목록
  - `claudecode/stream.go`: StreamEventType, StreamMessage, ContentBlock, ResultData, StreamCallback, processStream, 에러 변수
  - `claudecode/client.go`: Client 구조체, NewClient, Query 메소드, buildCommand, ErrProcessStartFailed, ErrProcessExitError
  - `claudecode/binary_test.go`: 6개 테스트
  - `claudecode/stream_test.go`: 10개 테스트
  - `claudecode/client_test.go`: 7개 테스트
- Modified files:
  - `go.mod`: github.com/google/uuid v1.6.0 의존성 추가
  - `go.sum`: 해당 checksum 추가
- New public interfaces:
  - `claudecode.Client` struct
  - `claudecode.NewClient(apiKey, workingDir string) (*Client, error)`
  - `(*Client).Query(systemPrompt, userPrompt, jsonSchema string, callback StreamCallback) (ResultData, error)`
  - `claudecode.StreamMessage`, `claudecode.ContentBlock`, `claudecode.ResultData` structs
  - `claudecode.StreamCallback` type
  - `claudecode.StreamEventType` type with 5 constants
  - 5 public error variables

# Verification (Build & Tests)
- 실행한 명령: `go build ./... && go vet ./... && go test ./... -timeout 60s`
- 결과: 전체 빌드 성공, vet 통과, 전체 테스트 통과 (app, claudecode, ui 패키지 모두)
- claudecode 패키지: 22개 테스트 모두 PASS (0.004s)

# Timeout & Execution Safety
- 테스트 실행 시 `timeout --signal=TERM --kill-after=15s 90` 외부 타임아웃과 `go test -timeout 60s` 내부 타임아웃을 동시 적용함.
- 모든 테스트는 0.004s 이내에 완료되어 타임아웃 이슈 없음.

# Debug / Temporary Instrumentation
- 디버그 로그 추가 없음. 첫 실행에서 2개 테스트 실패 후 테스트 접근 방식을 수정하여 2번째 실행에서 전체 통과.

# Guardrails & Interim Report
- 해당 없음. 2번의 테스트 반복으로 전체 통과.

# Known Issues / Technical Debt
- 없음.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `9e5988f`: `feat: Add claudecode package for Claude Code CLI client`