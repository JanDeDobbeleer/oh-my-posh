package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Aws struct {
	props properties.Properties
	env   platform.Environment

	Profile string
	Region  string
}

const (
	defaultUser = "default"
)

func (a *Aws) Template() string {
	return " {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} "
}

func (a *Aws) Init(props properties.Properties, env platform.Environment) {
	a.props = props
	a.env = env
}

func (a *Aws) Enabled() bool {
	getEnvFirstMatch := func(envs ...string) string {
		for _, env := range envs {
			value := a.env.Getenv(env)
			if value != "" {
				return value
			}
		}
		return ""
	}
	displayDefaultUser := a.props.GetBool(properties.DisplayDefault, true)
	a.Profile = getEnvFirstMatch("AWS_VAULT", "AWS_PROFILE")
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}
	a.Region = getEnvFirstMatch("AWS_REGION", "AWS_DEFAULT_REGION")
	if a.Profile != "" && a.Region != "" {
		return true
	}
	if a.Profile == "" && a.Region != "" && displayDefaultUser {
		a.Profile = defaultUser
		return true
	}
	a.getConfigFileInfo()
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}
	return a.Profile != ""
}

func (a *Aws) getConfigFileInfo() {
	configPath := a.env.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.aws/config", a.env.Home())
	}
	config := a.env.FileContent(configPath)
	configSection := "[default]"
	if a.Profile != "" {
		configSection = fmt.Sprintf("[profile %s]", a.Profile)
	}
	configLines := strings.Split(config, "\n")
	var sectionActive bool
	for _, line := range configLines {
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
	if a.Profile == "" && a.Region != "" {
		a.Profile = defaultUser
	}
}
