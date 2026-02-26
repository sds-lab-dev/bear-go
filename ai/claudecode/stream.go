package claudecode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/log"
)

type streamEventType string

const (
	streamEventTypeAssistant streamEventType = "assistant"
	streamEventTypeUser      streamEventType = "user"
	streamEventTypeResult    streamEventType = "result"
)

type StreamMessage struct {
	// Type can be "assistant", "user", or "result".
	Type streamEventType `json:"type"`
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
	if msg.Type != streamEventTypeAssistant && msg.Type != streamEventTypeUser {
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
	if msg.Role != string(streamEventTypeUser) && msg.Role != string(streamEventTypeAssistant) {
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

	// For simplicity, we only convert the first content block to an
	// ai.StreamMessage.
	return ai.StreamMessage{
		Type:    msg.Content[0].StreamMessageType(),
		Content: normalizeNewlines(msg.Content[0].StreamMessageContent()),
	}
}

func normalizeNewlines(s string) string {
    // 1) Windows CRLF -> LF
    s = strings.ReplaceAll(s, "\r\n", "\n")
    // 2) CR -> LF
    s = strings.ReplaceAll(s, "\r", "\n")
    return s
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
	// Used when Type is "tool_result". This field can be both string and array
	// of JSON objects, so we use json.RawMessage.
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

func convertJSONToYAML(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // preserves large integers/precise numbers when possible

	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}

	out, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// PrettyPrintQuotedEscaped takes a string that includes the surrounding quotes
// (") and contains escape sequences like \n, \t, \".
//
// It returns a decoded string where \n becomes an actual newline, \t becomes a
// tab, and indentation is preserved.
//
// If any error occurs during processing, it returns the original input unchanged.
func PrettyPrintQuotedEscapedBytes(b []byte) (string, error) {
	original := string(b)

	s := strings.TrimSpace(original)

	decoded, err := strconv.Unquote(s)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	_, err = sb.WriteString(decoded)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
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
		name := content.Name
		if name == "" {
			log.Warning("tool_use content block has empty name")
			name = "empty tool_use name"
		}

		input, err := convertJSONToYAML(content.Input)
		if err != nil {
			log.Warning(fmt.Sprintf("failed to convert tool_use input from JSON to YAML: %v", err))
			input = string(content.Input)
		}
		if input == "" {
			log.Warning("tool_use content block has empty input")
			input = "empty tool_use input"
		}

		return fmt.Sprintf("%v: %v", name, input)
	case "tool_result":
		// We first try to convert the content from JSON to YAML for better
		// readability, and if that fails, we pretty print the original content
		// as a quoted and escaped string. If that also fails, use the original
		// content as a fallback string.
		result, err := convertJSONToYAML(content.Content)
		if err != nil {
			log.Debug("failed to process tool_result content: failed to convert to YAML from JSON")
			log.Debug("trying to pretty print tool_result content as quoted and escaped string...")
			result, err = PrettyPrintQuotedEscapedBytes(content.Content)
			if err != nil {
				log.Warning(fmt.Sprintf("failed to process tool_result content: %v", err))
				result = string(content.Content)
			}
		}
		if len(result) == 0 {
			log.Warning("tool_result content block has empty content")
			return "empty tool_result content"
		}
		return result
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

func processStream(
	reader io.Reader,
	streamCallback func(ai.StreamMessage),
) (json.RawMessage, error) {
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
				return nil, errors.New("claude code CLI error: authentication failed: run `/login` first in the CLI or set ANTHROPIC_API_KEY environment variable")
			}
			return nil, fmt.Errorf("claude code CLI error: %v", msg.Error)
		}

		switch msg.Type {
		case streamEventTypeAssistant, streamEventTypeUser:
			if streamCallback == nil {
				continue
			}
			streamCallback(msg.toAiStreamMessage())
		case streamEventTypeResult:
			// IsError can be true even if Subtype is "success", so we check the
			// Subtype and StructuredOutput to determine if this is a successful
			// result.
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
