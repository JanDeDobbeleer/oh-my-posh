package segments

import (
	"encoding/xml"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Umbraco struct {
	props properties.Properties
	env   platform.Environment

	FoundUmbraco    bool
	IsModernUmbraco bool
	IsLegacyUmbraco bool
	Version         string
}

const (
	umbracoFolderName = "umbraco"
	umbracoWebConfig  = "web.config"
)

type CSProj struct {
	PackageReferences []struct {
		Name    string `xml:"Include,attr"`
		Version string `xml:"Version,attr"`
	} `xml:"ItemGroup>PackageReference"`
}

type WebConfig struct {
	AppSettings []struct {
		Key   string `xml:"key,attr"` // TODO: What happens if the web.config has the attribute as uppercase Key="" ?
		Value string `xml:"value,attr"`
	} `xml:"appSettings>add"`
}

func (u *Umbraco) Enabled() bool {
	u.env.Debug("UMBRACO: Checking if we enable segment")

	// If the cwd does not contain a folder called 'umbraco'
	// Then get out of here...
	if !u.env.HasFolder(umbracoFolderName) {
		return false
	}

	// Check if we have a .csproj OR a web.config in the CWD
	// TODO: What if file name on disk is Web.config or web.Config?
	if !u.env.HasFiles("*.csproj") && !u.env.HasFiles("web.config") {
		u.env.Debug("UMBRACO: NO CSProj or web.config found")
		return false
	}

	// Modern .NET Core based Umbraco
	if u.TryFindModernUmbraco() {
		return true
	}

	// Legacy .NET Framework based Umbraco
	if u.TryFindLegacyUmbraco() {
		return true
	}

	// If we have got here then neither modern or legacy Umbraco was NOT found
	u.FoundUmbraco = false
	return false
}

func (u *Umbraco) Template() string {
	return "{{.Version}} "
}

func (u *Umbraco) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env

	u.env.Debug("HEY HEY HEY")
}

func (u *Umbraco) TryFindModernUmbraco() bool {
	// Check we have one or more .csproj files in the CWD
	if !u.env.HasFiles("*.csproj") {
		return false
	}

	// Get a list of all files that match the search pattern
	// Some folders could have multiple .csproj files in them
	searchPattern := "*.csproj"

	// Get a list of all files that match the search pattern
	files, err := filepath.Glob(searchPattern)

	if err != nil {
		u.env.Debug("UMBRACO: Error while searching for .csproj files")
		u.env.Debug(err.Error())
		return false
	}

	// Loop over all the files that have a .csproj extension
	for _, file := range files {
		u.env.Debug("UMBRACO: Trying to open file at " + file)

		// Read the file contents of the csproj file
		contents := u.env.FileContent(file)

		// XML Unmarshal - map the contents of the file to the CSProj struct
		csProjPackages := CSProj{}
		err := xml.Unmarshal([]byte(contents), &csProjPackages)

		if err != nil {
			u.env.Debug("UMBRACO: Error while trying to parse XML of .csproj file")
			u.env.Debug(err.Error())
		}

		// Loop over all the package references
		for _, packageReference := range csProjPackages.PackageReferences {
			if strings.EqualFold(packageReference.Name, "umbraco.cms") {
				u.IsModernUmbraco = true
				u.FoundUmbraco = true
				u.Version = packageReference.Version

				return true
			}
		}
	}

	return false
}

func (u *Umbraco) TryFindLegacyUmbraco() bool {
	if !u.env.HasFiles(umbracoWebConfig) {
		return false
	}

	// Read the file contents of the web.config in the CWD
	contents := u.env.FileContent(umbracoWebConfig)

	// XML Unmarshal - web.config all AppSettings keys
	webConfigAppSettings := WebConfig{}
	err := xml.Unmarshal([]byte(contents), &webConfigAppSettings)

	if err != nil {
		u.env.Debug("UMBRACO: Error while trying to parse XML of web.config file")
		u.env.Debug(err.Error())
	}

	// Loop over all the package references
	for _, appSetting := range webConfigAppSettings.AppSettings {
		if strings.EqualFold(appSetting.Key, "umbraco.core.configurationstatus") {
			u.IsLegacyUmbraco = true
			u.FoundUmbraco = true

			if appSetting.Value == "" {
				u.Version = "Unknown"
			} else {
				u.Version = appSetting.Value
			}

			return true
		}
	}

	return false
}
