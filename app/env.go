package app

import "errors"

var ErrAPIKeyNotSet = errors.New("CLAUDE_CODE_API_KEY environment variable is not set or empty")

func ValidateAPIKeyEnv(lookupEnv func(string) (string, bool)) (string, error) {
	value, ok := lookupEnv("CLAUDE_CODE_API_KEY")
	if !ok || value == "" {
		return "", ErrAPIKeyNotSet
	}

	return value, nil
}
