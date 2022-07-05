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

type Info struct {
	Percentage int
	State      State
}

type NoBatteryError struct{}

func (m *NoBatteryError) Error() string {
	return "no battery"
}

// State type enumerates possible battery states.
type State int

var states = [...]string{
	Unknown:     "Unknown",
	Empty:       "Empty",
	Full:        "Full",
	Charging:    "Charging",
	Discharging: "Discharging",
	NotCharging: "Not Charging",
}

func (s State) String() string {
	return states[s]
}

// Possible state values.
// Unknown can mean either controller returned unknown, or
// not able to retrieve state due to some error.
const (
	Unknown State = iota
	Empty
	Full
	Charging
	Discharging
	NotCharging
)
