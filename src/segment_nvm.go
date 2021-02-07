package main

import (
	"bufio"
	"fmt"
	"strings"
)

type nvm struct {
	props *properties
	env   environmentInfo
}

const (
	NodeIcon     Property = "node_icon"
	MatchColor   Property = "match_color"
	NoMatchColor Property = "no_match_color"
)

func (n *nvm) enabled() bool {
	extensions := []string{"*.js", "*.ts", "package.json", ".nvmrc", "jsx"}
	for i, extension := range extensions {
		if n.env.hasFiles(extension) {
			break
		}
		if i == len(extensions)-1 {
			return false
		}
	}

	return true
}

func (n *nvm) string() string {
	var nvm_version string
	content := n.env.getFileContent(".nvmrc")
	if len(content) > 0 {
		scanner := bufio.NewScanner(strings.NewReader(content))
		if scanner.Scan() {
			nvm_version = scanner.Text()
		}
	}

	version, err := n.env.runCommand("node", "--version")
	if _, ok := err.(*commandError); ok {
		version = "---"
	}

	n.props.background = n.props.getString(MatchColor, "#448C42")
	n.props.foreground = "#FFFFFF"
	res := version
	logo := n.props.getString(NodeIcon, "")
	if len(nvm_version) > 0 && nvm_version != version {
		n.props.background = n.props.getString(NoMatchColor, "red")
		res = nvm_version
	}

	return fmt.Sprintf("%s %s", logo, res)
}

func (s *nvm) init(props *properties, env environmentInfo) {
	s.props = props
	s.env = env
}
