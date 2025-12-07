package segments

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/spf13/pflag"
	yaml "go.yaml.in/yaml/v3"
)

const (
	argocdOptsEnv     = "ARGOCD_OPTS"
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
	CurrentContext string           `yaml:"current-context"`
	Contexts       []*ArgocdContext `yaml:"contexts"`
}

type Argocd struct {
	Base

	ArgocdContext
}

func (a *Argocd) Template() string {
	return NameTemplate
}

func (a *Argocd) Enabled() bool {
	// always parse config instead of using cli to save time
	configPath := a.getConfigPath()
	succeeded, err := a.parseConfig(configPath)
	if err != nil {
		log.Error(err)
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
	flags.ParseErrorsAllowlist.UnknownFlags = true
	// only care about config
	flags.String("config", "", "get config from opts")

	opts := a.env.Getenv(argocdOptsEnv)
	_ = flags.Parse(strings.Split(opts, " "))
	return flags.Lookup("config").Value.String()
}

func (a *Argocd) parseConfig(file string) (bool, error) {
	config := a.env.FileContent(file)
	// missing or empty file content
	if config == "" {
		return false, errors.New(argocdInvalidYaml)
	}

	var data ArgocdConfig
	err := yaml.Unmarshal([]byte(config), &data)
	if err != nil {
		log.Error(err)
		return false, errors.New(argocdInvalidYaml)
	}
	a.Name = data.CurrentContext
	for _, context := range data.Contexts {
		if context.Name == a.Name {
			// mandatory fields in yaml
			if context.Server == "" || context.User == "" {
				return false, errors.New(argocdInvalidYaml)
			}
			a.Server = context.Server
			a.User = context.User
			return true, nil
		}
	}
	return false, errors.New(argocdNoCurrent)
}
