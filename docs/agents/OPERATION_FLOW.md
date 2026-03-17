# Operation Flow

## 1. Requirements Gathering

- Users can enter requirements through a terminal user interface.
- The entered requirements are analyzed via the Claude Code CLI, and additional questions are asked to enable a clear specification.
- Users provide answers to the additional questions to refine and concretize the requirements.
- The additional-question loop continues until the AI model determines that no further questions are necessary.

## 2. Writing the Specification Document

- Once requirements gathering is complete, the **Specification Agent** produces a draft specification document and presents it to the user.
- The user can provide feedback on the draft specification and request revisions as needed.
- The feedback and revision loop continues until the user is satisfied with the specification.
- The final approved specification moves to the development planning stage.

## 3. Development Planning

- Based on the approved specification, the **Planning Agent** produces a draft development plan document and presents it to the user.
- The user can provide feedback on the development plan and request revisions as needed.
- The feedback and revision loop continues until the user is satisfied with the development plan.
- The individual tasks specified in the development plan are split so that AI agents can process them in parallel.
- If there are dependencies among tasks, the development plan must represent a DAG (Directed Acyclic Graph) as an adjacency list to specify the execution order.
- The final approved development plan moves to the code writing stage.

## 4. Code Writing

- Following the approved specification and development plan, _n_ **Coding Agents** run in multiple threads and write code in parallel.
- A dedicated agent is assigned to each individual task in the development plan. Each agent generates code independently for its assigned task.
- If there are inter-task dependencies, agents follow the DAG specified in the development plan and execute tasks in dependency order. For tasks with dependencies, the preceding task's session content is converted into a handoff document and passed to the subsequent task agents.
- Each agent uses the Claude Code CLI to write code.

## 5. Code Review

- The written code is examined by the **Review Agent**. The Review Agent runs in parallel on the same threads in which the Coding Agents executed.
- The Review Agent checks code quality, style, and whether functional requirements are satisfied, and if necessary sends revision requests to the Coding Agents.
- Coding Agents apply the review feedback and modify the code, after which the Review Agent reviews the updated code again.
- This loop continues until the code satisfies all review criteria or the maximum iterations (default: 5) are reached.
- If the criteria does not meet after the maximum iterations:
  - Approves the code as-is if it is reasonably close to the criteria and the remaining issues are minor.
  - Otherwise, explains the blockers to the user and request a guidance on how to proceed.

## 6. User Feedback and Approval Loop

- After all tasks in the development plan are completed and code review passes, the process requires the user's final approval.
- The user reviews the final code and can either approve it or request additional changes.
- If the user requests additional changes, the request is handed back to the Coding Agents for revision work. The revised code is reviewed again by the Review Agent, and this loop continues until the user approves.
- Once the user approves, the development process is complete.

## 7. Documentation

- Throughout the steps above, each agent records its work as documentation.
- These documents are provided together with the project's final deliverables and can be used later for maintenance and reference.