# Metadata
- Workspace: /workspace-bear-worktree-5b5c6af7-efef-478f-889d-0d7c7a89ec73
- Base Branch: main
- Base Commit: f1657ca (Improve wrapWords unit tests with stronger assertions)

# Task Summary
TASK-08: 앱 오케스트레이션(`app/app.go`)과 프로그램 엔트리포인트(`main.go`)를 구현하여 배너 출력 → 워크스페이스 프롬프트 → 완료 메시지의 전체 흐름을 연결했다.

# Scope Summary
- 이 태스크는 `app.Run` 함수를 구현하여 터미널 폭 조회, 배너 렌더링, bubbletea 워크스페이스 프롬프트 실행, 결과에 따른 확인 메시지 출력을 오케스트레이션하는 것이 목표.
- DONE 기준: `app/app.go`에 `Run` 함수가 구현되고, `main.go`에서 이를 호출하며, 빌드/vet/기존 테스트 17개가 모두 통과.
- OUT OF SCOPE: 새로운 단위 테스트 추가 (이 태스크에는 별도 테스트가 계획되지 않음), CLI 인자 처리, 설정 영속화.

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음
- Build/test commands used: `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./... -v -count=1`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64), bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0, golang.org/x/term v0.40.0
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: `queryTerminalWidth`를 별도 비공개 함수로 분리
  - Rationale: `Run` 함수의 단일 책임 원칙을 준수하고, stdout의 `*os.File` 타입 확인 및 fd 추출 로직을 캡슐화.

- Decision: bubbletea `NewProgram`에 `tea.WithOutput(stderr)`를 전달
  - Rationale: bubbletea의 UI 렌더링 출력은 stderr로 보내고, 배너와 확인 메시지는 stdout으로 보내어 파이프라인 호환성을 유지.

- Decision: `go mod tidy` 실행으로 indirect 의존성을 direct로 승격
  - Rationale: 이전 upstream 태스크들에서 보고된 `// indirect` 경고(bubbles, bubbletea, lipgloss, reflow, x/term)를 해결.

# Invariants (MUST HOLD)
- `app.Run`은 `io.Writer` 인터페이스로 stdout/stderr를 받아야 함.
- 터미널 폭이 `ui.MinTerminalWidth`(23) 미만이면 에러를 반환해야 함.
- 사용자가 Ctrl+C로 취소하면 `nil`을 반환하고 확인 메시지를 출력하지 않아야 함.
- `go build ./...`, `go vet ./...`, `go test ./... -v`가 에러 없이 통과해야 함.

# Prohibited Changes (DO NOT DO)
- `app.Run`의 시그니처 변경 금지 (`func Run(stdout, stderr io.Writer) error`)
- `ui.WorkspacePromptModel`의 공개 API 시그니처 변경 금지
- `ui` 패키지에서 `app` 패키지 직접 import 금지
- `bearArtWidth` 값(22) 변경 금지

# What Changed in the current task
- New/modified files:
  - `app/app.go` (신규 생성, 60줄): `Run`, `queryTerminalWidth` 함수
  - `main.go` (수정): `app.Run` 호출로 엔트리포인트 구현
  - `go.mod` (수정): indirect 의존성 5개를 direct로 승격 (go mod tidy)
  - `go.sum` (수정): go mod tidy에 의한 자동 업데이트
- New/changed public interfaces:
  - `app.Run(stdout, stderr io.Writer) error`
- Behavior changes: 프로그램 실행 시 배너 출력 → 워크스페이스 프롬프트 → 확인 메시지 출력의 전체 흐름이 동작함
- Tests added/updated: 없음 (이 태스크에 단위 테스트가 계획되지 않음)
- Migrations/config changes: 없음

# Verification (Build & Tests)
- `timeout --signal=TERM --kill-after=15s 60 go mod tidy` → 성공
- `timeout --signal=TERM --kill-after=15s 60 go vet ./...` → 성공 (에러/경고 없음)
- `timeout --signal=TERM --kill-after=15s 120 go build ./...` → 성공
- `timeout --signal=TERM --kill-after=15s 90 go test ./... -v -count=1` → 17/17 PASS (app 5개, ui 12개)

# Timeout & Execution Safety
- `go mod tidy`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go build`: `timeout --signal=TERM --kill-after=15s 120` 적용, 정상 완료
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, 정상 완료

# Debug / Temporary Instrumentation
- 없음 (모든 빌드/테스트가 첫 실행에서 통과)

# Guardrails & Interim Report
- 해당 없음 (모든 작업 정상 완료)

# Known Issues / Technical Debt
- 없음. 이전 태스크들에서 보고된 `go.mod` indirect 경고가 이번 `go mod tidy`로 모두 해결됨.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `74bd90f`: `Add app orchestration and main entrypoint`