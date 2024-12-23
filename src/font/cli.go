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
	cache_ "github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

var (
	program *tea.Program
	cache   cache_.Cache
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

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
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
	err  error
	list *list.Model
	Asset
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

func installLocalFontZIP(m *main) {
	data, err := os.ReadFile(m.URL)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	installFontZIP(data, m)
}

func installFontZIP(zipFile []byte, m *main) {
	families, err := InstallZIP(zipFile, m)
	if err != nil {
		program.Send(errMsg(err))
		return
	}

	program.Send(successMsg(families))
}

func (m *main) Init() tea.Cmd {
	isLocalZipFile := func() bool {
		return !strings.HasPrefix(m.URL, "https") && strings.HasSuffix(m.URL, ".zip")
	}

	resolveFontZipURL := func() error {
		if strings.HasPrefix(m.URL, "https") {
			return nil
		}

		fonts, err := Fonts()
		if err != nil {
			return err
		}

		var fontAsset *Asset
		for _, font := range fonts {
			if !strings.EqualFold(m.URL, font.Name) {
				continue
			}

			fontAsset = font
			break
		}

		if fontAsset == nil {
			return fmt.Errorf("no matching font found")
		}

		m.Asset = *fontAsset

		return nil
	}

	if len(m.URL) != 0 && !isLocalZipFile() {
		m.state = downloadFont

		if err := resolveFontZipURL(); err != nil {
			m.err = err
			return tea.Quit
		}

		defer func() {
			go downloadFontZip(m.URL)
		}()

		m.spinner.Spinner = spinner.Globe
		return m.spinner.Tick
	}

	defer func() {
		if isLocalZipFile() {
			go installLocalFontZIP(m)
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
		return textStyle.Render(fmt.Sprintf("%s Downloading %s%s", m.spinner.View(), m.Name, terminal.StartProgress()))
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

		var builder strings.Builder

		builder.WriteString(fmt.Sprintf("Successfully installed %s 🚀\n\n%s", m.Name, terminal.StopProgress()))
		builder.WriteString("The following font families are now available for configuration:\n\n")

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

func SetCache(c cache_.Cache) {
	cache = c
}

func Run(font string, ch cache_.Cache, root bool, zipFolder string) {
	main := &main{
		system: root,
		Asset: Asset{
			Name:   font,
			URL:    font,
			Folder: zipFolder,
		},
	}

	SetCache(ch)

	program = tea.NewProgram(main)
	if _, err := program.Run(); err != nil {
		print("Error running program: %v", err)
		os.Exit(70)
	}
}
