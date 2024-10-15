package segments

import (
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"gopkg.in/yaml.v3"
)

// Whether to use kubectl or read kubeconfig ourselves
const (
	ParseKubeConfig properties.Property = "parse_kubeconfig"
	ContextAliases  properties.Property = "context_aliases"
	kubectlCacheKey                     = "kubectl"
)

type Kubectl struct {
	props properties.Properties
	env   runtime.Environment
	KubeContext
	Context string
	dirty   bool
}

type KubeConfig struct {
	CurrentContext string `yaml:"current-context"`
	Contexts       []struct {
		Context *KubeContext `yaml:"context"`
		Name    string       `yaml:"name"`
	} `yaml:"contexts"`
}

type KubeContext struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace"`
}

func (k *Kubectl) Template() string {
	return " {{ .Context }}{{ if .Namespace }} :: {{ .Namespace }}{{ end }} "
}

func (k *Kubectl) Init(props properties.Properties, env runtime.Environment) {
	k.props = props
	k.env = env
}

func (k *Kubectl) Enabled() bool {
	parseKubeConfig := k.props.GetBool(ParseKubeConfig, true)

	if parseKubeConfig {
		return k.doParseKubeConfig()
	}

	return k.doCallKubectl()
}

func (k *Kubectl) doParseKubeConfig() bool {
	// Follow kubectl search rules (see https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable)
	// TL;DR: KUBECONFIG can contain a list of files. If it's empty ~/.kube/config is used. First file in list wins when merging keys.
	kubeconfigs := filepath.SplitList(k.env.Getenv("KUBECONFIG"))
	if len(kubeconfigs) == 0 {
		kubeconfigs = []string{filepath.Join(k.env.Home(), ".kube/config")}
	}

	contexts := make(map[string]*KubeContext)
	k.Context = ""

	for _, kubeconfig := range kubeconfigs {
		if len(kubeconfig) == 0 {
			continue
		}

		content := k.env.FileContent(kubeconfig)

		var config KubeConfig
		err := yaml.Unmarshal([]byte(content), &config)
		if err != nil {
			continue
		}

		for _, context := range config.Contexts {
			if _, exists := contexts[context.Name]; !exists {
				contexts[context.Name] = context.Context
			}
		}

		if len(k.Context) == 0 {
			k.Context = config.CurrentContext
		}

		context, exists := contexts[k.Context]
		if !exists {
			continue
		}

		if context != nil {
			k.KubeContext = *context
		}

		k.SetContextAlias()
		k.dirty = true

		return true
	}

	displayError := k.props.GetBool(properties.DisplayError, false)
	if !displayError {
		return false
	}
	k.setError("KUBECONFIG ERR")
	return true
}

func (k *Kubectl) doCallKubectl() bool {
	cmd := "kubectl"
	if !k.env.HasCommand(cmd) {
		return false
	}

	result, err := k.env.RunCommand(cmd, "config", "view", "--output", "yaml", "--minify")
	displayError := k.props.GetBool(properties.DisplayError, false)
	if err != nil && displayError {
		k.setError("KUBECTL ERR")
		return true
	}

	if err != nil {
		return false
	}

	var config KubeConfig
	err = yaml.Unmarshal([]byte(result), &config)
	if err != nil {
		return false
	}

	k.Context = config.CurrentContext
	k.SetContextAlias()
	k.dirty = true

	if len(config.Contexts) > 0 {
		k.KubeContext = *config.Contexts[0].Context
	}

	return true
}

func (k *Kubectl) setError(message string) {
	if len(k.Context) == 0 {
		k.Context = message
	}

	k.Namespace = message
	k.User = message
	k.Cluster = message
}

func (k *Kubectl) SetContextAlias() {
	aliases := k.props.GetKeyValueMap(ContextAliases, map[string]string{})
	if alias, exists := aliases[k.Context]; exists {
		k.Context = alias
	}
}
