package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestWinReg(t *testing.T) {
	cases := []struct {
		Err             error
		getWRKVOutput   *runtime.WindowsRegistryValue
		CaseDescription string
		Path            string
		Fallback        string
		ExpectedValue   string
		ExpectedSuccess bool
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
			getWRKVOutput:   &runtime.WindowsRegistryValue{ValueType: runtime.STRING, String: "xbox"},
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
			getWRKVOutput:   &runtime.WindowsRegistryValue{ValueType: runtime.STRING, String: ""},
			Fallback:        "anaconda",
			ExpectedSuccess: true,
			ExpectedValue:   "",
		},
		{
			CaseDescription: "Empty string value (no error) should display empty string",
			Path:            "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion\\InstallTime",
			getWRKVOutput:   &runtime.WindowsRegistryValue{ValueType: runtime.STRING, String: ""},
			ExpectedSuccess: true,
			ExpectedValue:   "",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(runtime.WINDOWS)
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
