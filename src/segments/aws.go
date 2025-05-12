package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Aws struct {
	base

	Profile string
	Region  string
}

const (
	defaultUser = "default"
)

func (a *Aws) Template() string {
	return " {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} "
}

func (a *Aws) Enabled() bool {
	getEnvFirstMatch := func(envs ...string) string {
		for _, env := range envs {
			value := a.env.Getenv(env)
			if len(value) != 0 {
				return value
			}
		}

		return ""
	}

	displayDefaultUser := a.props.GetBool(properties.DisplayDefault, true)
	a.Profile = getEnvFirstMatch("AWS_VAULT", "AWS_DEFAULT_PROFILE", "AWS_PROFILE")
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}

	a.Region = getEnvFirstMatch("AWS_REGION", "AWS_DEFAULT_REGION")
	if len(a.Profile) != 0 && len(a.Region) != 0 {
		return true
	}

	if len(a.Profile) == 0 && len(a.Region) != 0 && displayDefaultUser {
		a.Profile = defaultUser
		return true
	}

	a.getConfigFileInfo()
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}

	return len(a.Profile) != 0
}

func (a *Aws) getConfigFileInfo() {
	configPath := a.env.Getenv("AWS_CONFIG_FILE")
	if len(configPath) == 0 {
		configPath = fmt.Sprintf("%s/.aws/config", a.env.Home())
	}

	config := a.env.FileContent(configPath)
	configSection := "[default]"
	if len(a.Profile) != 0 {
		configSection = fmt.Sprintf("[profile %s]", a.Profile)
	}

	configLines := strings.SplitSeq(config, "\n")
	var sectionActive bool
	for line := range configLines {
		if strings.HasPrefix(line, configSection) {
			sectionActive = true
			continue
		}

		if sectionActive && strings.HasPrefix(line, "region") {
			splitted := strings.Split(line, "=")
			if len(splitted) >= 2 {
				a.Region = strings.TrimSpace(splitted[1])
				break
			}
		}
	}

	if len(a.Profile) == 0 && len(a.Region) != 0 {
		a.Profile = defaultUser
	}
}

func (a *Aws) RegionAlias() string {
	if len(a.Region) == 0 {
		return ""
	}

	splitted := strings.Split(a.Region, "-")
	if len(splitted) < 2 {
		return a.Region
	}

	splitted[1] = regex.ReplaceAllString(`orth|outh|ast|est|entral`, splitted[1], "")
	return strings.Join(splitted, "")
}
