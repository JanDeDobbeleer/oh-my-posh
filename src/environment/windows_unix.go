//go:build !darwin

package environment

import (
	"math"
	"strings"
	"time"

	"github.com/distatus/battery"
)

func mapMostLogicalState(currentState, newState battery.State) battery.State {
	switch currentState {
	case battery.Discharging, battery.NotCharging:
		return battery.Discharging
	case battery.Empty:
		return newState
	case battery.Charging:
		if newState == battery.Discharging {
			return battery.Discharging
		}
		return battery.Charging
	case battery.Unknown:
		return newState
	case battery.Full:
		return newState
	}
	return newState
}

func (env *ShellEnvironment) BatteryState() (*BatteryInfo, error) {
	defer env.Trace(time.Now(), "BatteryInfo")

	parseBatteryInfo := func(batteries []*battery.Battery) *BatteryInfo {
		var info BatteryInfo
		var current, total float64
		var state battery.State
		for _, bt := range batteries {
			current += bt.Current
			total += bt.Full
			state = mapMostLogicalState(state, bt.State)
		}
		batteryPercentage := current / total * 100
		info.Percentage = int(math.Min(100, batteryPercentage))
		info.State = state
		return &info
	}

	batteries, err := battery.GetAll()
	// actual error, return it
	if err != nil && len(batteries) == 0 {
		env.log(Error, "BatteryInfo", err.Error())
		return nil, err
	}
	// there are no batteries found
	if len(batteries) == 0 {
		return nil, &NoBatteryError{}
	}
	// some batteries fail to get retrieved, filter them out if present
	validBatteries := []*battery.Battery{}
	for _, batt := range batteries {
		if batt != nil {
			validBatteries = append(validBatteries, batt)
		}
	}
	// clean minor errors
	unableToRetrieveBatteryInfo := "A device which does not exist was specified."
	unknownChargeRate := "Unknown value received"
	var fatalErr battery.Errors
	ignoreErr := func(err error) bool {
		if e, ok := err.(battery.ErrPartial); ok {
			// ignore unknown charge rate value error
			if e.Current == nil &&
				e.Design == nil &&
				e.DesignVoltage == nil &&
				e.Full == nil &&
				e.State == nil &&
				e.Voltage == nil &&
				e.ChargeRate != nil &&
				e.ChargeRate.Error() == unknownChargeRate {
				return true
			}
		}
		return false
	}
	if batErr, ok := err.(battery.Errors); ok {
		for _, err := range batErr {
			if !ignoreErr(err) {
				fatalErr = append(fatalErr, err)
			}
		}
	}

	// when battery info fails to get retrieved but there is at least one valid battery, return it without error
	if len(validBatteries) > 0 && fatalErr != nil && strings.Contains(fatalErr.Error(), unableToRetrieveBatteryInfo) {
		return parseBatteryInfo(validBatteries), nil
	}
	// another error occurred (possibly unmapped use-case), return it
	if fatalErr != nil {
		env.log(Error, "BatteryInfo", fatalErr.Error())
		return nil, fatalErr
	}
	// everything is fine
	return parseBatteryInfo(validBatteries), nil
}
