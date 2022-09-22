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
	Error   string
}

func (g *Gcp) Template() string {
	return " {{ if not .Error }}{{ .Project }}{{ end }} "
}

func (g *Gcp) Init(props properties.Properties, env environment.Environment) {
	g.props = props
	g.env = env
}

func (g *Gcp) Enabled() bool {
	cfgDir := g.getConfigDirectory()
	configFile, err := g.getActiveConfig(cfgDir)
	if err != nil {
		g.Error = err.Error()
		return true
	}

	cfgpath := path.Join(cfgDir, "configurations", "config_"+configFile)

	cfg, err := ini.Load(cfgpath)
	if err != nil {
		g.Error = "GCLOUD CONFIG ERROR"
		return true
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
	if len(fileContent) == 0 {
		return "", errors.New("NO ACTIVE CONFIG FOUND")
	}
	return fileContent, nil
}

func (g *Gcp) getConfigDirectory() string {
	cfgDir := g.env.Getenv("CLOUDSDK_CONFIG")
	if len(cfgDir) == 0 {
		cfgDir = path.Join(g.env.Home(), ".config", "gcloud")
	}

	return cfgDir
}
