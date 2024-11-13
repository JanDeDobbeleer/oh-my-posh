package segments

import (
	"errors"
	"path"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"gopkg.in/ini.v1"
)

const (
	GCPNOACTIVECONFIG = "NO ACTIVE CONFIG FOUND"
)

type Gcp struct {
	base

	Account string
	Project string
	Region  string
}

func (g *Gcp) Template() string {
	return " {{ .Project }} "
}

func (g *Gcp) Enabled() bool {
	cfgDir := g.getConfigDirectory()
	configFile, err := g.getActiveConfig(cfgDir)
	if err != nil {
		log.Error(err)
		return false
	}

	cfgpath := path.Join(cfgDir, "configurations", "config_"+configFile)
	cfg := g.env.FileContent(cfgpath)

	if len(cfg) == 0 {
		log.Error(errors.New("config file is empty"))
		return false
	}

	data, err := ini.Load([]byte(cfg))
	if err != nil {
		log.Error(err)
		return false
	}

	g.Project = data.Section("core").Key("project").String()
	g.Account = data.Section("core").Key("account").String()
	g.Region = data.Section("compute").Key("region").String()

	return true
}

func (g *Gcp) getActiveConfig(cfgDir string) (string, error) {
	ap := path.Join(cfgDir, "active_config")
	fileContent := g.env.FileContent(ap)
	if len(fileContent) == 0 {
		return "", errors.New(GCPNOACTIVECONFIG)
	}
	return fileContent, nil
}

func (g *Gcp) getConfigDirectory() string {
	cfgDir := g.env.Getenv("CLOUDSDK_CONFIG")
	if len(cfgDir) != 0 {
		return cfgDir
	}

	if g.env.GOOS() == runtime.WINDOWS {
		return path.Join(g.env.Getenv("APPDATA"), "gcloud")
	}

	return path.Join(g.env.Home(), ".config", "gcloud")
}
