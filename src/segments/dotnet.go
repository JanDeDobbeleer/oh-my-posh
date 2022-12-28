package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/constants"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Dotnet struct {
	language

	Unsupported bool
}

func (d *Dotnet) Template() string {
	return " {{ if .Unsupported }}\uf071{{ else }}{{ .Full }}{{ end }} "
}

func (d *Dotnet) Init(props properties.Properties, env platform.Environment) {
	d.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.cs", "*.csx", "*.vb", "*.sln", "*.slnf", "*.csproj", "*.vbproj", "*.fs", "*.fsx", "*.fsproj", "global.json"},
		commands: []*cmd{
			{
				executable: "dotnet",
				args:       []string{"--version"},
				regex: `(?P<version>((?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
					`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?))`,
			},
		},
		versionURLTemplate: "https://github.com/dotnet/core/blob/master/release-notes/{{ .Major }}.{{ .Minor }}/{{ .Major }}.{{ .Minor }}.{{ substr 0 1 .Patch }}/{{ .Major }}.{{ .Minor }}.{{ substr 0 1 .Patch }}.md", //nolint: lll
	}
}

func (d *Dotnet) Enabled() bool {
	enabled := d.language.Enabled()
	if !enabled {
		return false
	}
	d.Unsupported = d.language.exitCode == constants.DotnetExitCode
	return true
}
