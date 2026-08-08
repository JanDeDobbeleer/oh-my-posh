package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestBatteryEnabled(t *testing.T) {
	cases := []struct {
		Case            string
		Info            *battery.Info
		Error           error
		DisplayError    bool
		ExpectedEnabled bool
		Expected        string
	}{
		{
			Case:            "detected battery at 0%",
			Info:            &battery.Info{Percentage: 0, State: battery.NotCharging},
			ExpectedEnabled: true,
			Expected:        "0",
		},
		{
			Case:            "discharging",
			Info:            &battery.Info{Percentage: 42, State: battery.Discharging},
			ExpectedEnabled: true,
			Expected:        "42",
		},
		{
			Case:            "full",
			Info:            &battery.Info{Percentage: 100, State: battery.Full},
			ExpectedEnabled: true,
			Expected:        "100",
		},
		{
			Case:            "no battery present",
			Info:            nil,
			Error:           &battery.NoBatteryError{},
			ExpectedEnabled: false,
		},
		{
			Case:            "retrieval error, display_error disabled",
			Info:            nil,
			Error:           errors.New("boom"),
			ExpectedEnabled: false,
		},
		{
			Case:            "retrieval error, display_error enabled",
			Info:            nil,
			Error:           errors.New("boom"),
			DisplayError:    true,
			ExpectedEnabled: true,
			Expected:        "boom",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)
		env.On("BatteryState").Return(tc.Info, tc.Error)

		props := options.Map{}
		if tc.DisplayError {
			props[options.DisplayError] = true
		}

		b := &Battery{}
		b.Init(props, env)

		got := b.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, got, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, b.Template(), b), tc.Case)
		}
	}
}
