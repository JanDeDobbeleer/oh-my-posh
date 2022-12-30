package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type WindowsRegistry struct {
	props properties.Properties
	env   platform.Environment

	Value string
}

const (
	// full path to the key; if ends in \, gets "(Default)" key in that path
	RegistryPath properties.Property = "path"
	// Fallback is the text to display if the key is not found
	Fallback properties.Property = "fallback"
)

func (wr *WindowsRegistry) Template() string {
	return " {{ .Value }} "
}

func (wr *WindowsRegistry) Init(props properties.Properties, env platform.Environment) {
	wr.props = props
	wr.env = env
}

func (wr *WindowsRegistry) Enabled() bool {
	if wr.env.GOOS() != platform.WINDOWS {
		return false
	}

	registryPath := wr.props.GetString(RegistryPath, "")
	wr.Value = wr.props.GetString(Fallback, "")

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
