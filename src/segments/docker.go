package segments

import (
	"encoding/json"
	"path/filepath"

	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	// FetchContext is the property used to fetch the current docker context
	FetchContext properties.Property = "fetch_context"
)

type DockerConfig struct {
	CurrentContext string `json:"currentContext"`
}

type Docker struct {
	base

	Context string
}

func (d *Docker) Template() string {
	return " \uf308 {{ .Context }} "
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
	extensions := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"Dockerfile",
	}

	extensions = d.props.GetStringArray(LanguageExtensions, extensions)

	displayMode := d.props.GetString(DisplayMode, DisplayModeContext)
	switch displayMode {
	case DisplayModeContext:
		return d.fetchContext()
	case DisplayModeFiles:
		if !slices.ContainsFunc(extensions, d.env.HasFiles) {
			return false
		}

		// always respect the context fetching
		if d.props.GetBool(FetchContext, true) {
			_ = d.fetchContext()
		}

		return true
	}

	return false
}

func (d *Docker) fetchContext() bool {
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
