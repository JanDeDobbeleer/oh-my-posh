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

type batteryQueryInformation struct {
	BatteryTag       uint32
	InformationLevel int32
	AtRate           int32
}

type batteryInformation struct {
	Capabilities        uint32
	Technology          uint8
	Reserved            [3]uint8
	Chemistry           [4]uint8
	DesignedCapacity    uint32
	FullChargedCapacity uint32
	DefaultAlert1       uint32
	DefaultAlert2       uint32
	CriticalBias        uint32
	CycleCount          uint32
}

type batteryWaitStatus struct {
	BatteryTag   uint32
	Timeout      uint32
	PowerState   uint32
	LowCapacity  uint32
	HighCapacity uint32
}

type batteryStatus struct {
	PowerState uint32
	Capacity   uint32
	Voltage    uint32
	Rate       int32
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type spDeviceInterfaceData struct {
	cbSize             uint32
	InterfaceClassGuid guid //nolint:revive
	Flags              uint32
	Reserved           uint
}

var guidDeviceBattery = guid{
	0x72631e54,
	0x78A4,
	0x11d0,
	[8]byte{0xbc, 0xf7, 0x00, 0xaa, 0x00, 0xb7, 0xb3, 0x2a},
}

func uint32ToFloat64(num uint32) (float64, error) {
	if num == 0xffffffff { // BATTERY_UNKNOWN_CAPACITY
		return 0, errors.New("Unknown value received")
	}
	return float64(num), nil
}

func setupDiSetup(proc *windows.LazyProc, nargs, a1, a2, a3, a4, a5, a6 uintptr) (uintptr, error) {
	r1, _, errno := syscall.Syscall6(proc.Addr(), nargs, a1, a2, a3, a4, a5, a6) //nolint:staticcheck
	if windows.Handle(r1) == windows.InvalidHandle {
		if errno != 0 {
			return 0, error(errno)
		}
		return 0, syscall.EINVAL
	}
	return r1, nil
}

func setupDiCall(proc *windows.LazyProc, nargs, a1, a2, a3, a4, a5, a6 uintptr) syscall.Errno {
	r1, _, errno := syscall.Syscall6(proc.Addr(), nargs, a1, a2, a3, a4, a5, a6) //nolint:staticcheck
	if r1 == 0 {
		if errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}
	return 0
}

var setupapi = &windows.LazyDLL{Name: "setupapi.dll", System: true}
var setupDiGetClassDevsW = setupapi.NewProc("SetupDiGetClassDevsW")
var setupDiEnumDeviceInterfaces = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
var setupDiGetDeviceInterfaceDetailW = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
var setupDiDestroyDeviceInfoList = setupapi.NewProc("SetupDiDestroyDeviceInfoList")

func readState(powerState uint32) State {
	switch {
	case powerState&0x00000004 != 0:
		return Charging
	case powerState&0x00000008 != 0:
		return Empty
	case powerState&0x00000002 != 0:
		return Discharging
	case powerState&0x00000001 != 0:
		return Full
	default:
		return Unknown
	}
}

func systemGet(idx int) (*battery, error) {
	hdev, err := setupDiSetup(
		setupDiGetClassDevsW,
		4,
		uintptr(unsafe.Pointer(&guidDeviceBattery)),
		0,
		0,
		2|16, // DIGCF_PRESENT|DIGCF_DEVICEINTERFACE
		0, 0,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _, _ = syscall.Syscall(setupDiDestroyDeviceInfoList.Addr(), 1, hdev, 0, 0) //nolint:staticcheck
	}()

	var did spDeviceInterfaceData
	did.cbSize = uint32(unsafe.Sizeof(did))
	errno := setupDiCall(
		setupDiEnumDeviceInterfaces,
		5,
		hdev,
		0,
		uintptr(unsafe.Pointer(&guidDeviceBattery)),
		uintptr(idx),
		uintptr(unsafe.Pointer(&did)),
		0,
	)
	if errno == 259 { // ERROR_NO_MORE_ITEMS
		return nil, ErrNotFound
	}
	if errno != 0 {
		return nil, errno
	}
	var cbRequired uint32
	errno = setupDiCall(
		setupDiGetDeviceInterfaceDetailW,
		6,
		hdev,
		uintptr(unsafe.Pointer(&did)),
		0,
		0,
		uintptr(unsafe.Pointer(&cbRequired)),
		0,
	)
	if errno != 0 && errno != 122 { // ERROR_INSUFFICIENT_BUFFER
		return nil, errno
	}
	// The god damn struct with ANYSIZE_ARRAY of utf16 in it is crazy.
	// So... let's emulate it with array of uint16 ;-D.
	// Keep in mind that the first two elements are actually cbSize.
	didd := make([]uint16, cbRequired/2)
	cbSize := (*uint32)(unsafe.Pointer(&didd[0]))
	if unsafe.Sizeof(uint(0)) == 8 {
		*cbSize = 8
	} else {
		*cbSize = 6
	}
	errno = setupDiCall(
		setupDiGetDeviceInterfaceDetailW,
		6,
		hdev,
		uintptr(unsafe.Pointer(&did)),
		uintptr(unsafe.Pointer(&didd[0])),
		uintptr(cbRequired),
		uintptr(unsafe.Pointer(&cbRequired)),
		0,
	)
	if errno != 0 {
		return nil, errno
	}
	devicePath := &didd[2:][0]

	handle, err := windows.CreateFile(
		devicePath,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	var dwOut uint32

	var dwWait uint32
	var bqi batteryQueryInformation
	err = windows.DeviceIoControl(
		handle,
		2703424, // IOCTL_BATTERY_QUERY_TAG
		(*byte)(unsafe.Pointer(&dwWait)),
		uint32(unsafe.Sizeof(dwWait)),
		(*byte)(unsafe.Pointer(&bqi.BatteryTag)),
		uint32(unsafe.Sizeof(bqi.BatteryTag)),
		&dwOut,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if bqi.BatteryTag == 0 {
		return nil, errors.New("BatteryTag not returned")
	}

	b := &battery{}

	var bi batteryInformation
	err = windows.DeviceIoControl(
		handle,
		2703428, // IOCTL_BATTERY_QUERY_INFORMATION
		(*byte)(unsafe.Pointer(&bqi)),
		uint32(unsafe.Sizeof(bqi)),
		(*byte)(unsafe.Pointer(&bi)),
		uint32(unsafe.Sizeof(bi)),
		&dwOut,
		nil,
	)
	if err != nil {
		return nil, err
	}

	b.Full = float64(bi.FullChargedCapacity)

	bws := batteryWaitStatus{BatteryTag: bqi.BatteryTag}
	var bs batteryStatus
	err = windows.DeviceIoControl(
		handle,
		2703436, // IOCTL_BATTERY_QUERY_STATUS
		(*byte)(unsafe.Pointer(&bws)),
		uint32(unsafe.Sizeof(bws)),
		(*byte)(unsafe.Pointer(&bs)),
		uint32(unsafe.Sizeof(bs)),
		&dwOut,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if b.Current, err = uint32ToFloat64(bs.Capacity); err != nil {
		return nil, err
	}
	if b.Voltage, err = uint32ToFloat64(bs.Voltage); err != nil {
		return nil, err
	}
	b.Voltage /= 1000
	b.State = readState(bs.PowerState)

	return b, nil
}

func systemGetAll() ([]*battery, error) {
	var batteries []*battery
	var i int
	var errs Errors

	for i = 0; ; i++ {
		b, err := systemGet(i)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		batteries = append(batteries, b)
	}

	if i == 0 {
		return nil, &NoBatteryError{}
	}

	if len(batteries) == 0 {
		return nil, errs
	}

	return batteries, nil
}
