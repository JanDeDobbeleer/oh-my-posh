package segments

import (
	"encoding/json"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type DockerConfig struct {
	CurrentContext string `json:"currentContext"`
}

type Docker struct {
	props properties.Properties
	env   platform.Environment

	Context string
}

func (d *Docker) Template() string {
	return " \uf308 {{ .Context }} "
}

func (d *Docker) Init(props properties.Properties, env platform.Environment) {
	d.props = props
	d.env = env
}

func (d *Docker) envVars() []string {
	return []string{"DOCKER_MACHINE_NAME", "DOCKER_HOST", "DOCKER_CONTEXT"}
}

func (d *Docker) configFiles() []string {
	files := []string{
		filepath.Join(d.env.Home(), ".docker/config.json"),
	}

	dockerConfig := d.env.Getenv("DOCKER_CONFIG")
	if len(dockerConfig) > 0 {
		files = append(files, filepath.Join(dockerConfig, "config.json"))
	}

	return files
}

func (d *Docker) Enabled() bool {
	// Check if there is a non-empty environment variable named `DOCKER_HOST` or `DOCKER_CONTEXT`
	// These variables are set by the docker CLI and override the config file
	// Return the current context if it is not empty and not `default`
	for _, v := range d.envVars() {
		context := d.env.Getenv(v)
		if len(context) > 0 && context != "default" {
			d.Context = context
			return true
		}
	}

	// Check if there is a file named `$HOME/.docker/config.json` or `$DOCKER_CONFIG/config.json`
	// Return the current context if it is not empty and not `default`
	for _, f := range d.configFiles() {
		data := d.env.FileContent(f)
		if len(data) == 0 {
			continue
		}

		var cfg DockerConfig
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			continue
		}

		if len(cfg.CurrentContext) > 0 && cfg.CurrentContext != "default" {
			d.Context = cfg.CurrentContext
			return true
		}
	}

	return false
}
