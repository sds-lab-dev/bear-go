package main

import "os"

const (
	ANTHROPIC_API_KEY_ENV_VAR = "BEAR_ANTHROPIC_API_KEY"
	LOG_DIR_ENV_VAR           = "BEAR_LOG_DIR"
)

type config struct{}

func loadEnvironmentVariable(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if ok && value != "" {
		return value
	}
	return defaultValue
}

func (c config) AnthropicAPIKey() string {
	return loadEnvironmentVariable(ANTHROPIC_API_KEY_ENV_VAR, "")
}

func (c config) LogDir() string {
	return loadEnvironmentVariable(LOG_DIR_ENV_VAR, "/tmp/bear_logs")
}
