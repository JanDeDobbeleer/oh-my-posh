package segments

import (
	"errors"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type CfTarget struct {
	base

	CfTargetDetails
}

type CfTargetDetails struct {
	URL   string
	User  string
	Org   string
	Space string
}

func (c *CfTarget) Template() string {
	return "{{if .Org }}{{ .Org }}{{ end }}{{if .Space }}/{{ .Space }}{{ end }}"
}

func (c *CfTarget) Enabled() bool {
	if !c.env.HasCommand("cf") {
		return false
	}

	displayMode := c.props.GetString(DisplayMode, DisplayModeAlways)
	if displayMode != DisplayModeFiles {
		return c.setCFTargetStatus()
	}

	files := c.props.GetStringArray(properties.Files, []string{"manifest.yml"})
	for _, file := range files {
		manifest, err := c.env.HasParentFilePath(file, false)
		if err != nil || manifest.IsDir {
			continue
		}

		return c.setCFTargetStatus()
	}

	return false
}

func (c *CfTarget) setCFTargetStatus() bool {
	output, err := c.getCFTargetCommandOutput()

	if err != nil {
		return false
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		splitted := strings.SplitN(line, ":", 2)
		if len(splitted) < 2 {
			continue
		}
		key := splitted[0]
		value := strings.TrimSpace(splitted[1])
		switch key {
		case "API endpoint":
			c.URL = value
		case "user":
			c.User = value
		case "org":
			c.Org = value
		case "space":
			c.Space = value
		}
	}

	return true
}

func (c *CfTarget) getCFTargetCommandOutput() (string, error) {
	output, err := c.env.RunCommand("cf", "target")

	if err != nil {
		return "", err
	}

	if len(output) == 0 {
		return "", errors.New("cf command output is empty")
	}

	return output, nil
}
