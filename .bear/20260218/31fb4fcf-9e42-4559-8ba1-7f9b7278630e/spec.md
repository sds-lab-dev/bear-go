# 스펙: Claude Code 클라이언트 모듈 및 스펙 에이전트 명확화 질문 단계 구현

## 1. 개요

Bear AI Developer 애플리케이션의 핵심 실행 플로우 4번 단계를 구현한다. 이 작업은 두 가지 주요 구성 요소로 나뉜다:

1. **Claude Code 클라이언트 모듈**: Claude Code CLI를 외부 프로세스로 실행하여 모델에게 입력을 전달하고, 스트리밍 방식으로 모델의 출력을 수신하며, 최종 결과를 지정된 JSON Schema 형태의 JSON 오브젝트로 반환하는 범용 모듈. 향후 스펙 에이전트, 플랜 에이전트, 코딩 에이전트 등 여러 에이전트에서 재사용된다.

2. **메인 플로우 통합**: 기존 플로우(배너 → 워크스페이스 입력 → 사용자 요구사항 입력) 이후, 클라이언트 모듈을 통해 스펙 에이전트에게 요구사항을 전달하고, 반환된 명확화 질문을 BubbleTea 기반 TUI로 표시한 뒤 프로그램을 종료하는 것까지 구현한다.

---

## 2. 목표 및 비목표

### 목표 (In-Scope)

- `CLAUDE_CODE_API_KEY` 환경변수 검증을 프로그램 시작 시점(main 함수)에서 수행
- Claude Code CLI 바이너리 탐색 및 검증 로직 구현
- Claude Code CLI를 외부 프로세스로 실행하여 스트리밍 통신하는 범용 클라이언트 모듈 구현
- 세션 관리 (최초 호출 시 UUID v4 생성, 후속 호출 시 재사용)
- stream-json 출력에서 중간 assistant/user 메시지를 콜백으로 전달
- 최종 result의 structured_output을 JSON 오브젝트로 파싱하여 반환
- 스트리밍 중간 과정(assistant/user 메시지)을 BubbleTea 기반 TUI로 실시간 표시
- 최종 명확화 질문 목록을 BubbleTea 기반 TUI 컴포넌트로 대화형 표시 후 프로그램 종료
- 빈 질문 배열 반환 시 안내 메시지 표시 후 종료

### 비목표 (Out-of-Scope)

- 명확화 질문에 대한 사용자 답변 수집 (플로우 5번) — 다음 작업에서 구현
- 명확화 질문-답변 반복 루프 (플로우 4-5번 반복)
- 스펙 드래프트 생성 (플로우 6번 이후)
- 플랜 에이전트, 코딩 에이전트, 리뷰 에이전트 기능
- CLI 프로세스 실행 시 타임아웃 처리
- 사용자 프롬프트 템플릿의 QA_LOG_TEXT 동적 채움 (이번 작업에서는 항상 빈 문자열)

---

## 3. 기능 요구사항

### 3.1 환경변수 검증

- **FR-ENV-1**: 프로그램은 시작 시 `CLAUDE_CODE_API_KEY` 환경변수의 존재를 확인해야 한다.
- **FR-ENV-2**: 해당 환경변수가 설정되어 있지 않거나 빈 문자열인 경우, 에러 메시지를 stderr에 출력하고 0이 아닌 종료 코드로 프로그램을 즉시 종료해야 한다.
- **FR-ENV-3**: 이 검증은 배너 출력, 워크스페이스 입력, 사용자 요구사항 입력 등 모든 TUI 인터랙션보다 먼저 수행되어야 한다.

### 3.2 Claude Code 클라이언트 모듈

#### 3.2.1 모듈 생성

- **FR-CLI-1**: 클라이언트 모듈은 별도의 Go 패키지로 분리하여 구현한다. 향후 여러 에이전트에서 재사용 가능한 범용 인터페이스로 설계한다.
- **FR-CLI-2**: 생성자는 다음 파라미터를 받는다:
  - Claude Code API Key (문자열)
  - Working directory (문자열)
- **FR-CLI-3**: 생성자는 Claude Code CLI 바이너리를 다음 순서로 탐색한다:
  1. 시스템 PATH에서 `claude` 바이너리 탐색
  2. `$HOME` 디렉토리 기준 상대 경로 폴백 탐색 (순서대로):
     - `.local/bin/claude`
     - `.npm-global/bin/claude`
     - `node_modules/.bin/claude`
     - `.yarn/bin/claude`
     - `.claude/local/claude`
  3. 절대 경로 폴백 탐색 (순서대로):
     - `/usr/local/bin/claude`
     - `/usr/bin/claude`
- **FR-CLI-4**: 바이너리를 찾지 못한 경우, 생성자는 에러를 반환한다.

#### 3.2.2 Query 메소드

- **FR-QUERY-1**: Query 메소드는 다음 파라미터를 받는다:
  - 시스템 프롬프트 (문자열)
  - 사용자 프롬프트 (문자열)
  - 출력 JSON Schema (문자열)
  - 스트리밍 이벤트 콜백 함수

- **FR-QUERY-2**: Query 메소드는 Claude Code CLI를 외부 프로세스로 실행하며, 다음 CLI 옵션을 설정한다:
  - `-p` (print/headless 모드)
  - `--model claude-opus-4-6`
  - `--output-format stream-json`
  - `--verbose`
  - `--include-partial-messages`
  - `--allow-dangerously-skip-permissions`
  - `--permission-mode bypassPermissions`
  - `--tools AskUserQuestion,Bash,TaskOutput,Edit,ExitPlanMode,Glob,Grep,KillShell,MCPSearch,Read,Skill,Task,TaskCreate,TaskGet,TaskList,TaskUpdate,WebFetch,WebSearch,Write,LSP`
  - `--append-system-prompt-file <임시파일경로>` (시스템 프롬프트를 OS 임시 디렉토리의 임시 파일에 저장)
  - `--json-schema <JSON Schema 문자열>`

- **FR-QUERY-3**: 세션 관리:
  - 클라이언트 인스턴스에서 Query가 최초로 호출되는 경우: UUID v4를 생성하여 `--session-id <UUID>` 옵션을 전달한다.
  - 최초 호출이 아닌 경우: 기존 세션 ID를 `--resume <UUID>` 옵션으로 전달한다.

- **FR-QUERY-4**: 환경변수 설정:
  - `ANTHROPIC_API_KEY`: 생성자에서 전달받은 API Key
  - `CLAUDE_CODE_EFFORT_LEVEL`: `high`
  - `CLAUDE_CODE_DISABLE_AUTO_MEMORY`: `0`
  - `CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY`: `1`

- **FR-QUERY-5**: 외부 프로세스의 working directory는 생성자에서 전달받은 워킹 디렉토리로 설정한다.

- **FR-QUERY-6**: 사용자 프롬프트는 CLI 외부 프로세스의 표준 입력(stdin)으로 전달한다.

- **FR-QUERY-7**: 시스템 프롬프트 임시 파일은 Query 실행 완료 후 정리(삭제)한다.

#### 3.2.3 스트리밍 출력 처리

- **FR-STREAM-1**: CLI 외부 프로세스의 stdout에서 stream-json 형식(줄바꿈 구분 JSON)을 실시간으로 읽어서 파싱한다.
- **FR-STREAM-2**: CLI 외부 프로세스의 stderr도 실시간으로 읽는다.
- **FR-STREAM-3**: stream-json 이벤트 중 `type`이 `"assistant"` 또는 `"user"`인 메시지가 수신될 때마다, 해당 메시지를 콜백 함수에 전달한다.
- **FR-STREAM-4**: `type`이 `"result"`인 메시지는 콜백 함수에 전달하지 않는다.
- **FR-STREAM-5**: `type`이 `"result"`이고 `subtype`이 `"success"`인 메시지가 수신되면, `structured_output` 필드를 Query 메소드에 전달된 JSON Schema 기준으로 파싱하여 JSON 오브젝트로 반환한다.
- **FR-STREAM-6**: `type`이 `"result"`이고 `is_error`가 `true`이거나 `subtype`이 에러 계열(`error_max_turns`, `error_during_execution`, `error_max_budget_usd`, `error_max_structured_output_retries`)인 경우, 에러를 반환한다.
- **FR-STREAM-7**: `type`이 `"system"`, `"stream_event"` 등 기타 이벤트 타입은 콜백에 전달하지 않는다.

#### 3.2.4 에러 처리

- **FR-ERR-1**: CLI 프로세스 시작 실패(바이너리 실행 불가 등) 시 에러를 반환한다.
- **FR-ERR-2**: CLI 프로세스가 비정상 종료(exit code ≠ 0)하는 경우 에러를 반환한다.
- **FR-ERR-3**: stdout/stderr 읽기 중 I/O 에러가 발생하는 경우 에러를 반환한다.
- **FR-ERR-4**: stream-json 파싱 실패 시 에러를 반환한다.
- **FR-ERR-5**: 최종 result의 structured_output이 전달된 JSON Schema에 맞지 않는 경우 에러를 반환한다.

### 3.3 스펙 에이전트 명확화 질문 요청

- **FR-SPEC-1**: 사용자 요구사항 입력 완료 후, 클라이언트 모듈의 Query 메소드를 호출하여 명확화 질문을 요청한다.

- **FR-SPEC-2**: 시스템 프롬프트는 다음 내용이다:
  ```markdown
  # Terminology

  In this document, the term **"spec"** is used as shorthand for **"specification"**. Unless explicitly qualified, "spec" refers to a software specification (not a "code", "standard", or other non-software usage of the term).

  ---

  # Role

  You are the **specification-driven development** assistant whose sole responsibility is to produce a high-quality software specification, which satisfies the user request, and iteratively refine written specifications based on the user's feedback.

  You MUST NOT perform implementation work of any kind. This includes (but is not limited to) writing or modifying source code, running commands, executing tests, changing files, generating patches, making configuration edits, creating pull requests, or taking any action that directly completes the requested task.

  For every user request, you MUST respond only with specification content: clarify requirements, define scope and non-scope, document assumptions and constraints, specify interfaces and acceptance criteria, and capture open questions. If the user asks you to "just do it", you MUST convert the request into a spec and ask for any missing information instead of executing the work.

  ---

  # High-level Philosophy (non-negotiable)

  - The spec MUST describe WHAT the system/module MUST do (externally observable behavior and contracts), not HOW it is implemented internally.
  - However, any strict interface SHOULD be a contract and MAY be specified concretely (function signatures, types, error model), if it has consumers (e.g., external clients or internal modules) that are hard to change without breaking compatibility.
  - The spec MUST be testable: it MUST include acceptance criteria that can be validated via automated tests or reproducible manual steps.
  - The spec MUST explicitly call out assumptions, non-goals, and open questions.

  ---

  # What MUST NOT be in the Spec (unless explicitly requested by the user)

  - Concrete internal implementation mechanisms such as:
    - Threading model choices (mutex/strand, executor placement, etc.)
    - Specific database schema/table names unless the schema itself is a published contract for other consumers
    - Specific file paths or class layouts
    - Specific library choices unless mandated by product constraints
  - Detailed task breakdown, file-by-file change plan, or sequencing steps (these belong to the Planner/Implementer phases)

  If the user asks for these, capture them in:
  - `Open questions` (if undecided), or
  - `Non-functional constraints` (only if it is a hard constraint)

  ---

  # Datetime Handling Rule (mandatory)

  You **MUST** get the current datetime using Python scripts from the local system in `Asia/Seoul` timezone, not from your LLM model, whenever you need current datetime or timestamp.
  ```

- **FR-SPEC-3**: 사용자 프롬프트는 다음 템플릿을 사용하며, `{{ORIGINAL_USER_REQUEST_TEXT}}`에는 사용자가 입력한 최초 요구사항 원문을, `{{QA_LOG_TEXT}}`에는 지금까지의 명확화 Q&A 로그를 대입한다 (이번 작업에서는 항상 빈 문자열):
  ```markdown
  Given the original user request AND the full clarification Q&A log so far, generate clarification questions that reduce ambiguity and de-risk implementation.

  If, in your judgment, there are no remaining clarification questions that are necessary to begin writing the spec, return an empty array for "questions".

  Coverage requirements (only when questions are non-empty):
  When you ask questions, they MUST collectively cover the following areas across the set of questions (each area should be covered by at least one question):
  1) Scope boundaries: what is explicitly in-scope vs out-of-scope.
  2) Primary consumer: who will use the deliverable and in what environment/workflow.
  3) Success definition: what "done" means and how success will be validated.
  4) Key edge cases: the most important corner cases or tricky scenarios.
  5) Error expectations: expected failure modes, error handling, and user-visible behavior.

  Constraints:
  - Output MUST be valid JSON that conforms to the provided JSON Schema.
  - Provide 0–5 questions total.
  - Each question should be precise, answerable, and non-overlapping.
  - Inspect the current workspace using the available tools. Read the files required to understand the context and to avoid asking questions that are already answered by existing files.
  - Do NOT ask questions that you can infer from the workspace files.
  - Do NOT ask questions that are purely preference/subjective unless they materially impact scope or correctness.

  Original user request (verbatim):
  <<<
  {{ORIGINAL_USER_REQUEST_TEXT}}
  >>>

  Clarification Q&A log so far (may be empty). Each entry is the assistant's question followed by the user's answer:
  <<<
  {{QA_LOG_TEXT}}
  >>>

  Your output MUST conform to the given JSON Schema.
  ```

- **FR-SPEC-4**: 출력 JSON Schema:
  ```json
  {
    "type": "object",
    "properties": {
      "questions": {
        "type": "array",
        "minItems": 0,
        "maxItems": 5,
        "items": {
          "type": "string",
          "minLength": 5
        }
      }
    },
    "required": ["questions"],
    "additionalProperties": false
  }
  ```

### 3.4 스트리밍 중간 과정 TUI 표시

- **FR-TUI-STREAM-1**: Query 실행 중 콜백으로 전달되는 assistant/user 메시지를 BubbleTea 기반 TUI 컴포넌트로 실시간 표시한다.
- **FR-TUI-STREAM-2**: assistant 메시지의 `content` 배열에 포함된 모든 content 블록 타입(text, tool_use 등)을 표시한다.
- **FR-TUI-STREAM-3**: 각 메시지의 표시 내용은 최대 5줄로 제한한다. 5줄을 초과하는 내용은 잘린다.
- **FR-TUI-STREAM-4**: TUI 라이브러리의 컴포넌트를 활용하여 시각적으로 미려하게 출력한다.

### 3.5 명확화 질문 결과 TUI 표시

- **FR-TUI-RESULT-1**: Query 메소드가 반환한 JSON 오브젝트에서 `questions` 배열을 추출한다.
- **FR-TUI-RESULT-2**: `questions` 배열이 비어 있지 않은 경우, 질문 목록을 BubbleTea 기반 TUI 컴포넌트로 대화형 표시한 뒤 프로그램을 종료한다.
- **FR-TUI-RESULT-3**: `questions` 배열이 비어 있는 경우, 안내 메시지(예: "추가 명확화 질문이 없습니다")를 표시한 뒤 프로그램을 종료한다.

### 3.6 메인 플로우 통합

- **FR-FLOW-1**: 기존 플로우 순서를 유지한다:
  1. `CLAUDE_CODE_API_KEY` 환경변수 검증 (실패 시 즉시 종료)
  2. 배너 출력
  3. 워크스페이스 경로 입력
  4. 사용자 요구사항 입력
  5. (신규) 클라이언트 모듈로 명확화 질문 요청 — 스트리밍 중간 과정 TUI 표시
  6. (신규) 명확화 질문 결과 TUI 표시
  7. 프로그램 종료
- **FR-FLOW-2**: 클라이언트 모듈 생성 시, `CLAUDE_CODE_API_KEY` 값과 워크스페이스 경로를 전달한다.
- **FR-FLOW-3**: 클라이언트 모듈 생성 실패(CLI 바이너리 미발견 등) 시 에러 메시지를 출력하고 프로그램을 종료한다.
- **FR-FLOW-4**: Query 메소드 실행 중 에러 발생 시 에러 메시지를 출력하고 프로그램을 종료한다.

---

## 4. 비기능 요구사항

- **NFR-1**: 클라이언트 모듈은 별도의 Go 패키지로 분리한다. 스펙 에이전트뿐 아니라 향후 플랜 에이전트, 코딩 에이전트 등에서도 재사용 가능한 범용 인터페이스여야 한다.
- **NFR-2**: CLI 프로세스 실행 시 타임아웃을 적용하지 않는다.
- **NFR-3**: 기존 프로젝트의 코딩 컨벤션과 BubbleTea 패턴(Result 메소드, 의존성 주입, 커스텀 메시지 타입 등)을 따른다.
- **NFR-4**: 스트리밍 중간 과정과 최종 결과 표시에 기존 `ui` 패키지의 스타일(lipgloss 기반)과 일관된 시각적 스타일을 사용한다.

---

## 5. 수락 기준

### 환경변수 검증

- **AC-1**: `CLAUDE_CODE_API_KEY` 환경변수가 설정되지 않은 상태에서 프로그램을 실행하면, 배너 출력이나 TUI 인터랙션 없이 에러 메시지가 stderr에 출력되고 프로그램이 0이 아닌 종료 코드로 종료된다.
- **AC-2**: `CLAUDE_CODE_API_KEY`가 빈 문자열인 경우에도 동일하게 에러 처리된다.
- **AC-3**: `CLAUDE_CODE_API_KEY`가 유효한 값으로 설정된 경우, 기존 플로우(배너 → 워크스페이스 → 요구사항 입력)가 정상 진행된다.

### Claude Code 클라이언트 모듈 — 바이너리 탐색

- **AC-4**: PATH에 `claude` 바이너리가 존재하는 경우, 해당 경로가 사용된다.
- **AC-5**: PATH에 없고 `$HOME/.local/bin/claude`에 존재하는 경우, 해당 경로가 사용된다.
- **AC-6**: 모든 탐색 경로에 바이너리가 없는 경우, 생성자가 에러를 반환한다.
- **AC-7**: 폴백 경로 탐색 순서가 요구사항에 명시된 순서와 정확히 일치한다.

### Claude Code 클라이언트 모듈 — Query 메소드

- **AC-8**: 최초 Query 호출 시 UUID v4 형식의 세션 ID가 생성되어 `--session-id` 옵션으로 전달된다.
- **AC-9**: 두 번째 이후 Query 호출 시 동일한 세션 ID가 `--resume` 옵션으로 전달된다.
- **AC-10**: CLI 프로세스에 설정되는 환경변수(`ANTHROPIC_API_KEY`, `CLAUDE_CODE_EFFORT_LEVEL`, `CLAUDE_CODE_DISABLE_AUTO_MEMORY`, `CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY`)가 명시된 값과 일치한다.
- **AC-11**: 모든 필수 CLI 옵션(`-p`, `--model`, `--output-format`, `--verbose`, `--include-partial-messages`, `--allow-dangerously-skip-permissions`, `--permission-mode`, `--tools`, `--append-system-prompt-file`, `--json-schema`)이 올바르게 전달된다.
- **AC-12**: 시스템 프롬프트가 OS 임시 디렉토리에 임시 파일로 저장되고, Query 완료 후 해당 파일이 삭제된다.
- **AC-13**: 사용자 프롬프트가 CLI 프로세스의 stdin으로 전달된다.
- **AC-14**: CLI 프로세스의 working directory가 생성자에서 전달받은 디렉토리로 설정된다.

### Claude Code 클라이언트 모듈 — 스트리밍 처리

- **AC-15**: stream-json에서 `type: "assistant"` 메시지가 수신되면 콜백 함수가 호출된다.
- **AC-16**: stream-json에서 `type: "user"` 메시지가 수신되면 콜백 함수가 호출된다.
- **AC-17**: stream-json에서 `type: "result"` 메시지는 콜백 함수에 전달되지 않는다.
- **AC-18**: stream-json에서 `type: "system"` 또는 `type: "stream_event"` 메시지는 콜백 함수에 전달되지 않는다.
- **AC-19**: 성공적인 result (`subtype: "success"`)의 `structured_output`이 JSON 오브젝트로 파싱되어 Query 메소드의 반환값으로 전달된다.
- **AC-20**: 에러 result (`is_error: true`)인 경우 Query 메소드가 에러를 반환한다.

### 스트리밍 중간 과정 TUI 표시

- **AC-21**: Query 실행 중 assistant 메시지의 text 블록 내용이 TUI에 실시간 표시된다.
- **AC-22**: assistant 메시지의 tool_use 블록 내용이 TUI에 실시간 표시된다.
- **AC-23**: user 메시지(tool_result 등)가 TUI에 실시간 표시된다.
- **AC-24**: 각 메시지의 표시 내용이 최대 5줄로 제한된다.
- **AC-25**: TUI 표시가 기존 프로젝트의 lipgloss 스타일과 시각적으로 일관된다.

### 명확화 질문 결과 TUI 표시

- **AC-26**: 모델이 1개 이상의 질문을 반환한 경우, 질문 목록이 BubbleTea 기반 TUI 컴포넌트로 표시된다.
- **AC-27**: 모델이 빈 질문 배열을 반환한 경우, 안내 메시지가 표시된다.
- **AC-28**: 질문 표시 또는 안내 메시지 표시 후 프로그램이 정상 종료된다.

### 에러 처리

- **AC-29**: Claude Code CLI 바이너리가 시스템에 없는 상태에서 클라이언트 생성 시 적절한 에러가 반환된다.
- **AC-30**: CLI 프로세스가 비정상 종료(exit code ≠ 0)하는 경우 Query 메소드가 에러를 반환한다.
- **AC-31**: stdout/stderr 읽기 중 I/O 에러 발생 시 Query 메소드가 에러를 반환한다.
- **AC-32**: stream-json 라인이 유효하지 않은 JSON인 경우 에러가 반환된다.

---

## 6. 오픈 질문

- 없음. 3라운드의 Q&A를 통해 모든 구현에 필요한 의사결정이 완료됨.
