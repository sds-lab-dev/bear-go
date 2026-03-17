# Bear AI Developer Application

## 1. App Introduction

Read the [App Introduction](docs/agents/APP_INTRODUCTION.md) document to understand the core
features and operation flow overview of the Bear AI Developer (a Go-based TUI application built
on the Claude Code CLI). 

You MUST refer to this document when you need to grasp the project's overall architecture, or
before starting a new task to align with the specification-driven workflow (planning,
implementation, review, documentation).

## 2. Operation Flow

Read the [Operation Flow](docs/agents/OPERATION_FLOW.md) document to understand the 7-step
development pipeline (from requirements to documentation) and the roles of specialized agents
including DAG-based parallel task execution. 

You MUST refer to this document when you need to determine the current project phase, move to the
next step, or understand how tasks and handoffs are structured among different agents.

---

# Mandatory Rules for AI Assistants

## 1. Language Rules

Read the [Language Rules](docs/agents/AI_LANGUAGE_RULES.md) document to understand the strict
bilingual output guidelines (Korean for user chat/reasoning, English for code, comments, docs, and
commits). 

You MUST refer to this document whenever you generate any form of text or code to ensure the correct
language is applied based on the target context (chat window vs. artifacts).

## 2. Coding Conventions

Read the [Coding Conventions](docs/agents/CODING_CONVENTIONS.md) document to understand the strict
guidelines prioritizing code readability, naming rules, structural refactoring, and commenting.

You MUST refer to this document whenever you write, modify, or review code to ensure it adheres to
the single-responsibility principle and prioritizes clear structure over clever one-liners or
performance optimizations.

## 3. Coding Guardrails

Read the [Coding Guardrails](docs/agents/CODING_GUARDRAILS.md) document to understand the strict
boundaries for defensive programming, simplicity, surgical (minimal) code changes, and goal-driven
test verification. 

You MUST refer to this document before making architectural decisions, implementing features, or
modifying existing code to avoid over-engineering, prevent scope creep, and ensure verifiable
correctness.

## 4. Commit Conventions

Read the [Commit Conventions](docs/agents/COMMIT_CONVENTIONS.md) document to understand the strict
formatting rules for git commits (Conventional Commits, 72-character limits, English-only,
explaining the "why" instead of "what"). 

You MUST refer to this document whenever you are about to generate a commit message or execute a
git commit to ensure a standardized, single cohesive message for all workspace changes.
