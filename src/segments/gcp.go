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
	cfgName, err := g.getActiveConfig(cfgDir)
	if err != nil {
		log.Error(err)
		return false
	}

	cfgPath := path.Join(cfgDir, "configurations", "config_"+cfgName)
	cfg := g.env.FileContent(cfgPath)

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
	activeCfg := g.env.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")
	if len(activeCfg) != 0 {
		return activeCfg, nil
	}

	ap := path.Join(cfgDir, "active_config")
	activeCfg = g.env.FileContent(ap)
	if len(activeCfg) == 0 {
		return "", errors.New(GCPNOACTIVECONFIG)
	}

	return activeCfg, nil
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
