package main

type dotnet struct {
	language *language
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

func (d *dotnet) init(props *properties, env environmentInfo) {
	d.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"*.cs", "*.csx", "*.vb", "*.sln", "*.csproj", "*.vbproj", "*.fs", "*.fsx", "*.fsproj"},
		commands: []*cmd{
			{
				executable: "dotnet",
				args:       []string{"--version"},
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?:\d{2})(?P<patch>[0-9]{1}))))`,
			},
		},
		versionURLTemplate: "[%1s](https://github.com/dotnet/core/blob/master/release-notes/%[2]s.%[3]s/%[2]s.%[3]s.%[4]s/%[2]s.%[3]s.%[4]s.md)",
	}
}

func (d *dotnet) enabled() bool {
	return d.language.enabled()
}
