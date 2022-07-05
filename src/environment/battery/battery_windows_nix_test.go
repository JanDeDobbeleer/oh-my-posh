//go:build !darwin

// battery
// Copyright (C) 2016-2017 Karol 'Kenji Takahashi' Woźniak
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package battery

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestMapBatteriesState(t *testing.T) {
	cases := []struct {
		Case          string
		ExpectedState State
		CurrentState  State
		NewState      State
	}{
		{Case: "charging > charged", ExpectedState: Charging, CurrentState: Full, NewState: Charging},
		{Case: "charging < discharging", ExpectedState: Discharging, CurrentState: Discharging, NewState: Charging},
		{Case: "charging == charging", ExpectedState: Charging, CurrentState: Charging, NewState: Charging},
		{Case: "discharging > charged", ExpectedState: Discharging, CurrentState: Full, NewState: Discharging},
		{Case: "discharging > unknown", ExpectedState: Discharging, CurrentState: Unknown, NewState: Discharging},
		{Case: "discharging > full", ExpectedState: Discharging, CurrentState: Full, NewState: Discharging},
		{Case: "discharging > charging 2", ExpectedState: Discharging, CurrentState: Charging, NewState: Discharging},
		{Case: "discharging > empty", ExpectedState: Discharging, CurrentState: Empty, NewState: Discharging},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.ExpectedState, mapMostLogicalState(tc.CurrentState, tc.NewState), tc.Case)
	}
}
