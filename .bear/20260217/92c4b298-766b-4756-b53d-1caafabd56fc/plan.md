
# 사용자 요구사항 입력 프롬프트 구현 계획

## 1. Overview

### 목표
워크스페이스 경로 설정(2단계) 완료 직후, 사용자에게 요구사항을 자유 형식 텍스트로 입력받는 TUI 프롬프트(3단계)를 구현한다. 이 프롬프트는 기존 `WorkspacePromptModel`과 동일한 textarea 기반이되, 별도의 독립적인 Bubble Tea 모델로 구현한다. 외부 에디터 연동(Ctrl+G), 빈 입력 검증, 반응형 레이아웃 등을 포함한다.

### 비목표
- 다음 단계(스펙 에이전트 호출, 명확화 질문 등)의 구현은 포함하지 않는다.
- 입력된 텍스트에 대한 후속 처리 로직은 구현하지 않는다.
- 기존 `WorkspacePromptModel`의 동작 변경은 하지 않는다.
- 입력 이력 관리, 자동 저장, 자동 완성 등 부가 기능은 포함하지 않는다.

### 맥락
현재 `app/app.go`의 `Run` 함수는 배너 출력 → 워크스페이스 프롬프트 실행 → 결과 출력 후 종료하는 흐름이다. 이번 작업에서는 워크스페이스 프롬프트 완료 후 요구사항 입력 프롬프트를 순차적으로 실행하도록 `Run` 함수를 확장하고, 새로운 `UserRequestPromptModel`을 추가한다.

---

## 2. Assumptions & Open Questions

### 가정
- 기존 `WorkspacePromptModel`의 패턴(Init/Update/View/Result 구조, WindowSizeMsg 처리, ready 플래그 등)을 그대로 따른다.
- Bubble Tea의 `tea.ExecProcess`를 사용하여 외부 에디터를 실행한다. 이 기능은 Bubble Tea에 내장되어 있으며 TUI 일시 중단/재개를 자동 처리한다.
- `exec.LookPath`를 사용하여 시스템 명령 존재 여부를 확인한다 (Go 표준 라이브러리).
- 새 의존성 추가 없이 기존 의존성(bubbletea, bubbles, lipgloss)만으로 구현 가능하다.
- `EDITOR` 환경변수가 공백이 포함된 명령(예: `"code --wait"`)일 수 있으므로 적절히 파싱한다.

### 미결 사항
- 없음. 명세가 충분히 상세하다.

---

## 3. Proposed Design

### 아키텍처
기존 코드베이스의 패턴을 그대로 따른다:
- `ui/` 패키지에 새로운 Bubble Tea 모델 `UserRequestPromptModel`을 추가한다.
- `app/app.go`의 `Run` 함수에서 워크스페이스 프롬프트 완료 후 순차적으로 요구사항 프롬프트를 실행한다.
- 외부 에디터 해석 로직은 별도의 함수로 분리한다.

### 주요 설계 결정

**1. 외부 에디터 실행 방식: `tea.ExecProcess` 사용**
- Bubble Tea가 제공하는 `tea.ExecProcess`는 TUI를 일시 중단하고 외부 프로세스를 실행한 뒤 자동으로 TUI를 재개한다.
- 대안으로 직접 `os/exec`을 호출하고 `tea.Suspend`/`tea.Resume`를 수동 관리하는 방법이 있으나, `tea.ExecProcess`가 더 안전하고 간결하다.
- Bubble Tea가 `tea.ExecProcess` 완료 후 `tea.Msg`를 반환하므로, 에디터 종료 후 임시 파일 읽기 로직을 해당 메시지 핸들러에서 처리한다.

**2. 모델 구조: `WorkspacePromptModel`과 동일한 독립 모델 패턴**
- `WorkspacePromptModel`처럼 `Init`, `Update`, `View`, `Result` 메서드를 가진 독립 모델이다.
- `app.Run`에서 별도의 `tea.NewProgram`으로 실행한다.

**3. 에디터 명령 파싱**
- `EDITOR` 환경변수 값을 공백으로 분할하여 명령과 인자를 분리한다. 첫 번째 토큰이 실행 파일, 나머지가 인자이다.
- `code`를 사용하는 경우 `--wait` 플래그를 자동 추가한다.

### 에러 처리
- 빈 입력 시 Enter: 에러 메시지 표시 후 프롬프트 유지 (FR-5).
- 에디터 미발견: 영어 에러 메시지 표시 후 프롬프트 복귀 (FR-6-1).
- 에디터 실행 실패: 영어 에러 메시지 표시 후 프롬프트 복귀 (FR-6-4).
- 임시 파일 생성/읽기 실패: 에러 메시지 표시 후 프롬프트 복귀.

### Edge Cases
- `EDITOR` 환경변수가 빈 문자열인 경우: 미설정으로 간주하고 fallback 우선순위를 따른다.
- 에디터에서 파일을 저장하지 않고 종료한 경우: 기존 textarea 내용 유지(임시 파일에 이미 기존 내용이 있으므로 읽어도 동일).
- 매우 큰 텍스트 붙여넣기: textarea의 `CharLimit = 0`(무제한)이므로 문제없다.
- Ctrl+G 누를 때 textarea가 비어있는 경우: 빈 임시 파일을 생성하고 에디터를 연다.

---

## 4. Implementation

### TASK-00: UserRequestPromptModel 구현 및 app 통합

**목적:** 사용자 요구사항을 입력받는 새로운 Bubble Tea 모델을 구현하고, 외부 에디터 연동 기능을 포함하며, `app/app.go`의 `Run` 함수에 통합한다. 단위 테스트를 포함한다.

**의존성:** none

#### 변경 파일 목록

---

**파일 1: `ui/user_request_prompt.go` (신규 생성)**

**목적:** 사용자 요구사항 입력을 위한 독립적인 Bubble Tea 모델을 구현한다.

**새로 도입할 심볼:**
- `UserRequestPromptResult` (구조체)
- `UserRequestPromptModel` (구조체)
- `NewUserRequestPromptModel` (생성자 함수)
- `editorFinishedMsg` (에디터 완료 메시지 타입)

**편집 의도:**

1. `UserRequestPromptResult` 구조체 정의
   - 입력된 텍스트를 담는 필드
   - 사용자 취소 여부를 나타내는 필드

2. `UserRequestPromptModel` 구조체 정의
   - textarea 컴포넌트 (bubbles textarea.Model)
   - 에러 메시지 문자열
   - 확인 완료 여부 플래그
   - 확인된 텍스트
   - 터미널 너비
   - ready 플래그
   - 에디터 해석 함수 (테스트 시 주입 가능하도록)
   - 임시 파일 경로 (에디터 연동 시 사용)

3. `NewUserRequestPromptModel` 생성자 함수
   - textarea를 초기화한다: ShowLineNumbers false, CharLimit 0(무제한), InsertNewline 비활성화, Focus 호출
   - 기본 에디터 해석 함수를 설정한다

4. `Init` 메서드
   - `textarea.Blink` 커맨드를 반환한다 (기존 `WorkspacePromptModel`과 동일)

5. `Update` 메서드
   - `tea.WindowSizeMsg` 처리: 너비 설정, textarea 너비 조정, ready 플래그 설정
   - ready 상태가 아니면 메시지 무시
   - `tea.KeyMsg`이면 키 처리 메서드로 위임
   - `editorFinishedMsg` 처리: 에디터 종료 후 임시 파일 읽기 및 textarea 반영 처리 메서드로 위임
   - 그 외 메시지는 textarea에 전달

6. 키 처리 메서드 (handleKey에 해당)
   
   ```pseudocode
   IF <key is ctrl+c>
       RETURN <quit command>
   ENDIF

   IF <key is enter>
       Delegate to enter handler
       RETURN
   ENDIF

   IF <key is shift+enter or alt+enter>
       Insert newline into textarea
       Clear error message
       Adjust textarea height to match line count
       RETURN
   ENDIF

   IF <key is ctrl+g>
       Delegate to editor launch handler
       RETURN
   ENDIF

   Clear error message
   Pass key to textarea for default processing
   Adjust textarea height to match line count
   RETURN
   ```

7. Enter 처리 메서드 (handleEnter에 해당)

   ```pseudocode
   FUNCTION <enter handler>:
   INPUTS: <current model state>
   OUTPUTS: <updated model, command>
   FAILURE: <empty input -> set error message and remain in prompt>

   Trim whitespace from textarea value
   IF <trimmed value is empty>
       Set error message to indicate empty input is not allowed
       RETURN <model with error, no command>
   ENDIF

   Set confirmed text to trimmed value
   Set confirmed flag to true
   Clear error message
   RETURN <model, quit command>
   ```

8. 에디터 실행 메서드 (handleEditorLaunch에 해당)

   ```pseudocode
   FUNCTION <editor launch handler>:
   INPUTS: <current model state>
   OUTPUTS: <updated model, exec command>
   FAILURE: <no editor found or temp file error -> set error message and return>

   Call editor resolver function to get editor command and arguments
   IF <editor resolver returns error>
       Set error message from resolver error
       RETURN <model with error, no command>
   ENDIF

   Create temp file with md extension
   IF <temp file creation fails>
       Set error message about temp file failure
       RETURN <model with error, no command>
   ENDIF

   Write current textarea content to temp file
   IF <write fails>
       Remove temp file
       Set error message about write failure
       RETURN <model with error, no command>
   ENDIF

   Store temp file path in model
   Build exec command by appending temp file path to editor arguments
   RETURN <model, tea exec process command with editor finished message callback>
   ```

9. 에디터 종료 처리 메서드 (handleEditorFinished에 해당)

   ```pseudocode
   FUNCTION <editor finished handler>:
   INPUTS: <editor finished message containing possible error>
   OUTPUTS: <updated model, command>
   FAILURE: <read error or editor error -> set error message>

   IF <editor finished message has error>
       Set error message about editor failure
       Clean up temp file
       RETURN <model with error, no command>
   ENDIF

   Read content from temp file
   IF <read fails>
       Set error message about read failure
       Clean up temp file
       RETURN <model with error, no command>
   ENDIF

   Set textarea value to file content
   Adjust textarea height to match line count
   Clear error message
   Clean up temp file
   RETURN <model, no command>
   ```

10. `View` 메서드

    ```pseudocode
    FUNCTION <view>:
    INPUTS: <current model state>
    OUTPUTS: <rendered string>

    IF <not ready>
        RETURN <initializing message>
    ENDIF

    IF <confirmed>
        RETURN <empty string>
    ENDIF

    Build string with:
        - Label text styled with PromptLabelStyle indicating user request input
        - Help text showing key bindings: Enter to confirm, Shift+Enter or Alt+Enter for newline, Ctrl+G for external editor
        - Empty line
        - Textarea view

    IF <error message is not empty>
        Append error message styled with ErrorStyle
    ENDIF

    Append trailing newline
    RETURN <built string>
    ```

11. `Result` 메서드
    - `UserRequestPromptResult`를 반환한다
    - confirmed 여부에 따라 Cancelled 필드 설정
    - confirmed 텍스트를 Text 필드에 설정

---

**파일 2: `ui/editor.go` (신규 생성)**

**목적:** 외부 에디터 해석 로직을 분리하여 단독 테스트 가능하게 한다.

**새로 도입할 심볼:**
- `EditorCommand` (구조체: 실행 파일명과 인자 슬라이스를 담는다)
- `resolveEditor` (함수)
- `ErrNoEditorFound` (에러 변수)

**편집 의도:**

1. `EditorCommand` 구조체 정의
   - 실행 파일명을 담는 필드
   - 인자 슬라이스를 담는 필드

2. `ErrNoEditorFound` 에러 변수 정의
   - 영어 메시지: "No editor found. Please set the EDITOR environment variable or install 'code' or 'vi'."

3. `resolveEditor` 함수

   ```pseudocode
   FUNCTION <resolve editor>:
   INPUTS: <environment variable lookup function, command existence check function>
   OUTPUTS: <editor command structure>
   FAILURE: <no editor available -> return ErrNoEditorFound>

   Look up EDITOR environment variable
   IF <EDITOR is set and non-empty>
       Split EDITOR value by whitespace into tokens
       RETURN <first token as executable, remaining tokens as arguments>
   ENDIF

   IF <code command exists on system>
       RETURN <code as executable, --wait as argument>
   ENDIF

   IF <vi command exists on system>
       RETURN <vi as executable, no arguments>
   ENDIF

   RETURN <error: ErrNoEditorFound>
   ```

   이 함수는 환경변수 조회 함수와 명령 존재 확인 함수를 인자로 받아서 테스트 시 모킹할 수 있도록 한다.

---

**파일 3: `ui/editor_test.go` (신규 생성)**

**목적:** 에디터 해석 로직의 단위 테스트를 작성한다.

**새로 도입할 심볼:**
- `TestResolveEditor_EditorEnvSet` (테스트 함수)
- `TestResolveEditor_EditorEnvWithArgs` (테스트 함수)
- `TestResolveEditor_EditorEnvEmpty` (테스트 함수)
- `TestResolveEditor_FallbackToCode` (테스트 함수)
- `TestResolveEditor_FallbackToVi` (테스트 함수)
- `TestResolveEditor_NoEditorAvailable` (테스트 함수)

**편집 의도:**

각 테스트에서 환경변수 조회 함수와 명령 존재 확인 함수를 모킹하여 `resolveEditor`를 테스트한다:

1. `TestResolveEditor_EditorEnvSet`: `EDITOR`가 "vim"으로 설정된 경우, 실행 파일이 "vim"이고 인자가 비어있는지 확인
2. `TestResolveEditor_EditorEnvWithArgs`: `EDITOR`가 "code --wait"로 설정된 경우, 실행 파일이 "code"이고 인자가 "--wait"인지 확인
3. `TestResolveEditor_EditorEnvEmpty`: `EDITOR`가 빈 문자열인 경우, fallback 우선순위가 적용되는지 확인
4. `TestResolveEditor_FallbackToCode`: `EDITOR` 미설정, `code` 명령 존재 시, "code"와 "--wait" 인자가 반환되는지 확인
5. `TestResolveEditor_FallbackToVi`: `EDITOR` 미설정, `code` 미존재, `vi` 존재 시, "vi"가 반환되는지 확인
6. `TestResolveEditor_NoEditorAvailable`: `EDITOR` 미설정, `code`와 `vi` 모두 미존재 시, `ErrNoEditorFound`가 반환되는지 확인

---

**파일 4: `ui/user_request_prompt_test.go` (신규 생성)**

**목적:** `UserRequestPromptModel`의 핵심 동작에 대한 단위 테스트를 작성한다.

**새로 도입할 심볼:**
- `TestUserRequestPromptModel_InitReturnsBlinkCmd` (테스트 함수)
- `TestUserRequestPromptModel_EmptyInputShowsError` (테스트 함수)
- `TestUserRequestPromptModel_WhitespaceOnlyInputShowsError` (테스트 함수)
- `TestUserRequestPromptModel_ValidInputConfirms` (테스트 함수)
- `TestUserRequestPromptModel_CtrlCCancels` (테스트 함수)
- `TestUserRequestPromptModel_ResultAfterConfirm` (테스트 함수)
- `TestUserRequestPromptModel_ResultAfterCancel` (테스트 함수)
- `TestUserRequestPromptModel_ErrorClearsOnInput` (테스트 함수)
- `TestUserRequestPromptModel_ViewShowsLabelAndHelp` (테스트 함수)
- `TestUserRequestPromptModel_ViewShowsInitializingBeforeReady` (테스트 함수)
- `TestUserRequestPromptModel_ViewShowsErrorMessage` (테스트 함수)
- `TestUserRequestPromptModel_WindowSizeUpdatesWidth` (테스트 함수)

**편집 의도:**

Bubble Tea 모델을 직접 호출하여 테스트한다 (Bubble Tea 프로그램을 실행하지 않고 `Update`와 `View`를 직접 호출):

1. `TestUserRequestPromptModel_InitReturnsBlinkCmd`: `Init`이 nil이 아닌 커맨드를 반환하는지 확인
2. `TestUserRequestPromptModel_EmptyInputShowsError`: `WindowSizeMsg` 전송 후 Enter 입력 시 에러 메시지가 설정되고 confirmed가 false인지 확인
3. `TestUserRequestPromptModel_WhitespaceOnlyInputShowsError`: 공백만 입력 후 Enter 시 에러 메시지 설정 확인
4. `TestUserRequestPromptModel_ValidInputConfirms`: 텍스트 입력 후 Enter 시 confirmed가 true이고 텍스트가 올바르게 저장되는지 확인
5. `TestUserRequestPromptModel_CtrlCCancels`: Ctrl+C 시 취소 상태가 되는지 확인
6. `TestUserRequestPromptModel_ResultAfterConfirm`: 확인 후 Result 메서드가 올바른 텍스트와 Cancelled=false를 반환하는지 확인
7. `TestUserRequestPromptModel_ResultAfterCancel`: 취소 후 Result 메서드가 Cancelled=true를 반환하는지 확인
8. `TestUserRequestPromptModel_ErrorClearsOnInput`: 에러 상태에서 일반 키 입력 시 에러 메시지가 비워지는지 확인
9. `TestUserRequestPromptModel_ViewShowsLabelAndHelp`: View 출력에 레이블과 키 바인딩 도움말이 포함되는지 확인 (`stripANSI` 활용)
10. `TestUserRequestPromptModel_ViewShowsInitializingBeforeReady`: `WindowSizeMsg` 전송 전 View가 "Initializing..." 문자열을 반환하는지 확인
11. `TestUserRequestPromptModel_ViewShowsErrorMessage`: 에러 상태에서 View에 에러 메시지가 포함되는지 확인
12. `TestUserRequestPromptModel_WindowSizeUpdatesWidth`: `WindowSizeMsg` 전송 후 ready 상태가 되고 너비가 설정되는지 확인

참고: 외부 에디터 연동 관련 테스트(`tea.ExecProcess` 포함)는 실제 프로세스 실행이 필요하므로 단위 테스트로 다루기 어렵다. 대신 에디터 해석 로직은 `editor_test.go`에서, 에디터 완료 메시지 처리 로직은 `editorFinishedMsg`를 직접 전달하여 테스트할 수 있다. 이를 위한 추가 테스트:

- `TestUserRequestPromptModel_EditorFinishedUpdatesTextarea`: 모킹된 임시 파일에 내용을 쓰고, `editorFinishedMsg`를 전달하여 textarea가 업데이트되는지 확인
- `TestUserRequestPromptModel_EditorFinishedWithError`: 에러가 포함된 `editorFinishedMsg`를 전달하여 에러 메시지가 설정되는지 확인

---

**파일 5: `app/app.go` (수정)**

**목적:** 워크스페이스 프롬프트 완료 후 요구사항 입력 프롬프트를 순차적으로 실행하도록 `Run` 함수를 확장한다.

**수정할 심볼:** `Run`

**수정 위치:** `Run` 함수 내, `fmt.Fprintf(stdout, "Workspace set to: %s\n", result.Path)` 라인 이후

**편집 의도:**

현재 `Run` 함수의 마지막 부분에서 워크스페이스 경로를 출력한 뒤 바로 `return nil`을 하는데, 이 부분을 수정하여 요구사항 입력 프롬프트를 추가로 실행한다:

```pseudocode
FUNCTION <extended Run flow after workspace confirmation>:
INPUTS: <workspace result path, stdout, stderr>
OUTPUTS: <error or nil>
FAILURE: <prompt failure -> return wrapped error>

Print workspace path to stdout

Create UserRequestPromptModel
Create new Bubble Tea program with the model and stderr output
Run the program
IF <program run fails>
    RETURN <wrapped error>
ENDIF

Extract UserRequestPromptResult from final model
IF <user cancelled>
    RETURN <nil>
ENDIF

RETURN <nil>
```

현재 코드에서 `fmt.Fprintf(stdout, "Workspace set to: %s\n", result.Path)` 라인과 `return nil` 라인 사이에 요구사항 프롬프트 실행 로직을 삽입한다.

---

**파일 6: `ui/styles.go` (수정 — 필요한 경우)**

**목적:** 요구사항 프롬프트에서 사용할 새 스타일이 필요한 경우 추가한다.

**수정 위치:** 기존 스타일 변수 선언 블록 끝 부분

**편집 의도:**

기존 스타일(`PromptLabelStyle`, `ErrorStyle` 등)이 이미 요구사항 프롬프트에 필요한 스타일을 충분히 제공하므로, 추가 스타일이 필요하지 않을 가능성이 높다. 단, 도움말 텍스트(키 바인딩 안내)에 별도 스타일이 필요하다고 판단되면 `HelpTextStyle`을 추가한다. 기존 `WorkspacePromptModel`의 View에서 도움말 텍스트는 스타일 없이 일반 텍스트로 출력하고 있으므로, 일관성을 위해 요구사항 프롬프트도 동일하게 스타일 없이 출력한다. 따라서 이 파일의 변경은 불필요할 것으로 예상된다.

---

## 5. Testing & Validation

### 단위 테스트

| 테스트 파일 | 검증 대상 | 핵심 시나리오 |
|---|---|---|
| `ui/editor_test.go` | `resolveEditor` 함수 | EDITOR 환경변수 설정/미설정, 인자 포함, 빈 문자열, code/vi fallback, 에디터 없음 |
| `ui/user_request_prompt_test.go` | `UserRequestPromptModel` | 초기화, 빈 입력 검증, 유효 입력 확정, 취소, 에러 클리어, View 렌더링, 반응형 너비, 에디터 완료 메시지 처리 |

### 수동 테스트 체크리스트

1. 프로그램 실행 → 워크스페이스 확정 → 요구사항 프롬프트 표시 확인
2. 안내 레이블과 키 바인딩 도움말 표시 확인
3. 자유 형식 텍스트 입력 후 화면 표시 확인
4. 터미널 가로 폭 초과 텍스트의 자동 wrap 확인
5. 클립보드에서 긴 텍스트 붙여넣기 확인
6. Shift+Enter로 줄바꿈 삽입 확인
7. Alt+Enter로 줄바꿈 삽입 확인
8. 방향키(상하좌우)로 커서 이동 확인
9. 빈 입력 상태에서 Enter → 에러 메시지 표시 확인
10. 공백만 입력 후 Enter → 에러 메시지 표시 확인
11. 에러 메시지 표시 후 타이핑 시 에러 메시지 소멸 확인
12. 유효 텍스트 입력 후 Enter → 정상 종료 확인
13. Ctrl+C → 프로그램 종료 확인
14. Ctrl+G → 외부 에디터 실행 확인 (EDITOR 환경변수 설정 시)
15. 에디터에서 편집 후 저장/닫기 → textarea에 내용 반영 확인
16. 에디터에서 기존 textarea 내용이 임시 파일에 미리 채워져 있는지 확인
17. EDITOR 미설정, code 명령 있을 때 code --wait로 열리는지 확인
18. EDITOR 미설정, code 없고 vi 있을 때 vi로 열리는지 확인
19. EDITOR 미설정, code/vi 모두 없을 때 에러 메시지 표시 후 프롬프트 복귀 확인
20. 터미널 크기 변경 시 textarea 너비 조정 확인
21. 워크스페이스 프롬프트에서 Ctrl+C 시 요구사항 프롬프트 표시 안 됨 확인

### 자동 테스트 실행 명령
- `go test ./ui/... ./app/...`
- `go vet ./...`

---

## 6. Risk Analysis

### 기술적 리스크

| 리스크 | 영향 | 완화 방안 |
|---|---|---|
| `tea.ExecProcess`가 특정 터미널에서 올바르게 동작하지 않을 수 있음 | 외부 에디터 기능 불능 | 에러 핸들링을 통해 실패 시 프롬프트로 복귀하도록 처리. 에디터 기능은 핵심 입력 흐름에 영향 없음 |
| Bubble Tea textarea 컴포넌트의 Shift+Enter 기본 동작 | 줄바꿈 입력이 안 될 수 있음 | `WorkspacePromptModel`에서 이미 `InsertNewline`을 비활성화하고 Alt+Enter로 수동 처리하는 패턴을 사용 중. Shift+Enter도 동일하게 키 핸들러에서 수동 처리 |
| 임시 파일이 정리되지 않을 수 있음 (비정상 종료 시) | 임시 파일 잔존 | OS 임시 디렉토리에 생성하므로 OS가 정기적으로 정리함. 추가로 `defer`를 사용하여 가능한 한 정리 |

### 보안 고려사항
- `EDITOR` 환경변수를 통한 명령 주입: `exec.Command`에 인자를 배열로 전달하므로 shell injection 위험 없음.
- 임시 파일은 `os.CreateTemp`를 사용하여 기본 보안 권한(0600)으로 생성됨.

### Rollback 전략
- 새 파일 추가와 `app.go`의 소규모 수정으로 구성되므로, rollback 시 새 파일 삭제 및 `app.go` 원복으로 충분하다.
- 기존 `WorkspacePromptModel`과 기타 코드는 변경하지 않으므로 기존 기능에 영향 없음.

---

## 7. Implementation Notes

### 빌드 및 테스트 명령
- 빌드: `go build ./...`
- 테스트: `go test ./ui/... ./app/...`
- vet: `go vet ./...`

### 리포지토리 컨벤션
- Bubble Tea 모델 패턴: `Init`/`Update`/`View`/`Result` 4개 메서드 구조
- `Update`에서 `tea.WindowSizeMsg`를 최우선 처리하고, ready 플래그가 false이면 다른 메시지 무시
- 키 처리는 별도 private 메서드로 분리 (예: `handleKey`, `handleEnter`)
- 스타일은 `ui/styles.go`의 전역 변수로 정의
- 에러 메시지는 View에서 `ErrorStyle`로 렌더링
- 테스트에서 ANSI 코드 제거가 필요한 경우 `stripANSI` 함수 활용 (이미 `banner_test.go`에 존재하나 export되어 있지 않으므로 테스트 파일 내에서 필요시 동일하게 정의하거나, 테스트 헬퍼로 공유)
- 에러 변수는 `var ErrXxx = errors.New(...)` 형태로 선언
- 함수 50줄 이하 유지, 단일 책임 원칙 준수, 3단계 이상 들여쓰기 금지

### Shift+Enter 처리 관련 참고사항
- Bubble Tea에서 Shift+Enter는 터미널에 따라 다르게 보고될 수 있다. 일반적으로 `shift+enter`로 보고되지만, 일부 터미널에서는 다른 키 시퀀스를 보낼 수 있다. `WorkspacePromptModel`에서는 이미 Alt+Enter만 줄바꿈으로 처리하고 있으므로, 요구사항 프롬프트에서는 `shift+enter`와 `alt+enter` 모두를 줄바꿈 키로 처리한다. 구현 시 Bubble Tea의 `msg.String()` 반환값을 확인하여 적절한 키 문자열을 매칭해야 한다.
