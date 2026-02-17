# Metadata
- Workspace: /workspace-bear-worktree-3a49c943-ccc9-42d5-bb60-fc44b3f33186
- Base Branch: main
- Base Commit: a833b33
- Task Branch: bear/task/TASK-03-cf807c6a-415f-4ead-ae86-8d334fc480a9

# Task Summary
워크스페이스 경로 검증 순수 함수(`ValidateWorkspacePath`)와 센티널 에러 3종(`ErrRelativePath`, `ErrPathNotExist`, `ErrNotDirectory`)을 `app/workspace.go`에 구현하고, 5가지 시나리오에 대한 단위 테스트를 `app/workspace_test.go`에 작성했다.

# Scope Summary
- 이 태스크는 사용자가 입력한 워크스페이스 경로를 검증하는 순수 비즈니스 로직 함수를 구현하는 것이다.
- DONE 조건: `ValidateWorkspacePath` 함수가 절대 경로 여부, 경로 존재 여부, 디렉토리 여부를 순서대로 검증하고 적절한 센티널 에러를 반환하며, 모든 단위 테스트가 통과한다.
- OUT OF SCOPE: UI 모듈(bubbletea 프롬프트), 배너 렌더링, 앱 오케스트레이션, main 함수 연결 등은 별도 태스크에서 구현한다.

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음
- Build/test commands used: `go vet ./app/...`, `go test ./app/... -v`
- Runtime assumptions: Go 1.25+, Linux
- Any required secrets/credentials handling notes: 없음

# Key Decisions (and rationale)
- Decision: `filepath.IsAbs()`를 사용하여 절대 경로 여부를 판별
  - Rationale: Go 표준 라이브러리에서 제공하는 크로스플랫폼 절대 경로 판별 함수
  - Alternatives: `strings.HasPrefix(path, "/")` — OS 종속적이라 채택하지 않음
- Decision: `os.Stat` 에러 중 `os.ErrNotExist`만 센티널 에러로 변환하고 나머지는 `fmt.Errorf`로 wrap
  - Rationale: 플랜의 pseudocode에서 명시한 대로, not-exist와 기타 에러를 분리 처리
- Decision: 빈 문자열 입력 시 `ErrRelativePath` 반환
  - Rationale: `filepath.IsAbs("")`가 `false`를 반환하므로 자연스럽게 첫 번째 검증에서 잡힘. 플랜의 테스트 케이스에서도 "상대 경로 에러 또는 적절한 에러 반환"이라고 명시

# Invariants (MUST HOLD)
- `ValidateWorkspacePath`는 UI에 의존하지 않는 순수 함수여야 한다.
- 세 센티널 에러의 메시지 문자열은 명세에 정의된 그대로 유지해야 한다.
- 검증 순서: 절대 경로 → 존재 여부 → 디렉토리 여부 (이 순서가 바뀌면 안 됨).

# Prohibited Changes (DO NOT DO)
- `ValidateWorkspacePath`에 UI 의존성(bubbletea, lipgloss 등)을 추가하지 말 것.
- 센티널 에러 메시지 문자열을 변경하지 말 것 (다른 모듈에서 `errors.Is`로 비교에 사용됨).
- `app` 패키지에 `ui` 패키지를 import하지 말 것.

# What Changed in the current task
- New/modified files:
  - `app/workspace.go` (신규 생성)
  - `app/workspace_test.go` (신규 생성)
- New/changed public interfaces:
  - `func ValidateWorkspacePath(path string) error`
  - `var ErrRelativePath error`
  - `var ErrPathNotExist error`
  - `var ErrNotDirectory error`
- Behavior changes: 워크스페이스 경로 검증 기능이 새로 추가됨
- Tests added/updated: 5개 단위 테스트 추가
- Migrations/config changes: 없음

# Verification (Build & Tests)
- `go vet ./app/...` — 통과 (에러 없음)
- `go test ./app/... -v` — 5개 테스트 모두 PASS (0.002s)
  - TestValidateWorkspacePath_AbsoluteDirectoryPath: PASS
  - TestValidateWorkspacePath_RelativePath: PASS
  - TestValidateWorkspacePath_NonExistentPath: PASS
  - TestValidateWorkspacePath_FilePath: PASS
  - TestValidateWorkspacePath_EmptyPath: PASS

# Timeout & Execution Safety
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용
- 타임아웃 이슈 없음

# Debug / Temporary Instrumentation
- 디버그 로그 추가 없음 (모든 테스트가 첫 실행에서 통과)

# Guardrails & Interim Report
- 해당 없음 (첫 실행에서 모든 테스트 통과)

# Known Issues / Technical Debt
- 없음

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `abc1913`: `Add workspace path validation logic`