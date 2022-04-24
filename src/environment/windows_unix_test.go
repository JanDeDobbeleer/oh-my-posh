//go:build !darwin

package environment

import (
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, tc.ExpectedState, mapMostLogicalState(tc.CurrentState, tc.NewState), tc.Case)
	}
}
