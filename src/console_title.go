package main

import "fmt"

type consoleTitle struct {
	env      environmentInfo
	settings *Settings
	formats  *ansiFormats
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
	var title string
	switch t.settings.ConsoleTitleStyle {
	case FullPath:
		title = t.env.getcwd()
	case FolderName:
		fallthrough
	default:
		title = base(t.env.getcwd(), t.env)
	}
	return fmt.Sprintf(t.formats.title, title)
}
