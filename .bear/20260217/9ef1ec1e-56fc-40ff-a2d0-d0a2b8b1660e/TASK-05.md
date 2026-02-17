# Metadata
- Workspace: /workspace-bear-worktree-c4e01efd-3a98-4dc6-96d3-935914285e04
- Base Branch: main
- Base Commit: bdcb955 (Fix bearArtWidth off-by-one and add alignment test)

# Task Summary
TASK-05: bubbletea inline 모드로 동작하는 워크스페이스 경로 입력 프롬프트(`WorkspacePromptModel`)를 `ui/workspace_prompt.go`에 구현했다.

# Scope Summary
- 이 태스크는 bubbletea Model 인터페이스를 구현하는 `WorkspacePromptModel` 구조체, `NewWorkspacePromptModel` 생성 함수, `WorkspacePromptResult` 결과 구조체, `Result()` 메서드를 구현하는 것이 목표.
- DONE 기준: `WorkspacePromptModel`이 `tea.Model` 인터페이스를 구현하고, Enter로 확인/Alt+Enter로 줄바꿈/Ctrl+C로 취소를 처리하며, 경로 검증 함수를 주입받아 사용하고, 빌드와 기존 테스트가 모두 통과.
- OUT OF SCOPE: 앱 오케스트레이션(TASK-6), 테스트 코드(이 태스크에는 별도 단위 테스트가 계획에 포함되지 않음), main 함수 연결.

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음
- Build/test commands used: `go build ./ui/...`, `go vet ./ui/...`, `go test ./... -v`
- Runtime assumptions: Go 1.25+, bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: `textarea.KeyMap.InsertNewline.SetEnabled(false)`로 Enter 키 바인딩을 비활성화
  - Rationale: Enter를 제출 키로 사용하기 위해 textarea의 기본 Enter→줄바꿈 동작을 비활성화. Alt+Enter에서 `InsertString("\n")`으로 줄바꿈을 직접 삽입.
  - Alternatives: `key.NewBinding()`으로 빈 키로 교체 — SetEnabled이 더 명시적이라 채택.

- Decision: `msg.String()` 패턴 매칭으로 키 이벤트 처리
  - Rationale: bubbletea v1 공식 가이드에서 권장하는 방식 중 하나로, "ctrl+c", "enter", "alt+enter" 등의 문자열로 직관적 매칭 가능.

- Decision: `handleKey`, `handleEnter`로 메서드를 분리
  - Rationale: CLAUDE.md 코딩 컨벤션에 따라 함수 단일 책임 원칙을 준수하고 Update 메서드의 중첩 깊이를 최소화.

# Invariants (MUST HOLD)
- `WorkspacePromptModel`은 `tea.Model` 인터페이스를 구현해야 함 (Init, Update, View 메서드).
- `validatePath` 함수는 외부에서 주입되어야 하며, UI 모듈이 비즈니스 로직을 직접 참조하지 않아야 함 (NFR-01).
- Enter 입력 시 빈 값이면 `currentDir`을 확정, 비어있지 않으면 검증 후 확정.
- Ctrl+C 입력 시 `tea.Quit`을 반환하고 `confirmed`는 false로 남아 취소 상태를 나타냄.
- `go build ./ui/...`와 `go vet ./ui/...`가 에러 없이 통과해야 함.

# Prohibited Changes (DO NOT DO)
- `WorkspacePromptModel`의 공개 API(NewWorkspacePromptModel, Result, WorkspacePromptResult) 시그니처 변경 금지 (TASK-6에서 이 API를 사용함)
- `validatePath` 필드를 구체 타입으로 변경 금지 (함수 타입으로 유지하여 의존성 주입 패턴 보존)
- `ui` 패키지에서 `app` 패키지를 직접 import 금지

# What Changed in the current task
- New/modified files:
  - `ui/workspace_prompt.go` (신규 생성, 136줄)
- New/changed public interfaces:
  - `ui.WorkspacePromptModel` (struct, tea.Model 구현)
  - `ui.NewWorkspacePromptModel(currentDir string, validatePath func(string) error) WorkspacePromptModel`
  - `ui.WorkspacePromptResult` (struct: Path string, Cancelled bool)
  - `(WorkspacePromptModel).Result() WorkspacePromptResult`
- Behavior changes: 워크스페이스 경로 입력 프롬프트 UI 모델이 새로 추가됨
- Tests added/updated: 없음 (이 태스크에 단위 테스트가 계획되지 않음; bubbletea Model은 수동 테스트 대상)
- Migrations/config changes: 없음

# Verification (Build & Tests)
- `go build ./ui/...` — 성공 (에러 없음)
- `go vet ./ui/...` — 성공 (에러 없음)
- `go test ./... -v` — 17개 테스트 모두 PASS (app 5개, ui 12개)
- 이 태스크 자체의 단위 테스트는 계획에 포함되지 않음 (bubbletea Model은 터미널 상호작용이 필요하여 수동 테스트 체크리스트 대상).

# Timeout & Execution Safety
- `go build`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, 정상 완료

# Debug / Temporary Instrumentation
- 없음 (모든 빌드/테스트가 첫 실행에서 통과)

# Guardrails & Interim Report
- 해당 없음 (모든 작업 정상 완료)

# Known Issues / Technical Debt
- go.mod에서 `github.com/charmbracelet/bubbles`와 `github.com/charmbracelet/bubbletea`가 indirect로 마킹되어 있어 "should be direct" 경고 발생. TASK-6에서 `go mod tidy` 실행 시 자동 해결.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `b2859dd`: `Add bubbletea workspace prompt model`