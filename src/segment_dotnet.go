package main

type dotnet struct {
	language

	Unsupported bool
}

func (d *dotnet) string() string {
	segmentTemplate := d.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) != 0 {
		return d.language.renderTemplate(segmentTemplate, d)
	}
	if d.Unsupported {
		return d.language.props.getString(UnsupportedDotnetVersionIcon, "\uf071 ")
	}
	return d.language.string()
}

func (d *dotnet) init(props Properties, env Environment) {
	d.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.cs", "*.csx", "*.vb", "*.sln", "*.csproj", "*.vbproj", "*.fs", "*.fsx", "*.fsproj", "global.json"},
		commands: []*cmd{
			{
				executable: "dotnet",
				args:       []string{"--version"},
				regex: `(?P<version>((?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
					`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?))`,
			},
		},
		versionURLTemplate: "https://github.com/dotnet/core/blob/master/release-notes/{{ .Major }}.{{ .Minor }}/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/{{ .Major }}.{{ .Minor }}.{{ .Patch }}.md)", // nolint: lll
	}
}

func (d *dotnet) enabled() bool {
	enabled := d.language.enabled()
	if !enabled {
		return false
	}
	d.Unsupported = d.language.exitCode == dotnetExitCode
	return true
}
