package main

import (
	"errors"
	"fmt"
)

type winreg struct {
	props properties
	env   environmentInfo

	Value string
}

const (
	// full path to the key; if ends in \, gets "(Default)" key in that path
	RegistryPath Property = "path"
	// Fallback is the text to display if the key is not found
	Fallback Property = "fallback"
)

func (wr *winreg) init(props properties, env environmentInfo) {
	wr.props = props
	wr.env = env
}

func (wr *winreg) enabled() bool {
	if wr.env.getRuntimeGOOS() != windowsPlatform {
		return false
	}

	registryPath := wr.props.getString(RegistryPath, "")
	fallback := wr.props.getString(Fallback, "")

	var regValue *windowsRegistryValue
	regValue, _ = wr.env.getWindowsRegistryKeyValue(registryPath)

	if regValue != nil {
		switch regValue.valueType {
		case regString:
			wr.Value = regValue.str
			return true
		case regDword:
			wr.Value = fmt.Sprintf("0x%08X", regValue.dword)
			return true
		case regQword:
			wr.Value = fmt.Sprintf("0x%016X", regValue.qword)
			return true
		}
	}

	if len(fallback) > 0 {
		wr.Value = fallback
		return true
	}

	return false
}

func (wr *winreg) string() string {
	segmentTemplate := wr.props.getString(SegmentTemplate, "{{ .Value }}")
	return wr.templateString(segmentTemplate)
}

func (wr *winreg) templateString(segmentTemplate string) string {
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  wr,
		Env:      wr.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (wr winreg) GetRegistryString(path string) (string, error) {
	regValue, err := wr.env.getWindowsRegistryKeyValue(path)

	if regValue != nil {
		if regValue.valueType == regString {
			return regValue.str, nil
		}

		return "", errors.New("type mismatch, registry value is not a string")
	}

	return "", err
}

func (wr winreg) GetRegistryDword(path string) (uint32, error) {
	regValue, err := wr.env.getWindowsRegistryKeyValue(path)

	if regValue != nil {
		if regValue.valueType == regDword {
			return regValue.dword, nil
		}

		return 0, errors.New("type mismatch, registry value is not a dword or qword")
	}

	return 0, err
}

func (wr winreg) GetRegistryQword(path string) (uint64, error) {
	regValue, err := wr.env.getWindowsRegistryKeyValue(path)

	if regValue != nil {
		if regValue.valueType == regQword {
			return regValue.qword, nil
		}

		return 0, errors.New("type mismatch, registry value is not a qword")
	}

	return 0, err
}
