package ui

import (
	"fmt"
	"strings"

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
	specWriter   ai.SpecWriter
	eventCh      chan tea.Msg
	state        specPromptModelState
	errorMessage string
}

func NewSpecPromptModel(userRequest string, specWriter ai.SpecWriter) SpecPromptModel {
	terminalSize := GetTerminalSize()

	ta := textarea.New()
	ta.Placeholder = "Answer the clarifying questions to help the agent understand your request better"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(terminalSize.Width)
	ta.SetHeight(terminalSize.Height / 2)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Focus()

	model := SpecPromptModel{
		textarea:     ta,
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
	return func() tea.Msg {
		log.Debug("waiting for event in Init()")
		v := <-m.eventCh
		log.Debug(fmt.Sprintf("received event in Init(): %#v", v))
		return v
	}
}

func (m SpecPromptModel) waitForNext() tea.Cmd {
	return func() tea.Msg {
		log.Debug("waiting for event in waitForNext()")
		v := <-m.eventCh
		log.Debug(fmt.Sprintf("received event in waitForNext(): %#v", v))
		return v
	}
}

func (m SpecPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("received update message in SpecPromptModel: %#v", msg))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Debug(fmt.Sprintf("received window size message: width=%d, height=%d", msg.Width, msg.Height))
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height / 2)
		return m, nil
	case streamEventMsg:
		// TODO: 이벤트 메시지 유형에 따라 적절히 처리
		log.Debug(fmt.Sprintf("received stream event message: type=%v, content=%v", msg.Type, msg.Content))
		cmd := tea.Sequence(
			tea.Printf("%v\n", msg.Content),
			m.waitForNext(),
		)
		return m, cmd
	case clarifyingQuestionsMsg:
		log.Debug(fmt.Sprintf("received clarifying questions message: %v", msg.questions))
		m.state = specStateWaitUserAnswers
		cmd := tea.Printf("Clarifying questions:\n%v\n", strings.Join(msg.questions, "\n"))
		return m, cmd
	case userAnswersMsg:
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
	case userFeedbackMsg:
		log.Debug(fmt.Sprintf("received user feedback message: %v", msg.feedback))
		cmd := tea.Sequence(
			tea.Printf("Your feedback:\n%v\n", msg.feedback),
			func() tea.Msg {
				go m.reviseSpec(msg.feedback)
				return <-m.eventCh
			},
		)
		return m, cmd
	case clarifyingQuestionsDoneMsg:
		log.Debug("received clarifying questions done message")
		m.state = specStateSpecDrafting
		cmd := tea.Sequence(
			tea.Println(SuccessStyle.Render("No more clarifying questions.")),
			func() tea.Msg {
				go m.draftSpec()
				return <-m.eventCh
			},
		)
		return m, cmd
	case specDraftMsg:
		log.Debug(fmt.Sprintf("received spec draft message: %v", msg.draft))
		m.state = specStateWaitUserFeedback
		return m, tea.Println(SuccessStyle.Render("Draft spec:\n" + msg.draft))
	case specApprovedMsg:
		log.Debug(fmt.Sprintf("received spec approved message: %v", msg.spec))
		m.state = specStateSpecApproved
		return m, tea.Println(SuccessStyle.Render("Approved spec:\n" + msg.spec))
	case streamErrorMsg:
		log.Debug(fmt.Sprintf("received stream error message: %v", msg.err))
		return m, func() tea.Msg {
			return SpecPromptResult{
				Err:          msg.err,
				ApprovedSpec: "",
			}
		}
	case tea.KeyMsg:
		log.Debug(fmt.Sprintf("received key message: type=%v", msg.String()))

		switch msg.String() {
		// Go to next step on Enter
		case "enter":
			return m.handleEnter()
		case "shift+enter", "alt+enter":
			m.textarea.InsertString("\n")
			return m, nil
		}
		// TODO: external editor (Ctrl+G)
	}

	m.errorMessage = ""
	var cmd tea.Cmd
	// For all other messages, we pass them to the textarea component to handle them in the text
	// area.
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
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
		b.WriteString(renderAgentActivePrompt("Analyzing your request to find out if there are any clarifying questions..."))
		b.WriteByte('\n')
		return b.String()
	case specStateWaitUserAnswers:
		b.WriteString(renderAgentActivePrompt("Please answer the clarifying questions above. Press Enter when you're done."))
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString(m.textarea.View())
		if m.errorMessage != "" {
			b.WriteByte('\n')
			b.WriteString(ErrorStyle.Render(m.errorMessage))
		}
		return b.String()
	case specStateSpecDrafting:
		b.WriteString(renderAgentActivePrompt("Drafting the spec based on your request and answers..."))
		b.WriteByte('\n')
		return b.String()
	case specStateWaitUserFeedback:
		b.WriteString(renderAgentActivePrompt("Please review the drafted spec above and provide your feedback. Press Enter when you're done."))
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString(m.textarea.View())
		if m.errorMessage != "" {
			b.WriteByte('\n')
			b.WriteString(ErrorStyle.Render(m.errorMessage))
		}
		return b.String()
	case specStateSpecApproved:
		return ""
	default:
		panic(fmt.Sprintf("Unexpected state in View(): %v", m.state))
	}
}
