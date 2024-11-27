package prompt

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
)

func TestCanWriteRPrompt(t *testing.T) {
	cases := []struct {
		TerminalWidthError error
		Case               string
		TerminalWidth      int
		PromptLength       int
		RPromptLength      int
		Expected           bool
	}{
		{Case: "Width Error", Expected: false, TerminalWidthError: errors.New("burp")},
		{Case: "Terminal > Prompt enabled", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 10},
		{Case: "Terminal > Prompt enabled edge", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 70},
		{Case: "Prompt > Terminal enabled", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 70},
		{Case: "Terminal > Prompt disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 100, RPromptLength: 71},
		{Case: "Prompt > Terminal disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 300, RPromptLength: 80},
		{Case: "Prompt > Terminal disabled no room", Expected: true, TerminalWidth: 200, PromptLength: 400, RPromptLength: 80},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalWidthError)
		engine := &Engine{
			Env:               env,
			rpromptLength:     tc.RPromptLength,
			currentLineLength: tc.PromptLength,
			rprompt:           "hello",
		}

		_, got := engine.canWriteRightBlock(tc.RPromptLength, true)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestPrintPWD(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Config   string
		Pwd      string
		Shell    string
		Cygwin   bool
	}{
		{Case: "Empty PWD"},
		{Case: "OSC99", Config: terminal.OSC99, Expected: "\x1b]9;9;pwd\x1b\\"},
		{Case: "OSC99 - Elvish", Config: terminal.OSC99, Shell: shell.ELVISH},
		{Case: "OSC7", Config: terminal.OSC7, Expected: "\x1b]7;file://host/pwd\x1b\\"},
		{Case: "OSC51", Config: terminal.OSC51, Expected: "\x1b]51;Auser@host:pwd\x1b\\"},
		{Case: "Template (empty)", Config: "{{ if eq .Shell \"pwsh\" }}osc7{{ end }}"},
		{Case: "Template (non empty)", Shell: shell.GENERIC, Config: "{{ if eq .Shell \"shell\" }}osc7{{ end }}", Expected: "\x1b]7;file://host/pwd\x1b\\"},
		{
			Case:     "OSC99 Cygwin",
			Pwd:      `C:\Users\user\Documents\GitHub\oh-my-posh`,
			Config:   terminal.OSC99,
			Cygwin:   true,
			Expected: "\x1b]9;9;C:/Users/user/Documents/GitHub/oh-my-posh\x1b\\",
		},
		{
			Case:     "OSC99 Windows",
			Pwd:      `C:\Users\user\Documents\GitHub\oh-my-posh`,
			Config:   terminal.OSC99,
			Expected: "\x1b]9;9;C:\\Users\\user\\Documents\\GitHub\\oh-my-posh\x1b\\",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		if len(tc.Pwd) == 0 {
			tc.Pwd = "pwd"
		}

		env.On("Pwd").Return(tc.Pwd)
		env.On("User").Return("user")
		env.On("Shell").Return(tc.Shell)
		env.On("IsCygwin").Return(tc.Cygwin)
		env.On("Host").Return("host", nil)

		template.Cache = &cache.Template{
			Shell:    tc.Shell,
			Segments: maps.NewConcurrent(),
		}
		template.Init(env, nil)

		terminal.Init(shell.GENERIC)

		engine := &Engine{
			Env: env,
			Config: &config.Config{
				PWD: tc.Config,
			},
		}

		engine.pwd()
		got := engine.string()

		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func BenchmarkEngineRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engineRender()
	}
}

func engineRender() {
	cfg := config.Load("", shell.GENERIC, false)

	env := &runtime.Terminal{}
	env.Init(nil)

	defer env.Close()

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent(),
	}
	template.Init(env, nil)

	terminal.Init(shell.GENERIC)
	terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	terminal.Colors = cfg.MakeColors(env)

	engine := &Engine{
		Config: cfg,
		Env:    env,
	}

	engine.Primary()
}

func TestGetTitle(t *testing.T) {
	cases := []struct {
		Template      string
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
		Root          bool
	}{
		{
			Template:      "{{.Env.USERDOMAIN}} :: {{.PWD}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Root:          true,
			Expected:      "\x1b]0;MyCompany :: C:\\vagrant :: Admin :: PowerShell\a",
		},
		{
			Template:      "{{.Folder}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Cwd:           "C:\\vagrant",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;vagrant :: PowerShell\a",
		},
		{
			Template:      "{{.UserName}}@{{.HostName}}{{if .Root}} :: Admin{{end}} :: {{.Shell}}",
			Root:          true,
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@MyHost :: Admin :: PowerShell\a",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Getenv", "USERDOMAIN").Return("MyCompany")
		env.On("Shell").Return(tc.ShellName)

		terminal.Init(shell.GENERIC)

		template.Cache = &cache.Template{
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "MyHost",
			PWD:      tc.Cwd,
			Folder:   "vagrant",
			Segments: maps.NewConcurrent(),
		}
		template.Init(env, nil)

		engine := &Engine{
			Config: &config.Config{
				ConsoleTitleTemplate: tc.Template,
			},
			Env: env,
		}

		title := engine.getTitleTemplateText()
		got := terminal.FormatTitle(title)

		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetConsoleTitleIfGethostnameReturnsError(t *testing.T) {
	cases := []struct {
		Template      string
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
		Root          bool
	}{
		{
			Template:      "Not using Host only {{.UserName}} and {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;Not using Host only MyUser and PowerShell\a",
		},
		{
			Template:      "{{.UserName}}@{{.HostName}} :: {{.Shell}}",
			User:          "MyUser",
			PathSeparator: "\\",
			ShellName:     "PowerShell",
			Expected:      "\x1b]0;MyUser@ :: PowerShell\a",
		},
		{
			Template: "\x1b[93m[\x1b[39m\x1b[96mconsole-title\x1b[39m\x1b[96m ≡\x1b[39m\x1b[31m +0\x1b[39m\x1b[31m ~1\x1b[39m\x1b[31m -0\x1b[39m\x1b[31m !\x1b[39m\x1b[93m]\x1b[39m",
			Expected: "\x1b]0;[console-title ≡ +0 ~1 -0 !]\a",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("Getenv", "USERDOMAIN").Return("MyCompany")
		env.On("Shell").Return(tc.ShellName)

		terminal.Init(shell.GENERIC)

		template.Cache = &cache.Template{
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "",
			Segments: maps.NewConcurrent(),
		}
		template.Init(env, nil)

		engine := &Engine{
			Config: &config.Config{
				ConsoleTitleTemplate: tc.Template,
			},
			Env: env,
		}

		title := engine.getTitleTemplateText()
		got := terminal.FormatTitle(title)

		assert.Equal(t, tc.Expected, got)
	}
}

func TestShouldFill(t *testing.T) {
	cases := []struct {
		Case           string
		Overflow       config.Overflow
		ExpectedFiller string
		Block          config.Block
		Padding        int
		ExpectedBool   bool
	}{
		{
			Case:           "Plain single character with no padding",
			Padding:        0,
			ExpectedFiller: "",
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "-",
			},
		},
		{
			Case:           "Plain single character with 1 padding",
			Padding:        1,
			ExpectedFiller: "-",
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "-",
			},
		},
		{
			Case:           "Plain single character with lots of padding",
			Padding:        200,
			ExpectedFiller: strings.Repeat("-", 200),
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "-",
			},
		},
		{
			Case:           "Plain multi-character with some padding",
			Padding:        20,
			ExpectedFiller: strings.Repeat("-^-", 6) + "  ",
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "-^-",
			},
		},
		{
			Case:           "Template conditional on overflow with no overflow",
			Padding:        3,
			ExpectedFiller: strings.Repeat("X", 3),
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "{{ if .Overflow -}} O {{- else -}} X {{- end }}",
			},
		},
		{
			Case:           "Template conditional on overflow with an overflow",
			Overflow:       config.Break,
			Padding:        3,
			ExpectedFiller: strings.Repeat("O", 3),
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Hide,
				Filler:   "{{ if .Overflow -}} O {{- else -}} X {{- end }}",
			},
		},
		{
			Case:           "Template conditional on overflow break",
			Overflow:       config.Break,
			Padding:        3,
			ExpectedFiller: strings.Repeat("O", 3),
			ExpectedBool:   true,
			Block: config.Block{
				Overflow: config.Break,
				Filler:   `{{ if eq .Overflow "break" -}} O {{- else -}} X {{- end }}`,
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return(shell.GENERIC)

		engine := &Engine{
			Env:      env,
			Overflow: tc.Overflow,
		}

		template.Cache = &cache.Template{
			Shell:    shell.GENERIC,
			Segments: maps.NewConcurrent(),
		}
		template.Init(env, nil)

		terminal.Init(shell.GENERIC)
		terminal.Plain = true
		terminal.Colors = &color.Defaults{}

		gotFiller, gotBool := engine.shouldFill(tc.Block.Filler, tc.Padding)

		assert.Equal(t, tc.ExpectedFiller, gotFiller, tc.Case)
		assert.Equal(t, tc.ExpectedBool, gotBool, tc.Case)
	}
}
