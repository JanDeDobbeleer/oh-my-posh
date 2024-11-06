package segments

type Zig struct {
	language
}

func (zig *Zig) Template() string {
	return languageTemplate
}

func (zig *Zig) Enabled() bool {
	zig.extensions = []string{"*.zig", "*.zon"}
	zig.projectFiles = []string{"build.zig"}
	zig.commands = []*cmd{
		{
			executable: "zig",
			args:       []string{"version"},
			regex:      `(?P<version>(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)`, //nolint:lll
		},
	}

	zig.versionURLTemplate = "https://ziglang.org/download/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/release-notes.html"

	return zig.language.Enabled()
}

func (zig *Zig) InProjectDir() bool {
	return zig.projectRoot != nil
}
