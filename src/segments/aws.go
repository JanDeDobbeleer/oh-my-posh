package segments

import (
	"fmt"
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
	defaultUser = "default"

	// DisplayAccountID toggles populating the AccountID convenience field
	// from sso_account_id (preferred) or aws_account_id.
	DisplayAccountID options.Option = "display_account_id"
	// DisplayAccessKeyID toggles populating the AccessKeyID convenience field
	// from AWS_ACCESS_KEY_ID, the credentials file, or the config file.
	DisplayAccessKeyID options.Option = "display_access_key_id"

	// AWS shared config keys we promote to convenience fields.
	awsKeyRegion       = "region"
	awsKeyAccessKeyID  = "aws_access_key_id"
	awsKeyAccountID    = "aws_account_id"
	awsKeySSOAccountID = "sso_account_id"
	awsKeySSOSession   = "sso_session"

	awsConfigSectionDefault    = "default"
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
	displayAccountID := a.options.Bool(DisplayAccountID, true)
	displayAccessKeyID := a.options.Bool(DisplayAccessKeyID, false)

	a.Profile = getEnvFirstMatch("AWS_VAULT", "AWS_DEFAULT_PROFILE", "AWS_PROFILE")
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}

	a.Region = getEnvFirstMatch("AWS_REGION", "AWS_DEFAULT_REGION")
	if displayAccessKeyID {
		a.AccessKeyID = a.env.Getenv("AWS_ACCESS_KEY_ID")
	}

	a.loadConfigFile()
	a.loadCredentialsFile()

	if a.Region == "" {
		a.Region = a.Settings[awsKeyRegion]
	}

	if displayAccountID && a.AccountID == "" {
		a.AccountID = firstNonEmpty(a.Settings[awsKeySSOAccountID], a.Settings[awsKeyAccountID])
	}

	if displayAccessKeyID && a.AccessKeyID == "" {
		a.AccessKeyID = a.Settings[awsKeyAccessKeyID]
	}

	if a.Profile == "" && a.Region != "" {
		a.Profile = defaultUser
	}

	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}

	return a.Profile != ""
}

func (a *Aws) loadConfigFile() {
	configPath := a.env.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.aws/config", a.env.Home())
	}

	cfg, ok := a.parseINI(configPath)
	if !ok {
		return
	}

	sectionName := awsConfigSectionDefault
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
		credentialsPath = fmt.Sprintf("%s/.aws/credentials", a.env.Home())
	}

	cfg, ok := a.parseINI(credentialsPath)
	if !ok {
		return
	}

	sectionName := awsConfigSectionDefault
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
