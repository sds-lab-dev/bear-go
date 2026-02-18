# Metadata
- Workspace: /workspace-bear-worktree-9031592a-3c9c-4b2b-a74a-83af81476ab2
- Base Branch: bear/integration/31fb4fcf-9e42-4559-8ba1-7f9b7278630e-c4862f2d-786e-42a8-a8c6-2195c5981080
- Base Commit: 9939caf (Add implementation report for TASK-00)

# Task Summary
TASK-01: `CLAUDE_CODE_API_KEY` 환경변수 검증, 스트리밍 진행 과정과 결과를 표시하는 BubbleTea 모델, 그리고 이 모든 것을 기존 메인 플로우에 통합하는 작업을 구현했다.

# Scope Summary
- 이 태스크는 프로그램 시작 시 `CLAUDE_CODE_API_KEY` 환경변수 검증, 스펙 에이전트 시스템/사용자 프롬프트 상수 정의, BubbleTea 기반 스트리밍 TUI 모델 구현, 그리고 기존 메인 플로우에 이들을 통합하는 것이다.
- DONE 기준: `go test ./app/... ./ui/...` 모든 테스트 통과 (app: 8개, ui: 40개), `go build ./...` 빌드 성공, `go vet ./...` 통과.
- OUT OF SCOPE: 명확화 질문에 대한 사용자 답변 수집, 질문-답변 반복 루프, 스펙 드래프트 생성, QA_LOG_TEXT 동적 채움 (이번 작업에서는 항상 빈 문자열).

# Current System State (as of end of the current task)
- Feature flags / environment variables: `CLAUDE_CODE_API_KEY` 필수 환경변수 추가 (미설정 시 프로그램 시작 불가)
- Build/test commands used: `go build ./...`, `go test ./app/... ./ui/... -v -timeout 60s`, `go vet ./...`
- Runtime assumptions: Go 1.25, Linux. Claude Code CLI 바이너리가 PATH 또는 폴백 경로에 존재해야 함.
- Secrets/credentials: `CLAUDE_CODE_API_KEY`는 환경변수로 전달되며 CLI 인자에 포함하지 않음 (보안).

# Key Decisions (and rationale)
- Decision: `ValidateAPIKeyEnv`에 `lookupEnv` 함수를 DI로 주입
  - Rationale: 플랜에 명시. 테스트에서 환경변수 모킹 가능.

- Decision: `SpecStreamModel`에서 buffered channel(64) 기반 메시지 전달
  - Rationale: goroutine에서 queryFunc를 실행하고 channel을 통해 BubbleTea 루프로 이벤트 전달. buffered channel로 타이밍 이슈 방지.

- Decision: 시스템 프롬프트에서 backtick 포함 부분은 Go string concatenation으로 처리
  - Rationale: raw string literal 내에 backtick을 포함할 수 없으므로 해당 부분만 일반 문자열 연결.

- Decision: `formatStreamMessage`가 content block 타입별로 표시 텍스트 생성
  - Rationale: text, tool_use, tool_result 등 다양한 블록 타입을 사용자에게 의미 있게 표시.

# Invariants (MUST HOLD)
- `CLAUDE_CODE_API_KEY` 환경변수 검증은 배너 출력/TUI 인터랙션보다 먼저 수행되어야 한다.
- `SpecStreamModel`의 메시지 표시는 maxDisplayLines(5)줄을 초과하지 않아야 한다.
- `SpecStreamResult`의 Cancelled/Err/Questions 중 하나만 의미 있는 상태여야 한다.
- `BuildSpecClarificationUserPrompt`의 템플릿 플레이스홀더가 정확히 치환되어야 한다.

# Prohibited Changes (DO NOT DO)
- `claudecode` 패키지의 공개 인터페이스를 변경하지 말 것 (TASK-00에서 확립).
- `ValidateAPIKeyEnv`의 환경변수 이름 `CLAUDE_CODE_API_KEY`를 변경하지 말 것.
- `SpecAgentSystemPrompt`, `SpecClarificationJSONSchema`의 내용을 스펙(FR-SPEC-2, FR-SPEC-4)과 다르게 변경하지 말 것.
- `maxDisplayLines` 상수(5)를 임의로 변경하지 말 것.

# What Changed in the current task
- New files:
  - `app/env.go`: ValidateAPIKeyEnv 함수, ErrAPIKeyNotSet sentinel 에러
  - `app/env_test.go`: 3개 테스트 (ValidKey, NotSet, EmptyString)
  - `app/prompts.go`: SpecAgentSystemPrompt 상수, SpecClarificationJSONSchema 상수, BuildSpecClarificationUserPrompt 함수
  - `ui/spec_stream.go`: SpecStreamResult 구조체, SpecStreamModel 구조체, NewSpecStreamModel 생성자, streamEventMsg/streamDoneMsg/streamErrorMsg 메시지 타입, formatStreamMessage 헬퍼, parseQuestions 헬퍼
  - `ui/spec_stream_test.go`: 8개 테스트
- Modified files:
  - `app/app.go`: Run 함수에 환경변수 검증 추가 (시작 부분), claudecode.NewClient/Query 호출 및 SpecStreamModel 실행 (요구사항 입력 후)
  - `ui/styles.go`: AgentActivityStyle, QuestionStyle, InfoStyle 3개 스타일 변수 추가
- New public interfaces:
  - `app.ValidateAPIKeyEnv(lookupEnv func(string) (string, bool)) (string, error)`
  - `app.ErrAPIKeyNotSet` error variable
  - `app.SpecAgentSystemPrompt` constant
  - `app.SpecClarificationJSONSchema` constant
  - `app.BuildSpecClarificationUserPrompt(originalUserRequest, qaLogText string) string`
  - `ui.SpecStreamResult` struct
  - `ui.SpecStreamModel` struct
  - `ui.NewSpecStreamModel(queryFunc) SpecStreamModel`
  - `ui.AgentActivityStyle`, `ui.QuestionStyle`, `ui.InfoStyle` style variables

# Verification (Build & Tests)
- 실행한 명령:
  - `go build ./...` — 성공
  - `go vet ./...` — 통과
  - `go test ./app/... ./ui/... -v -timeout 60s` — 전체 통과
- app 패키지: 8개 테스트 PASS (0.002s)
- ui 패키지: 40개 테스트 PASS (0.006s)

# Timeout & Execution Safety
- 빌드: `timeout --signal=TERM --kill-after=15s 60` 적용
- 테스트: `timeout --signal=TERM --kill-after=15s 90` 외부 + `go test -timeout 60s` 내부 타임아웃 적용
- 모든 명령이 정상 시간 내 완료됨.

# Debug / Temporary Instrumentation
- 디버그 로그 추가 없음. 첫 실행에서 모든 테스트 통과.

# Guardrails & Interim Report
- 해당 없음. 1번의 빌드/테스트 반복으로 전체 통과.

# Known Issues / Technical Debt
- 없음.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `f49477d`: `feat: Add env validation, streaming TUI model, and main flow integration`