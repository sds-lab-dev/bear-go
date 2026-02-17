# Metadata
- Workspace: /workspace-bear-worktree-35f0cd85-e3ec-4c9a-8f33-f4816c9219e3
- Base Branch: main
- Base Commit: b2859dd (Add bubbletea workspace prompt model)

# Task Summary
TASK-07: `wrapWords` 함수의 다양한 입력에 대한 동작을 검증하는 단위 테스트 5개를 `ui/wordwrap_test.go`에 구현하였다.

# Scope Summary
- 이 태스크는 `wrapWords` 유틸리티 함수에 대한 단위 테스트를 작성하는 것이 목표.
- DONE 기준: 계획에서 명시한 5개 테스트 함수(`TestWrapWords_NormalText`, `TestWrapWords_EmptyText`, `TestWrapWords_ZeroWidth`, `TestWrapWords_SingleLongWord`, `TestWrapWords_ExactFit`)가 존재하고, 모두 PASS해야 한다.
- OUT OF SCOPE: `wrapWords` 함수 구현 변경(TASK-02), 배너 렌더링(TASK-03), 다른 UI 컴포넌트.

# Current System State
- Feature flags / environment variables: 없음
- Build/test commands used: `go build ./ui/...`, `go test ./ui/... -v -count=1 -run "TestWrapWords"`, `go vet ./ui/...`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64)
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: 상위 태스크(TASK-02)에서 이미 작성된 테스트 파일을 기반으로 어설션을 강화하는 방식으로 구현
- Rationale: 기존 테스트의 구조와 커버리지를 활용하되, 일부 약한 어설션(단어 보존 미검증, 단일 긴 단어 줄 수 미검증 등)을 보강
- Alternatives: 전체 재작성 — 기존 코드와 의미가 동일하므로 기각

- Decision: `TestWrapWords_NegativeWidth`를 별도 함수가 아닌 `TestWrapWords_ZeroWidth` 내 서브케이스로 통합
- Rationale: 계획에서 명시한 5개 심볼에 `NegativeWidth`가 별도로 없고, `ZeroWidth`에서 "폭 0 입력 테스트"로 음수 폭도 함께 검증하는 것이 자연스러움

# Invariants (MUST HOLD)
- `TestWrapWords_NormalText`: 줄바꿈 후 각 줄이 maxWidth를 초과하지 않고, 모든 원본 단어가 보존되어야 함
- `TestWrapWords_EmptyText`: 빈 텍스트 입력 시 빈(nil) 슬라이스 반환
- `TestWrapWords_ZeroWidth`: 폭 0 또는 음수 입력 시 빈(nil) 슬라이스 반환
- `TestWrapWords_SingleLongWord`: 단일 긴 단어는 분리되지 않고 1줄로 유지
- `TestWrapWords_ExactFit`: 정확히 폭에 맞는 텍스트는 1줄로 반환

# Prohibited Changes (DO NOT DO)
- `wrapWords` 함수의 시그니처 변경 금지 (TASK-03에서 사용 예정)
- `ui` 패키지명 변경 금지
- `muesli/reflow` 의존성 제거 금지

# What Changed in the current task
- New/modified files:
  - `ui/wordwrap_test.go` (수정): 테스트 어설션 강화, NegativeWidth를 ZeroWidth로 통합, 단어 보존 검증 추가, ExactFit에 여러 단어 서브케이스 추가
- New/changed public interfaces: 없음 (테스트 코드만 변경)
- Behavior changes: 없음
- Tests added/updated: 5개 테스트 함수 (기존 6개 → 5개로 통합/강화)
  - `TestWrapWords_NormalText` — 단어 보존 검증 추가
  - `TestWrapWords_EmptyText` — 기존 유지
  - `TestWrapWords_ZeroWidth` — 음수 폭 서브케이스 통합
  - `TestWrapWords_SingleLongWord` — 줄 수=1 검증 강화, 내용 정확성 검증
  - `TestWrapWords_ExactFit` — 여러 단어 서브케이스 추가
- Migrations/config changes: 없음

# Verification (Build & Tests)
- 실행한 명령:
  - `timeout --signal=TERM --kill-after=15s 60 go vet ./ui/...` → 통과 (출력 없음)
  - `timeout --signal=TERM --kill-after=15s 120 go build ./ui/...` → 통과 (출력 없음)
  - `timeout --signal=TERM --kill-after=15s 90 go test ./ui/... -v -count=1 -run "TestWrapWords"` → 5/5 테스트 PASS (0.002s)
- 결과: 모두 성공

# Timeout & Execution Safety
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go build`: `timeout --signal=TERM --kill-after=15s 120` 적용, 정상 완료
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, 0.002초에 완료

# Debug / Temporary Instrumentation
- 없음. 디버그 로그 추가 불필요.

# Guardrails & Interim Report
- 해당 없음. 첫 시도에서 모든 테스트 통과.

# Known Issues / Technical Debt
- go.mod 진단에서 일부 의존성이 `should be direct`로 표시됨. 이는 다른 태스크의 소스 코드가 모두 작성된 후 `go mod tidy` 실행 시 자동 수정됨.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `f1657ca`: `Improve wrapWords unit tests with stronger assertions`