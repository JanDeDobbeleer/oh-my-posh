package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzSegment(t *testing.T) {
	cases := []struct {
		Case               string
		ExpectedEnabled    bool
		ExpectedString     string
		EnvEnvironmentName string
		EnvUserName        string
		AccountName        string
		EnvSubscriptionID  string
		CLIExists          bool
		CLIEnvironmentname string
		CLISubscriptionID  string
		CLIAccountName     string
		CLIUserName        string
		Template           string
	}{
		{
			Case:               "display account name",
			ExpectedEnabled:    true,
			ExpectedString:     "foobar",
			CLIExists:          true,
			CLIEnvironmentname: "foo",
			CLISubscriptionID:  "bar",
			CLIAccountName:     "foobar",
			Template:           "{{.Name}}",
		},
		{
			Case:               "envvars present",
			ExpectedEnabled:    true,
			ExpectedString:     "foo$bar",
			EnvEnvironmentName: "foo",
			EnvUserName:        "bar",
			CLIExists:          false,
			Template:           "{{.EnvironmentName}}${{.User.Name}}",
		},
		{
			Case:               "envvar environment name present",
			ExpectedEnabled:    true,
			ExpectedString:     "foo",
			EnvEnvironmentName: "foo",
			CLIExists:          false,
			Template:           "{{.EnvironmentName}}",
		},
		{
			Case:            "envvar user name present",
			ExpectedEnabled: true,
			ExpectedString:  "bar",
			EnvUserName:     "bar",
			CLIExists:       false,
			Template:        "{{.User.Name}}",
		},
		{
			Case:              "envvar subscription id",
			ExpectedEnabled:   true,
			ExpectedString:    "foobar",
			EnvSubscriptionID: "foobar",
			EnvUserName:       "bar",
			CLIExists:         false,
			Template:          "{{.ID}}",
		},
		{
			Case:            "cli not found",
			ExpectedEnabled: false,
			ExpectedString:  "",
			CLIExists:       false,
		},
		{
			Case:               "cli contains data",
			ExpectedEnabled:    true,
			ExpectedString:     "foo$bar",
			CLIExists:          true,
			CLIEnvironmentname: "foo",
			CLISubscriptionID:  "bar",
			Template:           "{{.EnvironmentName}}${{.ID}}",
		},
		{
			Case:               "print only environment ame",
			ExpectedEnabled:    true,
			ExpectedString:     "foo",
			CLIExists:          true,
			CLIEnvironmentname: "foo",
			CLISubscriptionID:  "bar",
			Template:           "{{.EnvironmentName}}",
		},
		{
			Case:               "print only id",
			ExpectedEnabled:    true,
			ExpectedString:     "bar",
			CLIExists:          true,
			CLIEnvironmentname: "foo",
			CLISubscriptionID:  "bar",
			Template:           "{{.ID}}",
		},
		{
			Case:               "print none",
			ExpectedEnabled:    true,
			CLIExists:          true,
			CLIEnvironmentname: "foo",
			CLISubscriptionID:  "bar",
		},
		{
			Case:               "update needed",
			ExpectedEnabled:    true,
			ExpectedString:     updateMessage,
			CLIExists:          true,
			CLIEnvironmentname: "Do you want to continue? (Y/n): Visual Studio Enterprise",
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getenv", "AZ_ENVIRONMENT_NAME").Return(tc.EnvEnvironmentName)
		env.On("getenv", "AZ_USER_NAME").Return(tc.EnvUserName)
		env.On("getenv", "AZ_SUBSCRIPTION_ID").Return(tc.EnvSubscriptionID)
		env.On("getenv", "AZ_ACCOUNT_NAME").Return(tc.AccountName)
		env.On("hasCommand", "az").Return(tc.CLIExists)
		env.On("runCommand", "az", []string{"account", "show"}).Return(
			fmt.Sprintf(`{
				"environmentName": "%s",
				"homeTenantId": "8d934305-ac9f-46fe-b0e7-50fd32ad2acf",
				"id": "%s",
				"isDefault": true,
				"managedByTenants": [],
				"name": "%s",
				"state": "Enabled",
				"tenantId": "8d934305-ac9f-46fe-b0e7-50fd32ad2acf",
				"user": {
				  "name": "%s",
				  "type": "user"
				}
			  }`, tc.CLIEnvironmentname, tc.CLISubscriptionID, tc.CLIAccountName, tc.CLIUserName),
			nil,
		)
		props := &properties{
			values: map[Property]interface{}{
				SegmentTemplate: tc.Template,
			},
		}

		az := &az{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.ExpectedEnabled, az.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, az.string(), tc.Case)
	}
}
