package upgrade

import (
	"fmt"
	"strings"

	progress_ "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/progress"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

var (
	program   *tea.Program
	textStyle = lipgloss.NewStyle().Margin(1, 0, 2, 2)
)

type resultMsg string

type stateMsg state

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

type model struct {
	error    error
	config   *Config
	spinner  *spinner.Model
	progress *progress.Model
	message  string
	state    state
}

func initialModel(cfg *Config) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	p := progress.NewModel()

	return &model{spinner: &s, config: cfg, progress: p}
}

func (m *model) Init() tea.Cmd {
	go m.start()

	return m.spinner.Tick
}

func (m *model) start() {
	if err := install(m.config); err != nil {
		m.error = err
		log.Debug("failed to install")
		program.Send(resultMsg(fmt.Sprintf(" ❌ upgrade failed: %v", err)))
		return
	}

	current := fmt.Sprintf("v%s", build.Version)
	message := fmt.Sprintf("🚀 Upgraded from %s to %s", current, m.config.Latest)

	if current != m.config.Latest {
		log.Debug("new version installed, user needs to restart shell")
		message += ", restart your shell to take full advantage of the new functionality"
	}

	program.Send(resultMsg(message))
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

	case progress.Message:
		return m, m.progress.SetPercent(float64(msg))

	case progress_.FrameMsg:
		return m, m.progress.Update(msg)

	default:
		s, cmd := m.spinner.Update(msg)
		m.spinner = &s
		return m, cmd
	}
}

func (m *model) View() string {
	if len(m.message) > 0 {
		return textStyle.Render(m.message)
	}

	var message string
	m.spinner.Spinner = spinner.Dot

	switch m.state {
	case validating:
		message = "Validating current installation"
	case downloading:
		message = fmt.Sprintf("Downloading %s from %s...\n%s", m.config.Latest, m.config.Source.String(), m.progress.View())
		return textStyle.Render(message)
	case verifying:
		m.spinner.Spinner = spinner.Moon
		message = "Verifying download"
	case installing:
		m.spinner.Spinner = spinner.Jump
		message = "Installing"
	}

	return textStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), message))
}

func Run(cfg *Config) error {
	program = tea.NewProgram(initialModel(cfg))
	resultModel, _ := program.Run()

	programModel, OK := resultModel.(*model)
	if !OK {
		log.Debug("failed to cast model")
		return nil
	}

	return programModel.error
}

func IsMajorUpgrade(current, latest string) bool {
	if current == "" {
		return false
	}

	getMajorNumber := func(version string) string {
		return strings.Split(version, ".")[0]
	}

	return getMajorNumber(current) != getMajorNumber(latest)
}
