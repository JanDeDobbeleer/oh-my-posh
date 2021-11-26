package main

import (
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Whether to use kubectl or read kubeconfig ourselves
const ParseKubeConfig Property = "parse_kubeconfig"

type kubectl struct {
	props   properties
	env     environmentInfo
	Context string
	KubeConfigContext
}

type KubeConfigContext struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace"`
}

type KubeConfig struct {
	CurrentContext string `yaml:"current-context"`
	Contexts       []struct {
		Context KubeConfigContext `yaml:"context"`
		Name    string            `yaml:"name"`
	} `yaml:"contexts"`
}

func (k *kubectl) string() string {
	segmentTemplate := k.props.getString(SegmentTemplate, "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  k,
		Env:      k.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (k *kubectl) init(props properties, env environmentInfo) {
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
	kubeconfigs := filepath.SplitList(k.env.getenv("KUBECONFIG"))
	if len(kubeconfigs) == 0 {
		kubeconfigs = []string{filepath.Join(k.env.homeDir(), ".kube/config")}
	}
	contexts := make(map[string]KubeConfigContext)
	k.Context = ""
	for _, kubeconfig := range kubeconfigs {
		if len(kubeconfig) == 0 {
			continue
		}

		content := k.env.getFileContent(kubeconfig)

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
		if exists {
			k.KubeConfigContext = context
			return true
		}
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
	if !k.env.hasCommand(cmd) {
		return false
	}
	result, err := k.env.runCommand(cmd, "config", "view", "--minify", "--output", "jsonpath={..current-context},{..namespace},{..context.user},{..context.cluster}")
	displayError := k.props.getBool(DisplayError, false)
	if err != nil && displayError {
		k.setError("KUBECTL ERR")
		return true
	}
	if err != nil {
		return false
	}

	values := strings.Split(result, ",")
	k.Context = values[0]
	k.Namespace = values[1]
	k.User = values[2]
	k.Cluster = values[3]
	return len(k.Context) > 0
}

func (k *kubectl) setError(message string) {
	if len(k.Context) == 0 {
		k.Context = message
	}
	k.Namespace = message
	k.User = message
	k.Cluster = message
}
