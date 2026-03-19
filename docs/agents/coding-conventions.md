# Coding Conventions

---

## 1. Mandatory Conventions

You MUST follow the coding conventions in order:
1. Follow the given instructions described below.
2. Follow the project's existing coding conventions unless explicitly stated in the given
   instructions.
3. Follow the idiomatic style of the programming language used in the project if no existing
   conventions are present and no explicit instructions are given:
   - Do NOT add unnecessary comments if the code is self-explanatory on its own.
   ```rust
   // Incorrect example
   let count = items.len(); // Counts the number of items
   ```
   - You MUST add comments if it is not immediately obvious what the code does by reading it.
   ```rust
   // Correct example
   // Append the LLM response to the next request.
   self.request
       .append_message(response.choices[0].message.clone())
       .map_err(|err| EmailAnalyzerError::InvalidRequest(err.to_string()))?;
   ```

---

## 2. Code Readability

- Code readability is the highest priority.
  - Performance or idiomatic language expression is a consideration only after readability is
    secured.
  - If code written in multiple lines is easier to understand than a one-liner that uses syntactic
    sugar, you must choose the former.
- Do NOT use abbreviations just to keep names short.
  - Use long names as-is to improve readability, even if they are lengthy.
  - Exceptions are allowed for universally understood abbreviations such as `db`, or widely used
    loop indices such as `i` or `j`.
- You MUST refactor the structure immediately if blocks, conditionals, loops, and similar constructs
  cause indentation nesting of three levels or more.
  - Note that "three levels of indentation" is not a magic number.
  - Depending on the case, even a single level of indentation can harm readability, so handle it
    appropriately on a case-by-case basis.
- Every function MUST follow the single-responsibility principle and do only one thing.
  - An exception is when the function's purpose is orchestration by composing multiple components.
- Long functions or methods are NOT allowed:
  - If it exceeds 50 lines, you MUST consider refactoring.
  - If it exceeds 100 lines, you MUST refactor it unless you can justify that it has a single,
    cohesive responsibility.
- Do NOT prefer creating new types first; check whether the existing types in the codebase can be
  used before introducing new ones.
- Do NOT avoid modifying the existing application structure.
  - If the application already has "two eyes," do NOT take the easy route by attaching a "third eye"
    on the side.
  - Instead, modify the structure of the existing eyes so it can be done with only the two eyes.
