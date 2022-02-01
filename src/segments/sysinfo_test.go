package segments

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
)

func TestSysInfo(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		ExpectDisabled bool
		SysInfo        SystemInfo
		Precision      int
		Template       string
	}{
		{
			Case:           "physical mem",
			ExpectedString: "50",
			SysInfo:        SystemInfo{PhysicalPercentUsed: 50},
			Template:       "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "physical mem 2 digits",
			ExpectedString: "60.51",
			SysInfo:        SystemInfo{Precision: 2, PhysicalPercentUsed: 60.51},
			Template:       "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "physical meme rounded",
			ExpectedString: "61",
			SysInfo:        SystemInfo{Precision: 0, PhysicalPercentUsed: 61},
			Template:       "{{ round .PhysicalPercentUsed .Precision }}",
		},
		{
			Case:           "load",
			ExpectedString: "0.22 0.12 0",
			Template:       "{{ round .Load1 .Precision }} {{round .Load5 .Precision }} {{round .Load15 .Precision }}",
			SysInfo:        SystemInfo{Precision: 2, Load1: 0.22, Load5: 0.12, Load15: 0}},
		{Case: "not enabled", ExpectDisabled: true, SysInfo: SystemInfo{PhysicalPercentUsed: 0, SwapPercentUsed: 0}},
		{
			Case:           "2 physical cpus",
			ExpectedString: "1200 1200",
			Template:       "{{range $cpu := .CPU}}{{round $cpu.Mhz 2 }} {{end}}",
			SysInfo:        SystemInfo{CPU: []cpu.InfoStat{{Mhz: 1200}, {Mhz: 1200}}},
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		tc.SysInfo.env = env
		tc.SysInfo.props = properties.Map{
			Precision: tc.Precision,
		}
		enabled := tc.SysInfo.Enabled()
		if tc.ExpectDisabled {
			assert.Equal(t, false, enabled, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, tc.SysInfo), tc.Case)
		}
	}
}
