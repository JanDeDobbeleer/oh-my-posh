package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Node struct {
	PackageManagerIcon string
	PackageManagerName string

	Language
}

const (
	// PnpmIcon illustrates PNPM is used
	PnpmIcon options.Option = "pnpm_icon"
	// YarnIcon illustrates Yarn is used
	YarnIcon options.Option = "yarn_icon"
	// NPMIcon illustrates NPM is used
	NPMIcon options.Option = "npm_icon"
	// BunIcon illustrates Bun is used
	BunIcon options.Option = "bun_icon"
	// FetchPackageManager shows if Bun, NPM, PNPM, or Yarn is used
	FetchPackageManager options.Option = "fetch_package_manager"
)

func (n *Node) Template() string {
	return " {{ if .PackageManagerIcon }}{{ .PackageManagerIcon }} {{ end }}{{ .Full }} "
}

func (n *Node) Enabled() bool {
	n.extensions = []string{"*.js", "*.ts", "package.json", ".nvmrc", "pnpm-workspace.yaml", ".pnpmfile.cjs", ".vue"}
	n.commands = []*cmd{
		{
			executable: "node",
			args:       []string{"--version"},
			regex:      `(?:v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	n.versionURLTemplate = "https://github.com/nodejs/node/blob/master/doc/changelogs/CHANGELOG_V{{ .Major }}.md#{{ .Full }}"
	n.Language.matchesVersionFile = n.matchesVersionFile
	n.Language.loadContext = n.loadContext

	return n.Language.Enabled()
}

func (n *Node) loadContext() {
	if !n.options.Bool(FetchPackageManager, false) {
		return
	}

	packageManagerDefinitions := []struct {
		fileName     string
		name         string
		iconProperty options.Option
		defaultIcon  string
	}{
		{
			fileName:     "pnpm-lock.yaml",
			name:         "pnpm",
			iconProperty: PnpmIcon,
			defaultIcon:  "\U000F02C1",
		},
		{
			fileName:     "yarn.lock",
			name:         "yarn",
			iconProperty: YarnIcon,
			defaultIcon:  "\U000F011B",
		},
		{
			fileName:     "bun.lockb",
			name:         "bun",
			iconProperty: BunIcon,
			defaultIcon:  "\ue76f",
		},
		{
			fileName:     "bun.lock",
			name:         "bun",
			iconProperty: BunIcon,
			defaultIcon:  "\ue76f",
		},
		{
			fileName:     "package-lock.json",
			name:         "npm",
			iconProperty: NPMIcon,
			defaultIcon:  "\uE71E",
		},
		{
			fileName:     "package.json",
			name:         "npm",
			iconProperty: NPMIcon,
			defaultIcon:  "\uE71E",
		},
	}

	for _, pm := range packageManagerDefinitions {
		if n.env.HasFiles(pm.fileName) {
			n.PackageManagerName = pm.name
			n.PackageManagerIcon = n.options.String(pm.iconProperty, pm.defaultIcon)
			break
		}
	}
}

func (n *Node) matchesVersionFile() (string, bool) {
	fileVersion := n.env.FileContent(".nvmrc")
	if fileVersion == "" {
		return "", true
	}

	fileVersion = strings.TrimSpace(fileVersion)

	if strings.HasPrefix(fileVersion, "lts/") {
		fileVersion = strings.ToLower(fileVersion)
		codeName := strings.TrimPrefix(fileVersion, "lts/")
		switch codeName {
		case "argon":
			fileVersion = "4.9.1"
		case "boron":
			fileVersion = "6.17.1"
		case "carbon":
			fileVersion = "8.17.0"
		case "dubnium":
			fileVersion = "10.24.1"
		case "erbium":
			fileVersion = "12.22.12"
		case "fermium":
			fileVersion = "14.21.3"
		case "gallium":
			fileVersion = "16.20.2"
		case "hydrogen":
			fileVersion = "18.20.8"
		case "iron":
			fileVersion = "20.19.3"
		case "jod":
			fileVersion = "22.17.0"
		}
	}

	re := fmt.Sprintf(
		`(?im)^v?%s(\.?%s)?(\.?%s)?$`,
		n.Major,
		n.Minor,
		n.Patch,
	)

	version := strings.TrimSpace(fileVersion)
	version = strings.TrimPrefix(version, "v")

	return version, regex.MatchString(re, fileVersion)
}
