package main

import (
	"errors"
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

const (
	expectedColor = "#768954"
)

func setupBatteryTests(state battery.State, batteryLevel float64, props *properties) *batt {
	env := &MockedEnvironment{}
	bt := &battery.Battery{
		State:   state,
		Full:    100,
		Current: batteryLevel,
	}
	batteries := []*battery.Battery{
		bt,
	}
	env.On("getBatteryInfo", nil).Return(batteries, nil)
	b := &batt{
		props: props,
		env:   env,
	}
	b.enabled()
	return b
}

func TestBatteryCharging(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			ChargingIcon: "charging ",
		},
	}
	b := setupBatteryTests(battery.Charging, 80, props)
	assert.Equal(t, "charging 80", b.string())
}

func TestBatteryCharged(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			ChargedIcon: "charged ",
		},
	}
	b := setupBatteryTests(battery.Full, 100, props)
	assert.Equal(t, "charged 100", b.string())
}

func TestBatteryDischarging(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			DischargingIcon: "going down ",
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	assert.Equal(t, "going down 70", b.string())
}

func TestBatteryBackgroundColor(t *testing.T) {
	expected := expectedColor
	props := &properties{
		background: "#111111",
		values: map[Property]interface{}{
			DischargingIcon:  "going down ",
			ColorBackground:  true,
			DischargingColor: expected,
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	b.string()
	assert.Equal(t, expected, props.background)
}

func TestBatteryBackgroundColorInvalid(t *testing.T) {
	expected := expectedColor
	props := &properties{
		background: expected,
		values: map[Property]interface{}{
			DischargingIcon:  "going down ",
			ColorBackground:  true,
			DischargingColor: "derp",
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	b.string()
	assert.Equal(t, expected, props.background)
}

func TestBatteryForegroundColor(t *testing.T) {
	expected := expectedColor
	props := &properties{
		foreground: "#111111",
		values: map[Property]interface{}{
			DischargingIcon:  "going down ",
			ColorBackground:  false,
			DischargingColor: expected,
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	b.string()
	assert.Equal(t, expected, props.foreground)
}

func TestBatteryForegroundColorInvalid(t *testing.T) {
	expected := expectedColor
	props := &properties{
		foreground: expected,
		values: map[Property]interface{}{
			DischargingIcon:  "going down ",
			ColorBackground:  false,
			DischargingColor: "derp",
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	b.string()
	assert.Equal(t, expected, props.foreground)
}

func TestBatteryError(t *testing.T) {
	env := &MockedEnvironment{}
	err := errors.New("oh snap")
	batteries := []*battery.Battery{}
	env.On("getBatteryInfo", nil).Return(batteries, err)
	b := &batt{
		props: &properties{
			values: map[Property]interface{}{
				DisplayError: true,
			},
		},
		env: env,
	}
	assert.True(t, b.enabled())
	assert.Equal(t, "oh snap", b.string())
}

func TestBatteryErrorHidden(t *testing.T) {
	env := &MockedEnvironment{}
	err := errors.New("oh snap")
	batteries := []*battery.Battery{}
	env.On("getBatteryInfo", nil).Return(batteries, err)
	b := &batt{
		props: &properties{
			values: map[Property]interface{}{
				DisplayError: false,
			},
		},
		env: env,
	}
	assert.False(t, b.enabled())
}

func TestBatteryNoBattery(t *testing.T) {
	env := &MockedEnvironment{}
	err := &noBatteryError{}
	batteries := []*battery.Battery{}
	env.On("getBatteryInfo", nil).Return(batteries, err)
	b := &batt{
		props: &properties{
			values: map[Property]interface{}{
				DisplayError: true,
			},
		},
		env: env,
	}
	assert.False(t, b.enabled())
}

func TestBatteryDischargingAndDisplayChargingDisabled(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			DischargingIcon: "going down ",
			DisplayCharging: false,
		},
	}
	b := setupBatteryTests(battery.Discharging, 70, props)
	assert.Equal(t, true, b.enabled())
	assert.Equal(t, "going down 70", b.string())
}

func TestBatteryChargingAndDisplayChargingDisabled(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			ChargingIcon:    "charging ",
			DisplayCharging: false,
		},
	}
	b := setupBatteryTests(battery.Charging, 80, props)
	assert.Equal(t, false, b.enabled())
}

func TestBatteryChargedAndDisplayChargingDisabled(t *testing.T) {
	props := &properties{
		values: map[Property]interface{}{
			ChargedIcon:     "charged ",
			DisplayCharging: false,
		},
	}
	b := setupBatteryTests(battery.Full, 100, props)
	assert.Equal(t, false, b.enabled())
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
		segment := &Segment{
			writer: &batt{
				Percentage: tc.Percentage,
				Battery:    tc.Battery,
			},
		}
		segment.Foreground = tc.DefaultColor
		segment.ForegroundTemplates = tc.Templates
		color := segment.foreground()
		assert.Equal(t, tc.ExpectedColor, color, tc.Case)
	}
}
