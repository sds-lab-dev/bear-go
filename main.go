package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/ai/claudecode"
	"github.com/sds-lab-dev/bear-go/app"
	"github.com/sds-lab-dev/bear-go/ui"
)

const API_KEY_ENV_VAR = "ANTHROPIC_API_KEY"

var (
	// buildVersion will be set at build time by Makefile. If the script fails
	// for any reason, it will default to "unknown".
	buildVersion = "unknown"
)

func main() {
	session := aiSession{
		apiKey: getAPIKeyFromEnvVar(),
	}
	app.Run(buildVersion, uuid.New().String(), session)
}

type aiSession struct {
	apiKey string
}

func (r aiSession) NewSession(workingDir string) (ai.Session, error) {
	return claudecode.NewClient(r.apiKey, workingDir)
}

func getAPIKeyFromEnvVar() string {
	value, ok := os.LookupEnv(API_KEY_ENV_VAR)
	if ok && value != "" {
		return value
	}

	fmt.Println(
		ui.ErrorStyle.Render(
			fmt.Sprintf(
				"%v environment variable is not set or empty; trying to use a subscription plan, but this may fail if the key is required for authentication", API_KEY_ENV_VAR,
			),
		),
	)

	return ""
}
