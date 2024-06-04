//go:build windows || darwin

package upgrade

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

var (
	program   *tea.Program
	textStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type resultMsg string

type state int

const (
	validating state = iota
	downloading
	installing
)

type stateMsg state

type model struct {
	spinner spinner.Model
	message string
	state   state
}

func initialModel() *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	return &model{spinner: s}
}

func (m *model) Init() tea.Cmd {
	defer func() {
		go func() {
			if err := install(); err != nil {
				message := fmt.Sprintf("⚠️  %s", err)
				program.Send(resultMsg(message))
				return
			}

			program.Send(resultMsg(successMsg))
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
		return textStyle.Render(m.message)
	}

	var message string
	m.spinner.Spinner = spinner.Dot
	switch m.state {
	case validating:
		message = "Validating current installation"
	case downloading:
		m.spinner.Spinner = spinner.Globe
		message = "Downloading latest version"
	case installing:
		message = "Installing"
	}

	return textStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), message))
}

func Run() {
	program = tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func downloadAsset(asset string) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/%s", asset)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := platform.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to download installer: %s", url)
	}

	return resp.Body, nil
}
