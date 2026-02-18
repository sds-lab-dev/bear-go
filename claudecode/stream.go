package claudecode

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type StreamEventType string

const (
	StreamEventTypeAssistant   StreamEventType = "assistant"
	StreamEventTypeUser        StreamEventType = "user"
	StreamEventTypeResult      StreamEventType = "result"
	StreamEventTypeSystem      StreamEventType = "system"
	StreamEventTypeStreamEvent StreamEventType = "stream_event"
)

type StreamMessage struct {
	Type             StreamEventType `json:"type"`
	Subtype          string          `json:"subtype,omitempty"`
	IsError          bool            `json:"is_error,omitempty"`
	Content          []ContentBlock  `json:"content,omitempty"`
	StructuredOutput json.RawMessage `json:"structured_output,omitempty"`
	RawJSON          string          `json:"-"`
}

type ContentBlock struct {
	Type    string          `json:"type,omitempty"`
	Text    string          `json:"text,omitempty"`
	Name    string          `json:"name,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Content string          `json:"content,omitempty"`
}

type ResultData struct {
	Output json.RawMessage
}

type StreamCallback func(StreamMessage)

var (
	ErrStreamParseFailed = errors.New("failed to parse stream JSON line")
	ErrResultError       = errors.New("result returned an error")
	ErrNoResultReceived  = errors.New("stream ended without a result message")
)

var errorSubtypes = map[string]bool{
	"error_max_turns":                    true,
	"error_during_execution":             true,
	"error_max_budget_usd":               true,
	"error_max_structured_output_retries": true,
}

func processStream(reader io.Reader, callback StreamCallback) (ResultData, error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return ResultData{}, fmt.Errorf("%w: %s", ErrStreamParseFailed, line)
		}
		msg.RawJSON = line

		switch msg.Type {
		case StreamEventTypeAssistant, StreamEventTypeUser:
			callback(msg)
		case StreamEventTypeResult:
			if msg.IsError || errorSubtypes[msg.Subtype] {
				return ResultData{}, fmt.Errorf("%w: subtype=%s", ErrResultError, msg.Subtype)
			}
			if msg.Subtype == "success" {
				return ResultData{Output: msg.StructuredOutput}, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ResultData{}, fmt.Errorf("stream read error: %w", err)
	}

	return ResultData{}, ErrNoResultReceived
}
