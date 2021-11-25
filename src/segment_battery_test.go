package main

import (
	"errors"
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

const (
	chargingColor    = "#123456"
	dischargingColor = "#765432"
	chargedColor     = "#248644"
)

func TestBatterySegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		Batteries       []*battery.Battery
		ExpectedString  string
		ExpectedEnabled bool
		ExpectedColor   string
		ColorBackground bool
		DisplayError    bool
		Error           error
		DisableCharging bool
		DisableCharged  bool
	}{
		{Case: "80% charging", Batteries: []*battery.Battery{{Full: 100, State: battery.Charging, Current: 80}}, ExpectedString: "charging 80", ExpectedEnabled: true},
		{Case: "battery full", Batteries: []*battery.Battery{{Full: 100, State: battery.Full, Current: 100}}, ExpectedString: "charged 100", ExpectedEnabled: true},
		{Case: "70% discharging", Batteries: []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}}, ExpectedString: "going down 70", ExpectedEnabled: true},
		{
			Case:            "discharging background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   dischargingColor,
		},
		{
			Case:            "charging background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging, Current: 70}},
			ExpectedString:  "charging 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   chargingColor,
		},
		{
			Case:            "charged background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full, Current: 70}},
			ExpectedString:  "charged 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   chargedColor,
		},
		{
			Case:            "discharging foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			ExpectedColor:   dischargingColor,
		},
		{
			Case:            "charging foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging, Current: 70}},
			ExpectedString:  "charging 70",
			ExpectedEnabled: true,
			ExpectedColor:   chargingColor,
		},
		{
			Case:            "charged foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full, Current: 70}},
			ExpectedString:  "charged 70",
			ExpectedEnabled: true,
			ExpectedColor:   chargedColor,
		},
		{Case: "battery error", DisplayError: true, Error: errors.New("oh snap"), ExpectedString: "oh snap", ExpectedEnabled: true},
		{Case: "battery error disabled", Error: errors.New("oh snap")},
		{Case: "no batteries", DisplayError: true, Error: &noBatteryError{}},
		{Case: "no batteries without error"},
		{Case: "display charging disabled: charging", Batteries: []*battery.Battery{{Full: 100, State: battery.Charging}}, DisableCharging: true},
		{Case: "display charged disabled: charged", Batteries: []*battery.Battery{{Full: 100, State: battery.Full}}, DisableCharged: true},
		{
			Case:            "display charging disabled/display charged enabled: charging",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging}},
			DisableCharging: true,
			DisableCharged:  false},
		{
			Case:            "display charged disabled/display charging enabled: charged",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full}},
			DisableCharged:  true,
			DisableCharging: false},
		{
			Case:            "display charging disabled: discharging",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			DisableCharging: true,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			background: "#111111",
			foreground: "#ffffff",
			values: map[Property]interface{}{
				ChargingIcon:     "charging ",
				ChargedIcon:      "charged ",
				DischargingIcon:  "going down ",
				DischargingColor: dischargingColor,
				ChargedColor:     chargedColor,
				ChargingColor:    chargingColor,
				ColorBackground:  tc.ColorBackground,
				DisplayError:     tc.DisplayError,
			},
		}
		// default values
		if tc.DisableCharging {
			props.values[DisplayCharging] = false
		}
		if tc.DisableCharged {
			props.values[DisplayCharged] = false
		}
		env.On("getBatteryInfo", nil).Return(tc.Batteries, tc.Error)
		b := &batt{
			props: props,
			env:   env,
		}
		enabled := b.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}
		assert.Equal(t, tc.ExpectedString, b.string(), tc.Case)
		if len(tc.ExpectedColor) == 0 {
			continue
		}
		actualColor := b.props.foreground
		if tc.ColorBackground {
			actualColor = b.props.background
		}
		assert.Equal(t, tc.ExpectedColor, actualColor, tc.Case)
	}
}

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
		batt := &batt{
			Percentage: tc.Percentage,
		}
		if tc.Battery != nil {
			batt.Battery = *tc.Battery
		}
		segment := &Segment{
			writer: batt,
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
