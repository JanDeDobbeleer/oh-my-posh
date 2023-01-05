package segments

import (
	"regexp"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type CfTarget struct {
	props properties.Properties
	env   platform.Environment

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

func (c *CfTarget) Init(props properties.Properties, env platform.Environment) {
	c.props = props
	c.env = env
}

func (c *CfTarget) Enabled() bool {
	return c.setCFTargetStatus()
}

func (c *CfTarget) getCFTargetCommandOutput() string {
	if !c.env.HasCommand("cf") {
		return ""
	}

	output, err := c.env.RunCommand("cf", "target")

	if err != nil {
		return ""
	}

	return output
}

func (c *CfTarget) setCFTargetStatus() bool {
	output := c.getCFTargetCommandOutput()

	if output == "" {
		return false
	}

	regex := regexp.MustCompile(`API endpoint:\s*(?P<api_url>http[s].*)|user:\s*(?P<user>.*)|org:\s*(?P<org>.*)|space:\s*(?P<space>(.*))`)
	match := regex.FindAllStringSubmatch(output, -1)
	result := make(map[string]string)

	for i, name := range regex.SubexpNames() {
		if i == 0 || len(name) == 0 {
			continue
		}

		for j, val := range match[i-1] {
			if j == 0 {
				continue
			}

			if val != "" {
				result[name] = val
			}
		}
	}

	c.URL = result["api_url"]
	c.Org = result["org"]
	c.Space = result["space"]
	c.User = result["user"]

	return true
}
