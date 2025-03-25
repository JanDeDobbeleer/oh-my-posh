package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Node struct {
	PackageManagerIcon string
	PackageManagerName string

	language
}

const (
	// PnpmIcon illustrates PNPM is used
	PnpmIcon properties.Property = "pnpm_icon"
	// YarnIcon illustrates Yarn is used
	YarnIcon properties.Property = "yarn_icon"
	// NPMIcon illustrates NPM is used
	NPMIcon properties.Property = "npm_icon"
	// FetchPackageManager shows if NPM, PNPM, or Yarn is used
	FetchPackageManager properties.Property = "fetch_package_manager"
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
	n.language.matchesVersionFile = n.matchesVersionFile
	n.language.loadContext = n.loadContext

	return n.language.Enabled()
}

func (n *Node) loadContext() {
	if !n.props.GetBool(FetchPackageManager, false) {
		return
	}

	if n.env.HasFiles("pnpm-lock.yaml") {
		n.PackageManagerName = "pnpm"
		n.PackageManagerIcon = n.props.GetString(PnpmIcon, "\U000F02C1")
		return
	}

	if n.env.HasFiles("yarn.lock") {
		n.PackageManagerName = "yarn"
		n.PackageManagerIcon = n.props.GetString(YarnIcon, "\U000F011B")
		return
	}

	if n.env.HasFiles("package-lock.json") || n.env.HasFiles("package.json") {
		n.PackageManagerName = "npm"
		n.PackageManagerIcon = n.props.GetString(NPMIcon, "\uE71E")
	}
}

func (n *Node) matchesVersionFile() (string, bool) {
	fileVersion := n.env.FileContent(".nvmrc")
	if len(fileVersion) == 0 {
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
			fileVersion = "18.20.3"
		case "iron":
			fileVersion = "20.14.0"
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
