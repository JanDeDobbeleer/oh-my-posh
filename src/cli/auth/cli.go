package auth

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var (
	program   *tea.Program
	textStyle = lipgloss.NewStyle().Margin(1, 0, 2, 2)
)

type stateMsg state

type state int

const (
	code state = iota
	token
	done
)

func setState(message state) {
	if program == nil {
		return
	}

	program.Send(stateMsg(message))
}

type model struct {
	env     runtime.Environment
	err     error
	spinner *spinner.Model
	status  func(error) string
	code    string
	state   state
}

func (m *model) Init() tea.Cmd {
	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	m.spinner = &s

	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stateMsg:
		m.state = state(msg)
		if m.state == done {
			return m, tea.Quit
		}

		return m, nil

	default:
		s, cmd := m.spinner.Update(msg)
		m.spinner = &s
		return m, cmd
	}
}

func (m *model) View() string {
	var message string

	switch m.state {
	case code:
		message = fmt.Sprintf("%s Fetching code for authentication", m.spinner.View())
	case token:
		message = fmt.Sprintf("%s Fetching token with code: %s", m.spinner.View(), m.code)
	case done:
		message = m.status(m.err)
	}

	return textStyle.Render(message)
}

func Run(m tea.Model) error {
	program = tea.NewProgram(m)
	resultModel, _ := program.Run()

	programModel, OK := resultModel.(*model)
	if !OK {
		log.Debug("failed to cast model")
		return nil
	}

	return programModel.err
}
