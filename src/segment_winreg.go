package main

import (
	"errors"
	"fmt"
	"oh-my-posh/environment"
)

type winreg struct {
	props Properties
	env   environment.Environment

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

func (wr *winreg) init(props Properties, env environment.Environment) {
	wr.props = props
	wr.env = env
}

func (wr *winreg) enabled() bool {
	if wr.env.GOOS() != environment.WindowsPlatform {
		return false
	}

	registryPath := wr.props.getString(RegistryPath, "")
	fallback := wr.props.getString(Fallback, "")

	var regValue *environment.WindowsRegistryValue
	regValue, _ = wr.env.WindowsRegistryKeyValue(registryPath)

	if regValue != nil {
		switch regValue.ValueType {
		case environment.RegString:
			wr.Value = regValue.Str
			return true
		case environment.RegDword:
			wr.Value = fmt.Sprintf("0x%08X", regValue.Dword)
			return true
		case environment.RegQword:
			wr.Value = fmt.Sprintf("0x%016X", regValue.Qword)
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

	if regValue.ValueType != environment.RegString {
		return "", errors.New("type mismatch, registry value is not a string")
	}

	return regValue.Str, nil
}

func (wr winreg) GetRegistryDword(path string) (uint32, error) {
	regValue, err := wr.env.WindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.ValueType != environment.RegDword {
		return 0, errors.New("type mismatch, registry value is not a dword")
	}

	return regValue.Dword, nil
}

func (wr winreg) GetRegistryQword(path string) (uint64, error) {
	regValue, err := wr.env.WindowsRegistryKeyValue(path)

	if regValue == nil {
		return 0, err
	}

	if regValue.ValueType != environment.RegQword {
		return 0, errors.New("type mismatch, registry value is not a qword")
	}

	return regValue.Qword, nil
}
