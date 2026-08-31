package segments

import (
	"encoding/json"
	"errors"
	"path"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/jandedobbeleer/oh-my-posh/src/ini"
)

const (
	GCPNOACTIVECONFIG = "NO ACTIVE CONFIG FOUND"

	gcpAuthStatusAuthorized   = "authorized"
	gcpAuthStatusUnauthorized = "unauthorized"
	gcpAuthStatusUnknown      = "unknown"
	gcpCommand                = "gcloud"
)

var gcpAuthFields = []string{"Authorized", "AuthStatus", "TokenExpiry", "AuthError"}

type Gcp struct {
	Base
	Account      string
	Project      string
	Region       string
	ActiveConfig string
	TokenExpiry  string
	AuthError    string
	AuthStatus   string
	FieldRefs
	Authorized bool
}

type gcpConfigHelper struct {
	Credential struct {
		AccessToken string `json:"access_token"`
		TokenExpiry string `json:"token_expiry"`
	} `json:"credential"`
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

	g.ActiveConfig = cfgName
	cfgPath := path.Join(cfgDir, "configurations", "config_"+cfgName)
	cfg := g.env.FileContent(cfgPath)

	if cfg == "" {
		log.Error(errors.New("config file is empty"))
		return false
	}

	data, err := ini.Load(cfg)
	if err != nil {
		log.Error(err)
		return false
	}

	g.Project = data.Section("core").Key("project").String()
	g.Account = data.Section("core").Key("account").String()
	g.Region = data.Section("compute").Key("region").String()

	if !g.fetchUnit(gcpAuthFields...) {
		return true
	}

	g.loadAuthStatus()
	return true
}

func (g *Gcp) getActiveConfig(cfgDir string) (string, error) {
	activeCfg := g.env.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")
	if len(activeCfg) != 0 {
		return activeCfg, nil
	}

	ap := path.Join(cfgDir, "active_config")
	activeCfg = g.env.FileContent(ap)
	if activeCfg == "" {
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

func (g *Gcp) loadAuthStatus() {
	g.AuthStatus = gcpAuthStatusUnknown

	if !g.env.HasCommand(gcpCommand) {
		return
	}

	output, err := g.env.RunCommand(gcpCommand, "config", "config-helper", "--format=json")
	if err != nil {
		g.AuthStatus = gcpAuthStatusUnauthorized
		g.AuthError = err.Error()
		return
	}

	if output == "" {
		g.AuthStatus = gcpAuthStatusUnauthorized
		g.AuthError = "empty auth response"
		return
	}

	var helper gcpConfigHelper
	err = json.Unmarshal([]byte(output), &helper)
	if err != nil {
		log.Error(err)
		g.AuthError = err.Error()
		return
	}

	if helper.Credential.AccessToken == "" {
		g.AuthStatus = gcpAuthStatusUnauthorized
		return
	}

	g.Authorized = true
	g.AuthStatus = gcpAuthStatusAuthorized
	g.TokenExpiry = helper.Credential.TokenExpiry
}
