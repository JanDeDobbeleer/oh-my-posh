package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Node struct {
	PackageManagerIcon string
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

func (n *Node) Init(props properties.Properties, env runtime.Environment) {
	n.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.js", "*.ts", "package.json", ".nvmrc", "pnpm-workspace.yaml", ".pnpmfile.cjs", ".vue"},
		commands: []*cmd{
			{
				executable: "node",
				args:       []string{"--version"},
				regex:      `(?:v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "https://github.com/nodejs/node/blob/master/doc/changelogs/CHANGELOG_V{{ .Major }}.md#{{ .Full }}",
		matchesVersionFile: n.matchesVersionFile,
		loadContext:        n.loadContext,
	}
}

func (n *Node) Enabled() bool {
	return n.language.Enabled()
}

func (n *Node) loadContext() {
	if !n.language.props.GetBool(FetchPackageManager, false) {
		return
	}
	if n.language.env.HasFiles("pnpm-lock.yaml") {
		n.PackageManagerIcon = n.language.props.GetString(PnpmIcon, "\U000F02C1")
		return
	}
	if n.language.env.HasFiles("yarn.lock") {
		n.PackageManagerIcon = n.language.props.GetString(YarnIcon, "\U000F011B")
		return
	}
	if n.language.env.HasFiles("package-lock.json") || n.language.env.HasFiles("package.json") {
		n.PackageManagerIcon = n.language.props.GetString(NPMIcon, "\uE71E")
	}
}

func (n *Node) matchesVersionFile() (string, bool) {
	fileVersion := n.language.env.FileContent(".nvmrc")
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
		n.language.version.Major,
		n.language.version.Minor,
		n.language.version.Patch,
	)

	version := strings.TrimSpace(fileVersion)
	version = strings.TrimPrefix(version, "v")

	return version, regex.MatchString(re, fileVersion)
}
