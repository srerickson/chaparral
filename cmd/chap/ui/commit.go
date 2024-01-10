package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmState int

const (
	// state (input focus)
	commitNameInput    confirmState = iota // user name
	commitEmailInput                       // user email address
	commitMessageInput                     // commit message
	commitFinalInput                       // yes/no to commit
)

type CommitMeta struct {
	state confirmState

	Name    textinput.Model
	Email   textinput.Model
	Message textarea.Model
	YesNo   textinput.Model
}

func NewCommit(name, email, message string) CommitMeta {
	var m CommitMeta

	m.Name = textinput.New()
	m.Name.CharLimit = 32
	m.Name.Cursor.Style = cursorStyle
	m.Name.Prompt = "Name"
	m.Name.Placeholder = "required"
	m.Name.PlaceholderStyle = faintText
	m.Name.SetValue(name)

	m.Email = textinput.New()
	m.Email.Prompt = "Email"
	m.Email.Placeholder = "required"
	m.Email.PlaceholderStyle = faintText
	m.Email.CharLimit = 32
	m.Email.Cursor.Style = cursorStyle
	m.Email.SetValue(email)

	m.Message = textarea.New()
	m.Message.ShowLineNumbers = false
	m.Message.Cursor.Style = cursorStyle
	m.Message.BlurredStyle = textarea.Style{
		Prompt:      faintText,
		Placeholder: faintText,
	}
	m.Message.FocusedStyle = textarea.Style{
		Prompt:      faintText,
		Placeholder: faintText,
	}
	m.Message.Placeholder = "..."
	m.Message.SetHeight(3)
	m.Message.Prompt = ""
	m.Message.SetValue(message)

	m.YesNo = textinput.New()
	m.YesNo.Prompt = "Confirm"
	m.YesNo.Placeholder = "yes/no"
	m.YesNo.PlaceholderStyle = faintText
	m.YesNo.CharLimit = 3
	m.YesNo.Cursor.Style = cursorStyle

	return m
}

func (m CommitMeta) Update(msg tea.Msg) (CommitMeta, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case confirmFocusMsg:
		m.blurInputs()
		if m.Name.Value() == "" {
			m.state = commitNameInput
		} else if m.Email.Value() == "" {
			m.state = commitEmailInput
		} else if m.Message.Value() == "" {
			m.state = commitMessageInput
		} else {
			m.state = commitFinalInput
		}
		m.setInputFocus()
		return m, tea.Batch(textarea.Blink, textinput.Blink)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, m.Cancel()
		case "tab", "down":
			m.blurInputs()
			m.cycleFocus(true)
			m.setInputFocus()
		case "up":
			m.blurInputs()
			m.cycleFocus(false)
			m.setInputFocus()
		case "enter":
			if m.state != commitFinalInput {
				m.blurInputs()
				m.cycleFocus(true)
				m.setInputFocus()
				break
			}
			switch m.YesNo.Value() {
			case "yes", "y", "Y":
				m.blurInputs()
				if m.Name.Value() == "" {
					m.state = commitNameInput
					m.setInputFocus()
					break
				}
				if m.Email.Value() == "" {
					m.state = commitEmailInput
					m.setInputFocus()
					break
				}
				if m.Message.Value() == "" {
					m.state = commitMessageInput
					m.setInputFocus()
					break
				}
				return m, m.Confirm()
			case "no", "n", "N":
				return m, m.Cancel()
			default:
				return m, nil
			}
		}
	}
	cmd = m.updateInputs(msg)
	return m, cmd
}

func (m CommitMeta) View() string {
	// Name: (required)
	// Email: (required)
	// Message:
	// ...
	// Confirm:(yes/no)
	b := strings.Builder{}
	b.WriteString("Finalize commit:\n")
	b.WriteString(m.Name.View() + "\n")
	b.WriteString(m.Email.View() + "\n")
	style := promptBlurStyle
	if m.state == commitMessageInput {
		style = promptFocusStyle
	}
	b.WriteString(style.Render("Message"))
	b.WriteString(m.Message.View() + "\n")
	b.WriteString(m.YesNo.View())
	b.WriteString(faintText.Render("\n(tab/enter to change field, 'yes' to confirm, 'ctrl+c' to cancel)"))
	return b.String()
}

type confirmFocusMsg struct{}

func CommitFocus() tea.Cmd {
	return func() tea.Msg {
		return confirmFocusMsg{}
	}
}

func (m *CommitMeta) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd
	m.Name, cmd = m.Name.Update(msg)
	cmds = append(cmds, cmd)
	m.Email, cmd = m.Email.Update(msg)
	cmds = append(cmds, cmd)
	m.Message, cmd = m.Message.Update(msg)
	cmds = append(cmds, cmd)
	m.YesNo, cmd = m.YesNo.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m *CommitMeta) blurInputs() {
	m.Name.Blur()
	m.Name.PromptStyle = promptBlurStyle
	m.Email.Blur()
	m.Email.PromptStyle = promptBlurStyle
	m.Message.Blur()
	m.YesNo.Blur()
	m.YesNo.PromptStyle = promptBlurStyle
}
func (m *CommitMeta) cycleFocus(forward bool) {
	switch forward {
	case true:
		m.state = (m.state + 1) % (commitFinalInput + 1)
	case false:
		m.state = m.state - 1
		if m.state < 0 {
			m.state = commitFinalInput
		}
	}
}

func (m *CommitMeta) setInputFocus() {
	switch m.state {
	case commitNameInput:
		m.Name.Focus()
		m.Name.PromptStyle = promptFocusStyle
	case commitEmailInput:
		m.Email.Focus()
		m.Email.PromptStyle = promptFocusStyle
	case commitMessageInput:
		m.Message.Focus()
	case commitFinalInput:
		m.YesNo.Focus()
		m.YesNo.PromptStyle = promptFocusStyle

	}
}

type CommitResult struct {
	Confirm bool
	Name    string
	Email   string
	Message string
}

func (m *CommitMeta) Cancel() tea.Cmd {
	return func() tea.Msg {
		return CommitResult{}
	}
}

func (m *CommitMeta) Confirm() tea.Cmd {
	return func() tea.Msg {
		return CommitResult{
			Confirm: true,
			Name:    m.Name.Value(),
			Email:   m.Email.Value(),
			Message: m.Message.Value(),
		}
	}
}
