package ai

import (
	_ "embed"
	"strings"
)

//go:embed prompts/clarification_system.md
var RawClarificationSystemPrompt string

//go:embed prompts/clarification_user_initial_request.md
var RawClarificationUserPromptForInitialRequest string

func ClarificationSystemPrompt() string {
	return RawClarificationSystemPrompt
}

func ClarificationUserPromptForInitialRequest(initialUserRequest string) string {
	return strings.ReplaceAll(RawClarificationUserPromptForInitialRequest, "{{INITIAL_USER_REQUEST_TEXT}}", initialUserRequest)
}
