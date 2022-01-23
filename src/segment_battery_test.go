package main

import (
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

func TestGetBatteryColors(t *testing.T) {
	cases := []struct {
		Case          string
		ExpectedColor string
		Templates     []string
		DefaultColor  string
		Battery       *battery.Battery
		Percentage    int
	}{
		{
			Case:          "Percentage lower",
			ExpectedColor: "color2",
			DefaultColor:  "color",
			Templates: []string{
				"{{if (lt .Percentage 60)}}color2{{end}}",
				"{{if (gt .Percentage 60)}}color3{{end}}",
			},
			Percentage: 50,
		},
		{
			Case:          "Percentage higher",
			ExpectedColor: "color3",
			DefaultColor:  "color",
			Templates: []string{
				"{{if (lt .Percentage 60)}}color2{{end}}",
				"{{if (gt .Percentage 60)}}color3{{end}}",
			},
			Percentage: 70,
		},
		{
			Case:          "Charging",
			ExpectedColor: "color2",
			DefaultColor:  "color",
			Templates: []string{
				"{{if eq \"Charging\" .Battery.State.String}}color2{{end}}",
				"{{if eq \"Discharging\" .Battery.State.String}}color3{{end}}",
				"{{if eq \"Full\" .Battery.State.String}}color4{{end}}",
			},
			Battery: &battery.Battery{
				State: battery.Charging,
			},
		},
		{
			Case:          "Discharging",
			ExpectedColor: "color3",
			DefaultColor:  "color",
			Templates: []string{
				"{{if eq \"Charging\" .Battery.State.String}}color2{{end}}",
				"{{if eq \"Discharging\" .Battery.State.String}}color3{{end}}",
				"{{if eq \"Full\" .Battery.State.String}}color2{{end}}",
			},
			Battery: &battery.Battery{
				State: battery.Discharging,
			},
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.onTemplate()
		batt := &batt{
			Percentage: tc.Percentage,
		}
		if tc.Battery != nil {
			batt.Battery = *tc.Battery
		}
		segment := &Segment{
			writer: batt,
			env:    env,
		}
		segment.Foreground = tc.DefaultColor
		segment.ForegroundTemplates = tc.Templates
		color := segment.foreground()
		assert.Equal(t, tc.ExpectedColor, color, tc.Case)
	}
}

func TestMapBatteriesState(t *testing.T) {
	cases := []struct {
		Case          string
		ExpectedState battery.State
		CurrentState  battery.State
		NewState      battery.State
	}{
		{Case: "charging > charged", ExpectedState: battery.Charging, CurrentState: battery.Full, NewState: battery.Charging},
		{Case: "charging < discharging", ExpectedState: battery.Discharging, CurrentState: battery.Discharging, NewState: battery.Charging},
		{Case: "charging == charging", ExpectedState: battery.Charging, CurrentState: battery.Charging, NewState: battery.Charging},
		{Case: "discharging > charged", ExpectedState: battery.Discharging, CurrentState: battery.Full, NewState: battery.Discharging},
		{Case: "discharging > unknown", ExpectedState: battery.Discharging, CurrentState: battery.Unknown, NewState: battery.Discharging},
		{Case: "discharging > full", ExpectedState: battery.Discharging, CurrentState: battery.Full, NewState: battery.Discharging},
		{Case: "discharging > charging 2", ExpectedState: battery.Discharging, CurrentState: battery.Charging, NewState: battery.Discharging},
		{Case: "discharging > empty", ExpectedState: battery.Discharging, CurrentState: battery.Empty, NewState: battery.Discharging},
	}
	for _, tc := range cases {
		batt := &batt{}
		assert.Equal(t, tc.ExpectedState, batt.mapMostLogicalState(tc.CurrentState, tc.NewState), tc.Case)
	}
}
