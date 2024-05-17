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
)

var program *tea.Program

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
	spinner  spinner.Model
	list     *list.Model
	system   bool
	font     string
	state    state
	err      error
	families []string
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
		return textStyle.Render(fmt.Sprintf("%s Downloading font list", m.spinner.View()))
	case selectFont:
		return "\n" + m.list.View()
	case downloadFont:
		return textStyle.Render(fmt.Sprintf("%s Downloading %s", m.spinner.View(), m.font))
	case unzipFont:
		return textStyle.Render(fmt.Sprintf("%s Extracting %s", m.spinner.View(), m.font))
	case installFont:
		return textStyle.Render(fmt.Sprintf("%s Installing %s", m.spinner.View(), m.font))
	case quit:
		return textStyle.Render("No need to install a new font? That's cool.")
	case done:
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Successfully installed %s 🚀\n\n", m.font))
		builder.WriteString("The following font families are now available for configuration:\n")
		for _, family := range m.families {
			builder.WriteString(fmt.Sprintf("  • %s\n", family))
		}
		return textStyle.Render(builder.String())
	}
	return ""
}

func Run(font string, system bool) {
	main := &main{
		font:   font,
		system: system,
	}
	program = tea.NewProgram(main)
	if _, err := program.Run(); err != nil {
		print("Error running program: %v", err)
		os.Exit(70)
	}
}
