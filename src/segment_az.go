package main

import (
	"fmt"
	"strings"
)

type az struct {
	props *properties
	env   environmentInfo
	name  string
	id    string
}

const (
	// SubscriptionInfoSeparator is put between the name and ID
	SubscriptionInfoSeparator Property = "info_separator"
	// DisplaySubscriptionID hides or show the subscription GUID
	DisplaySubscriptionID Property = "display_id"
	// DisplaySubscriptionName hides or shows the subscription display name
	DisplaySubscriptionName Property = "display_name"

	updateConsentNeeded = "Do you want to continue?"
	updateMessage       = "AZ CLI: Update needed!"
	updateForeground    = "#ffffff"
	updateBackground    = "#ff5349"
)

func (a *az) string() string {
	separator := ""
	if a.idEnabled() && a.nameEnabled() {
		separator = a.props.getString(SubscriptionInfoSeparator, " | ")
	}

	return fmt.Sprintf("%s%s%s", a.getName(), separator, a.getID())
}

func (a *az) init(props *properties, env environmentInfo) {
	a.props = props
	a.env = env
}

func (a *az) enabled() bool {
	var enabled bool
	a.name, a.id, enabled = a.getFromEnvVars()
	if enabled {
		return enabled
	}

	a.name, a.id, enabled = a.getFromAzCli()
	return enabled
}

func (a *az) getFromEnvVars() (string, string, bool) {
	name := a.env.getenv("AZ_SUBSCRIPTION_NAME")
	id := a.env.getenv("AZ_SUBSCRIPTION_ID")

	if name == "" && id == "" {
		return "", "", false
	}

	return name, id, true
}

func (a *az) getFromAzCli() (string, string, bool) {
	cmd := "az"
	if (!a.idEnabled() && !a.nameEnabled()) || !a.env.hasCommand(cmd) {
		return "", "", false
	}

	output, _ := a.env.runCommand(cmd, "account", "show", "--query=[name,id]", "-o=tsv")
	if output == "" {
		return "", "", false
	}

	if strings.Contains(output, updateConsentNeeded) {
		a.props.foreground = updateForeground
		a.props.background = updateBackground
		return updateMessage, "", true
	}

	splittedOutput := strings.Split(output, "\n")
	if len(splittedOutput) < 2 {
		return "", "", false
	}

	name := strings.TrimSpace(splittedOutput[0])
	id := strings.TrimSpace(splittedOutput[1])

	return name, id, true
}

func (a *az) getID() string {
	if !a.idEnabled() {
		return ""
	}

	return a.id
}

func (a *az) getName() string {
	if !a.nameEnabled() {
		return ""
	}

	return a.name
}

func (a *az) idEnabled() bool {
	return a.props.getBool(DisplaySubscriptionID, false)
}

func (a *az) nameEnabled() bool {
	return a.props.getBool(DisplaySubscriptionName, true)
}
