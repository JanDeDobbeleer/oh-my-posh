package main

import (
	"encoding/json"
	"strings"
)

type az struct {
	props properties
	env   environmentInfo

	EnvironmentName string     `json:"environmentName"`
	HomeTenantID    string     `json:"homeTenantId"`
	ID              string     `json:"id"`
	IsDefault       bool       `json:"isDefault"`
	Name            string     `json:"name"`
	State           string     `json:"state"`
	TenantID        string     `json:"tenantId"`
	User            *AzureUser `json:"user"`
}

const (
	updateConsentNeeded = "Do you want to continue?"
	updateMessage       = "AZ CLI: Update needed!"
	updateForeground    = "#ffffff"
	updateBackground    = "#ff5349"
)

type AzureUser struct {
	Name string `json:"name"`
}

func (a *az) string() string {
	if a != nil && a.Name == updateMessage {
		return updateMessage
	}
	segmentTemplate := a.props.getString(SegmentTemplate, "{{.Name}}")
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

func (a *az) init(props properties, env environmentInfo) {
	a.props = props
	a.env = env
}

func (a *az) enabled() bool {
	if a.getFromEnvVars() {
		return true
	}

	return a.getFromAzCli()
}

func (a *az) getFromEnvVars() bool {
	environmentName := a.env.getenv("AZ_ENVIRONMENT_NAME")
	userName := a.env.getenv("AZ_USER_NAME")
	id := a.env.getenv("AZ_SUBSCRIPTION_ID")
	accountName := a.env.getenv("AZ_ACCOUNT_NAME")

	if userName == "" && environmentName == "" {
		return false
	}

	a.EnvironmentName = environmentName
	a.Name = accountName
	a.ID = id
	a.User = &AzureUser{
		Name: userName,
	}

	return true
}

func (a *az) getFromAzCli() bool {
	cmd := "az"
	if !a.env.hasCommand(cmd) {
		return false
	}

	output, _ := a.env.runCommand(cmd, "account", "show")
	if len(output) == 0 {
		return false
	}

	if strings.Contains(output, updateConsentNeeded) {
		a.props[ForegroundOverride] = updateForeground
		a.props[BackgroundOverride] = updateBackground
		a.Name = updateMessage
		return true
	}

	err := json.Unmarshal([]byte(output), a)
	return err == nil
}
