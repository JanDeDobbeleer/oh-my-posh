package upgrade

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
)

var (
	program   *tea.Program
	textStyle = lipgloss.NewStyle().Margin(1, 0, 2, 0)
	title     string
)

type resultMsg string

type state int

const (
	validating state = iota
	downloading
	verifying
	installing
)

func setState(message state) {
	if program == nil {
		return
	}

	program.Send(stateMsg(message))
}

type stateMsg state

type model struct {
	error   error
	message string
	tag     string
	spinner spinner.Model
	state   state
}

func initialModel(tag string) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	return &model{spinner: s, tag: tag}
}

func (m *model) Init() tea.Cmd {
	defer func() {
		go func() {
			if err := install(m.tag); err != nil {
				m.error = err
				program.Send(resultMsg(fmt.Sprintf("❌ upgrade failed: %v", err)))
				return
			}

			program.Send(resultMsg("🚀 Upgrade successful, restart your shell to take full advantage of the new functionality."))
		}()
	}()

	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil
		}

	case resultMsg:
		m.message = string(msg)
		return m, tea.Quit

	case stateMsg:
		m.state = state(msg)
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *model) View() string {
	if len(m.message) > 0 {
		return title + textStyle.Render(m.message)
	}

	var message string
	m.spinner.Spinner = spinner.Dot

	switch m.state {
	case validating:
		message = "Validating current installation"
	case downloading:
		m.spinner.Spinner = spinner.Globe
		message = "Downloading latest version"
	case verifying:
		m.spinner.Spinner = spinner.Moon
		message = "Verifying download"
	case installing:
		m.spinner.Spinner = spinner.Jump
		message = "Installing"
	}

	return title + textStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), message))
}

func Run(latest string) error {
	titleStyle := lipgloss.NewStyle().Margin(1, 0, 1, 0)
	title = "📦 Upgrading Oh My Posh"

	current := build.Version
	if len(current) == 0 {
		current = "dev"
	}

	title = fmt.Sprintf("%s from %s to %s", title, current, latest)
	title = titleStyle.Render(title)

	program = tea.NewProgram(initialModel(latest))
	resultModel, _ := program.Run()

	programModel, OK := resultModel.(*model)
	if !OK {
		return nil
	}

	return programModel.error
}

func IsMajorUpgrade(current, latest string) bool {
	if len(current) == 0 {
		return false
	}

	getMajorNumber := func(version string) string {
		return strings.Split(version, ".")[0]
	}

	return getMajorNumber(current) != getMajorNumber(latest)
}
