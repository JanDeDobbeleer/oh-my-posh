package font

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

var (
	program     *tea.Program
	environment runtime.Environment
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1)
	textStyle         = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type loadMsg []*Asset

type zipMsg []byte

type successMsg []string

type errMsg error

type state int

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) { //nolint: gocritic
	i, ok := listItem.(*Asset)
	if !ok {
		return
	}

	str := fmt.Sprintf(i.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

const (
	getFonts state = iota
	selectFont
	downloadFont
	unzipFont
	installFont
	quit
	done
)

type main struct {
	err      error
	list     *list.Model
	font     string
	families []string
	spinner  spinner.Model
	state    state
	system   bool
}

func (m *main) buildFontList(nerdFonts []*Asset) {
	var items []list.Item
	for _, font := range nerdFonts {
		items = append(items, font)
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select font"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m.list = &l
}

func getFontsList() {
	fonts, err := Fonts()
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	program.Send(loadMsg(fonts))
}

func downloadFontZip(location string) {
	zipFile, err := Download(location)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	program.Send(zipMsg(zipFile))
}

func installLocalFontZIP(zipFile string, user bool) {
	data, err := os.ReadFile(zipFile)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	installFontZIP(data, user)
}

func installFontZIP(zipFile []byte, user bool) {
	families, err := InstallZIP(zipFile, user)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	program.Send(successMsg(families))
}

func (m *main) Init() tea.Cmd {
	isLocalZipFile := func() bool {
		return !strings.HasPrefix(m.font, "https") && strings.HasSuffix(m.font, ".zip")
	}

	if len(m.font) != 0 && !isLocalZipFile() {
		m.state = downloadFont
		if !strings.HasPrefix(m.font, "https") {
			m.font = fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", m.font)
		}
		defer func() {
			go downloadFontZip(m.font)
		}()
		m.spinner.Spinner = spinner.Globe
		return m.spinner.Tick
	}

	defer func() {
		if isLocalZipFile() {
			go installLocalFontZIP(m.font, m.system)
			return
		}
		go getFontsList()
	}()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	m.spinner = s
	m.state = getFonts
	if isLocalZipFile() {
		m.state = unzipFont
	}

	return m.spinner.Tick
}

func (m *main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadMsg:
		m.state = selectFont
		m.buildFontList(msg)
		return m, nil

	case tea.WindowSizeMsg:
		if m.list == nil {
			return m, nil
		}
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q", "esc":
			m.state = quit
			return m, tea.Quit

		case "enter":
			if len(m.font) != 0 || m.list == nil || m.list.SelectedItem() == nil {
				return m, nil
			}
			var font *Asset
			var ok bool
			if font, ok = m.list.SelectedItem().(*Asset); !ok {
				m.err = fmt.Errorf("no font selected")
				return m, tea.Quit
			}
			m.state = downloadFont
			m.font = font.Name
			defer func() {
				go downloadFontZip(font.URL)
			}()
			m.spinner.Spinner = spinner.Globe
			return m, m.spinner.Tick
		}

	case zipMsg:
		m.state = installFont
		defer func() {
			go installFontZIP(msg, m.system)
		}()
		m.spinner.Spinner = spinner.Dot
		return m, m.spinner.Tick

	case successMsg:
		m.state = done
		m.families = msg
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.list == nil {
		return m, nil
	}

	lst, cmd := m.list.Update(msg)
	m.list = &lst
	return m, cmd
}

func (m *main) View() string {
	if m.err != nil {
		return textStyle.Render(m.err.Error())
	}

	switch m.state {
	case getFonts:
		return textStyle.Render(fmt.Sprintf("%s Downloading font list%s", m.spinner.View(), terminal.StartProgress()))
	case selectFont:
		return fmt.Sprintf("\n%s%s", m.list.View(), terminal.StopProgress())
	case downloadFont:
		return textStyle.Render(fmt.Sprintf("%s Downloading %s%s", m.spinner.View(), m.font, terminal.StartProgress()))
	case unzipFont:
		return textStyle.Render(fmt.Sprintf("%s Extracting %s", m.spinner.View(), m.font))
	case installFont:
		return textStyle.Render(fmt.Sprintf("%s Installing %s", m.spinner.View(), m.font))
	case quit:
		return textStyle.Render(fmt.Sprintf("No need to install a new font? That's cool.%s", terminal.StopProgress()))
	case done:
		var builder strings.Builder

		builder.WriteString(fmt.Sprintf("Successfully installed %s 🚀\n\n%s", m.font, terminal.StopProgress()))
		builder.WriteString("The following font families are now available for configuration:\n")

		for i, family := range m.families {
			builder.WriteString(fmt.Sprintf("  • %s", family))

			if i < len(m.families)-1 {
				builder.WriteString("\n")
			}
		}

		return textStyle.Render(builder.String())
	}

	return ""
}

func Run(font string, env runtime.Environment) {
	main := &main{
		font:   font,
		system: env.Root(),
	}

	environment = env

	program = tea.NewProgram(main)
	if _, err := program.Run(); err != nil {
		print("Error running program: %v", err)
		os.Exit(70)
	}
}
