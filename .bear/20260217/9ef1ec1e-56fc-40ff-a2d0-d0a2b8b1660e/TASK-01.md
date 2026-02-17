# Metadata
- Workspace: /workspace-bear-worktree-1ca6cd23-19ca-437a-bd33-f86771cfe911
- Base Branch: main
- Base Commit: 4648041 (Add placeholder main.go for build/test validation)

# Task Summary
TASK-01: lipgloss 스타일 정의. 앱 전체에서 재사용할 lipgloss 스타일을 `ui/styles.go`에 패키지 수준 변수로 정의함.

# Scope Summary
- 이 태스크는 TUI 컴포넌트들이 공통으로 사용할 lipgloss 스타일을 한 곳에 정의하는 것이 목표.
- DONE 기준: 7개 스타일 변수(BearArtStyle, SloganStyle, DescriptionStyle, SeparatorStyle, ErrorStyle, PromptLabelStyle, SuccessStyle)가 패키지 수준에서 정의되고, ANSI 16색 팔레트를 사용하며, 빌드가 성공적으로 통과.
- OUT OF SCOPE: 스타일을 사용하는 배너/프롬프트 렌더링 로직, 테스트 코드 작성, 다른 ui/ 파일 생성.

# Current System State (as of end of the current task)
- Feature flags / environment variables: 없음
- Build/test commands used: `go build ./ui/...`, `go vet ./ui/...`
- Runtime assumptions: Go 1.25+ (실제: Go 1.26.0 linux/arm64), lipgloss v1.1.0
- Secrets/credentials: 없음

# Key Decisions (and rationale)
- Decision: ANSI 16색 번호를 문자열로 사용 (예: `"3"`, `"6"`, `"8"`)
- Rationale: 계획서에 명시된 대로 ANSI 16색 번호를 사용하여 폭넓은 터미널 호환성 확보.
- Alternatives: lipgloss.AdaptiveColor 또는 256색/TrueColor 사용 — 호환성이 낮아 채택하지 않음.

- Decision: PromptLabelStyle은 Bold만 적용, 전경색 미지정
- Rationale: 계획서에서 PromptLabelStyle의 색상이 별도로 명시되지 않았으므로 Bold만 적용. 후속 태스크에서 필요시 색상 추가 가능.

# Invariants (MUST HOLD)
- `ui/styles.go`에 7개 스타일 변수가 패키지 수준에서 export되어야 함: BearArtStyle, SloganStyle, DescriptionStyle, SeparatorStyle, ErrorStyle, PromptLabelStyle, SuccessStyle.
- 색상 매핑: Yellow="3", Cyan="6", DarkGrey="8", Red="1", Green="2".
- SloganStyle은 Bold(true)가 적용되어야 함.
- PromptLabelStyle은 Bold(true)가 적용되어야 함.
- `go build ./ui/...`와 `go vet ./ui/...`가 에러 없이 통과해야 함.

# Prohibited Changes (DO NOT DO)
- 스타일 변수명 변경 금지 (다른 태스크에서 이 이름을 참조함)
- 색상 번호 변경 금지 (명세서에 정의된 값)
- `ui` 패키지명 변경 금지

# What Changed in the current task
- New/modified files:
  - `ui/styles.go` (신규 생성): 7개 lipgloss 스타일 변수 정의
- New/changed public interfaces:
  - `ui.BearArtStyle` (lipgloss.Style)
  - `ui.SloganStyle` (lipgloss.Style)
  - `ui.DescriptionStyle` (lipgloss.Style)
  - `ui.SeparatorStyle` (lipgloss.Style)
  - `ui.ErrorStyle` (lipgloss.Style)
  - `ui.PromptLabelStyle` (lipgloss.Style)
  - `ui.SuccessStyle` (lipgloss.Style)
- Behavior changes: 없음 (스타일 정의만)
- Tests added/updated: 없음 (이 태스크에 테스트 없음)
- Migrations/config changes: 없음

# Verification (Build & Tests)
- 실행한 명령: `go build ./ui/...` — 성공 (에러 없음)
- 실행한 명령: `go vet ./ui/...` — 성공 (에러 없음)
- 이 태스크에는 테스트 코드가 없음 (스타일 정의는 순수 선언적 코드이므로 별도 단위 테스트가 계획에 포함되지 않음).

# Timeout & Execution Safety
- `go build`와 `go vet` 명령에 `timeout --signal=TERM --kill-after=15s 60` 적용.
- 모든 명령이 타임아웃 없이 정상 완료됨.

# Debug / Temporary Instrumentation
- 없음.

# Guardrails & Interim Report
- 해당 없음. 모든 작업 정상 완료.

# Known Issues / Technical Debt
- go.mod 경고: `go mod tidy` 실행 시 lipgloss가 direct dependency로 전환되어야 하나, 다른 의존성들이 아직 사용되지 않아 tidy 실행 시 제거될 수 있음. 후속 태스크에서 모든 소스 파일 작성 후 tidy 실행 권장.

# Unfinished Work / Continuation Plan
NONE

# Git Commit
- `3ef2e31`: `Add lipgloss style definitions for TUI components`