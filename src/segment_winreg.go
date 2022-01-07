package main

import (
	"errors"
	"fmt"
)

type winreg struct {
	props Properties
	env   Environment

	Value string
}

const (
	// full path to the key; if ends in \, gets "(Default)" key in that path
	RegistryPath Property = "path"
	// Fallback is the text to display if the key is not found
	Fallback Property = "fallback"
)

func (wr *winreg) init(props Properties, env Environment) {
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

	if regValue == nil {
		return "", err
	}

	if regValue.valueType != regString {
		return "", errors.New("type mismatch, registry value is not a string")
	}

	return regValue.str, nil
}

func (wr winreg) GetRegistryDword(path string) (uint32, error) {
	regValue, err := wr.env.getWindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.valueType != regDword {
		return 0, errors.New("type mismatch, registry value is not a dword")
	}

	return regValue.dword, nil
}

func (wr winreg) GetRegistryQword(path string) (uint64, error) {
	regValue, err := wr.env.getWindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.valueType != regQword {
		return 0, errors.New("type mismatch, registry value is not a qword")
	}

	return regValue.qword, nil
}
