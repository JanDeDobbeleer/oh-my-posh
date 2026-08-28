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
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// systemPowerStatus mirrors the Win32 SYSTEM_POWER_STATUS struct. GetSystemPowerStatus
// reads a value the OS already maintains, so unlike the battery miniclass IOCTLs
// (IOCTL_BATTERY_QUERY_TAG/INFORMATION/STATUS) it never round-trips to the embedded
// controller, which is what made the previous SetupAPI-based implementation slow.
type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

const (
	batteryFlagCharging  = 0x08
	batteryFlagCritical  = 0x04
	batteryFlagNoBattery = 0x80
	batteryFlagUnknown   = 0xFF

	batteryLifePercentUnknown = 0xFF

	acLineOnline = 1
)

var kernel32 = &windows.LazyDLL{Name: "kernel32.dll", System: true}
var getSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")

func systemGetAll() ([]*battery, error) {
	var sps systemPowerStatus

	r1, _, errno := syscall.SyscallN(getSystemPowerStatus.Addr(), uintptr(unsafe.Pointer(&sps)))
	if r1 == 0 {
		if errno != 0 {
			return nil, error(errno)
		}
		return nil, syscall.EINVAL
	}

	if sps.BatteryFlag&batteryFlagNoBattery != 0 || sps.BatteryFlag == batteryFlagUnknown {
		return nil, &NoBatteryError{}
	}

	if sps.BatteryLifePercent == batteryLifePercentUnknown {
		return nil, errors.New("unknown value received")
	}

	b := &battery{Full: 100, Current: float64(min(sps.BatteryLifePercent, 100))}

	switch {
	case sps.BatteryFlag&batteryFlagCharging != 0:
		b.State = Charging
	case sps.BatteryLifePercent == 100 && sps.ACLineStatus == acLineOnline:
		b.State = Full
	case sps.ACLineStatus == acLineOnline:
		b.State = NotCharging
	case sps.BatteryFlag&batteryFlagCritical != 0:
		b.State = Empty
	default:
		b.State = Discharging
	}

	return []*battery{b}, nil
}
