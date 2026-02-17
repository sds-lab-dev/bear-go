# Metadata
- Workspace: /workspace-bear-worktree-7fa6f02e-e3fe-4ced-a580-5255feaa64f9
- Base Branch: main
- Base Commit: f79c2ce (Add TUI dependencies for Bear AI Developer)

# Task Summary
TASK-00: 프로젝트 의존성 설치. bubbletea v1, bubbles v1, lipgloss v1, reflow, golang.org/x/term 패키지를 Go 모듈 의존성으로 추가하여 후속 태스크에서 사용 가능하도록 함.

# Scope Summary
- 이 태스크는 후속 TUI 구현 태스크들이 사용할 Go 모듈 의존성을 설치하는 것이 목표.
- DONE 기준: 모든 필수 패키지(bubbletea, bubbles, lipgloss, reflow, x/term)가 go.mod에 등록되고, go.sum에 체크섬이 존재하며, `go list`로 import 가능 확인.
- OUT OF SCOPE: 실제 Go 소스 코드 작성, UI/비즈니스 로직 구현, 테스트 코드 작성.

# Current System State
- Feature flags / environment variables: 없음
- Build/test commands used: `go get`, `go mod tidy`, `go mod download`, `go list`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64), 네트워크 접근 가능
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: 누락된 전이 의존성 `github.com/atotto/clipboard` v0.1.4 추가
- Rationale: `bubbles/textarea` 패키지가 `atotto/clipboard`를 import하는데, 기존 go.sum에 해당 모듈의 체크섬이 누락되어 `go list`에서 에러 발생. 후속 태스크에서 textarea를 사용할 때 컴파일 실패를 방지하기 위해 추가.
- Alternatives: 없음. 이 의존성은 필수적이며 누락은 빌드 실패를 유발.

# Invariants (MUST HOLD)
- go.mod에 다음 모듈이 반드시 존재해야 함:
  - github.com/charmbracelet/bubbletea v1.3.10
  - github.com/charmbracelet/bubbles v1.0.0
  - github.com/charmbracelet/lipgloss v1.1.0
  - github.com/muesli/reflow v0.3.0
  - golang.org/x/term v0.40.0
  - github.com/atotto/clipboard v0.1.4
- go.sum에 위 모듈들의 체크섬이 존재해야 함.
- `go list` 명령으로 위 패키지들이 정상적으로 resolve 되어야 함.

# Prohibited Changes (DO NOT DO)
- go.mod의 모듈 경로(`github.com/sds-lab-dev/bear-go`) 변경 금지
- Go 버전(`go 1.25.0`) 변경 금지
- 기존 의존성 버전 다운그레이드 금지
- 현재 시점에서 `go mod tidy` 실행 금지 (Go 소스 파일이 없어서 모든 의존성이 제거됨)

# What Changed in the current task
- New/modified files:
  - `go.mod`: `github.com/atotto/clipboard v0.1.4 // indirect` 추가
  - `go.sum`: `github.com/atotto/clipboard v0.1.4` 체크섬 2줄 추가
- New/changed public interfaces: 없음
- Behavior changes: 없음 (의존성 설정만)
- Tests added/updated: 없음 (이 태스크에 테스트 없음)
- Migrations/config changes: 없음

# Verification (Build & Tests)
- 실행한 명령: `go list github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles/textarea github.com/charmbracelet/lipgloss github.com/muesli/reflow/wordwrap golang.org/x/term`
- 결과: 5개 패키지 모두 정상 resolve (에러 없음)
- 이 태스크에는 빌드할 Go 소스 코드가 없으므로 `go build` 실행 불가. 대신 `go list`로 모든 의존성의 import 가능 여부를 검증함.

# Timeout & Execution Safety
- `go get`, `go mod download`, `go list` 명령에 `timeout --signal=TERM --kill-after=15s` 적용 (30~60초)
- 모든 명령이 타임아웃 없이 정상 완료됨.

# Debug / Temporary Instrumentation
- 없음.

# Guardrails & Interim Report
- 해당 없음. 모든 작업 정상 완료.

# Known Issues / Technical Debt
- 모든 의존성이 `// indirect`로 표시됨. 이는 현재 프로젝트에 Go 소스 파일이 없기 때문이며, 후속 태스크에서 import하면 자동으로 direct dependency로 전환됨.
- `go mod tidy`를 실행하면 소스 파일이 없어서 모든 의존성이 제거됨. 후속 태스크에서 소스 파일 작성 후에 tidy를 실행해야 함.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `0ab738f`: `Add missing transitive dependency for bubbles/textarea`