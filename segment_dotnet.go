package main

import (
	"errors"
)

type dotnet struct {
	props              *properties
	env                environmentInfo
	activeVersion      string
	unsupportedVersion bool
}

const (
	// UnsupportedDotnetVersionIcon is displayed when the dotnet version in
	// the current folder isn't supported by the installed dotnet SDK set.
	UnsupportedDotnetVersionIcon Property = "unsupported_version_icon"
)

func (d *dotnet) string() string {
	if d.unsupportedVersion {
		return d.props.getString(UnsupportedDotnetVersionIcon, "\u2327")
	}

	if d.props.getBool(DisplayVersion, true) {
		return d.activeVersion
	}

	return ""
}

func (d *dotnet) init(props *properties, env environmentInfo) {
	d.props = props
	d.env = env
}

func (d *dotnet) enabled() bool {
	if !d.env.hasCommand("dotnet") {
		return false
	}

	output, err := d.env.runCommand("dotnet", "--version")
	if err == nil {
		d.activeVersion = output
		return true
	}

	// Exit code 145 is a special indicator that dotnet
	// ran, but the current project config settings specify
	// use of an SDK that isn't installed.
	var exerr *commandError
	if errors.As(err, &exerr) && exerr.exitCode == 145 {
		d.unsupportedVersion = true
		return true
	}

	return false
}
