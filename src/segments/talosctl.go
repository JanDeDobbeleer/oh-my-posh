package segments

import (
	"errors"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	yaml "go.yaml.in/yaml/v3"
)

type TalosCTL struct {
	Base

	Context string `yaml:"context"`
}

func (t *TalosCTL) Template() string {
	return " {{ .Context}} "
}

func (t *TalosCTL) Enabled() bool {
	cfgDir := filepath.Join(t.env.Home(), ".talos")
	configFile, err := t.getActiveConfig(cfgDir)
	if err != nil {
		log.Error(err)
		return false
	}

	err = yaml.Unmarshal([]byte(configFile), t)
	if err != nil {
		log.Error(err)
		return false
	}

	if t.Context == "" {
		return false
	}

	return true
}

func (t *TalosCTL) getActiveConfig(cfgDir string) (string, error) {
	activeConfigFile := filepath.Join(cfgDir, "config")
	activeConfigData := t.env.FileContent(activeConfigFile)
	if activeConfigData == "" {
		return "", errors.New("no active config found")
	}
	return activeConfigData, nil
}
