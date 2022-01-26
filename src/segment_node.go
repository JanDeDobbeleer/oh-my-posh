package main

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
)

type node struct {
	language

	PackageManagerIcon string
}

const (
	// YarnIcon illustrates Yarn is used
	YarnIcon properties.Property = "yarn_icon"
	// NPMIcon illustrates NPM is used
	NPMIcon properties.Property = "npm_icon"
	// FetchPackageManager shows if NPM or Yarn is used
	FetchPackageManager properties.Property = "fetch_package_manager"
)

func (n *node) template() string {
	return "{{ if .PackageManagerIcon }}{{ .PackageManagerIcon }} {{ end }}{{ .Full }}"
}

func (n *node) init(props properties.Properties, env environment.Environment) {
	n.language = language{
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
	if !n.language.props.GetBool(FetchPackageManager, false) {
		return
	}
	if n.language.env.HasFiles("yarn.lock") {
		n.PackageManagerIcon = n.language.props.GetString(YarnIcon, " \uF61A")
		return
	}
	if n.language.env.HasFiles("package-lock.json") || n.language.env.HasFiles("package.json") {
		n.PackageManagerIcon = n.language.props.GetString(NPMIcon, " \uE71E")
	}
}

func (n *node) matchesVersionFile() bool {
	fileVersion := n.language.env.FileContent(".nvmrc")
	if len(fileVersion) == 0 {
		return true
	}

	re := fmt.Sprintf(
		`(?im)^v?%s(\.?%s)?(\.?%s)?$`,
		n.language.version.Major,
		n.language.version.Minor,
		n.language.version.Patch,
	)

	return regex.MatchString(re, fileVersion)
}
