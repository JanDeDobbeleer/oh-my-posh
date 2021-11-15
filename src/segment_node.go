package main

import "fmt"

type node struct {
	language           *language
	packageManagerIcon string
}

const (
	// YarnIcon illustrates Yarn is used
	YarnIcon Property = "yarn_icon"
	// NPMIcon illustrates NPM is used
	NPMIcon Property = "npm_icon"
	// DisplayPackageManager shows if NPM or Yarn is used
	DisplayPackageManager Property = "display_package_manager"
)

func (n *node) string() string {
	version := n.language.string()
	return fmt.Sprintf("%s%s", version, n.packageManagerIcon)
}

func (n *node) init(props *properties, env environmentInfo) {
	n.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"*.js", "*.ts", "package.json", ".nvmrc", "pnpm-workspace.yaml", ".pnpmfile.cjs", ".npmrc"},
		commands: []*cmd{
			{
				executable: "node",
				args:       []string{"--version"},
				regex:      `(?:v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%[1]s](https://github.com/nodejs/node/blob/master/doc/changelogs/CHANGELOG_V%[2]s.md#%[1]s)",
		matchesVersionFile: n.matchesVersionFile,
		loadContext:        n.loadContext,
	}
}

func (n *node) enabled() bool {
	return n.language.enabled()
}

func (n *node) loadContext() {
	if !n.language.props.getBool(DisplayPackageManager, false) {
		return
	}
	if n.language.env.hasFiles("yarn.lock") {
		n.packageManagerIcon = n.language.props.getString(YarnIcon, " \uF61A")
		return
	}
	if n.language.env.hasFiles("package-lock.json") || n.language.env.hasFiles("package.json") {
		n.packageManagerIcon = n.language.props.getString(NPMIcon, " \uE71E")
	}
}

func (n *node) matchesVersionFile() bool {
	fileVersion := n.language.env.getFileContent(".nvmrc")
	if len(fileVersion) == 0 {
		return true
	}

	regex := fmt.Sprintf(
		`(?im)^v?%s(\.?%s)?(\.?%s)?$`,
		n.language.activeCommand.version.Major,
		n.language.activeCommand.version.Minor,
		n.language.activeCommand.version.Patch,
	)

	return matchString(regex, fileVersion)
}
