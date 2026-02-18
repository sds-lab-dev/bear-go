# 프로그램 소개

우리가 만들고자 하는 프로그램은 **“Bear AI Developer”**라는 이름의 TUI 기반 애플리케이션이야. 이 프로그램은 Claude Code CLI 위에서 명세 기반 소프트웨어 개발을 지원하는 도구로, Go 언어로 작성될 예정이야.

# 구현 방식

앞으로 개발은 탑다운 방식을 사용할거야. 우선 프로그램의 핵심 실행 플로우를 따라가면서 요구사항들을 나열하고, 요구사항들을 구현하기 위해서 필요한 모듈들을 만드는데 각 모듈의 세부 기능은 일단 인터페이스만 정의해놓고 실제 동작은 나중에 구현할거야. 초기 단계에서 각 모듈이 구현해야 하는 모든 기능들을 다 정의하지는 않고 일단 핵심 플로우를 따라가면서 필요한 기능들부터 먼저 정의하고 나머지는 추가 구현 단계에서 정의할거야. 이렇게 하면 초기 단계에서 너무 많은 세부사항들을 고민하지 않고도 전체적인 구조와 흐름을 잡아갈 수 있을거야.

# 핵심 실행 플로우

1. (구현 완료) 사용자가 프로그램을 실행하면 먼저 배너가 출력된다. 
2. (구현 완료) 현재 워킹 디렉토리를 보여주면서 사용자로부터 워크스페이스 경로를 입력받는다.
3. (구현 완료) 사용자 요구사항을 입력 받는다.
4. (미구현) 스펙 에이전트가 요구사항을 분석해서 명확화 질문이 있으면 사용자에게 질문한다.
5. (미구현) 명확화 질문에 대한 사용자 답변을 받는다. 필요하다면 4번으로 돌아가서 추가 명확화 질문을 반복한다.
6. (미구현) 명확화가 완료되면 스펙 에이전트가 드래프트 스펙을 생성해서 사용자에게 보여준다.
7. (미구현) 사용자가 드래프트 스펙을 검토하고 피드백을 입력한다.
8. (미구현) 스펙 에이전트가 사용자 피드백을 반영해서 드래프트 스펙을 수정한다. 필요하다면 7번으로 돌아가서 추가 피드백을 반복한다.
9. (미구현) 최종 스펙이 확정되면 플랜 에이전트가 사용자 요규사항과 최종 스펙을 분석해서 명확화 질문이 있으면 사용자에게 질문한다.
10. (미구현) 명확화 질문에 대한 사용자 답변을 받는다. 필요하다면 9번으로 돌아가서 추가 명확화 질문을 반복한다.
11. (미구현) 명확화가 완료되면 플랜 에이전트가 드래프트 플랜을 생성해서 사용자에게 보여준다.
12. (미구현) 사용자가 드래프트 플랜을 검토하고 피드백을 입력한다.
13. (미구현) 플랜 에이전트가 사용자 피드백을 반영해서 드래프트 플랜을 수정한다. 필요하다면 12번으로 돌아가서 추가 피드백을 반복한다.
14. (미구현) 최종 플랜이 확정되면 코딩 에이전트가 플랜을 분석해서 개별 태스크들의 상호의존성을 고려한 최적의 실행 순서를 계획한다.
15. (미구현) 코딩 에이전트가 메인 브랜치로부터 통합 브랜치를 생성하고, 각각의 태스크별로 통합 브랜치를 베이스로 워크트리를 만들어서 코드를 구현한다.
16. (미구현) 리뷰어 에이전트가 스펙, 플랜, 구현된 코드를 검토하고 피드백을 제공한다.
17. (미구현) 코딩 에이전트가 리뷰어 피드백을 반영해서 코드를 수정한다. 16번으로 돌아가서 리뷰 승인될 때까지 반복한다.
18. (미구현) 구현 완료된 태스크의 워크트리는 통합 브랜치로 병합되고 모든 태스크가 완료되기를 대기한다.
19. (미구현) 모든 태스크가 완료되면 사용자가 최종 결과물을 검토하고 승인한다. 
20. (미구현) 만약 수정 사항이 있으면 통합 브랜치에서 새로운 워크트리를 만들고 코딩 에이전트가 사용자 피드백을 반영해서 코드를 수정한다. 19번으로 돌아가서 사용자 승인을 받을 때까지 반복한다. 
21. (미구현) 최종 결과물이 승인되면 (아직 통합 브랜치에 병합되지 않은 워크트리가 있다면 병합하고) 통합 브랜치를 메인 브랜치로 병합하고 종료한다.

# 작업 지시

- 이번 작업에서는 위 실행 플로우의 4번 "스펙 에이전트가 요구사항을 분석해서 명확화 질문이 있으면 사용자에게 질문" 단계만 구현한다.
- 우선 Claude Code CLI를 외부 프로세스로 실행해서 모델에게 입력을 전달하고, 모델의 출력을 지정된 JSON Schema 형태로 받아오는 클라이언트 모듈을 구현한다.
  - 모듈의 생성자에는 다음 항목들이 파라메터로 전달되어야 한다.
    - Claude Code API Key
    - Working directory
  - 모듈의 생성자는 다음과 같은 작업을 수행한다.
    - Claude Code CLI가 시스템에 설치되어 있는지 확인한다. 만약 설치되어 있지 않다면 에러를 반환한다.
    - Claude Code CLI 바이너리를 찾는 루틴은 아래 코드를 참고해서 동일하게 구현한다.
      ```rust
        const ABSOLUTE_FALLBACK_PATHS: &[&str] = &[
            "/usr/local/bin/claude",
            "/usr/bin/claude",
        ];

        const HOME_RELATIVE_FALLBACK_PATHS: &[&str] = &[
            ".local/bin/claude",
            ".npm-global/bin/claude",    
            "node_modules/.bin/claude",
            ".yarn/bin/claude",
            ".claude/local/claude"
        ];

        pub fn find_claude_binary() -> Result<PathBuf, ClaudeCodeClientError> {
            // PATH에서 먼저 찾아본다.
            if let Ok(path) = which::which("claude") {
                return Ok(path);
            }

            // 없으면 HOME 디렉토리 밑에서 폴백 바이너리 경로를 찾아본다.
            if let Some(home_directory) = std::env::var_os("HOME") {
                let home = PathBuf::from(home_directory);
                for relative_path in HOME_RELATIVE_FALLBACK_PATHS {
                    let candidate = home.join(relative_path);
                    if candidate.exists() {
                        return Ok(candidate);
                    }
                }
            }

            // 없으면 절대경로로 폴백 바이너리 경로를 찾아본다.
            for absolute_path in ABSOLUTE_FALLBACK_PATHS {
                let candidate = PathBuf::from(absolute_path);
                if candidate.exists() {
                    return Ok(candidate);
                }
            }

            Err(ClaudeCodeClientError::BinaryNotFound)
        }
      ```
  - Query 메소드는 Claude Code CLI를 호출해서 모델에게 입력을 전달하고 모델의 출력을 스트리밍 방식으로 받아온다.
    - Query 메소드의 입력 파라메터:
      - 모델에게 전달할 시스템 프롬프트
      - 모델에게 전달할 사용자 프롬프트
      - 모델의 출력 JSON Schema 
      - 모델의 스트리밍 출력을 실시간으로 처리할 콜백 함수
    - Query 메소드의 동작:
      - Claude Code CLI 바이너리를 외부 프로세스로 실행한다.
      - Claude Code CLI 외부 프로세스의 working directory는 모듈 생성자에서 전달받은 워킹 디렉토리로 설정한다.
      - 이 클라이언트 모듈에서 Query 메소드가 최초로 호출된 경우:
        - UUIDv4로 고유 세션 ID를 생성해서 `--session-id` 옵션에 전달한다.
      - 이 클라이언트 모듈에서 Query 메소드가 최초로 호출된게 아닌 경우:
        - 최초 호출 시 생성된 세션 ID를 재사용해서 `--resume` 옵션에 전달한다.
      - `--model` 옵션에는 `claude-opus-4-6`을 전달한다.
      - `ANTHROPIC_API_KEY` 환경변수에는 모듈 생성자에서 전달받은 API Key를 설정한다.
      - `CLAUDE_CODE_EFFORT_LEVEL` 환경변수에는 `high` 값을 설정한다.
      - `CLAUDE_CODE_DISABLE_AUTO_MEMORY` 환경변수에는 `0` 값을 설정한다.
      - `CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY` 환경변수에는 `1` 값을 설정한다.
      - `-p` 옵션을 설정한다.
      - `--allow-dangerously-skip-permissions` 옵션을 설정한다.
      - `--permission-mode` 옵션에 `bypassPermissions` 값을 설정한다.
      - `--tools` 옵션에 `AskUserQuestion,Bash,TaskOutput,Edit,ExitPlanMode,Glob,Grep,KillShell,MCPSearch,Read,Skill,Task,TaskCreate,TaskGet,TaskList,TaskUpdate,WebFetch,WebSearch,Write,LSP` 값을 설정한다.
      - 시스템 프롬프트는 OS 임시 디렉토리에 임시 파일을 만들어서 저장하고 `--append-system-prompt-file` 옵션에 해당 파일 경로를 전달한다.
      - 사용자 프롬프트는 Claude Code CLI 외부 프로세스의 표준 입력으로 전달한다.
      - 모델의 출력 JSON Schema는 `--json-schema` 옵션에 전달한다.
      - `--output-format` 옵션에 `stream-json` 값을 설정한다.
      - `--verbose` 옵션을 설정한다.
      - `--include-partial-messages` 옵션을 설정한다.
      - 모델의 출력이 에러로 설정되는 경우와, 프로세스 실행 및 통신 과정에서 에러가 발생하는 경우를 모두 처리한다.
      - Claude Code CLI 외부 프로세스의 표준 출력과 표준 에러를 둘 다 실시간으로 읽어서 처리한다.
      - 모델의 출력이 JSON Schema에 맞는 유효한 JSON 객체로 완성될 때마다 콜백 함수를 호출해서 처리한다.
      - 스트리밍 결과에서 최종 result가 전달되면 스트리밍이 완료된 것으로 간주한다.
- Claude Code API Key는 프로그램 실행 시 `CLAUDE_CODE_API_KEY` 환경변수를 읽어서 가져온다. 만약 해당 환경변수가 설정되어 있지 않다면, 즉시 에러 메시지를 출력하고 프로그램이 에러 상태 코드로 종료되어야 한다.
- 사용자 요구사항을 모델에게 전달해서 명확화 질문이 있는지 확인한다. 모델의 출력이 명확화 질문이 포함된 JSON Schema 형태로 전달된다고 가정한다.
  - 시스템 프롬프트:
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
  - 사용자 프롬프트:
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
    
    위 사용자 프롬프트에서 {{ORIGINAL_USER_REQUEST_TEXT}}와 {{QA_LOG_TEXT}}는 실제로 Query 메소드가 호출될 때 동적으로 채워지는 부분이다. `ORIGINAL_USER_REQUEST_TEXT`에는 사용자가 입력한 최초 요구사항이 그대로 들어가야 하고, `QA_LOG_TEXT`에는 지금까지의 명확화 질문과 답변이 순서대로 들어가야 한다. 만약 아직 명확화 질문이 한 번도 나오지 않은 상태라면 `QA_LOG_TEXT`는 빈 문자열이 될 것이다.

  - 모델의 출력 JSON Schema:
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
    
    `questions`: 명확화 질문들의 배열. 각 질문은 문자열이어야 한다. 명확화 질문이 더 이상 필요하지 않다고 판단되는 경우에는 빈 배열이 될 수 있다. 
- 이번 작업에서는 최초 사용자 요청을 모델에게 보여주고, 명확화 질문을 받아서 사용자에게 출력하고 프로그램 종료하는 부분까지만 구현한다.

# References

- https://code.claude.com/docs/en/cli-reference
- https://code.claude.com/docs/en/headless