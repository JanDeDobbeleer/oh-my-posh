package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
)

func TestSysInfo(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		ExpectDisabled bool
		SysInfo        platform.SystemInfo
		Precision      int
		Template       string
		Error          error
	}{
		{
			Case:           "Error",
			ExpectDisabled: true,
			Error:          errors.New("error"),
		},
		{
			Case:           "physical mem",
			ExpectedString: "50",
			SysInfo: platform.SystemInfo{
				Memory: platform.Memory{
					PhysicalPercentUsed: 50,
				},
			},
			Template: "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "physical mem 2 digits",
			ExpectedString: "60.51",
			SysInfo: platform.SystemInfo{
				Memory: platform.Memory{
					PhysicalPercentUsed: 60.51,
				},
			},
			Precision: 2,
			Template:  "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "physical meme rounded",
			ExpectedString: "61",
			SysInfo: platform.SystemInfo{
				Memory: platform.Memory{
					PhysicalPercentUsed: 61,
				},
			},
			Template: "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "load",
			ExpectedString: "0.22 0.12 0",
			Precision:      2,
			Template:       "{{ round .Load1 .Precision }} {{round .Load5 .Precision }} {{round .Load15 .Precision }}",
			SysInfo:        platform.SystemInfo{Load1: 0.22, Load5: 0.12, Load15: 0},
		},
		{
			Case:           "not enabled",
			ExpectDisabled: true,
			SysInfo: platform.SystemInfo{
				Memory: platform.Memory{
					PhysicalPercentUsed: 0,
					SwapPercentUsed:     0,
				},
			},
		},
		{
			Case:           "2 physical cpus",
			ExpectedString: "1200 1200",
			Template:       "{{range $cpu := .CPU}}{{round $cpu.Mhz 2 }} {{end}}",
			SysInfo:        platform.SystemInfo{CPU: []cpu.InfoStat{{Mhz: 1200}, {Mhz: 1200}}},
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("SystemInfo").Return(&tc.SysInfo, tc.Error)
		sysInfo := &SystemInfo{}
		props := properties.Map{
			Precision: tc.Precision,
		}
		sysInfo.Init(props, env)
		enabled := sysInfo.Enabled()
		if tc.ExpectDisabled {
			assert.Equal(t, false, enabled, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, sysInfo), tc.Case)
		}
	}
}
