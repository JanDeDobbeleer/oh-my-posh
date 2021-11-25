package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegQueryEnabled(t *testing.T) {
	cases := []struct {
		CaseDescription string
		Path            string
		Key             string
		Fallback        string
		ExpectedSuccess bool
		ExpectedValue   string
		Output          string
		Err             error
	}{
		{
			CaseDescription: "Error",
			Path:            "HKLLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "ProductName",
			Err:             errors.New("No match"),
			ExpectedSuccess: false,
		},
		{
			CaseDescription: "Value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "InstallTime",
			Output:          "xbox",
			ExpectedSuccess: true,
			ExpectedValue:   "xbox",
		},
		{
			CaseDescription: "Fallback value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "InstallTime",
			Output:          "no formatter",
			Fallback:        "cortana",
			Err:             errors.New("No match"),
			ExpectedSuccess: true,
			ExpectedValue:   "cortana",
		},
		{
			CaseDescription: "Fallback value on empty",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "InstallTime",
			Output:          "",
			Fallback:        "anaconda",
			ExpectedSuccess: true,
			ExpectedValue:   "anaconda",
		},
		{
			CaseDescription: "Empty no fallback disabled",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "InstallTime",
			Output:          "",
			ExpectedSuccess: false,
			ExpectedValue:   "",
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getRuntimeGOOS", nil).Return(windowsPlatform)
		env.On("getWindowsRegistryKeyValue", tc.Path, tc.Key).Return(tc.Output, tc.Err)
		props := &properties{
			values: map[Property]interface{}{
				RegistryPath: tc.Path,
				RegistryKey:  tc.Key,
				Fallback:     tc.Fallback,
			},
		}
		r := &winreg{
			env:   env,
			props: props,
		}

		assert.Equal(t, tc.ExpectedSuccess, r.enabled(), tc.CaseDescription)
		assert.Equal(t, tc.ExpectedValue, r.string(), tc.CaseDescription)
	}
}
