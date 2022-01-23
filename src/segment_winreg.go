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

func (wr *winreg) template() string {
	return "{{ .Value }}"
}

func (wr *winreg) init(props Properties, env Environment) {
	wr.props = props
	wr.env = env
}

func (wr *winreg) enabled() bool {
	if wr.env.GOOS() != windowsPlatform {
		return false
	}

	registryPath := wr.props.getString(RegistryPath, "")
	fallback := wr.props.getString(Fallback, "")

	var regValue *WindowsRegistryValue
	regValue, _ = wr.env.WindowsRegistryKeyValue(registryPath)

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

func (wr winreg) GetRegistryString(path string) (string, error) {
	regValue, err := wr.env.WindowsRegistryKeyValue(path)

	if regValue == nil {
		return "", err
	}

	if regValue.valueType != regString {
		return "", errors.New("type mismatch, registry value is not a string")
	}

	return regValue.str, nil
}

func (wr winreg) GetRegistryDword(path string) (uint32, error) {
	regValue, err := wr.env.WindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.valueType != regDword {
		return 0, errors.New("type mismatch, registry value is not a dword")
	}

	return regValue.dword, nil
}

func (wr winreg) GetRegistryQword(path string) (uint64, error) {
	regValue, err := wr.env.WindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.valueType != regQword {
		return 0, errors.New("type mismatch, registry value is not a qword")
	}

	return regValue.qword, nil
}
