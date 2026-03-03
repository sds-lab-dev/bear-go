package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/ai/claudecode"
	"github.com/sds-lab-dev/bear-go/app"
)

var (
	// buildVersion will be set at build time by Makefile. If the script fails
	// for any reason, it will default to "unknown".
	buildVersion = "unknown"
)

func main() {
	config := config{}
	if config.AnthropicAPIKey() == "" {
		fmt.Printf(
			"- WARNING:\n%v environment variable is not set or empty; trying to use a subscription plan, but this may fail if the key is required for authentication.\n\n",
			ANTHROPIC_API_KEY_ENV_VAR,
		)
	}

	app.Run(app.Config{
		BuildVersion: buildVersion,
		SessionID:    uuid.New().String(),
		AIPorts: aiSession{
			apiKey: config.AnthropicAPIKey(),
		},
		LogDir: config.LogDir(),
	})
}

type aiSession struct {
	apiKey string
}

func (r aiSession) NewSession(workingDir string) (ai.Session, error) {
	return claudecode.NewClient(r.apiKey, workingDir)
}
