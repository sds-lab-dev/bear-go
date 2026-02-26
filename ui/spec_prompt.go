package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sds-lab-dev/bear-go/ai"
	"github.com/sds-lab-dev/bear-go/log"
)

type SpecPromptResult struct {
	Err          error
	ApprovedSpec string
}

type streamEventMsg struct {
	ai.StreamMessage
}

type clarifyingQuestionsMsg struct {
	questions []string
}

type userAnswersMsg struct {
	answers string
}

type clarifyingQuestionsDoneMsg struct{}

type specDraftMsg struct {
	draft string
}

type userFeedbackMsg struct {
	feedback string
}

type specApprovedMsg struct {
	spec string
}

type streamErrorMsg struct {
	err error
}

type specPromptModelState int

const (
	specStatePrepareClarifyingQuestions specPromptModelState = iota
	specStateWaitUserAnswers
	specStateSpecDrafting
	specStateWaitUserFeedback
	specStateSpecApproved
)

type SpecPromptModel struct {
	textarea     textarea.Model
	spinner      spinner.Model
	specWriter   ai.SpecWriter
	eventCh      chan tea.Msg
	state        specPromptModelState
	errorMessage string
}

func NewSpecPromptModel(
	userRequest string,
	specWriter ai.SpecWriter,
) SpecPromptModel {
	terminalSize := GetTerminalSize()

	ta := textarea.New()
	ta.Placeholder = "Answer the clarifying questions to help the agent understand your request better"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(terminalSize.Width)
	ta.SetHeight(terminalSize.Height / 2)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot

	model := SpecPromptModel{
		textarea:     ta,
		spinner:      s,
		specWriter:   specWriter,
		eventCh:      make(chan tea.Msg, 64),
		state:        specStatePrepareClarifyingQuestions,
		errorMessage: "",
	}
	model.specWriter.SetStreamCallbackHandler(model.defaultStreamCallback)
	go model.getClarifyingQuestions(userRequest)

	return model
}

func (m SpecPromptModel) defaultStreamCallback(msg ai.StreamMessage) {
	log.Debug(fmt.Sprintf("sending a stream message to the event channel: %#v", msg))
	m.eventCh <- streamEventMsg{StreamMessage: msg}
	log.Debug("stream message sent to event channel")
}

func (m SpecPromptModel) getClarifyingQuestions(input string) {
	log.Debug(fmt.Sprintf("getting clarifying questions for input: %s", input))

	questions, err := m.specWriter.GetInitialClarifyingQuestions(input)
	if err != nil {
		log.Debug(fmt.Sprintf("sending streamErrorMsg to the event channel: %#v", err))
		m.eventCh <- streamErrorMsg{err: err}
		return
	}
	log.Debug(fmt.Sprintf("received clarifying questions: %v", questions))

	if len(questions) > 0 {
		m.eventCh <- clarifyingQuestionsMsg{questions: questions}
	} else {
		m.eventCh <- clarifyingQuestionsDoneMsg{}
	}
}

func (m SpecPromptModel) draftSpec() {
	log.Debug("drafting spec")

	spec, err := m.specWriter.DraftSpec()
	if err != nil {
		m.eventCh <- streamErrorMsg{err: err}
		return
	}
	log.Debug(fmt.Sprintf("received drafted spec: %v", spec))

	log.Debug("sending drafted spec to event channel")
	m.eventCh <- specDraftMsg{draft: spec}
	log.Debug("drafted spec sent to event channel")
}

func (m SpecPromptModel) reviseSpec(feedback string) {
	log.Debug(fmt.Sprintf("revising spec with user feedback: %s", feedback))

	spec, err := m.specWriter.ReviseSpec(feedback)
	if err != nil {
		m.eventCh <- streamErrorMsg{err: err}
		return
	}
	log.Debug(fmt.Sprintf("received revised spec: %v", spec))

	log.Debug("sending revised spec to event channel")
	m.eventCh <- specDraftMsg{draft: spec}
	log.Debug("revised spec sent to event channel")
}

func (m SpecPromptModel) Init() tea.Cmd {
	cmd := tea.Sequence(
		m.spinner.Tick,
		func() tea.Msg {
			log.Debug("waiting for event in Init()")
			v := <-m.eventCh
			log.Debug(fmt.Sprintf("received event in Init(): %#v", v))
			return v
		},
	)
	return cmd
}

func (m SpecPromptModel) waitForNext() tea.Cmd {
	return func() tea.Msg {
		log.Debug("waiting for event in waitForNext()")
		v := <-m.eventCh
		log.Debug(fmt.Sprintf("received event in waitForNext(): %#v", v))
		return v
	}
}

func (m SpecPromptModel) handleWindowSizeMsg(
	msg tea.WindowSizeMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf(
		"received window size message: width=%d, height=%d", msg.Width, msg.Height))
	m.textarea.SetWidth(msg.Width)
	m.textarea.SetHeight(msg.Height / 2)
	return m, nil
}

func (m SpecPromptModel) handleStreamEventMsg(
	msg streamEventMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf(
		"received stream event message: type=%v, content=%v", msg.Type, msg.Content))

	var content string
	switch msg.Type {
	case ai.StreamMessageTypeThinking:
		content = renderStreamMessageThinking(msg.Content)
	case ai.StreamMessageTypeToolCall:
		content = renderStreamMessageToolCall(msg.Content)
	case ai.StreamMessageTypeToolCallResult:
		content = renderStreamMessageToolCallResult(msg.Content)
	case ai.StreamMessageTypeText:
		// Fallthrough to default case.
	default:
		content = renderStreamMessageText(msg.Content)
	}
	cmd := tea.Sequence(
		tea.Printf("%v\n", content),
		m.waitForNext(),
	)
	return m, cmd
}

func (m SpecPromptModel) handleClarifyingQuestionsMsg(
	msg clarifyingQuestionsMsg,
) (tea.Model, tea.Cmd) {
	if len(msg.questions) == 0 {
		panic("clarifying questions array should have at least one element")
	}
	log.Debug(fmt.Sprintf(
		"received clarifying questions message: %v", msg.questions))
	m.state = specStateWaitUserAnswers
	var b strings.Builder
	for i, s := range msg.questions {
		fmt.Fprintf(&b, "%v. %v\n", i+1, s)
		if i+1 < len(msg.questions) {
			fmt.Fprint(&b, "\n")
		}
	}
	questions := fmt.Sprintf("%v\n%v",
		renderAgentInactivePrompt(
			successStyle.Render("Clarifying questions:"), true,
		), b.String(),
	)
	return m, tea.Println(questions)
}

func (m SpecPromptModel) handleUserAnswersMsg(
	msg userAnswersMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received user answers message: %v", msg.answers))
	// Go to next state to prepare clarifying questions based on user's answers.
	m.state = specStatePrepareClarifyingQuestions
	cmd := tea.Sequence(
		tea.Printf("Your answers:\n%v\n", msg.answers),
		func() tea.Msg {
			go m.getClarifyingQuestions(msg.answers)
			return <-m.eventCh
		},
	)
	return m, cmd
}

func (m SpecPromptModel) handleUserFeedbackMsg(
	msg userFeedbackMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received user feedback message: %v", msg.feedback))
	cmd := tea.Sequence(
		tea.Printf("Your feedback:\n%v\n", msg.feedback),
		func() tea.Msg {
			go m.reviseSpec(msg.feedback)
			return <-m.eventCh
		},
	)
	return m, cmd
}

func (m SpecPromptModel) handleClarifyingQuestionsDoneMsg() (tea.Model, tea.Cmd) {
	log.Debug("received clarifying questions done message")
	m.state = specStateSpecDrafting
	cmd := tea.Sequence(
		tea.Println(successStyle.Render("No more clarifying questions.")),
		func() tea.Msg {
			go m.draftSpec()
			return <-m.eventCh
		},
	)
	return m, cmd
}

func (m SpecPromptModel) handleSpecDraftMsg(
	msg specDraftMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received spec draft message: %v", msg.draft))
	m.state = specStateWaitUserFeedback
	return m, tea.Println(successStyle.Render("Draft spec:\n" + msg.draft))
}

func (m SpecPromptModel) handleSpecApprovedMsg(
	msg specApprovedMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received spec approved message: %v", msg.spec))
	m.state = specStateSpecApproved
	return m, tea.Println(successStyle.Render("Approved spec:\n" + msg.spec))
}

func (m SpecPromptModel) handleStreamErrorMsg(
	msg streamErrorMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received stream error message: %v", msg.err))
	return m, func() tea.Msg {
		return SpecPromptResult{
			Err:          msg.err,
			ApprovedSpec: "",
		}
	}
}

// TODO: external editor (Ctrl+G)
func (m SpecPromptModel) handleKeyMsg(
	msg tea.KeyMsg,
) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received key message: type=%v", msg.String()))

	switch msg.String() {
	// Go to next step on Enter
	case "enter":
		return m.handleEnter()
	case "shift+enter", "alt+enter":
		m.textarea.InsertString("\n")
		return m, nil
	default:
		return m, nil
	}
}

func (m SpecPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received update message in SpecPromptModel: %#v", msg))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case streamEventMsg:
		return m.handleStreamEventMsg(msg)
	case clarifyingQuestionsMsg:
		return m.handleClarifyingQuestionsMsg(msg)
	case userAnswersMsg:
		return m.handleUserAnswersMsg(msg)
	case userFeedbackMsg:
		return m.handleUserFeedbackMsg(msg)
	case clarifyingQuestionsDoneMsg:
		return m.handleClarifyingQuestionsDoneMsg()
	case specDraftMsg:
		return m.handleSpecDraftMsg(msg)
	case specApprovedMsg:
		return m.handleSpecApprovedMsg(msg)
	case streamErrorMsg:
		return m.handleStreamErrorMsg(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// For all other messages, we pass them to the child components.
	m.errorMessage = ""
	var cmd tea.Cmd
	var sequence []tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	sequence = append(sequence, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
	sequence = append(sequence, cmd)

	return m, tea.Sequence(sequence...)
}

func (m SpecPromptModel) handleEnter() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textarea.Value())
	if value == "" {
		m.errorMessage = "Please enter your feedback."
		return m, nil
	}

	switch m.state {
	case specStateWaitUserAnswers:
		return m, func() tea.Msg {
			return userAnswersMsg{answers: value}
		}
	case specStateWaitUserFeedback:
		return m, func() tea.Msg {
			return userFeedbackMsg{feedback: value}
		}
	default:
		panic(fmt.Sprintf("Enter key pressed in unexpected state: %v", m.state))
	}
}

func (m SpecPromptModel) View() string {
	log.Debug(fmt.Sprintf("rendering spec prompt view: state=%v, errorMessage=%v", m.state, m.errorMessage))

	var b strings.Builder

	switch m.state {
	case specStatePrepareClarifyingQuestions:
		b.WriteString(
			renderAgentActivePrompt(
				fmt.Sprintf("%vAnalyzing your request to find out if there are any clarifying questions...", m.spinner.View()),
				false,
			),
		)
		b.WriteByte('\n')
		return b.String()
	case specStateWaitUserAnswers:
		b.WriteString(
			renderAgentActivePrompt(
				"Please answer the clarifying questions above. Press Enter when you're done.",
				true,
			),
		)
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString(m.textarea.View())
		if m.errorMessage != "" {
			b.WriteByte('\n')
			b.WriteString(errorStyle.Render(m.errorMessage))
		}
		return b.String()
	case specStateSpecDrafting:
		b.WriteString(
			renderAgentActivePrompt(
				fmt.Sprintf("%vDrafting the spec based on your request and answers...", m.spinner.View()),
				false,
			),
		)
		b.WriteByte('\n')
		return b.String()
	case specStateWaitUserFeedback:
		b.WriteString(
			renderAgentActivePrompt(
				"Please review the drafted spec above and provide your feedback. Press Enter when you're done.",
				true,
			),
		)
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString(m.textarea.View())
		if m.errorMessage != "" {
			b.WriteByte('\n')
			b.WriteString(errorStyle.Render(m.errorMessage))
		}
		return b.String()
	case specStateSpecApproved:
		return ""
	default:
		panic(fmt.Sprintf("Unexpected state in View(): %v", m.state))
	}
}
