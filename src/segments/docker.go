package segments

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	// FetchContext is the property used to fetch the current docker context
	FetchContext options.Option = "fetch_context"
	// DockerCommand is the property used to specify the docker command to use
	DockerCommand options.Option = "docker_command"
	// Filter is the property used to specify a filter to apply when fetching the current docker context, see https://docs.docker.com/reference/cli/docker/container/ls/#filter
	Filter options.Option = "filter"
)

type DockerConfig struct {
	CurrentContext string `json:"currentContext"`
	DockerCommand  string
	Filter         string
}

type Docker struct {
	Base

	Context    string
	Containers []Container

	command *cmd
}

type Container struct {
	ID      string
	Image   string
	Command string
	Created string
	Status  string
	Ports   string
	Names   string
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
		"compose.yml",
		"compose.yaml",
		"docker-compose.yml",
		"docker-compose.yaml",
		"Dockerfile",
	}

	extensions = d.options.StringArray(LanguageExtensions, extensions)

	displayMode := d.options.String(DisplayMode, DisplayModeContext)

	switch displayMode {
	case DisplayModeContext:
		return d.fetchContext()
	case DisplayModeFiles:
		if !slices.ContainsFunc(extensions, d.env.HasFiles) {
			return false
		}

		// always respect the context fetching
		if d.options.Bool(FetchContext, true) {
			_ = d.fetchContext()
		}

		return true
	case DisplayModeEnvironment:
		// always fetch context first
		hasContext := d.fetchContext()

		dockerCommand := d.options.String(DockerCommand, "docker")
		if !d.env.HasCommand(dockerCommand) {
			return hasContext
		}

		filter := d.options.String(Filter, "")
		// Use Go template formatting with tab separation
		format := `{{.ID}}\t{{.Image}}\t{{.Command}}\t{{.CreatedAt}}\t{{.Status}}\t{{.Ports}}\t{{.Names}}`
		args := []string{"ps", "--format", format}
		if len(filter) > 0 {
			args = append(args, "--filter", filter)
		}

		d.command = &cmd{
			executable: dockerCommand,
			args:       args,
		}

		containers, err := d.fetchContainers()
		if err != nil {
			return hasContext
		}

		d.Containers = containers

		return hasContext || len(d.Containers) > 0
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
		if data == "" {
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

func (d *Docker) fetchContainers() ([]Container, error) {
	if d.command == nil {
		return nil, nil
	}

	output, err := d.env.RunCommand(d.command.executable, d.command.args...)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil, nil
	}

	containers := make([]Container, len(lines))
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != 7 {
			continue // Or handle error for malformed line
		}
		containers[i] = Container{
			ID:      fields[0],
			Image:   fields[1],
			Command: fields[2],
			Created: fields[3],
			Status:  fields[4],
			Ports:   fields[5],
			Names:   fields[6],
		}
	}

	return containers, nil
}
