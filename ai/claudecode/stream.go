package claudecode

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/log"
)

type StreamEventType string

const (
	StreamEventTypeAssistant StreamEventType = "assistant"
	StreamEventTypeUser      StreamEventType = "user"
	StreamEventTypeResult    StreamEventType = "result"
)

type StreamMessage struct {
	// Type can be "assistant", "user", or "result".
	Type StreamEventType `json:"type"`
	// Used when Type is "result".
	Subtype string `json:"subtype"`
	// Used when Type is "result".
	IsError bool `json:"is_error"`
	// Used when Type is "result".
	Result string `json:"result"`
	// Used when Type is "user" or "assistant".
	Message Message `json:"message"`
	// Used when Type is "result".
	StructuredOutput json.RawMessage `json:"structured_output"`
	// Used when Claude Code CLI does not log in.
	Error string `json:"error"`
	// Internal field to hold the original JSON line for debugging purposes.
	RawJSON string `json:"-"`
}

func (msg StreamMessage) toAiStreamMessage() ai.StreamMessage {
	if msg.Type != StreamEventTypeAssistant && msg.Type != StreamEventTypeUser {
		panic(fmt.Sprintf("unexpected StreamMessage type for toAiStreamMessage: %v", msg.Type))
	}
	return msg.Message.toAiStreamMessage()
}

type Message struct {
	// Role can be "user" or "assistant".
	Role string `json:"role"`
	// Used when Role is "assistant".
	Model string `json:"model"`
	// Used when Role is "user" or "assistant".
	Content []ContentBlock `json:"content"`
}

func (msg Message) toAiStreamMessage() ai.StreamMessage {
	if msg.Role != string(StreamEventTypeUser) && msg.Role != string(StreamEventTypeAssistant) {
		log.Warning(fmt.Sprintf("unexpected stream message role: %v", msg.Role))
		// Fallback
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error(fmt.Sprintf("failed to marshal stream message: %v", err))
			return ai.StreamMessage{}
		}
		return ai.StreamMessage{
			Type:    ai.StreamMessageTypeText,
			Content: string(jsonBytes),
		}
	}

	if len(msg.Content) == 0 {
		log.Warning("stream message has no content blocks")
		// Fallback
		return ai.StreamMessage{
			Type:    ai.StreamMessageTypeText,
			Content: "empty stream message",
		}
	}

	// For simplicity, we only convert the first content block to an ai.StreamMessage.
	return ai.StreamMessage{
		Type:    msg.Content[0].StreamMessageType(),
		Content: msg.Content[0].StreamMessageContent(),
	}
}

type ContentBlock struct {
	// Type can be "text", "tool_use", "tool_result", or "thinking".
	Type string `json:"type"`
	// Used when Type is "text".
	Text string `json:"text"`
	// Used when Type is "tool_use".
	Name string `json:"name"`
	// Used when Type is "tool_use".
	Input json.RawMessage `json:"input"`
	// Used when Type is "tool_result". This field can be both string and array of JSON
	// objects, so we use json.RawMessage.
	Content json.RawMessage `json:"content"`
	// Used when Type is "thinking".
	Thinking string `json:"thinking"`
}

func (content ContentBlock) StreamMessageType() ai.StreamMessageType {
	switch content.Type {
	case "text":
		return ai.StreamMessageTypeText
	case "tool_use":
		return ai.StreamMessageTypeToolCall
	case "tool_result":
		return ai.StreamMessageTypeToolCallResult
	case "thinking":
		return ai.StreamMessageTypeThinking
	default:
		// Fallback
		return ai.StreamMessageTypeText
	}
}

func (content ContentBlock) StreamMessageContent() string {
	switch content.Type {
	case "text":
		if content.Text == "" {
			log.Warning("text content block has empty text")
			return "empty text content"
		}
		return content.Text
	case "tool_use":
		if content.Name == "" {
			log.Warning("tool_use content block has empty name")
			return "tool_use with empty name"
		}
		if len(content.Input) == 0 {
			log.Warning("tool_use content block has empty input")
			return "tool_use with empty input"
		}
		return fmt.Sprintf("%v: %v", content.Name, string(content.Input))
	case "tool_result":
		if len(content.Content) == 0 {
			log.Warning("tool_result content block has empty content")
			return "empty tool_result content"
		}
		return string(content.Content)
	case "thinking":
		if content.Thinking == "" {
			log.Warning("thinking content block has empty thinking text")
			return "empty thinking content"
		}
		return content.Thinking
	default:
		// Fallback
		jsonBytes, err := json.Marshal(content)
		if err != nil {
			log.Error(fmt.Sprintf("failed to marshal content block: %v", err))
			return "failed to marshal content block"
		}
		return string(jsonBytes)
	}
}

var (
	ErrStreamParseFailed = errors.New("failed to parse stream JSON line")
	ErrResultError       = errors.New("result returned an error")
	ErrNoResultReceived  = errors.New("stream ended without a result message")
)

func processStream(reader io.Reader, streamCallback func(ai.StreamMessage)) (json.RawMessage, error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		log.Debug(fmt.Sprintf("received raw stream line: %v", line))

		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrStreamParseFailed, line)
		}
		msg.RawJSON = line
		log.Debug(fmt.Sprintf("parsed stream message: %#v", msg))

		if len(msg.Error) > 0 {
			if msg.Error == "authentication_failed" {
				return nil, errors.New("claude code CLI error: authentication failed: run `/login` first")
			}
			return nil, fmt.Errorf("claude code CLI error: %v", msg.Error)
		}

		switch msg.Type {
		case StreamEventTypeAssistant, StreamEventTypeUser:
			if streamCallback == nil {
				continue
			}
			streamCallback(msg.toAiStreamMessage())
		case StreamEventTypeResult:
			// IsError can be true even if Subtype is "success", so we check the Subtype and
			// StructuredOutput to determine if this is a successful result.
			if msg.Subtype == "success" && len(msg.StructuredOutput) > 0 {
				return msg.StructuredOutput, nil
			}
			return nil, fmt.Errorf("%w: %v", ErrResultError, msg.Result)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("stream read error: %w", err)
	}

	return nil, ErrNoResultReceived
}
