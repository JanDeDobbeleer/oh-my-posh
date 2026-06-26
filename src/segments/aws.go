package segments

import (
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"gopkg.in/ini.v1"
)

type Aws struct {
	Base

	// Settings holds every key/value pair from the active profile in the AWS shared
	// config and credentials files. Credential-file entries take precedence over
	// config-file entries for the same key, mirroring the AWS SDK's resolution order.
	// Templates can read any AWS-recognized setting via {{ .Settings.<key> }}, e.g.
	// {{ .Settings.role_arn }} or {{ .Settings.sso_role_name }}.
	Settings map[string]string

	// SSOSession holds the resolved [sso-session <name>] section keys when the
	// active profile references one via the sso_session key. Use as
	// {{ .SSOSession.sso_start_url }}, etc.
	SSOSession map[string]string

	Profile     string
	Region      string
	AccountID   string
	AccessKeyID string
}

const (
	defaultStr = "default"

	// AWS shared config keys we promote to convenience fields.
	awsKeyRegion       = "region"
	awsKeyAccessKeyID  = "aws_access_key_id"
	awsKeyAccountID    = "aws_account_id"
	awsKeySSOAccountID = "sso_account_id"
	awsKeySSOSession   = "sso_session"

	awsConfigSectionPrefix     = "profile "
	awsConfigSectionSSOSession = "sso-session "
)

func (a *Aws) Template() string {
	return " {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} "
}

func (a *Aws) Enabled() bool {
	a.Settings = map[string]string{}
	a.SSOSession = map[string]string{}

	getEnvFirstMatch := func(envs ...string) string {
		for _, env := range envs {
			if value := a.env.Getenv(env); value != "" {
				return value
			}
		}

		return ""
	}

	displayDefaultUser := a.options.Bool(options.DisplayDefault, true)

	a.Profile = getEnvFirstMatch("AWS_VAULT", "AWS_DEFAULT_PROFILE", "AWS_PROFILE")
	if !displayDefaultUser && a.Profile == defaultStr {
		return false
	}

	a.Region = getEnvFirstMatch("AWS_REGION", "AWS_DEFAULT_REGION")
	a.AccessKeyID = a.env.Getenv("AWS_ACCESS_KEY_ID")

	a.loadConfigFile()
	a.loadCredentialsFile()

	if a.Region == "" {
		a.Region = a.Settings[awsKeyRegion]
	}

	if a.AccountID == "" {
		a.AccountID = firstNonEmpty(a.Settings[awsKeySSOAccountID], a.Settings[awsKeyAccountID])
	}

	if a.AccessKeyID == "" {
		a.AccessKeyID = a.Settings[awsKeyAccessKeyID]
	}

	if a.Profile == "" && a.Region != "" {
		a.Profile = defaultStr
	}

	if !displayDefaultUser && a.Profile == defaultStr {
		return false
	}

	return a.Profile != ""
}

func (a *Aws) loadConfigFile() {
	configPath := a.env.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join(a.env.Home(), ".aws", "config")
	}

	cfg, ok := a.parseINI(configPath)
	if !ok {
		return
	}

	sectionName := defaultStr
	if a.Profile != "" {
		sectionName = awsConfigSectionPrefix + a.Profile
	}

	a.copySection(cfg, sectionName, a.Settings)

	if sessionName := a.Settings[awsKeySSOSession]; sessionName != "" {
		a.copySection(cfg, awsConfigSectionSSOSession+sessionName, a.SSOSession)
	}
}

func (a *Aws) loadCredentialsFile() {
	credentialsPath := a.env.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if credentialsPath == "" {
		credentialsPath = filepath.Join(a.env.Home(), ".aws", "credentials")
	}

	cfg, ok := a.parseINI(credentialsPath)
	if !ok {
		return
	}

	sectionName := defaultStr
	if a.Profile != "" {
		sectionName = a.Profile
	}

	a.copySection(cfg, sectionName, a.Settings)
}

func (a *Aws) parseINI(path string) (*ini.File, bool) {
	content := a.env.FileContent(path)
	if content == "" {
		return nil, false
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, []byte(content))
	if err != nil {
		log.Error(err)
		return nil, false
	}

	return cfg, true
}

func (a *Aws) copySection(cfg *ini.File, name string, dest map[string]string) {
	section, err := cfg.GetSection(name)
	if err != nil {
		return
	}

	for _, key := range section.Keys() {
		dest[key.Name()] = strings.TrimSpace(key.Value())
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

func (a *Aws) RegionAlias() string {
	if a.Region == "" {
		return ""
	}

	splitted := strings.Split(a.Region, "-")
	if len(splitted) < 2 {
		return a.Region
	}

	splitted[1] = regex.ReplaceAllString(`orth|outh|ast|est|entral`, splitted[1], "")
	return strings.Join(splitted, "")
}
