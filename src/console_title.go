package main

import "fmt"

type consoleTitle struct {
	env      environmentInfo
	settings *Settings
}

// ConsoleTitleStyle defines how to show the title in the console window
type ConsoleTitleStyle string

const (
	// FolderName show the current folder name
	FolderName ConsoleTitleStyle = "folder"
	// FullPath show the current path
	FullPath ConsoleTitleStyle = "path"
)

func (t *consoleTitle) getConsoleTitle() string {
	switch t.settings.ConsoleTitleStyle {
	case FullPath:
		return t.formatConsoleTitle(t.env.getcwd())
	case FolderName:
		fallthrough
	default:
		return t.formatConsoleTitle(base(t.env.getcwd(), t.env))
	}
}

func (t *consoleTitle) formatConsoleTitle(title string) string {
	var format string
	switch t.env.getShellName() {
	case zsh:
		format = "%%{\033]0;%s\007%%}"
	case bash:
		format = "\\[\033]0;%s\007\\]"
	default:
		format = "\033]0;%s\007"
	}
	return fmt.Sprintf(format, title)
}
