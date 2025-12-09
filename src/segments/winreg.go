package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type WindowsRegistry struct {
	Base

	Value string
}

const (
	// full path to the key; if ends in \, gets "(Default)" key in that path
	RegistryPath options.Option = "path"
	// Fallback is the text to display if the key is not found
	Fallback options.Option = "fallback"
)

func (wr *WindowsRegistry) Template() string {
	return " {{ .Value }} "
}

func (wr *WindowsRegistry) Enabled() bool {
	if wr.env.GOOS() != runtime.WINDOWS {
		return false
	}

	registryPath := wr.options.String(RegistryPath, "")
	wr.Value = wr.options.String(Fallback, "")

	regValue, err := wr.env.WindowsRegistryKeyValue(registryPath)
	if err == nil {
		wr.Value = regValue.String
		return true
	}
	if len(wr.Value) > 0 {
		// we have fallback value
		return true
	}
	return false
}
