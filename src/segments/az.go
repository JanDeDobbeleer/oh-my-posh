package segments

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Az struct {
	props  properties.Properties
	env    runtime.Environment
	Origin string
	AzureSubscription
}

const (
	Source properties.Property = "source"

	Pwsh       = "pwsh"
	Cli        = "cli"
	FirstMatch = "first_match"
	azureEnv   = "POSH_AZURE_SUBSCRIPTION"
)

type AzureConfig struct {
	InstallationID string               `json:"installationId"`
	Subscriptions  []*AzureSubscription `json:"subscriptions"`
}

type AzureSubscription struct {
	User              *AzureUser `json:"user"`
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	State             string     `json:"state"`
	TenantID          string     `json:"tenantId"`
	TenantDisplayName string     `json:"tenantDisplayName"`
	EnvironmentName   string     `json:"environmentName"`
	HomeTenantID      string     `json:"homeTenantId"`
	ManagedByTenants  []any      `json:"managedByTenants"`
	IsDefault         bool       `json:"isDefault"`
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
		ID   string `json:"Id"`
		Name string `json:"Name"`
	} `json:"Tenant"`
}

func (a *Az) Template() string {
	return NameTemplate
}

func (a *Az) Init(props properties.Properties, env runtime.Environment) {
	a.props = props
	a.env = env
}

func (a *Az) Enabled() bool {
	source := a.props.GetString(Source, FirstMatch)
	switch source {
	case FirstMatch:
		return a.getCLISubscription() || a.getModuleSubscription()
	case Pwsh:
		return a.getModuleSubscription()
	case Cli:
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
	a.TenantDisplayName = config.Tenant.Name

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
