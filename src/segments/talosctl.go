package segments

import (
	"errors"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"gopkg.in/yaml.v3"
)

type TalosCTL struct {
	props properties.Properties
	env   platform.Environment

	Context string `yaml:"context"`
}

func (t *TalosCTL) Template() string {
	return " {{ .Context}} "
}

func (t *TalosCTL) Init(props properties.Properties, env platform.Environment) {
	t.props = props
	t.env = env
}

func (t *TalosCTL) Enabled() bool {
	cfgDir := filepath.Join(t.env.Home(), ".talos")
	configFile, err := t.getActiveConfig(cfgDir)
	if err != nil {
		t.env.Error(err)
		return false
	}

	err = yaml.Unmarshal([]byte(configFile), t)
	if err != nil {
		t.env.Error(err)
		return false
	}

	if len(t.Context) == 0 {
		return false
	}

	return true
}

func (t *TalosCTL) getActiveConfig(cfgDir string) (string, error) {
	activeConfigFile := filepath.Join(cfgDir, "config")
	activeConfigData := t.env.FileContent(activeConfigFile)
	if len(activeConfigData) == 0 {
		return "", errors.New("NO ACTIVE CONFIG FOUND")
	}
	return activeConfigData, nil
}
