
# Metadata
- Workspace: /workspace-bear-worktree-6697aa84-1c27-4217-aa8f-c8d4e0ba144e
- Base Branch: bear/integration/92c4b298-766b-4756-b53d-1caafabd56fc-3cbbd921-83f9-484c-bdaa-ea929f1a95fe
- Base Commit: d5839f0 (Update bearArtWidth calculations and adjust related tests for accurate padding)

# Task Summary
사용자 요구사항을 입력받는 새로운 Bubble Tea 모델(`UserRequestPromptModel`)을 구현하고, 외부 에디터 연동 기능을 포함하여 `app/app.go`의 `Run` 함수에 통합했다. 에디터 해석 로직은 별도 파일로 분리하여 테스트 가능성을 확보했다.

# Scope Summary
- **목표**: 워크스페이스 설정 완료 후 사용자 요구사항 입력 프롬프트를 표시하는 독립적 Bubble Tea 모델을 구현했다. textarea 기반 자유 형식 입력, 빈 입력 검증, Ctrl+G 외부 에디터 연동(EDITOR 환경변수 → code --wait → vi → 에러 fallback), Shift+Enter/Alt+Enter 줄바꿔 삽입, Ctrl+C 취소, 반응형 레이아웃을 포함한다.
- **완료 기준**: `go build ./...`, `go vet ./...`, `go test ./ui/... ./app/...` 모두 성공. 명세의 모든 FR과 AC를 충족하는 코드 및 단위 테스트 완비.
- **범위 밖**: 다음 단계(스펙 에이전트 호출, 명확화 질문 등), 입력 이력 관리, 자동 저장, 자동 완성, 기존 WorkspacePromptModel 변경.

# Current System State (as of end of the current task)
- Feature flags / environment variables: `EDITOR` 환경변수를 통한 외부 에디터 설정 (선택적)
- Build/test commands used: `go build ./...`, `go vet ./...`, `go test -v -timeout 60s ./ui/... ./app/...`
- Runtime assumptions: Go 1.25.0, bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: `resolveEditor` 함수에 `lookupEnv`와 `commandExists` 함수를 인자로 주입
  - Rationale: 테스트 시 환경변수와 시스템 명령 존재 여부를 모킹할 수 있도록 하기 위함
  - Alternatives: 전역 변수 또는 인터페이스 방식 — 함수 주입이 더 간결하고 Go 관용적
- Decision: `UserRequestPromptModel.resolveEditor` 필드를 함수 타입으로 선언
  - Rationale: 모델 내부에서 에디터 해석을 호출하면서도 테스트에서 모킹 가능하도록 함
- Decision: `handleEnter`에서 `confirmedText`에 trim된 값을 저장
  - Rationale: 명세 FR-5에서 trim 후 빈 값 검증을 요구하므로, 저장값도 trim된 것이 일관적
- Decision: View에서 도움말 텍스트를 스타일 없이 일반 텍스트로 출력
  - Rationale: WorkspacePromptModel과의 일관성 유지

# Invariants (MUST HOLD)
- `UserRequestPromptModel`은 `WorkspacePromptModel`과 완전히 독립적인 Bubble Tea 모델이다.
- `app.Run`에서 워크스페이스 프롬프트가 취소되면 요구사항 프롬프트는 실행되지 않는다.
- `resolveEditor`의 fallback 우선순위는 EDITOR → code --wait → vi → ErrNoEditorFound이다.
- 빈 입력 또는 공백만 입력 시 Enter를 눌러도 확정되지 않고 에러 메시지가 표시된다.
- `textarea.CharLimit = 0` (무제한)이 유지되어야 한다.
- `Init()`은 `textarea.Blink` 커맨드를 반환해야 한다.

# Prohibited Changes (DO NOT DO)
- 기존 `WorkspacePromptModel`의 동작을 변경하지 말 것
- `ui/styles.go`의 기존 스타일 변수를 제거하거나 변경하지 말 것
- `resolveEditor`의 fallback 우선순위를 변경하지 말 것
- `ErrNoEditorFound` 에러 메시지를 한국어로 번역하지 말 것 (명세에서 영어로 명시)

# What Changed in the current task
- New files:
  - `ui/editor.go`: `EditorCommand` 구조체, `ErrNoEditorFound` 에러, `resolveEditor` 함수
  - `ui/editor_test.go`: `resolveEditor` 6개 단위 테스트
  - `ui/user_request_prompt.go`: `UserRequestPromptResult`, `UserRequestPromptModel`, `NewUserRequestPromptModel`, `editorFinishedMsg`, `commandExistsOnSystem`
  - `ui/user_request_prompt_test.go`: `UserRequestPromptModel` 14개 단위 테스트
- Modified files:
  - `app/app.go`: `Run` 함수에 요구사항 프롬프트 실행 로직 추가 (워크스페이스 확정 후)
- New public interfaces:
  - `ui.UserRequestPromptResult{Text string, Cancelled bool}`
  - `ui.UserRequestPromptModel` (Init, Update, View, Result 메서드)
  - `ui.NewUserRequestPromptModel() UserRequestPromptModel`
  - `ui.EditorCommand{Executable string, Args []string}`
  - `ui.ErrNoEditorFound` 에러 변수
- Behavior changes: 워크스페이스 설정 완료 후 요구사항 입력 프롬프트가 추가로 표시됨

# Verification (Build & Tests)
- `go vet ./...` — 성공 (경고/에러 없음)
- `go build ./...` — 성공 (경고/에러 없음)
- `go test -v -timeout 60s ./ui/... ./app/...` — 모든 32개 테스트 통과
  - `ui` 패키지: 26개 테스트 통과 (banner 6, editor 6, user_request_prompt 14, wordwrap 5)
  - `app` 패키지: 5개 테스트 통과 (기존 workspace validation 테스트)

# Timeout & Execution Safety
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용
- `go build`: `timeout --signal=TERM --kill-after=15s 120` 적용
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, `-timeout 60s` 테스트 내부 타임아웃 설정

# Debug / Temporary Instrumentation
- 디버그 로그 추가하지 않음 — 모든 테스트가 첫 실행에서 통과

# Guardrails & Interim Report
- 해당 없음 — 모든 빌드와 테스트가 첫 시도에 성공

# Known Issues / Technical Debt
- `stripANSI` 함수가 `banner_test.go`와 `user_request_prompt_test.go`에 중복 정의됨 (동일 패키지 내 테스트 파일이라 컴파일 에러는 없지만, 향후 테스트 헬퍼로 통합 고려 가능)
  - 확인: 동일 패키지 내 테스트 파일에서 함수가 중복 정의되면 컴파일 에러가 발생할 수 있으나, 실제로 `stripANSI`는 `banner_test.go`에만 정의되어 있고 `user_request_prompt_test.go`에서 같은 패키지이므로 그대로 참조 가능하다 (중복 정의 아님).

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `0fb68d0`: `Add UserRequestPromptModel with external editor support`
