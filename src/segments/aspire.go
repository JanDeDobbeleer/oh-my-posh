package segments

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Aspire struct {
	Base

	AppHostPath string
	Version     string
	Name        string
	Lang        string
	Running     bool
}

var aspireAppHostFiles = []string{
	"apphost.cs",
	"apphost.ts",
}

const aspireCommand = "aspire"

const (
	// FetchRunning controls whether the segment should query aspire ps.
	FetchRunning options.Option = "fetch_running"
)

type aspireAppHostSelection struct {
	SelectedProjectFile      string   `json:"selected_project_file"`
	AllProjectFileCandidates []string `json:"all_project_file_candidates"`
}

type aspireAppHostProcess struct {
	AppHostPath string `json:"appHostPath"`
}

func (a *Aspire) Template() string {
	return " \uf423 {{ .Name }}{{ if .Running }} \uf00c{{ end }} "
}

func (a *Aspire) Enabled() bool {
	appHostPath := a.resolveAppHostPath()
	if appHostPath == "" {
		return false
	}

	a.AppHostPath = filepath.Clean(appHostPath)
	a.Name = aspireDisplayName(a.AppHostPath)
	a.Lang = aspireLanguage(a.AppHostPath)
	a.resolveVersion()
	a.Running = a.options.Bool(FetchRunning, true) && a.isRunning()

	return true
}

func (a *Aspire) resolveAppHostPath() string {
	if a.env.HasCommand(aspireCommand) {
		appHostPath := a.resolveAppHostPathWithCLI()
		if appHostPath != "" {
			return appHostPath
		}
	}

	for _, candidate := range aspireAppHostFiles {
		file, err := a.env.HasParentFilePath(candidate, false)
		if err == nil {
			return file.Path
		}
	}

	return ""
}

func (a *Aspire) resolveAppHostPathWithCLI() string {
	output, err := a.env.RunCommand(aspireCommand, "extension", "get-apphosts")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		var selection aspireAppHostSelection
		if err := json.Unmarshal([]byte(trimmed), &selection); err != nil {
			continue
		}

		if selection.SelectedProjectFile != "" {
			return selection.SelectedProjectFile
		}

		if len(selection.AllProjectFileCandidates) == 1 {
			return selection.AllProjectFileCandidates[0]
		}
	}

	return ""
}

func (a *Aspire) isRunning() bool {
	if !a.env.HasCommand(aspireCommand) || a.AppHostPath == "" {
		return false
	}

	for _, args := range [][]string{{"ps", "--format", "json", "--resources"}, {"ps", "--format", "json"}} {
		output, err := a.env.RunCommand(aspireCommand, args...)
		if err != nil {
			continue
		}

		running, parsed := a.parseRunningAppHosts(output)
		if parsed {
			return running
		}
	}

	return false
}

func (a *Aspire) parseRunningAppHosts(output string) (bool, bool) {
	var appHosts []aspireAppHostProcess
	if err := json.Unmarshal([]byte(output), &appHosts); err != nil {
		return false, false
	}

	for _, appHost := range appHosts {
		if sameFilePath(appHost.AppHostPath, a.AppHostPath) {
			return true, true
		}
	}

	return false, true
}

func (a *Aspire) resolveVersion() {
	file, err := a.env.HasParentFilePath("Directory.Packages.props", false)
	if err != nil {
		return
	}

	propsContent := a.env.FileContent(file.Path)
	if propsContent == "" {
		return
	}

	if idx := strings.Index(propsContent, "Aspire.Hosting.AppHost"); idx >= 0 {
		a.Version = extractVersionAttribute(propsContent[idx:])
	}
}

func extractVersionAttribute(s string) string {
	versionKey := "Version=\""
	idx := strings.Index(s, versionKey)
	if idx < 0 {
		return ""
	}

	start := idx + len(versionKey)
	end := strings.Index(s[start:], "\"")
	if end < 0 {
		return ""
	}

	return s[start : start+end]
}

func aspireDisplayName(appHostPath string) string {
	fileName := filepath.Base(appHostPath)
	if fileName == "apphost.cs" || fileName == "apphost.ts" {
		return filepath.Base(filepath.Dir(appHostPath))
	}

	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func aspireLanguage(appHostPath string) string {
	switch strings.TrimPrefix(filepath.Ext(appHostPath), ".") {
	case "cs", "ts":
		return strings.TrimPrefix(filepath.Ext(appHostPath), ".")
	default:
		return ""
	}
}

func sameFilePath(left, right string) bool {
	cleanLeft := filepath.Clean(left)
	cleanRight := filepath.Clean(right)

	if cleanLeft == cleanRight {
		return true
	}

	return strings.EqualFold(cleanLeft, cleanRight)
}
