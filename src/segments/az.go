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

type AzurePowerShellConfig struct {
	DefaultContextKey string                                  `json:"DefaultContextKey"`
	Contexts          map[string]*AzurePowerShellSubscription `json:"Contexts"`
}

type AzurePowerShellSubscription struct {
	Account struct {
		ID         string      `json:"Id"`
		Credential interface{} `json:"Credential"`
		Type       string      `json:"Type"`
		TenantMap  struct {
		} `json:"TenantMap"`
		ExtendedProperties struct {
			Subscriptions string `json:"Subscriptions"`
			Tenants       string `json:"Tenants"`
			HomeAccountID string `json:"HomeAccountId"`
		} `json:"ExtendedProperties"`
	} `json:"Account"`
	Tenant struct {
		ID                 string      `json:"Id"`
		Directory          interface{} `json:"Directory"`
		IsHome             bool        `json:"IsHome"`
		ExtendedProperties struct {
		} `json:"ExtendedProperties"`
	} `json:"Tenant"`
	Subscription struct {
		ID                 string `json:"Id"`
		Name               string `json:"Name"`
		State              string `json:"State"`
		ExtendedProperties struct {
			HomeTenant          string `json:"HomeTenant"`
			AuthorizationSource string `json:"AuthorizationSource"`
			SubscriptionPolices string `json:"SubscriptionPolices"`
			Tenants             string `json:"Tenants"`
			Account             string `json:"Account"`
			Environment         string `json:"Environment"`
		} `json:"ExtendedProperties"`
	} `json:"Subscription"`
	Environment struct {
		Name string `json:"Name"`
	} `json:"Environment"`
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
	var content string
	configDir, err := a.ConfigDir(true)
	if err != nil {
		return false
	}
	profile := filepath.Join(configDir, "azureProfile.json")
	if content = a.FileContentWithoutBom(profile); len(content) == 0 {
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
	var content string
	configDir, err := a.ConfigDir(false)
	if err != nil {
		return false
	}
	profiles := []string{
		filepath.Join(configDir, "AzureRmContext.json"),
	}
	for _, profile := range profiles {
		if content = a.FileContentWithoutBom(profile); len(content) != 0 {
			break
		}
	}
	if len(content) == 0 {
		return false
	}
	var config AzurePowerShellConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return false
	}
	defaultContext := config.Contexts[config.DefaultContextKey]
	if defaultContext == nil {
		return false
	}
	a.IsDefault = true
	a.EnvironmentName = defaultContext.Environment.Name
	a.TenantID = defaultContext.Tenant.ID
	a.ID = defaultContext.Subscription.ID
	a.Name = defaultContext.Subscription.Name
	a.State = defaultContext.Subscription.State
	a.User = &AzureUser{
		Name: defaultContext.Subscription.ExtendedProperties.Account,
		Type: defaultContext.Account.Type,
	}
	a.Origin = "PWSH"
	return true
}

func (a *Az) ConfigDir(cli bool) (string, error) {
	configDirs := []string{
		filepath.Join(a.env.Home(), ".azure"),
		filepath.Join(a.env.Home(), ".Azure"),
	}
	if cli {
		configDirs = append([]string{a.env.Getenv("AZURE_CONFIG_DIR")}, configDirs...)
	}
	for _, dir := range configDirs {
		if len(dir) != 0 && a.env.HasFolder(dir) {
			return dir, nil
		}
	}
	return "", errors.New("azure config dir not found")
}
