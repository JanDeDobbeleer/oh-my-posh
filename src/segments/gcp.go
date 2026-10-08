package segments

import (
	"errors"
	"path"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/sqlite"

	"github.com/jandedobbeleer/oh-my-posh/src/ini"
)

const (
	GCPNOACTIVECONFIG = "NO ACTIVE CONFIG FOUND"

	// FetchAuthStatus, when enabled, reads access_tokens.db to populate
	// Authorized and TokenExpiresAt. Off by default: it's an extra file read
	// of a credentials cache that most users don't need.
	FetchAuthStatus options.Option = "fetch_auth_status"

	accessTokensDB = "access_tokens.db"
)

type Gcp struct {
	Base

	// TokenExpiresAt is the zero time when the cached access token's expiry
	// is unknown (feature disabled, or no cached token for Account - e.g.
	// service account/workload identity logins).
	TokenExpiresAt time.Time

	Account      string
	Project      string
	Region       string
	ActiveConfig string

	// Authorized is true unless a cached access token for Account was found
	// and confirmed expired; unknown status is treated as authorized.
	Authorized bool
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

	if g.options.Bool(FetchAuthStatus, false) {
		g.loadAuthStatus(cfgDir)
	}

	return true
}

// loadAuthStatus reads the cached access token's expiry for Account from
// access_tokens.db. Org "session length" reauth policies revoke the local
// session through an opaque RAPT token with no readable expiry, so this can't
// predict the exact reauth moment - it reports whether the cached access
// token itself is still valid, which stops refreshing once reauth is needed.
func (g *Gcp) loadAuthStatus(cfgDir string) {
	g.Authorized = true

	if g.Account == "" {
		return
	}

	content := g.env.FileContent(path.Join(cfgDir, accessTokensDB))
	if content == "" {
		return
	}

	row, found, err := sqlite.FindRow([]byte(content), "access_tokens", "account_id", g.Account)
	if err != nil {
		log.Error(err)
		return
	}

	if !found {
		return
	}

	expiresAt, err := parseTokenExpiry(row["token_expiry"])
	if err != nil {
		log.Error(err)
		return
	}

	g.TokenExpiresAt = expiresAt
	g.Authorized = time.Now().Before(expiresAt)
}

// parseTokenExpiry parses gcloud's token_expiry column, a naive UTC
// timestamp written by Python's sqlite3 datetime adapter, with or without
// a microseconds component.
func parseTokenExpiry(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty token expiry")
	}

	if t, err := time.ParseInLocation("2006-01-02 15:04:05.999999", value, time.UTC); err == nil {
		return t, nil
	}

	return time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
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
