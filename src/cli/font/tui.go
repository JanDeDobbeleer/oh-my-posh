package font

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	progress_ "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/progress"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

var (
	program *tea.Program
)

const listHeight = 14

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(3)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(3)
	helpStyle         = lipgloss.NewStyle().PaddingLeft(3).PaddingBottom(1)
	textStyle         = lipgloss.NewStyle().Margin(1, 0, 2, 2)
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

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("•" + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(i.Name))
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
	spinner  *spinner.Model
	progress *progress.Model
	Asset
	families []string
	state    state
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
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m.list = &l
}

func getFontsList() {
	fonts, err := fonts()
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

func installLocalFontZIP(m *main) {
	data, err := os.ReadFile(m.URL)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	installFontZIP(data, m)
}

func installFontZIP(zipFile []byte, m *main) {
	families, err := InstallZIP(zipFile, m.Folder)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	program.Send(successMsg(families))
}

func (m *main) Init() tea.Cmd {
	m.progress = progress.NewModel()

	s := spinner.New()
	m.spinner = &s

	if len(m.URL) != 0 && !IsLocalZipFile(m.URL) {
		m.state = downloadFont

		asset, err := ResolveFontAsset(m.URL)
		if err != nil {
			m.err = err
			return tea.Quit
		}

		m.Asset = *asset

		defer func() {
			go downloadFontZip(asset.URL)
		}()

		m.spinner.Spinner = spinner.Globe
		return m.spinner.Tick
	}

	defer func() {
		if IsLocalZipFile(m.URL) {
			go installLocalFontZIP(m)
			return
		}

		go getFontsList()
	}()

	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	m.state = getFonts

	if IsLocalZipFile(m.URL) {
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
			if len(m.URL) != 0 || m.list == nil || m.list.SelectedItem() == nil {
				return m, nil
			}

			var font *Asset
			var ok bool

			if font, ok = m.list.SelectedItem().(*Asset); !ok {
				m.err = fmt.Errorf("no font selected")
				return m, tea.Quit
			}

			m.state = downloadFont
			m.Asset = *font

			defer func() {
				go downloadFontZip(font.URL)
			}()

			m.spinner.Spinner = spinner.Globe
			return m, m.spinner.Tick

		case "up", "k":
			if m.list != nil {
				if m.list.Index() == 0 {
					m.list.Select(len(m.list.Items()) - 1)
				} else {
					m.list.Select(m.list.Index() - 1)
				}
			}
			return m, nil

		case "down", "j":
			if m.list != nil {
				if m.list.Index() == len(m.list.Items())-1 {
					m.list.Select(0)
				} else {
					m.list.Select(m.list.Index() + 1)
				}
			}
			return m, nil
		}

	case progress.Message:
		return m, m.progress.SetPercent(float64(msg))

	case progress_.FrameMsg:
		return m, m.progress.Update(msg)

	case zipMsg:
		m.state = installFont
		defer func() {
			go installFontZIP(msg, m)
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
		s, cmd := m.spinner.Update(msg)
		m.spinner = &s
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
		return textStyle.Render(fmt.Sprintf("Downloading %s...\n%s", m.Name, m.progress.View()))
	case unzipFont:
		return textStyle.Render(fmt.Sprintf("%s Extracting %s", m.spinner.View(), m.Name))
	case installFont:
		return textStyle.Render(fmt.Sprintf("%s Installing %s", m.spinner.View(), m.Name))
	case quit:
		return textStyle.Render(fmt.Sprintf("No need to install a new font? That's cool.%s", terminal.StopProgress()))
	case done:
		if len(m.families) == 0 {
			return textStyle.Render(fmt.Sprintf("No matching font families were installed. Try setting --zip-folder to the correct folder when using CascadiaCode (MS) or a custom font zip file. %s", terminal.StopProgress())) //nolint: lll
		}

		sb := text.NewBuilder()

		sb.WriteString(fmt.Sprintf("Successfully installed %s 🚀\n\n%s", m.Name, terminal.StopProgress()))
		sb.WriteString("The following font families are now available for configuration:\n\n")

		for i, family := range m.families {
			sb.WriteString(fmt.Sprintf("  • %s", family))

			if i < len(m.families)-1 {
				sb.WriteString("\n")
			}
		}

		return textStyle.Render(sb.String())
	}

	return ""
}

func Run(font, zipFolder string) (string, error) {
	main := &main{
		Asset: Asset{
			Name:   font,
			URL:    font,
			Folder: zipFolder,
		},
	}

	program = tea.NewProgram(main)
	_, err := program.Run()
	return main.Name, err
}
