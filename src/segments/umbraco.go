package segments

import (
	"encoding/xml"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Umbraco struct {
	props   properties.Properties
	env     runtime.Environment
	Version string
	Modern  bool
}

type CSProj struct {
	PackageReferences []struct {
		Name    string `xml:"include,attr"`
		Version string `xml:"version,attr"`
	} `xml:"ItemGroup>PackageReference"`
}

type WebConfig struct {
	AppSettings []struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	} `xml:"appSettings>add"`
}

func (u *Umbraco) Enabled() bool {
	var location string

	// Check if we have a folder called Umbraco or umbraco in the current directory or a parent directory
	folders := []string{"umbraco", "Umbraco"}
	for _, folder := range folders {
		if file, err := u.env.HasParentFilePath(folder, false); err == nil {
			location = file.ParentFolder
			break
		}
	}

	if len(location) == 0 {
		u.env.Debug("no umbraco folder found in parent directories")
		return false
	}

	files := u.env.LsDir(location)

	// Loop over files where we found the Umbraco folder
	// To see if we can find a web.config or *.csproj file
	// If we do then we can scan the file to see if Umbraco has been installed
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.EqualFold(file.Name(), "web.config") {
			if u.TryFindLegacyUmbraco(filepath.Join(location, file.Name())) {
				return true
			}

			// We may have found a web.config first before a *.csproj file
			// So we need to keep checking to see if modern Umbraco is installed if we come across a *.csproj file
			continue
		}

		if strings.EqualFold(filepath.Ext(file.Name()), ".csproj") {
			if u.TryFindModernUmbraco(filepath.Join(location, file.Name())) {
				return true
			}

			// We may have found a *.csproj first before a web.config file
			// So we need to keep checking if legacy Umbraco is installed (as the *.csproj could be for a non-Umbraco project)
			continue
		}
	}

	return false
}

func (u *Umbraco) Template() string {
	return "{{.Version}} "
}

func (u *Umbraco) Init(props properties.Properties, env runtime.Environment) {
	u.props = props
	u.env = env
}

func (u *Umbraco) TryFindModernUmbraco(configPath string) bool {
	// Check the passed in filepath is not empty
	if len(configPath) == 0 {
		u.env.Debug("no configPath provided")
		return false
	}

	// Read the file contents of the csproj file
	contents := u.env.FileContent(configPath)

	// As XML unmarshal does not support case insenstivity attributes
	// this is just a simple string replace to lowercase the attribute
	contents = strings.ReplaceAll(contents, "Include=", "include=")
	contents = strings.ReplaceAll(contents, "Version=", "version=")

	// XML Unmarshal - map the contents of the file to the CSProj struct
	csProjPackages := CSProj{}
	err := xml.Unmarshal([]byte(contents), &csProjPackages)

	if err != nil {
		u.env.Debug(err.Error())
	}

	// Loop over all the package references
	for _, packageReference := range csProjPackages.PackageReferences {
		if strings.EqualFold(packageReference.Name, "umbraco.cms") {
			u.Modern = true
			u.Version = packageReference.Version

			return true
		}
	}

	return false
}

func (u *Umbraco) TryFindLegacyUmbraco(configPath string) bool {
	// Check the passed in filepath is not empty
	if len(configPath) == 0 {
		u.env.Debug("no configPath provided")
		return false
	}

	// Read the file contents of the web.config
	contents := u.env.FileContent(configPath)

	// As XML unmarshal does not support case insenstivity attributes
	// this is just a simple string replace to lowercase the attribute
	contents = strings.ReplaceAll(contents, "Key=", "key=")
	contents = strings.ReplaceAll(contents, "Value=", "value=")

	// XML Unmarshal - web.config all AppSettings keys
	webConfigAppSettings := WebConfig{}
	err := xml.Unmarshal([]byte(contents), &webConfigAppSettings)

	if err != nil {
		u.env.Debug(err.Error())
	}

	// Loop over all the package references
	for _, appSetting := range webConfigAppSettings.AppSettings {
		if strings.EqualFold(appSetting.Key, "umbraco.core.configurationstatus") || strings.EqualFold(appSetting.Key, "umbracoConfigurationStatus") {
			u.Modern = false
			u.Version = appSetting.Value

			if len(u.Version) == 0 {
				u.Version = UNKNOWN
			}

			return true
		}
	}

	return false
}
