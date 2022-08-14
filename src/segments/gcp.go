package segments

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"path"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Gcp struct {
	props properties.Properties
	env   environment.Environment

	Account string
	Project string
	Region  string
}

func (g *Gcp) Template() string {
	return " {{ .Project }} "
}

func (g *Gcp) Init(props properties.Properties, env environment.Environment) {
	g.props = props
	g.env = env
}

func (g *Gcp) Enabled() bool {
	cfgDir := g.getConfigDirectory()
	configFile, err := g.getActiveConfig(cfgDir)
	if err != nil {
		return g.setError(err.Error())
	}

	cfgpath := path.Join(cfgDir, "configurations", "config_"+configFile)

	cfg, err := ini.Load(cfgpath)
	if err != nil {
		return g.setError("GCLOUD CONFIG ERROR")
	}

	g.Project = cfg.Section("core").Key("project").String()
	g.Account = cfg.Section("core").Key("account").String()
	g.Region = cfg.Section("compute").Key("region").String()

	return true
}

func (g *Gcp) getActiveConfig(cfgDir string) (string, error) {
	ap := path.Join(cfgDir, "active_config")
	absolutePath, err := filepath.Abs(ap)
	if err != nil {
		return "", err
	}

	fileContent := g.env.FileContent(absolutePath)
	if fileContent == "" {
		return "", errors.New("NO ACTIVE CONFIG")
	}
	return fileContent, nil
}

func (g *Gcp) getConfigDirectory() string {
	cfgDir := g.env.Getenv("CLOUDSDK_CONFIG")
	if cfgDir == "" {
		cfgDir = path.Join(g.env.Home(), ".config", "gcloud")
	}

	return cfgDir
}

func (g *Gcp) setError(message string) bool {
	displayError := g.props.GetBool(properties.DisplayError, false)
	if !displayError {
		return false
	}

	g.Project = message
	g.Account = message
	g.Region = message

	return true
}
