package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Bazel struct {
	Icon string
	language
}

const (
	// Bazel's icon
	Icon properties.Property = "icon"
)

func (b *Bazel) Template() string {
	return " {{ if .Error }}{{ .Icon }} {{ .Error }}{{ else }}{{ url .Icon .URL }} {{ .Full }}{{ end }} "
}

func (b *Bazel) Enabled() bool {
	b.extensions = []string{"*.bazel", "*.bzl", "BUILD", "WORKSPACE", ".bazelrc", ".bazelversion"}
	b.folders = []string{"bazel-bin", "bazel-out", "bazel-testlogs"}
	b.commands = []*cmd{
		{
			executable: "bazel",
			args:       []string{"--version"},
			regex:      `bazel (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	// Use the correct URL for Bazel >5.4.1, since they do not have the docs subdomain.
	b.versionURLTemplate = "https://{{ if lt .Major 6 }}docs.{{ end }}bazel.build/versions/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	b.Icon = b.props.GetString(Icon, "\ue63a")

	return b.language.Enabled()
}
