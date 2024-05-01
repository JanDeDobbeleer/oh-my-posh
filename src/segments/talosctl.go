package segments

import (
	"errors"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"gopkg.in/yaml.v3"
)

const (
	TALOSCTLNOACTIVECONFIG = "NO ACTIVE CONFIG FOUND"
)

type Talosctl struct {
	props properties.Properties
	env   platform.Environment

	Context string
}

type TalosctlData struct {
	ActiveContext string `yaml:"context"`
}

func (t *Talosctl) Template() string {
	return " {{ .Context}} "
}

func (t *Talosctl) Init(props properties.Properties, env platform.Environment) {
	t.props = props
	t.env = env
}

func (t *Talosctl) Enabled() bool {
	cfgDir := filepath.Join(t.env.Home(), ".talos")
	configFile, err := t.getActiveConfig(cfgDir)
	if err != nil {
		t.env.Error(err)
		return false
	}

	data, err := t.getTalosctlData(configFile)
	if err != nil {
		t.env.Error(err)
		return false
	}

	if len(data.ActiveContext) == 0 {
		return false
	}

	t.Context = data.ActiveContext
	return true
}

func (t *Talosctl) getActiveConfig(cfgDir string) (string, error) {
	activeConfigFile := filepath.Join(cfgDir, "config")
	activeConfigData := t.env.FileContent(activeConfigFile)
	if len(activeConfigData) == 0 {
		return "", errors.New(TALOSCTLNOACTIVECONFIG)
	}
	return activeConfigData, nil
}

func (t *Talosctl) getTalosctlData(configFile string) (*TalosctlData, error) {
	var data TalosctlData

	err := yaml.Unmarshal([]byte(configFile), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
