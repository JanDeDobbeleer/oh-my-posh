package main

import (
	"errors"
	"strconv"
	"strings"
)

type packagemanager struct {
	props                   properties
	env                     environmentInfo
	AvailablePackageUpdates string
	Error                   error
}

type PackageManagerType int

const (
	cacheKey string = "PackageManager_AvailablePackageUpdates"
)

const (
	Chocolatey PackageManagerType = iota
)

func (pmt PackageManagerType) string() string {
	switch pmt {
	case Chocolatey:
		return "chocolatey"
	}
	return "unknown"
}

func (pm *packagemanager) enabled() bool {
	availablePackageUpdates, error := pm.getAvailablePackageUpdates()
	pm.Error = error

	if error != nil {
		return false
	}

	pm.AvailablePackageUpdates = availablePackageUpdates
	return availablePackageUpdates != "0"
}

func (pm *packagemanager) string() string {
	segmentTemplate := pm.props.getString(SegmentTemplate, "\uF487 ↓{{.AvailablePackageUpdates}}")

	template := &textTemplate{
		Template: segmentTemplate,
		Context:  pm,
		Env:      pm.env,
	}

	text, err := template.render()

	if err != nil {
		return err.Error()
	}

	return text
}

func (pm *packagemanager) init(props properties, env environmentInfo) {
	pm.props = props
	pm.env = env
}

func (pm *packagemanager) getAvailablePackageUpdates() (string, error) {
	getCachedValue := func(key string) (string, error) {
		cachedValue, cachedValueFound := pm.env.cache().get(key)

		if cachedValueFound {
			return cachedValue, nil
		}

		return "", errors.New(key + " value not found in cache")
	}

	availablePackageUpdates, getCachedValueError := getCachedValue(cacheKey)

	if getCachedValueError != nil {
		var packageManagerError error

		switch pm.props.getString("PackageManagerType", "") {
		case Chocolatey.string():
			availablePackageUpdates, packageManagerError = pm.getAvailablePackageUpdatesFromChocolatey()
		default:
			return "", errors.New("Unknown package manager")
		}

		if packageManagerError != nil {
			return "", packageManagerError
		}

		_, conversionError := strconv.Atoi(availablePackageUpdates)
		if conversionError != nil {
			return "", errors.New("Available package updates value is non-numeric")
		}
	}

	pm.env.cache().set(cacheKey, availablePackageUpdates, 5)
	return availablePackageUpdates, nil
}

func (pm *packagemanager) getAvailablePackageUpdatesFromChocolatey() (string, error) {
	chocoOutdatedOutput := pm.env.runShellCommand("powershell", "choco outdated")

	chocoOutdatedOutputParts := strings.Split(chocoOutdatedOutput, "has determined ")
	if len(chocoOutdatedOutputParts) < 2 {
		return "", errors.New("Unexpected output from choco outdated")
	}

	return strings.Split(chocoOutdatedOutputParts[1], " ")[0], nil
}
