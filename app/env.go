package app

import "errors"

var ErrAPIKeyNotSet = errors.New("ANTHROPIC_API_KEY environment variable is not set or empty")

func ValidateAPIKeyEnv(lookupEnv func(string) (string, bool)) (string, error) {
	value, ok := lookupEnv("ANTHROPIC_API_KEY")
	if !ok || value == "" {
		return "", ErrAPIKeyNotSet
	}

	return value, nil
}
