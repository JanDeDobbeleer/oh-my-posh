package main

import (
	"strings"
)

type az struct {
	props     *properties
	env       environmentInfo
	name      string
	id        string
	account   string
	builder   strings.Builder
	separator string
}

const (
	// SubscriptionInfoSeparator is put between the name and ID
	SubscriptionInfoSeparator Property = "info_separator"
	// DisplaySubscriptionID hides or show the subscription GUID
	DisplaySubscriptionID Property = "display_id"
	// DisplaySubscriptionName hides or shows the subscription display name
	DisplaySubscriptionName Property = "display_name"
	// DisplaySubscriptionAccount hides or shows the subscription account name
	DisplaySubscriptionAccount Property = "display_account"

	updateConsentNeeded = "Do you want to continue?"
	updateMessage       = "AZ CLI: Update needed!"
	updateForeground    = "#ffffff"
	updateBackground    = "#ff5349"
)

func (a *az) string() string {
	a.separator = a.props.getString(SubscriptionInfoSeparator, " | ")
	writeValue := func(value string) {
		if len(value) == 0 {
			return
		}
		if a.builder.Len() > 0 {
			a.builder.WriteString(a.separator)
		}
		a.builder.WriteString(value)
	}
	if a.props.getBool(DisplaySubscriptionAccount, false) {
		writeValue(a.account)
	}
	if a.props.getBool(DisplaySubscriptionName, true) {
		writeValue(a.name)
	}
	if a.props.getBool(DisplaySubscriptionID, false) {
		writeValue(a.id)
	}

	return a.builder.String()
}

func (a *az) init(props *properties, env environmentInfo) {
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
	a.name = a.env.getenv("AZ_SUBSCRIPTION_NAME")
	a.id = a.env.getenv("AZ_SUBSCRIPTION_ID")
	a.account = a.env.getenv("AZ_SUBSCRIPTION_ID")

	if a.name == "" && a.id == "" {
		return false
	}

	return true
}

func (a *az) getFromAzCli() bool {
	cmd := "az"
	if !a.env.hasCommand(cmd) {
		return false
	}

	output, _ := a.env.runCommand(cmd, "account", "show", "--query=[name,id,user.name]", "-o=tsv")
	if len(output) == 0 {
		return false
	}

	if strings.Contains(output, updateConsentNeeded) {
		a.props.foreground = updateForeground
		a.props.background = updateBackground
		a.name = updateMessage
		return true
	}

	splittedOutput := strings.Split(output, "\n")
	if len(splittedOutput) < 3 {
		return false
	}

	a.name = strings.TrimSpace(splittedOutput[0])
	a.id = strings.TrimSpace(splittedOutput[1])
	a.account = strings.TrimSpace(splittedOutput[2])
	return true
}
