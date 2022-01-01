package main

type dotnet struct {
	language
}

const (
	// UnsupportedDotnetVersionIcon is displayed when the dotnet version in
	// the current folder isn't supported by the installed dotnet SDK set.
	UnsupportedDotnetVersionIcon Property = "unsupported_version_icon"
)

func (d *dotnet) string() string {
	version := d.language.string()

	exitCode := d.language.exitCode
	if exitCode == dotnetExitCode {
		return d.language.props.getString(UnsupportedDotnetVersionIcon, "\uf071 ")
	}

	return version
}

func (d *dotnet) init(props Properties, env environmentInfo) {
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
		versionURLTemplate: "[%1s](https://github.com/dotnet/core/blob/master/release-notes/%[2]s.%[3]s/%[2]s.%[3]s.%[4]s/%[2]s.%[3]s.%[4]s.md)",
	}
}

func (d *dotnet) enabled() bool {
	return d.language.enabled()
}
