package segments

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Aws struct {
	base

	Profile string
	Region  string
	RegionAlias string
}

const (
	defaultUser = "default"
)

var (
	// cardinalSuffixRegex converts cardinal regions to first letter aliases
	// Also ensures compounds like "southeast" are converted to "se"
	cardinalSuffixRegex = regexp.MustCompile(`orth|outh|ast|est|entral`)
)

func (a *Aws) Template() string {
	return " {{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }} "
}

func (a *Aws) Enabled() bool {
	getEnvFirstMatch := func(envs ...string) string {
		for _, env := range envs {
			value := a.env.Getenv(env)
			if value != "" {
				return value
			}
		}
		return ""
	}
	displayDefaultUser := a.props.GetBool(properties.DisplayDefault, true)
	a.Profile = getEnvFirstMatch("AWS_VAULT", "AWS_DEFAULT_PROFILE", "AWS_PROFILE")
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}
	a.Region = getEnvFirstMatch("AWS_REGION", "AWS_DEFAULT_REGION")
	a.RegionAlias = a.getRegionAlias(a.Region)
	if a.Profile != "" && a.Region != "" {
		return true
	}
	if a.Profile == "" && a.Region != "" && displayDefaultUser {
		a.Profile = defaultUser
		return true
	}
	a.getConfigFileInfo()
	if !displayDefaultUser && a.Profile == defaultUser {
		return false
	}
	return a.Profile != ""
}

func (a *Aws) getConfigFileInfo() {
	configPath := a.env.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.aws/config", a.env.Home())
	}
	config := a.env.FileContent(configPath)
	configSection := "[default]"
	if a.Profile != "" {
		configSection = fmt.Sprintf("[profile %s]", a.Profile)
	}
	configLines := strings.SplitSeq(config, "\n")
	var sectionActive bool
	for line := range configLines {
		if strings.HasPrefix(line, configSection) {
			sectionActive = true
			continue
		}
		if sectionActive && strings.HasPrefix(line, "region") {
			splitted := strings.Split(line, "=")
			if len(splitted) >= 2 {
				a.Region = strings.TrimSpace(splitted[1])
				a.RegionAlias = a.getRegionAlias(a.Region)
				break
			}
		}
	}
	if a.Profile == "" && a.Region != "" {
		a.Profile = defaultUser
	}
}

func (a *Aws) getRegionAlias(region string) string {
	splitted := strings.Split(region, "-")
	if len(splitted) < 2 {
		return region
	}
	splitted[1] = cardinalSuffixRegex.ReplaceAllString(splitted[1], "")
	return strings.Join(splitted, "")
}
