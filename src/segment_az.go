package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type az struct {
	props Properties
	env   environmentInfo

	AzureSubscription
}

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

func (a *az) string() string {
	segmentTemplate := a.props.getString(SegmentTemplate, "{{ .EnvironmentName }}")
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

func (a *az) init(props Properties, env environmentInfo) {
	a.props = props
	a.env = env
}

func (a *az) getFileContentWithoutBom(file string) string {
	const ByteOrderMark = "\ufeff"
	file = filepath.Join(a.env.homeDir(), file)
	config := a.env.getFileContent(file)
	return strings.TrimLeft(config, ByteOrderMark)
}

func (a *az) enabled() bool {
	return a.getAzureProfile() || a.getAzureRmContext()
}

func (a *az) getAzureProfile() bool {
	var content string
	if content = a.getFileContentWithoutBom(".azure/azureProfile.json"); len(content) == 0 {
		return false
	}
	var config AzureConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return false
	}
	for _, subscription := range config.Subscriptions {
		if subscription.IsDefault {
			a.AzureSubscription = *subscription
			return true
		}
	}
	return false
}

func (a *az) getAzureRmContext() bool {
	var content string
	if content = a.getFileContentWithoutBom(".azure/AzureRmContext.json"); len(content) == 0 {
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
	}
	return true
}
