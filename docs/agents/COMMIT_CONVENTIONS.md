# Commit Conventions

---

## 1. Core Rules

- All commit messages MUST be in English.
- Based on the workspace changes, produce EXACTLY ONE commit message in English, consisting of a
  single subject and a body explaining "why", NOT "what".

---

## 2. Commit Message Format

A commit message MUST be formatted as follows:

```
<SUBJECT>

<BODY>
```

**IMPORTANT:** 

- You MUST NEVER include the literal characters "\n" in the subject and body of the commit message.
  The subject and body MUST be separated by an actual newline character, NOT the literal string "\n".
- `<SUBJECT>` MUST be less than or equal to 72 characters in length. Do NOT exceed 72 characters.
- `<BODY>` MUST be hard-wrapped at 72 characters per line. Do NOT produce a single long line for
  the body. 

### `<SUBJECT>`

- Short single subject line including its type and description.
- MUST be <= 72 characters; Do NOT exceed 72 characters.
- MUST follow the Conventional Commits format:

  ```
  <TYPE>: <DESCRIPTION>
  ```

  **`<TYPE>`:**

  `<TYPE>` MUST be one of the following, based on the nature of the change:
  
  - `feat`: a new feature
  - `fix`: a bug fix
  - `docs`: documentation-only changes
  - `style`: formatting or style-only changes with no behavioral change
  - `refactor`: code changes that neither fix a bug nor add a feature
  - `perf`: performance improvements
  - `test`: adding or updating tests
  - `build`: build system or dependency changes
  - `ci`: CI/CD configuration changes
  - `chore`: maintenance tasks that do not fit other types
  - `revert`: revert a previous commit

  Choosing a type for mixed diffs:
  
  - Select ONE type only, based on the dominant overall intent of the entire change set.
  - Do NOT emit multiple types.
  - When several types could apply, prefer the type that best describes the primary outcome for
    users or maintainers rather than secondary edits.

  **`<DESCRIPTION>`:**
  
  `<DESCRIPTION>` MUST describe the change in an imperative, present tense style. For example,
  "add feature" and NOT "added feature" or "adding feature". The description should be concise
  yet descriptive enough to understand the change at a glance.

  - Use a short, imperative description.
  - Start the description with a lowercase letter.
  - Do not end the description with a period.
  - The subject MUST summarize the whole change set, not one file or one hunk, unless the
    entire diff is truly about that one thing.      
  - Prefer a more specific type such as `build`, `ci`, or `style` instead of `chore` when
    possible.
  - Include a `BREAKING CHANGE:` in the front of the description if the change introduces a
    breaking change.
- Subject examples:
  - "feat: add refresh token rotation"
  - "fix: BREAKING CHANGE: change empty input handling in parser"
  - "docs: add local development instructions"
  - "refactor: simplify request validation"
  - "test: add eviction policy tests"
  - "build: update tauri to 2.1.0"
  - "ci: run clippy on pull requests"
  - "chore: update editorconfig"
  - "feat: BREAKING CHANGE: change authentication response format"

### `<BODY>`

- Hard-wrap the body at 72 characters per line (do NOT produce a single long line).
- The body should explain the "why" behind the change, providing context and reasoning for future
  reference. It should not simply restate "what" was changed, as that should be clear from the code
  diff and the subject line.
- Explain the breaking changes in detail in the front of the body if they exist.
- For large or mixed diffs, use the body to unify the changes under one rationale instead of
  listing each individual modification.
- Do NOT turn the body into a changelog or bullet list of file-by-file edits.

---

## 3. Output Constraints

- Output EXACTLY ONE commit message for the entire set of workspace changes.
- NEVER output multiple commit messages.
- NEVER output a list of per-file, per-directory, or per-topic commit messages.
- NEVER output alternatives, options, variants, bullets, numbering, headings, explanations,
  prose before the message, or prose after the message.
- The output MUST be a subject line, followed by a blank line, followed by a body.
- When the diff is large or spans multiple areas, you MUST still synthesize the changes into
  ONE cohesive commit message that best represents the overall intent of the change set.
- Do NOT decompose the diff into several Conventional Commit entries even if the workspace
  contains multiple logical changes.
- If the diff contains mixed changes, choose the most important user-facing or repository-wide
  theme as the subject, and summarize the shared motivation in the body.
- If there is no single clearly dominant file-level change, prefer the highest-level intent that
  unifies the changes, such as:
  - aligning project structure
  - reorganizing desktop app paths
  - improve release/build workflow consistency
  - update documentation and tooling for the new layout
