package progress

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

type Message float64

func NewModel() *Model {
	p := progress.New(progress.WithScaledGradient("#800080", "#ffc0cb"))
	return &Model{Model: p}
}

type Model struct {
	progress.Model
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	model, cmd := m.Model.Update(msg)
	m.Model = model.(progress.Model)

	return cmd
}

func (m *Model) View() string {
	return m.Model.View() + terminal.SetProgress(int(m.Percent()*100))
}
