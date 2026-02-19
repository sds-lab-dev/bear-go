package app

import "os"

func getAPIKeyFromEnvVar(lookupEnvFunction func(string) (string, bool)) string {
	if lookupEnvFunction == nil {
		lookupEnvFunction = os.LookupEnv
	}

	value, ok := lookupEnvFunction("ANTHROPIC_API_KEY")
	if !ok || value == "" {
		return ""
	}

	return value
}
