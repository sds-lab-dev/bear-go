# Bear AI Developer — 초기 TUI (슬로건 + 워크스페이스) 구현 계획

## 개정 이력

### v3 (현재) 변경 사항
- **디렉토리 구조 변경**: `internal/app/`, `internal/ui/` 대신 `app/`, `ui/`로 변경. 사용자 피드백에 따라 `internal` 패키지 컨벤션을 사용하지 않음.
- 이전 결정 사항 유지: bubbletea v1 사용, Alt+Enter로 줄바꿈 통일.

### v2 변경 사항
- bubbletea v1 (안정 버전) 사용 결정.
- 줄바꿈 키를 Alt+Enter로 통일.

---

## 1. Overview

### 목표
- 프로그램 실행 시 Bear ASCII 아트 + 슬로건 배너를 화면 최상단에 출력한다.
- 배너 출력 후 사용자에게 워크스페이스 디렉토리를 확인받는다.
- 워크스페이스가 확정되면 설정 완료 메시지를 출력하고 프로그램을 정상 종료한다 (exit code 0).
- UI 모듈과 비즈니스 로직 모듈을 분리한다.

### 비목표 (Out-of-Scope)
- 요구사항 수집, 명세 작성, 개발 계획, 코드 작성, 코드 리뷰, 문서화 등 이후 단계 기능.
- Claude Code CLI 연동.
- AI 모델 응답 출력.
- 설정 파일 저장이나 워크스페이스 상태 영속화.
- CLI arguments 처리.

---

## 2. Assumptions & Open Questions

### 가정
- Go 1.25+ 환경에서 빌드한다 (go.mod에 `go 1.25.0` 명시됨).
- 모듈 경로는 `github.com/sds-lab-dev/bear-go`이다.
- bubbletea v1 (최신 안정 버전)과 bubbles v1, lipgloss v1을 사용한다.
- 줄바꿈 키는 Alt+Enter로 통일한다.
- Tab 문자는 bubbles textarea 기본 동작(4칸 스페이스 치환)을 그대로 따른다. 명세에서 "공백, 탭 등 모든 키 입력이 그대로 보존"이라 했으나, textarea 기본 동작이 탭을 스페이스로 치환하는 것이므로 이 단계에서는 기본 동작을 따르고, 리터럴 탭 보존이 필요하면 이후 단계에서 대응한다.
- Bear ASCII 아트의 고정 폭(왼쪽 영역)은 참조 Rust 코드의 가장 긴 줄 기준 21자이다.
- 터미널 최소 폭이 Bear ASCII 아트 폭 미만이면 에러 메시지 출력 후 종료한다.
- 모든 사용자 대면 메시지는 영어로 작성한다 (FR-TUI-09).
- `internal` 패키지 컨벤션은 사용하지 않으며, `app/`과 `ui/` 디렉토리를 워크스페이스 루트 바로 아래에 배치한다.

### 미결 사항
- 없음.

---

## 3. Proposed Design

### 3.1 아키텍처

UI 모듈과 비즈니스 로직 모듈을 명확히 분리한다.

디렉토리 구조:

```text
main.go                    -- 엔트리포인트
app/
  app.go                   -- 앱 오케스트레이션 (비즈니스 로직)
  workspace.go             -- 워크스페이스 경로 검증 로직
  workspace_test.go        -- 워크스페이스 검증 테스트
ui/
  banner.go                -- 슬로건 배너 렌더링
  banner_test.go           -- 배너 렌더링 테스트
  workspace_prompt.go      -- 워크스페이스 입력 프롬프트 (bubbletea Model)
  styles.go                -- 공통 lipgloss 스타일 정의
  wordwrap.go              -- 단어 단위 줄바꿈 유틸리티
  wordwrap_test.go         -- 줄바꿈 유틸리티 테스트
```

### 3.2 핵심 설계 결정

**배너 출력 방식**: 배너는 bubbletea 프로그램 시작 전에 `fmt.Fprint`로 직접 출력한다. bubbletea inline 모드에서 프로그램 시작 전 출력된 내용은 터미널 스크롤백에 그대로 남으므로, 배너가 고정되지 않고 자연스럽게 스크롤된다 (FR-TUI-02 충족). 이렇게 하면 bubbletea의 View()는 입력 프롬프트만 관리하면 되어 inline 모드의 동적 높이 변경 문제를 최소화한다.

**대안**: 배너를 View()에 포함시키는 방식. 매 렌더 시 배너를 다시 그리므로 비효율적이고, inline 모드에서 View 높이 변경 시 artifact 위험이 있어 채택하지 않음.

**워크스페이스 프롬프트**: bubbletea inline 모드로 textarea 기반 입력 프롬프트를 구현한다. Enter로 제출, Alt+Enter로 줄바꿈을 삽입한다. 마우스 캡처를 활성화하지 않으므로 네이티브 터미널 선택/복사가 가능하다 (FR-TUI-06 충족).

**경로 검증 로직**: `app/workspace.go`에 순수 함수로 구현한다. UI 모듈에서는 검증 함수를 주입받아 결과만 처리한다 (NFR-01 충족).

### 3.3 흐름

```pseudocode
FUNCTION main:
INPUTS: none
OUTPUTS: exit code

IF <terminal width query fails> THEN
    Print error and exit with non-zero code
ENDIF

IF <terminal width is less than minimum bear art width> THEN
    Print error and exit with non-zero code
ENDIF

Render banner to stdout
Print separator line to stdout

Start bubbletea program in inline mode for workspace prompt

IF <bubbletea returns error> THEN
    Print error and exit with non-zero code
ENDIF

Extract confirmed workspace path from final model
Print workspace confirmation message
Exit with code zero
```

### 3.4 에러 처리

- 터미널 폭 조회 실패: 에러 메시지 출력 후 exit code 1로 종료.
- 터미널 폭 부족: 에러 메시지 출력 후 exit code 1로 종료.
- 워크스페이스 경로 검증 실패: 에러 메시지를 bubbletea의 View에 표시하고 재입력 루프.
- 현재 워킹 디렉토리 조회 실패: 에러 메시지 출력 후 exit code 1로 종료.

---

## 4. Implementation

### TASK-0: 프로젝트 의존성 설치

**목적**: bubbletea v1, bubbles v1, lipgloss v1, reflow 패키지를 Go 모듈 의존성으로 추가한다.

**파일**: `go.mod`, `go.sum`

**작업 내용**:
- `go get`으로 다음 의존성을 추가:
  - `github.com/charmbracelet/bubbletea` (v1 최신)
  - `github.com/charmbracelet/bubbles` (v1 최신)
  - `github.com/charmbracelet/lipgloss` (v1 최신)
  - `github.com/muesli/reflow` (최신)
- 터미널 폭 조회를 위해 `golang.org/x/term` 패키지도 추가 (배너는 bubbletea 시작 전에 출력하므로 직접 터미널 폭을 조회해야 함).

**의존성**: 없음

---

### TASK-1: lipgloss 스타일 정의

**목적**: 앱 전체에서 재사용할 lipgloss 스타일을 한 곳에서 정의한다.

**파일**: `ui/styles.go` (신규 생성)

**새로운 심볼**:
- `BearArtStyle` — Bear ASCII 아트용 노란색 스타일
- `SloganStyle` — 슬로건용 시안색 + Bold 스타일
- `DescriptionStyle` — 설명 텍스트용 DarkGrey 스타일
- `SeparatorStyle` — 구분선용 DarkGrey 스타일
- `ErrorStyle` — 에러 메시지용 빨간색 스타일
- `PromptLabelStyle` — 프롬프트 레이블용 스타일
- `SuccessStyle` — 성공 메시지용 녹색 스타일

**작업 내용**:
- `lipgloss.NewStyle()`을 사용하여 각 스타일을 패키지 수준 변수로 정의.
- 색상은 ANSI 16색 번호 사용: Yellow=`"3"`, Cyan=`"6"`, DarkGrey=`"8"`, Red=`"1"`, Green=`"2"`.
- Bold는 `.Bold(true)` 체이닝.

**의존성**: TASK-0

---

### TASK-2: 단어 단위 줄바꿈 유틸리티

**목적**: 슬로건 오른쪽 영역 텍스트의 단어 단위 줄바꿈을 담당하는 함수를 구현한다.

**파일**: `ui/wordwrap.go` (신규 생성)

**새로운 심볼**:
- `wrapWords` — 주어진 텍스트를 최대 폭에 맞게 단어 단위로 줄바꿈하여 문자열 슬라이스를 반환하는 비공개 함수

**작업 내용**:
- `github.com/muesli/reflow/wordwrap` 패키지를 사용하여 텍스트를 주어진 최대 폭으로 단어 단위 줄바꿈.
- 결과를 줄 단위 문자열 슬라이스로 분할하여 반환.
- 빈 텍스트나 최대 폭이 0 이하인 경우 빈 슬라이스를 반환.

```pseudocode
FUNCTION wrapWords:
INPUTS: text string, max width integer
OUTPUTS: list of wrapped line strings

IF <text is empty or max width is zero or less> THEN
    RETURN <empty list>
ENDIF

Apply word-wrap to text with given max width
Split result by newline into list of lines
RETURN <list of lines>
```

**의존성**: TASK-0

---

### TASK-3: 슬로건 배너 렌더링

**목적**: Bear ASCII 아트와 슬로건/설명 텍스트를 조합하여 배너 문자열을 생성하는 함수를 구현한다.

**파일**: `ui/banner.go` (신규 생성)

**새로운 심볼**:
- `bearArtLines` — Bear ASCII 아트 7줄을 담은 패키지 수준 문자열 슬라이스 상수
- `bearArtWidth` — Bear ASCII 아트 영역의 고정 폭 상수 (가장 긴 줄 기준, 21)
- `sloganText` — 슬로건 원문 상수
- `descriptionText` — 설명 원문 상수
- `rightColumnStartOffset` — 오른쪽 영역 텍스트가 시작되는 Bear 아트 줄 오프셋 (참조 Rust 코드의 RIGHT_COLUMN_START에 대응)
- `MinTerminalWidth` — 배너 표시에 필요한 최소 터미널 폭 상수
- `RenderBanner` — 터미널 폭을 받아 배너 전체(아트 + 오른쪽 텍스트 + 구분선)를 렌더링한 문자열을 반환하는 공개 함수
- `buildRightColumn` — 오른쪽 영역(슬로건 + 빈 줄 + 설명)의 줄 목록을 생성하는 비공개 함수

**작업 내용**:

Bear ASCII 아트 정의:
- 참조 Rust 코드의 BEAR_TEXTS 배열을 Go 문자열 슬라이스로 변환.
- 7줄: 빈 줄, `       () _ _ ()`, `      / __  __ \`, `/@@\ /  o    o  \ /@@\`, `\ @ \|     ^    |/ @ /`, ` \   \    ___   /   /`, `  \   \________/   /`.

buildRightColumn 함수:

```pseudocode
FUNCTION buildRightColumn:
INPUTS: max width for right column
OUTPUTS: list of tuples each containing line text and style name

Wrap slogan text to max width using wrapWords
Wrap description text to max width using wrapWords

Create empty result list

LOOP <over each slogan wrapped line>:
    Append tuple of line text and slogan style identifier
ENDLOOP

IF <slogan lines exist and description lines exist> THEN
    Append tuple of empty string and no-style identifier
ENDIF

LOOP <over each description wrapped line>:
    Append tuple of line text and description style identifier
ENDLOOP

RETURN <result list>
```

RenderBanner 함수:

```pseudocode
FUNCTION RenderBanner:
INPUTS: terminal width
OUTPUTS: rendered banner string

Calculate right column width as terminal width minus bear art width

Build right column lines using buildRightColumn with right column width

Create a string builder

LOOP <over each bear art line with index>:
    Pad bear art line to bear art width with trailing spaces
    Apply yellow style to padded bear art line
    Append styled bear art to builder

    Calculate right column index as current index minus right column start offset
    IF <right column index is valid and within right column lines> THEN
        Apply appropriate style to right column line text
        Append styled right column text to builder
    ENDIF

    Append newline to builder
ENDLOOP

Create separator by repeating horizontal line character to terminal width
Apply dark grey style to separator
Append styled separator to builder
Append newline to builder

RETURN <builder content as string>
```

**의존성**: TASK-1, TASK-2

---

### TASK-4: 워크스페이스 경로 검증 로직 (비즈니스 로직)

**목적**: 사용자가 입력한 워크스페이스 경로를 검증하는 순수 함수를 구현한다. UI에 의존하지 않는다.

**파일**: `app/workspace.go` (신규 생성)

**새로운 심볼**:
- `ValidateWorkspacePath` — 경로 문자열을 검증하여 에러를 반환하는 공개 함수
- `ErrRelativePath` — 상대 경로 에러를 나타내는 센티널 에러
- `ErrPathNotExist` — 경로 미존재 에러를 나타내는 센티널 에러
- `ErrNotDirectory` — 디렉토리가 아닌 경로 에러를 나타내는 센티널 에러

**작업 내용**:
- 세 가지 센티널 에러를 `errors.New`로 정의.
  - `ErrRelativePath`: `"path must be absolute; relative paths are not allowed"`
  - `ErrPathNotExist`: `"path does not exist"`
  - `ErrNotDirectory`: `"path is not a directory; please enter a directory path"`
- ValidateWorkspacePath 함수:

```pseudocode
FUNCTION ValidateWorkspacePath:
INPUTS: path string
OUTPUTS: error or nil

IF <path is not absolute> THEN
    RETURN <ErrRelativePath>
ENDIF

Stat the path on the file system
IF <stat returns not-exist error> THEN
    RETURN <ErrPathNotExist>
ENDIF
IF <stat returns other error> THEN
    RETURN <that error wrapped>
ENDIF

IF <stat result is not a directory> THEN
    RETURN <ErrNotDirectory>
ENDIF

RETURN <nil>
```

**의존성**: 없음 (순수 Go 표준 라이브러리만 사용)

---

### TASK-5: 워크스페이스 입력 프롬프트 (bubbletea Model)

**목적**: bubbletea inline 모드로 동작하는 워크스페이스 경로 입력 프롬프트를 구현한다.

**파일**: `ui/workspace_prompt.go` (신규 생성)

**새로운 심볼**:
- `WorkspacePromptModel` — bubbletea Model 인터페이스를 구현하는 구조체
- `NewWorkspacePromptModel` — 현재 워킹 디렉토리와 검증 함수를 받아 모델을 생성하는 공개 함수
- `WorkspacePromptResult` — 프롬프트 결과를 담는 구조체 (확정된 워크스페이스 경로, 취소 여부)

**작업 내용**:

WorkspacePromptModel 구조체 설계:
- `textarea` 필드: bubbles textarea.Model (사용자 입력)
- `currentDir` 필드: 현재 워킹 디렉토리 경로 문자열
- `validatePath` 필드: 경로 검증 함수 (함수 타입, UI에서 비즈니스 로직을 직접 참조하지 않도록 함수를 주입)
- `errorMessage` 필드: 현재 표시할 에러 메시지 문자열 (빈 문자열이면 에러 없음)
- `confirmed` 필드: 워크스페이스 확정 여부 boolean
- `confirmedPath` 필드: 확정된 워크스페이스 경로 문자열
- `width` 필드: 터미널 폭 정수
- `ready` 필드: WindowSizeMsg 수신 여부 boolean

NewWorkspacePromptModel 함수:
- textarea를 생성하고 기본 설정 적용: placeholder 없음, ShowLineNumbers false, CharLimit 0 (무제한).
- textarea.KeyMap.InsertNewline을 비활성화 (Enter를 제출로 사용하기 위해).
- textarea를 Focus 상태로 설정.
- validatePath 함수를 주입받아 저장.

Init 메서드:
- `textarea.Blink` 커맨드를 반환하여 커서 깜박임 시작.

Update 메서드:

```pseudocode
FUNCTION WorkspacePromptModel Update:
INPUTS: message
OUTPUTS: updated model, command

IF <message is WindowSizeMsg> THEN
    Store width from message
    Set textarea width to message width
    Set ready to true
    RETURN <updated model, nil>
ENDIF

IF <not ready> THEN
    RETURN <model as-is, nil>
ENDIF

IF <message is KeyMsg> THEN
    IF <key is ctrl+c> THEN
        RETURN <model, quit command>
    ENDIF

    IF <key is enter> THEN
        Get trimmed value from textarea
        IF <value is empty> THEN
            Set confirmed path to current directory
        ELSE
            Run validate path function on value
            IF <validation returns error> THEN
                Set error message to error string
                RETURN <model, nil>
            ENDIF
            Set confirmed path to value
        ENDIF
        Set confirmed to true
        Clear error message
        RETURN <model, quit command>
    ENDIF

    IF <key is alt+enter> THEN
        Insert newline into textarea
        Clear error message
        RETURN <model, nil>
    ENDIF

    Clear error message
ENDIF

Delegate message to textarea Update
RETURN <updated model, textarea command>
```

View 메서드:

```pseudocode
FUNCTION WorkspacePromptModel View:
INPUTS: none
OUTPUTS: rendered string

IF <not ready> THEN
    RETURN <initializing message>
ENDIF

IF <confirmed> THEN
    RETURN <empty string>
ENDIF

Create string builder

Append prompt label line showing current directory
Append hint line showing Enter to confirm and Alt+Enter for newline
Append newline
Append textarea view

IF <error message is not empty> THEN
    Append newline
    Append error-styled error message
ENDIF

Append newline

RETURN <builder content>
```

프롬프트 레이블 예시: `"Current directory: /workspace"`
힌트 예시: `"Press Enter to confirm, or type an absolute path. (Alt+Enter for newline)"`

WorkspacePromptResult 접근:
- `WorkspacePromptModel`에 `Result` 메서드를 추가하여 `WorkspacePromptResult`를 반환. 확정된 경로와 취소 여부를 포함.

**의존성**: TASK-1, TASK-4

---

### TASK-6: 앱 오케스트레이션 및 엔트리포인트

**목적**: main 함수와 앱 오케스트레이션 로직을 구현하여 배너 출력 → 워크스페이스 프롬프트 → 완료 메시지의 전체 흐름을 연결한다.

**파일 1**: `app/app.go` (신규 생성)

**새로운 심볼**:
- `Run` — 앱 전체 실행 흐름을 오케스트레이션하는 공개 함수

**작업 내용**:

```pseudocode
FUNCTION Run:
INPUTS: stdout writer, stderr writer
OUTPUTS: error

Query terminal width from stdout file descriptor using x/term
IF <query fails> THEN
    RETURN <error>
ENDIF

IF <terminal width is less than MinTerminalWidth> THEN
    RETURN <terminal too narrow error>
ENDIF

Render banner with terminal width
Write banner to stdout

Get current working directory
IF <get cwd fails> THEN
    RETURN <error>
ENDIF

Create workspace prompt model with cwd and ValidateWorkspacePath function
Create bubbletea program in inline mode with the model
Run the program
IF <program returns error> THEN
    RETURN <error>
ENDIF

Extract result from final model
IF <result indicates cancellation> THEN
    RETURN <nil>
ENDIF

Print workspace confirmation message to stdout using confirmed path
RETURN <nil>
```

확인 메시지 예시: `"Workspace set to: /some/path"`

**파일 2**: `main.go` (신규 생성)

**새로운 심볼**:
- `main` — 프로그램 엔트리포인트

**작업 내용**:

```pseudocode
FUNCTION main:
INPUTS: none
OUTPUTS: none

Call app Run with os.Stdout and os.Stderr
IF <error returned> THEN
    Print error to stderr
    Exit with code 1
ENDIF
```

**의존성**: TASK-3, TASK-5

---

### TASK-7: 워크스페이스 검증 로직 단위 테스트

**목적**: `ValidateWorkspacePath` 함수의 모든 검증 경로를 테스트한다.

**파일**: `app/workspace_test.go` (신규 생성)

**새로운 심볼**:
- `TestValidateWorkspacePath_AbsoluteDirectoryPath` — 유효한 절대 디렉토리 경로 테스트
- `TestValidateWorkspacePath_RelativePath` — 상대 경로 입력 시 ErrRelativePath 반환 테스트
- `TestValidateWorkspacePath_NonExistentPath` — 존재하지 않는 경로 입력 시 ErrPathNotExist 반환 테스트
- `TestValidateWorkspacePath_FilePath` — 파일 경로 입력 시 ErrNotDirectory 반환 테스트
- `TestValidateWorkspacePath_EmptyPath` — 빈 문자열 입력 시 동작 테스트

**테스트 케이스**:

1. **유효한 절대 디렉토리 경로**: 임시 디렉토리를 생성하고 그 경로를 입력 → nil 반환 확인.
2. **상대 경로**: `"./some/path"` 입력 → `errors.Is(err, ErrRelativePath)` 확인.
3. **존재하지 않는 경로**: `"/nonexistent/path/abc123"` 입력 → `errors.Is(err, ErrPathNotExist)` 확인.
4. **파일 경로**: 임시 파일을 생성하고 그 경로를 입력 → `errors.Is(err, ErrNotDirectory)` 확인.
5. **빈 문자열**: `""` 입력 → 상대 경로 에러 또는 적절한 에러 반환 확인.

**의존성**: TASK-4

---

### TASK-8: 배너 렌더링 단위 테스트

**목적**: `RenderBanner` 함수가 올바른 배너 문자열을 생성하는지 테스트한다.

**파일**: `ui/banner_test.go` (신규 생성)

**새로운 심볼**:
- `TestRenderBanner_ContainsBearArt` — 배너에 Bear ASCII 아트가 포함되어 있는지 확인
- `TestRenderBanner_ContainsSlogan` — 배너에 슬로건 텍스트가 포함되어 있는지 확인
- `TestRenderBanner_ContainsDescription` — 배너에 설명 텍스트가 포함되어 있는지 확인
- `TestRenderBanner_ContainsSeparator` — 배너에 구분선이 포함되어 있는지 확인
- `TestRenderBanner_NarrowTerminal` — 좁은 터미널에서 오른쪽 영역 텍스트가 줄바꿈되는지 확인

**테스트 케이스**:

1. **넓은 터미널(120)**: RenderBanner 호출 → ANSI 코드를 제거한 결과에서 Bear 아트 각 줄의 핵심 문자열(`() _ _ ()`, `/@@\` 등)이 포함되어 있는지 확인.
2. **슬로건 포함**: 결과에서 `"Bear: The AI developer that saves your time."` 텍스트가 포함되어 있는지 확인.
3. **설명 포함**: 결과에서 description 텍스트의 일부가 포함되어 있는지 확인.
4. **구분선 포함**: 결과에서 `"─"` 문자가 포함되어 있는지 확인.
5. **좁은 터미널(60)**: RenderBanner 호출 → 오른쪽 영역 텍스트가 여러 줄로 줄바꿈되었는지 확인 (줄 수 증가).

**의존성**: TASK-3

---

### TASK-9: 단어 줄바꿈 유틸리티 단위 테스트

**목적**: `wrapWords` 함수의 다양한 입력에 대한 동작을 테스트한다.

**파일**: `ui/wordwrap_test.go` (신규 생성)

**새로운 심볼**:
- `TestWrapWords_NormalText` — 일반 텍스트 줄바꿈 테스트
- `TestWrapWords_EmptyText` — 빈 텍스트 입력 테스트
- `TestWrapWords_ZeroWidth` — 폭 0 입력 테스트
- `TestWrapWords_SingleLongWord` — 폭보다 긴 단일 단어 테스트
- `TestWrapWords_ExactFit` — 정확히 폭에 맞는 텍스트 테스트

**의존성**: TASK-2

---

## 5. Testing & Validation

### 단위 테스트 (자동화)
- TASK-7: 워크스페이스 경로 검증 — 5가지 시나리오 (유효 절대 경로, 상대 경로, 미존재 경로, 파일 경로, 빈 문자열).
- TASK-8: 배너 렌더링 — 5가지 시나리오 (아트 포함, 슬로건 포함, 설명 포함, 구분선 포함, 좁은 터미널 줄바꿈).
- TASK-9: 단어 줄바꿈 — 5가지 시나리오 (일반 텍스트, 빈 텍스트, 폭 0, 긴 단어, 정확히 맞는 텍스트).

### 수동 테스트 체크리스트
- [ ] `go run main.go` 실행 시 Bear ASCII 아트가 노란색으로 출력되는지 확인.
- [ ] 슬로건이 시안색 Bold, 설명이 DarkGrey로 출력되는지 확인.
- [ ] 터미널 폭을 줄였을 때 오른쪽 영역 텍스트가 줄바꿈되는지 확인.
- [ ] 배너 아래에 DarkGrey 구분선(`─`)이 터미널 전체 폭으로 출력되는지 확인.
- [ ] 워크스페이스 프롬프트에서 Enter만 누르면 현재 디렉토리로 확정되고 완료 메시지 출력 후 종료되는지 확인.
- [ ] 유효한 절대 디렉토리 경로를 입력하면 해당 경로로 확정되는지 확인.
- [ ] 상대 경로 입력 시 적절한 영어 에러 메시지가 표시되고 재입력 프롬프트로 돌아가는지 확인.
- [ ] 존재하지 않는 경로 입력 시 적절한 영어 에러 메시지가 표시되고 재입력 프롬프트로 돌아가는지 확인.
- [ ] 파일 경로 입력 시 적절한 영어 에러 메시지가 표시되고 재입력 프롬프트로 돌아가는지 확인.
- [ ] 에러 발생 후 유효한 경로를 입력하면 정상 확정되는지 확인.
- [ ] Alt+Enter로 줄바꿈 삽입이 가능한지 확인.
- [ ] 여러 줄 입력 상태에서 방향키로 커서 이동이 되는지 확인.
- [ ] 긴 텍스트 입력 시 자동 wrap되어 오른쪽 잘림 없이 표시되는지 확인.
- [ ] 마우스로 화면 텍스트 선택 및 복사가 가능한지 확인.
- [ ] 스크롤 시 슬로건을 포함한 화면 전체가 함께 움직이는지 확인.
- [ ] 터미널 폭이 Bear 아트보다 좁을 때 에러 메시지 출력 후 종료되는지 확인 (exit code != 0).
- [ ] Ctrl+C로 프로그램이 정상 종료되는지 확인.

### 인수 기준 매핑

| 인수 기준 | 구현 태스크 | 검증 방법 |
|---|---|---|
| AC-SLG-01 | TASK-3 | TASK-8 + 수동 확인 |
| AC-SLG-02 | TASK-3 | TASK-8 + 수동 확인 |
| AC-SLG-03 | TASK-3, TASK-2 | TASK-8 (좁은 터미널) + 수동 확인 |
| AC-SLG-04 | TASK-3 | TASK-8 + 수동 확인 |
| AC-SLG-05 | TASK-6 | 수동 확인 (좁은 터미널) |
| AC-WS-01 | TASK-5, TASK-6 | 수동 확인 |
| AC-WS-02 | TASK-5 | 수동 확인 |
| AC-WS-03 | TASK-5, TASK-4 | TASK-7 + 수동 확인 |
| AC-WS-04 | TASK-4, TASK-5 | TASK-7 + 수동 확인 |
| AC-WS-05 | TASK-4, TASK-5 | TASK-7 + 수동 확인 |
| AC-WS-06 | TASK-4, TASK-5 | TASK-7 + 수동 확인 |
| AC-WS-07 | TASK-5 | 수동 확인 |
| AC-TUI-01 | TASK-5, TASK-6 | 수동 확인 |
| AC-TUI-02 | TASK-6 | 수동 확인 |
| AC-TUI-03 | TASK-5 | 수동 확인 |
| AC-TUI-04 | TASK-5 | 수동 확인 |
| AC-TUI-05 | TASK-5 | 수동 확인 |
| AC-TUI-06 | TASK-5, TASK-6 | 수동 확인 |
| AC-TUI-07 | TASK-5 | 수동 확인 |
| AC-DES-01 | TASK-4, TASK-5 | 코드 리뷰 |

### 빌드 및 실행 명령
```text
go mod tidy
go build -o bear ./main.go
go test ./app/... ./ui/... -v
go vet ./...
```

---

## 6. Risk Analysis

### 기술적 리스크

| 리스크 | 영향 | 완화 |
|---|---|---|
| bubbletea v1 inline 모드에서 View 높이 변경 시 ghost line artifact | 에러 메시지 표시/해제 시 화면 깨짐 가능 | View 출력 높이를 최대한 일정하게 유지. 에러 메시지 영역을 항상 포함하되, 에러 없을 때는 빈 줄로 패딩. |
| Alt+Enter가 일부 터미널에서 다른 용도로 매핑되어 있을 수 있음 | 줄바꿈 삽입 불가 | 힌트 메시지에 키 바인딩을 명확히 안내. 이후 필요시 대체 키 추가 고려. |
| textarea의 탭 문자가 4칸 스페이스로 치환됨 | 명세의 "탭 보존" 요구와 다소 차이 | 이번 단계에서는 기본 동작을 수용. 이후 필요시 커스텀 Sanitizer 적용. |
| 터미널 폭 변경(리사이즈) 시 배너가 이미 출력된 상태라 재렌더링 불가 | 배너 레이아웃이 맞지 않을 수 있음 | 배너는 정적 출력이므로 리사이즈 대응은 비목표. 리사이즈 시에도 textarea는 WindowSizeMsg로 폭 갱신됨. |

### 보안 고려
- 사용자 입력 경로는 `filepath.IsAbs`와 `os.Stat`으로만 검증하며, 임의 명령 실행 경로가 없으므로 인젝션 위험 없음.
- 외부 네트워크 접근 없음.

### 롤백 전략
- 이번 구현은 신규 파일만 추가하므로, 해당 커밋을 revert하면 완전히 원복된다.
- 기존 파일 변경은 `go.mod`/`go.sum`에 의존성 추가뿐이므로 롤백이 간단하다.

---

## 7. Implementation Notes

### 빌드/테스트/린트 명령
```text
go mod tidy
go build ./...
go test ./app/... ./ui/... -v
go vet ./...
```

### 파일 목록 요약

| 파일 | 동작 | TASK |
|---|---|---|
| `go.mod` | 수정 (의존성 추가) | TASK-0 |
| `go.sum` | 수정 (자동 생성) | TASK-0 |
| `main.go` | 신규 생성 | TASK-6 |
| `app/app.go` | 신규 생성 | TASK-6 |
| `app/workspace.go` | 신규 생성 | TASK-4 |
| `app/workspace_test.go` | 신규 생성 | TASK-7 |
| `ui/styles.go` | 신규 생성 | TASK-1 |
| `ui/banner.go` | 신규 생성 | TASK-3 |
| `ui/banner_test.go` | 신규 생성 | TASK-8 |
| `ui/wordwrap.go` | 신규 생성 | TASK-2 |
| `ui/wordwrap_test.go` | 신규 생성 | TASK-9 |
| `ui/workspace_prompt.go` | 신규 생성 | TASK-5 |

### 의존성 DAG (인접 리스트)

```text
TASK-0: (없음)
TASK-1: TASK-0
TASK-2: TASK-0
TASK-3: TASK-1, TASK-2
TASK-4: (없음)
TASK-5: TASK-1, TASK-4
TASK-6: TASK-3, TASK-5
TASK-7: TASK-4
TASK-8: TASK-3
TASK-9: TASK-2
```

병렬 실행 가능 그룹:
- **1단계**: TASK-0, TASK-4 (동시 실행)
- **2단계**: TASK-1, TASK-2, TASK-7 (동시 실행, TASK-0/TASK-4 완료 후)
- **3단계**: TASK-3, TASK-5, TASK-9 (동시 실행, TASK-1/TASK-2 완료 후)
- **4단계**: TASK-6, TASK-8 (동시 실행, TASK-3/TASK-5 완료 후)

### 레포지토리 컨벤션
- `app/`과 `ui/` 디렉토리를 워크스페이스 루트 바로 아래에 배치한다 (`internal` 패키지 미사용).
- UI 코드는 `ui/`, 비즈니스 로직은 `app/`에 배치.
- 테스트 파일은 동일 패키지에 `_test.go` 접미사로 배치.
- CLAUDE.md 코딩 컨벤션 준수: 50줄 초과 함수 리팩토링 고려, 3단계 이상 들여쓰기 금지, 단일 책임 원칙.
