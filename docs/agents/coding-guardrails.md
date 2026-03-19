# Coding Guardrails

---

## 1. Think Before Coding

**Do NOT assume. Do NOT hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly.
- If you are uncertain, ask targeted questions.
- If multiple interpretations exist, present them. Do not pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, distinguish:
  - Blocking ambiguity: stop and ask.
  - Non-blocking ambiguity: choose a reasonable default, state it, and proceed.

---

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- Do NOT add extra features beyond what was asked.
- Avoid abstractions whose main purpose is hypothetical future flexibility.
- Prefer small helpers when they reduce duplication, clarify intent, or improve testability.
- Keep code proportional. If you wrote 200 lines and it could be 50, rewrite it.

Error handling:

- Do NOT add elaborate handling for truly unreachable states proven by invariants.
- Do add minimal, appropriate handling for:
  - External inputs (user input, files, network, databases)
  - Boundary cases (empty, null, out-of-range)
  - Concurrency and timeouts
- If you choose to omit handling, explain the invariant that makes it safe.

Ask yourself:
"Would a senior engineer say this is overcomplicated?"
If yes, simplify.

---

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Do NOT refactor unrelated code, comments, naming, or formatting.
- Match existing style and conventions.
- Exception:
  - If the code you touch (or its immediate dependencies) has a correctness bug, security weakness, or reliability hazard that affects the behavior of your change, fix it.
  - If your change reveals or triggers a pre-existing build/test failure, type error, lint failure, or static-analysis issue in the area you touch, apply the minimal fix needed to restore a green build.
  - If there is an obvious resource-safety issue in the path you modify (for example, leaks, missing cleanup, missing bounds checks, unsafe concurrency), make the smallest fix that prevents harm.
  - If a small local adjustment is required to keep the change consistent with surrounding invariants (for example, error-handling contract, nullability/ownership rules, API preconditions), make that adjustment.
- Guardrails for exceptions:
  - Keep the fix local to the files/functions you touched (or the smallest necessary surface area).
  - Do NOT do opportunistic cleanup.
  - Do NOT reformat or "improve" nearby code unless required for the fix.
  - Explain what was wrong, why it matters, and why the chosen fix is minimal.

When your changes create orphans:

- Remove imports/variables/functions that your changes made unused.
- Do NOT remove pre-existing dead code unless asked.
- Instead, mention it and point to locations if you notice unrelated dead code or suspicious logic.

The test:
Every changed line should trace directly to the user's request or to making the requested change correct and verifiable.

---

## 4. Goal-driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals. For example:

- "Add validation" -> "Write tests for valid and invalid inputs, then make them pass."
- "Fix the bug" -> "Write tests that reproduce the bug and verify no regression, then make them pass."
- "Refactor X" -> "Ensure tests pass before and after, and behavior is unchanged without regressions."
- "Implement feature Y" -> "Write tests that demonstrate the feature with several scenarios, then make them pass."

Verification rules:

- Prefer automated verification (tests, type checks, linting, builds).
- Prefer Testcontainers to verify external dependencies if available.
- If you cannot run verification in your environment, say so explicitly and provide:
  - Commands to run
  - Expected outputs or assertions
  - Any risks or edge cases to double-check
- If you find issues during verification, fix them and verify again.
- The verification loop continues until all criteria are met or the maximum number of iterations (default: 5) is reached.
- If the criteria does not meet after the maximum iterations:
  - Explains the blockers to the user and request a guidance on how to proceed.
