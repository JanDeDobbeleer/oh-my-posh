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
	"fmt"
	"io/ioutil" //nolint:staticcheck,nolintlint
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const sysfs = "/sys/class/power_supply"

func newState(name string) (State, error) {
	for i, state := range states {
		if strings.EqualFold(name, state) {
			return State(i), nil
		}
	}
	return Unknown, fmt.Errorf("Invalid state `%s`", name)
}

func readFloat(path, filename string) (float64, error) {
	str, err := ioutil.ReadFile(filepath.Join(path, filename))
	if err != nil {
		return 0, err
	}
	if len(str) == 0 {
		return 0, ErrNotFound
	}
	num, err := strconv.ParseFloat(string(str[:len(str)-1]), 64)
	if err != nil {
		return 0, err
	}
	return num / 1000, nil // Convert micro->milli
}

func readAmp(path, filename string, volts float64) (float64, error) {
	val, err := readFloat(path, filename)
	if err != nil {
		return 0, err
	}
	return val * volts, nil
}

func isBattery(path string) bool {
	t, err := ioutil.ReadFile(filepath.Join(path, "type"))
	return err == nil && string(t) == "Battery\n"
}

func getBatteryFiles() ([]string, error) {
	files, err := ioutil.ReadDir(sysfs)
	if err != nil {
		return nil, err
	}

	var bFiles []string
	for _, file := range files {
		path := filepath.Join(sysfs, file.Name())
		if isBattery(path) {
			bFiles = append(bFiles, path)
		}
	}

	if len(bFiles) == 0 {
		return nil, &NoBatteryError{}
	}

	return bFiles, nil
}

func getByPath(path string) (*battery, error) {
	b := &battery{}
	var err error

	if b.Current, err = readFloat(path, "energy_now"); err == nil {
		if b.Full, err = readFloat(path, "energy_full"); err != nil {
			return nil, errors.New("unable to parse energy_full")
		}
	} else {
		currentDoesNotExist := os.IsNotExist(err)
		if b.Voltage, err = readFloat(path, "voltage_now"); err != nil {
			return nil, errors.New("unable to parse voltage_now")
		}
		b.Voltage /= 1000
		if currentDoesNotExist {
			if b.Current, err = readAmp(path, "charge_now", b.Voltage); err != nil {
				return nil, errors.New("unable to parse charge_now")
			}
			if b.Full, err = readAmp(path, "charge_full", b.Voltage); err != nil {
				return nil, errors.New("unable to parse charge_full")
			}
		} else {
			if b.Full, err = readFloat(path, "energy_full"); err != nil {
				return nil, errors.New("unable to parse energy_full")
			}
		}
	}

	state, err := ioutil.ReadFile(filepath.Join(path, "status"))
	if err != nil || len(state) == 0 {
		return nil, errors.New("unable to parse or invalid status")
	}
	if b.State, err = newState(string(state[:len(state)-1])); err != nil {
		return nil, errors.New("unable to map to new state")
	}

	return b, nil
}

func systemGetAll() ([]*battery, error) {
	bFiles, err := getBatteryFiles()
	if err != nil {
		return nil, err
	}

	var batteries []*battery
	var errs Errors

	for _, bFile := range bFiles {
		b, err := getByPath(bFile)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		batteries = append(batteries, b)
	}

	if len(batteries) == 0 {
		return nil, errs
	}

	return batteries, nil
}
