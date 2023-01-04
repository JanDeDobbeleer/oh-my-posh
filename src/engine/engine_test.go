package engine

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/ansi"
	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/shell"

	"github.com/stretchr/testify/assert"
)

func TestCanWriteRPrompt(t *testing.T) {
	cases := []struct {
		Case               string
		Expected           bool
		TerminalWidth      int
		TerminalWidthError error
		PromptLength       int
		RPromptLength      int
	}{
		{Case: "Width Error", Expected: true, TerminalWidthError: errors.New("burp")},
		{Case: "Terminal > Prompt enabled", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 10},
		{Case: "Terminal > Prompt enabled edge", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 70},
		{Case: "Prompt > Terminal enabled", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 70},
		{Case: "Terminal > Prompt disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 100, RPromptLength: 71},
		{Case: "Prompt > Terminal disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 300, RPromptLength: 80},
		{Case: "Prompt > Terminal disabled no room", Expected: true, TerminalWidth: 200, PromptLength: 400, RPromptLength: 80},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalWidthError)
		engine := &Engine{
			Env:               env,
			rpromptLength:     tc.RPromptLength,
			currentLineLength: tc.PromptLength,
			rprompt:           "hello",
		}
		got := engine.canWriteRightBlock(true)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestPrintPWD(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		PWD      string
		OSC99    bool
	}{
		{Case: "Empty PWD"},
		{Case: "OSC99", PWD: ansi.OSC99, Expected: "\x1b]9;9;\"pwd\"\x1b\\"},
		{Case: "OSC7", PWD: ansi.OSC7, Expected: "\x1b]7;\"file://host/pwd\"\x1b\\"},
		{Case: "OSC51", PWD: ansi.OSC51, Expected: "\x1b]51;Auser@host:pwd\x1b\\"},
		{Case: "Deprecated OSC99", OSC99: true, Expected: "\x1b]9;9;\"pwd\"\x1b\\"},
		{Case: "Template (empty)", PWD: "{{ if eq .Shell \"pwsh\" }}osc7{{ end }}"},
		{Case: "Template (non empty)", PWD: "{{ if eq .Shell \"shell\" }}osc7{{ end }}", Expected: "\x1b]7;\"file://host/pwd\"\x1b\\"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return("pwd")
		env.On("Shell").Return("shell")
		env.On("User").Return("user")
		env.On("Host").Return("host", nil)
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env:   make(map[string]string),
			Shell: "shell",
		})

		writer := &ansi.Writer{}
		writer.Init(shell.GENERIC)
		engine := &Engine{
			Env: env,
			Config: &Config{
				PWD:   tc.PWD,
				OSC99: tc.OSC99,
			},
			Writer: writer,
		}
		engine.printPWD()
		got := engine.print()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func BenchmarkEngineRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engineRender()
	}
}

func engineRender() {
	env := &platform.Shell{}
	env.Init()
	defer env.Close()

	cfg := LoadConfig(env)
	defer testClearDefaultConfig()

	writerColors := cfg.MakeColors()
	writer := &ansi.Writer{
		TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
		AnsiColors:         writerColors,
	}
	writer.Init(shell.GENERIC)
	engine := &Engine{
		Config: cfg,
		Env:    env,
		Writer: writer,
	}

	engine.PrintPrimary()
}

func BenchmarkEngineRenderPalette(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engineRender()
	}
}

func TestGetTitle(t *testing.T) {
	cases := []struct {
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
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
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "MyHost",
			PWD:      tc.Cwd,
			Folder:   "vagrant",
		})
		writer := &ansi.Writer{}
		writer.Init(shell.GENERIC)
		engine := &Engine{
			Config: &Config{
				ConsoleTitleTemplate: tc.Template,
			},
			Writer: writer,
			Env:    env,
		}
		title := engine.getTitleTemplateText()
		got := writer.FormatTitle(title)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGetConsoleTitleIfGethostnameReturnsError(t *testing.T) {
	cases := []struct {
		Template      string
		Root          bool
		User          string
		Cwd           string
		PathSeparator string
		ShellName     string
		Expected      string
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
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return(tc.Cwd)
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: map[string]string{
				"USERDOMAIN": "MyCompany",
			},
			Shell:    tc.ShellName,
			UserName: "MyUser",
			Root:     tc.Root,
			HostName: "",
		})
		writer := &ansi.Writer{}
		writer.Init(shell.GENERIC)
		engine := &Engine{
			Config: &Config{
				ConsoleTitleTemplate: tc.Template,
			},
			Writer: writer,
			Env:    env,
		}
		title := engine.getTitleTemplateText()
		got := writer.FormatTitle(title)
		assert.Equal(t, tc.Expected, got)
	}
}
