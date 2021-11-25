package main

import (
	"errors"
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

// GIT Segment

func TestGetStatusDetailStringDefault(t *testing.T) {
	expected := "icon +1"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverride(t *testing.T) {
	expected := "<#123456>icon +1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor: "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideAndIconColorOverride(t *testing.T) {
	expected := "<#789123>work</> <#123456>+1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "<#789123>work</>",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideNoIconColorOverride(t *testing.T) {
	expected := "<#123456>work +1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "work",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatus(t *testing.T) {
	expected := "icon"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatusColorOverride(t *testing.T) {
	expected := "<#123456>icon</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
				WorkingColor:        "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusColorLocalChangesStaging(t *testing.T) {
	expected := changesColor
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
		Staging: &GitStatus{
			Changed: true,
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorLocalChangesWorking(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAheadAndBehind(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  3,
		props: &properties{
			values: map[Property]interface{}{
				AheadAndBehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAhead(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  0,
		props: &properties{
			values: map[Property]interface{}{
				AheadColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorBehind(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  5,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorDefault(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  0,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: changesColor,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor(expected))
}

func TestSetStatusColorForeground(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   false,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.foreground)
}

func TestSetStatusColorBackground(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   true,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.background)
}

func TestStatusColorsWithoutDisplayStatus(t *testing.T) {
	expected := changesColor
	context := &detachedContext{
		status: "## main...origin/main [ahead 33]\n M myfile",
	}
	g := setupHEADContextEnv(context)
	g.props = &properties{
		values: map[Property]interface{}{
			DisplayStatus:       false,
			StatusColorsEnabled: true,
			LocalChangesColor:   expected,
		},
	}
	g.string()
	assert.Equal(t, expected, g.props.background)
}

// EXIT Segement

func TestExitWriterDeprecatedString(t *testing.T) {
	cases := []struct {
		ExitCode        int
		Expected        string
		SuccessIcon     string
		ErrorIcon       string
		DisplayExitCode bool
		AlwaysNumeric   bool
	}{
		{ExitCode: 129, Expected: "SIGHUP", DisplayExitCode: true},
		{ExitCode: 5001, Expected: "5001", DisplayExitCode: true},
		{ExitCode: 147, Expected: "SIGSTOP", DisplayExitCode: true},
		{ExitCode: 147, Expected: "", DisplayExitCode: false},
		{ExitCode: 147, Expected: "147", DisplayExitCode: true, AlwaysNumeric: true},
		{ExitCode: 0, Expected: "wooopie", SuccessIcon: "wooopie"},
		{ExitCode: 129, Expected: "err SIGHUP", ErrorIcon: "err ", DisplayExitCode: true},
		{ExitCode: 129, Expected: "err", ErrorIcon: "err", DisplayExitCode: false},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(tc.ExitCode)
		props := &properties{
			foreground: "#111111",
			background: "#ffffff",
			values: map[Property]interface{}{
				SuccessIcon:     tc.SuccessIcon,
				ErrorIcon:       tc.ErrorIcon,
				DisplayExitCode: tc.DisplayExitCode,
				AlwaysNumeric:   tc.AlwaysNumeric,
			},
		}
		e := &exit{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.Expected, e.string())
	}
}

// Battery Segment

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
