package main

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWinReg(t *testing.T) {
	cases := []struct {
		CaseDescription string
		Path            string
		Fallback        string
		ExpectedSuccess bool
		ExpectedValue   string
		getWRKVOutput   *environment.WindowsRegistryValue
		Err             error
	}{
		{
			CaseDescription: "Error",
			Path:            "HKLLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\ProductName",
			Err:             errors.New("No match"),
			ExpectedSuccess: false,
		},
		{
			CaseDescription: "Value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &environment.WindowsRegistryValue{ValueType: environment.RegString, Str: "xbox"},
			ExpectedSuccess: true,
			ExpectedValue:   "xbox",
		},
		{
			CaseDescription: "Fallback value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			Fallback:        "cortana",
			Err:             errors.New("No match"),
			ExpectedSuccess: true,
			ExpectedValue:   "cortana",
		},
		{
			CaseDescription: "Empty string value (no error) should display empty string even in presence of fallback",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &environment.WindowsRegistryValue{ValueType: environment.RegString, Str: ""},
			Fallback:        "anaconda",
			ExpectedSuccess: true,
			ExpectedValue:   "",
		},
		{
			CaseDescription: "Empty string value (no error) should display empty string",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &environment.WindowsRegistryValue{ValueType: environment.RegString, Str: ""},
			ExpectedSuccess: true,
			ExpectedValue:   "",
		},
		{
			CaseDescription: "DWORD value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &environment.WindowsRegistryValue{ValueType: environment.RegDword, Dword: 0xdeadbeef},
			ExpectedSuccess: true,
			ExpectedValue:   "0xDEADBEEF",
		},
		{
			CaseDescription: "QWORD value",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &environment.WindowsRegistryValue{ValueType: environment.RegQword, Qword: 0x7eb199e57fa1afe1},
			ExpectedSuccess: true,
			ExpectedValue:   "0x7EB199E57FA1AFE1",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return(environment.WindowsPlatform)
		env.On("WindowsRegistryKeyValue", tc.Path).Return(tc.getWRKVOutput, tc.Err)
		r := &WindowsRegistry{
			env: env,
			props: properties.Map{
				RegistryPath: tc.Path,
				Fallback:     tc.Fallback,
			},
		}

		assert.Equal(t, tc.ExpectedSuccess, r.Enabled(), tc.CaseDescription)
		assert.Equal(t, tc.ExpectedValue, renderTemplate(env, r.Template(), r), tc.CaseDescription)
	}
}
