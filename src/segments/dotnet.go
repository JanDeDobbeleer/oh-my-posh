package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/constants"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type globalJSON struct {
	Sdk struct {
		Version string `json:"version"`
	} `json:"sdk"`
}

const (
	// FetchSDKVersion fetches the SDK version in global.json
	FetchSDKVersion properties.Property = "fetch_sdk_version"
)

type Dotnet struct {
	SDKVersion string
	language
	Unsupported bool
}

func (d *Dotnet) Template() string {
	return " {{ if .Unsupported }}\uf071{{ else }}{{ .Full }}{{ end }} "
}

func (d *Dotnet) Enabled() bool {
	d.extensions = []string{
		"*.cs",
		"*.csx",
		"*.vb",
		"*.sln",
		"*.slnf",
		"*.csproj",
		"*.vbproj",
		"*.fs",
		"*.fsx",
		"*.fsproj",
		"global.json",
	}
	d.commands = []*cmd{
		{
			executable: "dotnet",
			args:       []string{"--version"},
			regex: `(?P<version>((?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
				`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?))`,
		},
	}
	d.versionURLTemplate = "https://github.com/dotnet/core/blob/master/release-notes/{{ .Major }}.{{ .Minor }}/{{ .Major }}.{{ .Minor }}.{{ substr 0 1 .Patch }}/{{ .Major }}.{{ .Minor }}.{{ substr 0 1 .Patch }}.md" //nolint: lll

	enabled := d.language.Enabled()
	if !enabled {
		return false
	}

	d.Unsupported = d.exitCode == constants.DotnetExitCode

	if !d.props.GetBool(FetchSDKVersion, false) {
		return true
	}

	file, err := d.env.HasParentFilePath("global.json", false)
	if err != nil {
		return true
	}

	content := d.env.FileContent(file.Path)

	var globalJSON globalJSON
	if err := json.Unmarshal([]byte(content), &globalJSON); err == nil {
		d.SDKVersion = globalJSON.Sdk.Version
	}

	return true
}
