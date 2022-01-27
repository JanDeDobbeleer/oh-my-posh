package main

import (
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Whether to use kubectl or read kubeconfig ourselves
const ParseKubeConfig Property = "parse_kubeconfig"

type kubectl struct {
	props   Properties
	env     Environment
	Context string
	KubeContext
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

func (k *kubectl) template() string {
	return "{{ .Context }}{{ if .Namespace }} :: {{ .Namespace }}{{ end }}"
}

func (k *kubectl) init(props Properties, env Environment) {
	k.props = props
	k.env = env
}

func (k *kubectl) enabled() bool {
	parseKubeConfig := k.props.getBool(ParseKubeConfig, false)
	if parseKubeConfig {
		return k.doParseKubeConfig()
	}
	return k.doCallKubectl()
}

func (k *kubectl) doParseKubeConfig() bool {
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
		return true
	}

	displayError := k.props.getBool(DisplayError, false)
	if !displayError {
		return false
	}
	k.setError("KUBECONFIG ERR")
	return true
}

func (k *kubectl) doCallKubectl() bool {
	cmd := "kubectl"
	if !k.env.HasCommand(cmd) {
		return false
	}
	result, err := k.env.RunCommand(cmd, "config", "view", "--output", "yaml", "--minify")
	displayError := k.props.getBool(DisplayError, false)
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
	if len(config.Contexts) > 0 {
		k.KubeContext = *config.Contexts[0].Context
	}
	return true
}

func (k *kubectl) setError(message string) {
	if len(k.Context) == 0 {
		k.Context = message
	}
	k.Namespace = message
	k.User = message
	k.Cluster = message
}
