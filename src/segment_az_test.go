package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		EnvSubName      string
		EnvSubID        string
		EnvSubAccount   string
		CliExists       bool
		CliSubName      string
		CliSubID        string
		CliSubAccount   string
		InfoSeparator   string
		DisplayID       bool
		DisplayName     bool
		DisplayAccount  bool
	}{
		{
			Case:            "print only account",
			ExpectedEnabled: true,
			ExpectedString:  "foobar",
			CliExists:       true,
			CliSubName:      "foo",
			CliSubID:        "bar",
			CliSubAccount:   "foobar",
			InfoSeparator:   "$",
			DisplayID:       false,
			DisplayName:     false,
			DisplayAccount:  true,
		},
		{
			Case:            "envvars present",
			ExpectedEnabled: true,
			ExpectedString:  "foo$bar",
			EnvSubName:      "foo",
			EnvSubID:        "bar",
			CliExists:       false,
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     true,
		},
		{
			Case:            "envvar name present",
			ExpectedEnabled: true,
			ExpectedString:  "foo",
			EnvSubName:      "foo",
			CliExists:       false,
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     true,
		},
		{
			Case:            "envvar id present",
			ExpectedEnabled: true,
			ExpectedString:  "bar",
			EnvSubID:        "bar",
			CliExists:       false,
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     true,
		},
		{
			Case:            "envvar account present",
			ExpectedEnabled: true,
			ExpectedString:  "foobar",
			EnvSubAccount:   "foobar",
			EnvSubID:        "bar",
			CliExists:       false,
			InfoSeparator:   "$",
			DisplayAccount:  true,
		},
		{
			Case:            "cli not found",
			ExpectedEnabled: false,
			ExpectedString:  "",
			CliExists:       false,
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     true,
		},
		{
			Case:            "cli contains data",
			ExpectedEnabled: true,
			ExpectedString:  "foo$bar",
			CliExists:       true,
			CliSubName:      "foo",
			CliSubID:        "bar",
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     true,
		},
		{
			Case:            "print only name",
			ExpectedEnabled: true,
			ExpectedString:  "foo",
			CliExists:       true,
			CliSubName:      "foo",
			CliSubID:        "bar",
			InfoSeparator:   "$",
			DisplayID:       false,
			DisplayName:     true,
		},
		{
			Case:            "print only id",
			ExpectedEnabled: true,
			ExpectedString:  "bar",
			CliExists:       true,
			CliSubName:      "foo",
			CliSubID:        "bar",
			InfoSeparator:   "$",
			DisplayID:       true,
			DisplayName:     false,
		},
		{
			Case:            "print none",
			ExpectedEnabled: true,
			CliExists:       true,
			CliSubName:      "foo",
			CliSubID:        "bar",
			InfoSeparator:   "$",
		},
		{
			Case:            "update needed",
			ExpectedEnabled: true,
			ExpectedString:  updateMessage,
			CliExists:       true,
			CliSubName:      "Do you want to continue? (Y/n): Visual Studio Enterprise",
			DisplayID:       false,
			DisplayName:     true,
		},
		{
			Case:            "account info",
			ExpectedEnabled: true,
			ExpectedString:  updateMessage,
			CliExists:       true,
			CliSubName:      "Do you want to continue? (Y/n): Visual Studio Enterprise",
			DisplayID:       false,
			DisplayName:     true,
			DisplayAccount:  true,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getenv", "AZ_SUBSCRIPTION_NAME").Return(tc.EnvSubName)
		env.On("getenv", "AZ_SUBSCRIPTION_ID").Return(tc.EnvSubID)
		env.On("getenv", "AZ_SUBSCRIPTION_ACCOUNT").Return(tc.EnvSubAccount)
		env.On("hasCommand", "az").Return(tc.CliExists)
		env.On("runCommand", "az", []string{"account", "show", "--query=[name,id,user.name]", "-o=tsv"}).Return(
			fmt.Sprintf("%s\n%s\n%s\n", tc.CliSubName, tc.CliSubID, tc.CliSubAccount),
			nil,
		)
		props := &properties{
			values: map[Property]interface{}{
				SubscriptionInfoSeparator:  tc.InfoSeparator,
				DisplaySubscriptionID:      tc.DisplayID,
				DisplaySubscriptionName:    tc.DisplayName,
				DisplaySubscriptionAccount: tc.DisplayAccount,
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
