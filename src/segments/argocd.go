package segments

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	argocdOptsEnv     = "ARGOCD_OPTS"
	argocdInvalidFlag = "invalid flag"
	argocdInvalidYaml = "invalid yaml"
	argocdNoCurrent   = "no current context"

	NameTemplate = " {{ .Name }} "
)

type ArgocdContext struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	User   string `yaml:"user"`
}

type ArgocdConfig struct {
	Contexts       []*ArgocdContext `yaml:"contexts"`
	CurrentContext string           `yaml:"current-context"`
}

type Argocd struct {
	props properties.Properties
	env   platform.Environment

	ArgocdContext
}

func (a *Argocd) Template() string {
	return NameTemplate
}

func (a *Argocd) Init(props properties.Properties, env platform.Environment) {
	a.props = props
	a.env = env
}

func (a *Argocd) Enabled() bool {
	// always parse config instead of using cli to save time
	configPath := a.getConfigPath()
	succeeded, err := a.parseConfig(configPath)
	if err != nil {
		a.env.Error(err)
		return false
	}
	return succeeded
}

func (a *Argocd) getConfigPath() string {
	cp := path.Join(a.env.Home(), ".config", "argocd", "config")
	cpo := a.getConfigFromOpts()
	if len(cpo) > 0 {
		cp = cpo
	}
	return cp
}

func (a *Argocd) getConfigFromOpts() string {
	// don't exit/panic when encountering invalid flags
	flags := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	// ignore other valid and invalid flags
	flags.ParseErrorsWhitelist.UnknownFlags = true
	// only care about config
	flags.String("config", "", "get config from opts")

	opts := a.env.Getenv(argocdOptsEnv)
	_ = flags.Parse(strings.Split(opts, " "))
	return flags.Lookup("config").Value.String()
}

func (a *Argocd) parseConfig(file string) (bool, error) {
	config := a.env.FileContent(file)
	// missing or empty file content
	if len(config) == 0 {
		return false, errors.New(argocdInvalidYaml)
	}

	var data ArgocdConfig
	err := yaml.Unmarshal([]byte(config), &data)
	if err != nil {
		a.env.Error(err)
		return false, errors.New(argocdInvalidYaml)
	}
	a.Name = data.CurrentContext
	for _, context := range data.Contexts {
		if context.Name == a.Name {
			// mandatory fields in yaml
			if len(context.Server) == 0 || len(context.User) == 0 {
				return false, errors.New(argocdInvalidYaml)
			}
			a.Server = context.Server
			a.User = context.User
			return true, nil
		}
	}
	return false, errors.New(argocdNoCurrent)
}
