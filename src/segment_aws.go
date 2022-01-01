package main

import (
	"fmt"
	"strings"
)

type aws struct {
	props   Properties
	env     environmentInfo
	Profile string
	Region  string
}

const (
	defaultUser = "default"
)

func (a *aws) init(props Properties, env environmentInfo) {
	a.props = props
	a.env = env
}

func (a *aws) enabled() bool {
	getEnvFirstMatch := func(envs ...string) string {
		for _, env := range envs {
			value := a.env.getenv(env)
			if value != "" {
				return value
			}
		}
		return ""
	}
	displayDefaultUser := a.props.getBool(DisplayDefault, true)
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

func (a *aws) getConfigFileInfo() {
	configPath := a.env.getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.aws/config", a.env.homeDir())
	}
	config := a.env.getFileContent(configPath)
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

func (a *aws) string() string {
	segmentTemplate := a.props.getString(SegmentTemplate, "{{.Profile}}{{if .Region}}@{{.Region}}{{end}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  a,
		Env:      a.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}
