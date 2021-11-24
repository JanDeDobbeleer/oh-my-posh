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
		ExpectedSuccess bool
		Output          string
		Err             error
	}{
		{
			CaseDescription: "Error",
			Path:            "HKLLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "ProductName",
			ExpectedSuccess: false,
			Err:             errors.New("No match"),
		},
		{
			CaseDescription: "Value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
			Key:             "InstallTime",
			ExpectedSuccess: true,
			Output:          "no formatter",
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
			},
		}
		r := &winreg{
			env:   env,
			props: props,
		}

		assert.Equal(t, r.enabled(), tc.ExpectedSuccess, tc.CaseDescription)
	}
}
