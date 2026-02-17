# Metadata
- Workspace: /workspace-bear-worktree-67de9f9d-8965-48d6-85a2-d45340ca85ac
- Base Branch: main
- Base Commit: 3ef2e31 (Add lipgloss style definitions for TUI components)

# Task Summary
TASK-02: 단어 단위 줄바꿈 유틸리티. `ui/wordwrap.go`에 `wrapWords` 비공개 함수를 구현하여, 주어진 텍스트를 최대 폭에 맞게 단어 단위로 줄바꿈하고 문자열 슬라이스로 반환.

# Scope Summary
- 이 태스크는 슬로건 배너의 오른쪽 영역 텍스트를 단어 단위로 줄바꿈하는 유틸리티 함수를 구현하는 것이 목표.
- DONE 기준: `wrapWords(text string, maxWidth int) []string` 함수가 구현되고, 빈 텍스트/0 이하 폭 시 빈 슬라이스 반환, 정상 텍스트 시 줄바꿈된 문자열 슬라이스 반환, 단위 테스트 통과.
- OUT OF SCOPE: 배너 렌더링(TASK-3), 스타일 적용, 다른 UI 컴포넌트 구현.

# Current System State
- Feature flags / environment variables: 없음
- Build/test commands used: `go build ./ui/...`, `go test ./ui/... -v -count=1`, `go vet ./ui/...`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64)
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: `muesli/reflow/wordwrap.String()` 함수를 사용하여 줄바꿈 수행
- Rationale: 계획에서 명시한 패키지이며, go.mod에 이미 의존성으로 등록되어 있음.
- Alternatives: 직접 구현 — 불필요한 복잡성이므로 기각.

- Decision: 빈 텍스트나 0 이하 폭에서 `nil`을 반환 (빈 슬라이스가 아닌 nil)
- Rationale: Go 관용적 패턴. `nil` 슬라이스는 `len() == 0`으로 동작하므로 빈 슬라이스와 호환됨.

# Invariants (MUST HOLD)
- `wrapWords("", n)`은 nil을 반환해야 함 (모든 n에 대해).
- `wrapWords(text, 0)` 및 `wrapWords(text, -n)`은 nil을 반환해야 함.
- `wrapWords(text, maxWidth)`의 각 줄은 단일 단어보다 긴 경우를 제외하고 `maxWidth`를 초과하지 않아야 함.
- 반환된 줄을 `\n`으로 join하면 원본 텍스트의 모든 단어가 순서대로 보존되어야 함.

# Prohibited Changes (DO NOT DO)
- `wrapWords` 함수의 시그니처 변경 금지 (TASK-3에서 사용 예정)
- `ui` 패키지명 변경 금지
- `muesli/reflow` 의존성 제거 금지

# What Changed in the current task
- New/modified files:
  - `ui/wordwrap.go` (신규): `wrapWords` 함수 구현
  - `ui/wordwrap_test.go` (신규): 6개 테스트 케이스
- New/changed public interfaces: 없음 (비공개 함수만 추가)
- Behavior changes: 없음 (신규 기능)
- Tests added/updated: 6개 테스트 추가
  - `TestWrapWords_NormalText`
  - `TestWrapWords_EmptyText`
  - `TestWrapWords_ZeroWidth`
  - `TestWrapWords_NegativeWidth`
  - `TestWrapWords_SingleLongWord`
  - `TestWrapWords_ExactFit`
- Migrations/config changes: 없음

# Verification (Build & Tests)
- 실행한 명령:
  - `go vet ./ui/...` → 통과 (출력 없음)
  - `go build ./ui/...` → 통과 (출력 없음)
  - `go test ./ui/... -v -count=1` → 6/6 테스트 PASS
- 결과: 모두 성공

# Timeout & Execution Safety
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, 0.002초에 완료
- `go build`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료

# Debug / Temporary Instrumentation
- 없음. 디버그 로그 추가 불필요.

# Guardrails & Interim Report
- 해당 없음. 첫 시도에서 모든 테스트 통과.

# Known Issues / Technical Debt
- `wrapWords`는 현재 비공개(unexported) 함수로, IDE에서 "unused" 경고가 표시될 수 있으나 TASK-3(배너 렌더링)에서 사용 예정.
- go.mod 진단에서 `muesli/reflow`가 `should be direct`로 표시됨. 이는 import하는 소스가 있지만 go.mod에서 `// indirect`로 표시되어 있기 때문. `go mod tidy` 실행 시 자동 수정됨 (다른 태스크의 소스 코드가 모두 작성된 후 실행 권장).

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `a833b33`: `Add word-wrap utility for slogan banner text`