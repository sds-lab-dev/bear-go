# App Introduction

This project, named **“Bear AI Developer,”** is a TUI-based application that supports specification-driven software development on top of the Claude Code CLI. The application is written in Go language. You can build a consistent development environment using Dev Containers.

---

## 1. Requirements

- The Claude Code CLI must be installed, and its executable path must be available in `$PATH`.
- A valid Anthropic API key must be set in the `ANTHROPIC_API_KEY` environment variable.

---

## 2. Features

- Specification writing
- Development planning based on the specification
- Code writing and modification based on the specification and development plan
- Code review
- Documentation support

---

## 3. Operation Flow Overview

The Bear AI Developer application supports the following primary development flow:

1. Requirements gathering
2. Question and answer loop to refine requirements
3. Writing the specification document
4. Development planning based on the specification
5. Question and answer loop to refine the development plan
6. Code writing based on the specification and development plan
7. Code review
8. User feedback and approval loop
9. Documentation
