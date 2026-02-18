package app

import (
	"errors"
	"testing"
)

func TestValidateAPIKeyEnv_ValidKey(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "CLAUDE_CODE_API_KEY" {
			return "sk-test-key-123", true
		}
		return "", false
	}

	apiKey, err := ValidateAPIKeyEnv(lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "sk-test-key-123" {
		t.Errorf("expected 'sk-test-key-123', got %q", apiKey)
	}
}

func TestValidateAPIKeyEnv_NotSet(t *testing.T) {
	lookup := func(key string) (string, bool) {
		return "", false
	}

	_, err := ValidateAPIKeyEnv(lookup)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrAPIKeyNotSet) {
		t.Errorf("expected ErrAPIKeyNotSet, got %v", err)
	}
}

func TestValidateAPIKeyEnv_EmptyString(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "CLAUDE_CODE_API_KEY" {
			return "", true
		}
		return "", false
	}

	_, err := ValidateAPIKeyEnv(lookup)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrAPIKeyNotSet) {
		t.Errorf("expected ErrAPIKeyNotSet, got %v", err)
	}
}
