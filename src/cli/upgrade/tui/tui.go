// Package tui implements the Bubble Tea spinner and progress bar oh-my-posh
// shows while it upgrades itself. It lives one level below cli/upgrade
// because cli/upgrade holds the plain Config, CDN, and Source types that the
// config and segments packages need, and that import graph also has to
// compile for wasm — a target with no terminal to draw to, and one that
// bubbletea itself does not build for at all. Keeping this file here means
// importing cli/upgrade for its types never drags bubbletea along; this
// package imports cli/upgrade, never the other way around.
package tui

import (
	"fmt"

	progress_ "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/progress"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

var (
	program   *tea.Program
	textStyle = lipgloss.NewStyle().Margin(1, 0, 2, 2)
)

type resultMsg string

type stateMsg upgrade.Stage

type model struct {
	error    error
	config   *upgrade.Config
	spinner  *spinner.Model
	progress *progress.Model
	message  string
	state    upgrade.Stage
}

func initialModel(cfg *upgrade.Config) *model {
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
	if err := upgrade.Install(m.config); err != nil {
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
		m.state = upgrade.Stage(msg)
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
	case upgrade.StageValidating:
		message = "Validating current installation"
	case upgrade.StageDownloading:
		message = fmt.Sprintf("Downloading %s from %s...\n%s", m.config.Latest, m.config.Source.String(), m.progress.View())
		return textStyle.Render(message)
	case upgrade.StageVerifying:
		m.spinner.Spinner = spinner.Moon
		message = "Verifying download"
	case upgrade.StageInstalling:
		m.spinner.Spinner = spinner.Jump
		message = "Installing"
	}

	return textStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), message))
}

func Run(cfg *upgrade.Config) error {
	// Relay cli/upgrade's plain Stage/percent callbacks into this program's
	// message loop. cli/upgrade cannot send tea.Msg values itself without
	// importing bubbletea, which it must not do, so this program subscribes
	// on its behalf for the duration of the run.
	upgrade.SetStageReporter(func(stage upgrade.Stage) {
		if program == nil {
			return
		}

		program.Send(stateMsg(stage))
	})

	upgrade.SetProgressReporter(func(percent float64) {
		if program == nil {
			return
		}

		program.Send(progress.Message(percent))
	})

	defer upgrade.SetStageReporter(nil)
	defer upgrade.SetProgressReporter(nil)

	program = tea.NewProgram(initialModel(cfg))
	resultModel, _ := program.Run()

	programModel, OK := resultModel.(*model)
	if !OK {
		log.Debug("failed to cast model")
		return nil
	}

	return programModel.error
}
