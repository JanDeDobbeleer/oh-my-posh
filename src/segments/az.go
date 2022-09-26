package segments

import (
	"encoding/json"
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"path/filepath"
	"strings"
)

type Az struct {
	props properties.Properties
	env   environment.Environment

	AzureSubscription
	Origin string
}

const (
	Source properties.Property = "source"

	pwsh       = "pwsh"
	cli        = "cli"
	firstMatch = "first_match"
	azureEnv   = "POSH_AZURE_SUBSCRIPTION"
)

type AzureConfig struct {
	Subscriptions  []*AzureSubscription `json:"subscriptions"`
	InstallationID string               `json:"installationId"`
}

type AzureSubscription struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	State            string        `json:"state"`
	User             *AzureUser    `json:"user"`
	IsDefault        bool          `json:"isDefault"`
	TenantID         string        `json:"tenantId"`
	EnvironmentName  string        `json:"environmentName"`
	HomeTenantID     string        `json:"homeTenantId"`
	ManagedByTenants []interface{} `json:"managedByTenants"`
}

type AzureUser struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AzurePowerShellSubscription struct {
	Name    string `json:"Name"`
	Account struct {
		Type string `json:"Type"`
	} `json:"Account"`
	Environment struct {
		Name string `json:"Name"`
	} `json:"Environment"`
	Subscription struct {
		ID                 string `json:"Id"`
		Name               string `json:"Name"`
		State              string `json:"State"`
		ExtendedProperties struct {
			Account string `json:"Account"`
		} `json:"ExtendedProperties"`
	} `json:"Subscription"`
	Tenant struct {
		ID string `json:"Id"`
	} `json:"Tenant"`
}

func (a *Az) Template() string {
	return " {{ .Name }} "
}

func (a *Az) Init(props properties.Properties, env environment.Environment) {
	a.props = props
	a.env = env
}

func (a *Az) Enabled() bool {
	source := a.props.GetString(Source, firstMatch)
	switch source {
	case firstMatch:
		return a.getCLISubscription() || a.getModuleSubscription()
	case pwsh:
		return a.getModuleSubscription()
	case cli:
		return a.getCLISubscription()
	}
	return false
}

func (a *Az) FileContentWithoutBom(file string) string {
	config := a.env.FileContent(file)
	const ByteOrderMark = "\ufeff"
	return strings.TrimLeft(config, ByteOrderMark)
}

func (a *Az) getCLISubscription() bool {
	cfg, err := a.findConfig("azureProfile.json")
	if err != nil {
		return false
	}
	content := a.FileContentWithoutBom(cfg)
	if len(content) == 0 {
		return false
	}
	var config AzureConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return false
	}
	for _, subscription := range config.Subscriptions {
		if subscription.IsDefault {
			a.AzureSubscription = *subscription
			a.Origin = "CLI"
			return true
		}
	}
	return false
}

func (a *Az) getModuleSubscription() bool {
	envSubscription := a.env.Getenv(azureEnv)
	if len(envSubscription) == 0 {
		return false
	}
	var config AzurePowerShellSubscription
	if err := json.Unmarshal([]byte(envSubscription), &config); err != nil {
		return false
	}
	a.IsDefault = true
	a.EnvironmentName = config.Environment.Name
	a.TenantID = config.Tenant.ID
	a.ID = config.Subscription.ID
	a.Name = config.Subscription.Name
	a.State = config.Subscription.State
	a.User = &AzureUser{
		Name: config.Subscription.ExtendedProperties.Account,
		Type: config.Account.Type,
	}
	a.Origin = "PWSH"
	return true
}

func (a *Az) findConfig(fileName string) (string, error) {
	configDirs := []string{
		a.env.Getenv("AZURE_CONFIG_DIR"),
		filepath.Join(a.env.Home(), ".azure"),
		filepath.Join(a.env.Home(), ".Azure"),
	}
	for _, dir := range configDirs {
		if len(dir) != 0 && a.env.HasFilesInDir(dir, fileName) {
			return filepath.Join(dir, fileName), nil
		}
	}
	return "", errors.New("azure config dir not found")
}
