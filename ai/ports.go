package ai

type Ports interface {
	// NewSession creates a new AI session for the given working directory.
	//
	// AI session represents a single interaction session with the AI agent, and it can maintain
	// state across multiple interactions. You can create multiple isolated sessions if needed,
	// for example, to handle different tasks that require separate contexts.
	//
	// workingDir can be used by the AI session to perform file operations, such as
	// reading and writing files, executing shell commands, etc. If workingDir is empty, the AI
	// session can use the current working directory as the default.
	NewSession(workingDir string) (Session, error)
}

type Session interface {
	SpecWriter
}

// SpecWriter is the interface that defines the methods for generating a specification based
// on a user request.
type SpecWriter interface {
	StreamCallbackHandler

	// GetInitialClarifyingQuestions takes the initial user request to generate clarifying
	// questions.
	//
	// This function returns a list of clarifying questions, and the list can be empty if no
	// clarifying questions are needed.
	GetInitialClarifyingQuestions(initialUserRequest string) ([]string, error)

	// GetNextClarifyingQuestions takes the user's answer to the previous clarifying questions
	// to generate the next set of clarifying questions. This function can be called multiple
	// times in a loop until no more clarifying questions are needed, at which point the caller
	// will proceed to draft the spec based on the user's answers to all the clarifying questions.
	//
	// The userAnswer parameter is the user's answer to the previous set of clarifying
	// questions.
	//
	// This function returns a list of clarifying questions, and the list can be empty if no
	// clarifying questions are needed.
	GetNextClarifyingQuestions(userAnswer string) ([]string, error)

	// DraftSpec generates a draft specification based on the initial user request and the
	// clarifying Q&As between the user and the AI agent.
	DraftSpec() (string, error)

	// ReviseSpec takes the user's feedback on the previous drafted spec and generates a
	// revised spec.
	ReviseSpec(userFeedback string) (string, error)
}

// A stream callback handler function is used to send intermediate messages back to the caller,
// and it can be called multiple times before the final result is returned.
//
// The messages sent through the callback can be used to provide real-time feedback, such as
// what the agent is currently thinking, what tools it is calling, and what results it is getting
// from those tools.
//
// The stream callback handler is nil by default if no callback is explicitly set, in which
// case you can ignore the streaming messages.
type StreamCallbackHandler interface {
	SetStreamCallbackHandler(func(StreamMessage))
}

type StreamMessage struct {
	Role    StreamMessageRole
	Type    StreamMessageType
	Content string
}

type StreamMessageRole int

const (
	StreamMessageRoleAssistant StreamMessageRole = iota
	StreamMessageRoleUser
)

type StreamMessageType int

const (
	StreamMessageTypeThinking StreamMessageType = iota
	StreamMessageTypeToolCall
	StreamMessageTypeToolCallResult
	StreamMessageTypeText
)
