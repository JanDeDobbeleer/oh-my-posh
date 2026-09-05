package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Bazel struct {
	Icon template.Markup
	Language
}

const (
	Icon options.Option = "icon"
)

func (b *Bazel) Template() string {
	return " {{ if .Error }}{{ .Icon }} {{ .Error }}{{ else }}{{ url .Icon .URL }} {{ .Full }}{{ end }} "
}

const bazelToolName = "bazel"

func (b *Bazel) Enabled() bool {
	b.loadSpec()

	return b.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (b *Bazel) Activation() Activation {
	b.loadSpec()

	return b.activation()
}

func (b *Bazel) loadSpec() {
	b.extensions = []string{"*.bazel", "*.bzl", "BUILD", "WORKSPACE", ".bazelrc", ".bazelversion"}
	b.folders = []string{"bazel-bin", "bazel-out", "bazel-testlogs"}
	// Not marked versionCacheable: the officially recommended way to install
	// "bazel" is bazelisk, whose own file never changes but which resolves
	// the actual Bazel release to run from the project's .bazelversion (one
	// of this segment's own file triggers above) or MODULE.bazel - so the
	// same resolved binary can report a different version per project.
	b.tooling = map[string]*cmd{
		bazelToolName: {
			executable: bazelToolName,
			args:       []string{versionFlagArg},
			regex:      `bazel ` + versionRegex,
		},
	}
	b.defaultTooling = []string{bazelToolName}
	// Use the correct URL for Bazel >5.4.1, since they do not have the docs subdomain.
	b.versionURLTemplate = "https://{{ if lt .Major 6 }}docs.{{ end }}bazel.build/versions/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	b.Icon = b.options.Markup(Icon, "\ue63a")
}
