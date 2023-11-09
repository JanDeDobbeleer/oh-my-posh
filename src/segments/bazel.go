package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Bazel struct {
	language

	Icon string
}

const (
	// Bazel's icon
	Icon properties.Property = "icon"
)

func (c *Bazel) Template() string {
	return " {{ if .Error }}{{ .Icon }} {{ .Error }}{{ else }}{{ url .Icon .URL }} {{ .Full }}{{ end }} "
}

func (c *Bazel) Init(props properties.Properties, env platform.Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.bazel", "*.bzl", "BUILD", "WORKSPACE", ".bazelrc", ".bazelversion"},
		folders:    []string{"bazel-bin", "bazel-out", "bazel-testlogs"},
		commands: []*cmd{
			{
				executable: "bazel",
				args:       []string{"--version"},
				regex:      `bazel (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		// Use the correct URL for Bazel 6.x. it does not have the docs subdomain.
		versionURLTemplate: "https://{{ if ne .Major \"6\" }}docs.{{ end }}bazel.build/versions/{{ .Major }}.{{ .Minor }}.{{ .Patch }}",
	}
	c.Icon = props.GetString(Icon, "\ue63a")
}

func (c *Bazel) Enabled() bool {
	return c.language.Enabled()
}
