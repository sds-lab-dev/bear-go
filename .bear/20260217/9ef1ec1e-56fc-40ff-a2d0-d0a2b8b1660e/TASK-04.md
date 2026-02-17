# Metadata
- Workspace: /workspace-bear-worktree-d9e2db66-0965-433e-8245-0d430dab1f7e
- Base Branch: main
- Base Commit: cf73880 (Add slogan banner rendering with Bear ASCII art)

# Task Summary
TASK-04 리비전: 리뷰어 피드백에 따라 `bearArtWidth` off-by-one 버그를 수정하고, 컬럼 정렬 일관성 검증 테스트를 추가.

# Scope Summary
- 이 리비전은 리뷰에서 지적된 critical 버그(bearArtWidth=21이 실제 최장 줄 22자와 불일치)를 수정하는 것이 목표.
- DONE 기준: bearArtWidth가 22로 수정되고, MinTerminalWidth가 23으로 자동 보정되며, 컬럼 정렬 검증 테스트가 추가되어 모든 테스트(12개) 통과.
- OUT OF SCOPE: 오른쪽 컬럼이 bear art 줄 수를 초과할 때 오버플로 처리(리뷰어가 informational로 표시한 known limitation).

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음
- Build/test commands used: `go vet ./...`, `go test ./ui/... -v -count=1`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64), lipgloss v1.1.0
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: `bearArtWidth`를 21에서 22로 변경
- Rationale: 실제 bear art line 3(`/@@\ /  o    o  \ /@@\`)과 line 4(`\ @ \|     ^    |/ @ /`)가 22자. 계획서의 "21"은 오류였으며, 실제 문자열 길이에 맞게 수정.
- Alternatives: 아트 줄을 21자로 줄이는 방안 — 명세서에 정의된 아트를 변경하게 되므로 기각.

- Decision: `TestRenderBanner_ColumnAlignment` 테스트 추가
- Rationale: 리뷰어가 권장한 대로, ANSI 코드를 제거한 후 모든 bear art 줄에서 오른쪽 컬럼 텍스트가 `bearArtWidth` 위치에서 시작하는지 검증.

# Invariants (MUST HOLD)
- `bearArtWidth`는 22여야 함 (가장 긴 줄 기준).
- `MinTerminalWidth`는 `bearArtWidth + 1 = 23`이어야 함.
- 모든 bear art 줄이 `fmt.Sprintf("%-*s", bearArtWidth, line)`으로 패딩 시 정확히 22자 폭이어야 함.
- `go test ./ui/... -v`가 12개 테스트 모두 PASS.

# Prohibited Changes (DO NOT DO)
- `bearArtWidth` 값을 22 미만으로 변경 금지
- bear art 문자열 내용 변경 금지 (명세서에 정의)
- `MinTerminalWidth` 공개 상수명 변경 금지

# What Changed in the current task
- Modified files:
  - `ui/banner.go`: `bearArtWidth` 21 → 22 (line 19), `MinTerminalWidth` 22 → 23 (자동 반영)
  - `ui/banner_test.go`: `TestRenderBanner_ColumnAlignment` 테스트 함수 추가
- Behavior changes: 오른쪽 컬럼 폭이 `terminalWidth - 22`로 1자 줄어듦 (올바른 정렬)
- Tests added: `TestRenderBanner_ColumnAlignment` (컬럼 정렬 일관성 검증)

# Verification (Build & Tests)
- 실행한 명령: `timeout --signal=TERM --kill-after=15s 60 go vet ./...` → 성공
- 실행한 명령: `timeout --signal=TERM --kill-after=15s 90 go test ./ui/... -v -count=1` → 12/12 PASS (0.002s)

# Timeout & Execution Safety
- `go vet`: `timeout --signal=TERM --kill-after=15s 60` 적용, 정상 완료
- `go test`: `timeout --signal=TERM --kill-after=15s 90` 적용, 0.002초에 완료

# Debug / Temporary Instrumentation
- `cmd_debug_len_test.go`: bear art 줄 길이를 확인하기 위해 임시 생성 후 삭제. line 3=22, line 4=22 확인.
- `cmd_verify_test.go`: `bearArtWidth=22`에서 패딩 동작을 시각적으로 확인하기 위해 임시 생성 후 삭제.

# Guardrails & Interim Report
- 해당 없음. 1번째 시도에서 모든 테스트 통과.

# Known Issues / Technical Debt
- go.mod 경고: `lipgloss`와 `reflow`가 `// indirect`로 표시. 모든 태스크 완료 후 `go mod tidy` 실행 필요.
- 오른쪽 컬럼 오버플로: 매우 좁은 터미널에서 오른쪽 컬럼 줄이 6줄을 초과하면 잘림. TASK-6에서 MinTerminalWidth로 최소 폭을 보장하므로 실질적 문제는 아님.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `bdcb955`: `Fix bearArtWidth off-by-one and add alignment test`