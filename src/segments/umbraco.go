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

	Modern  bool
	Version string
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
		if file, err := u.env.HasParentFilePath(folder); err == nil {
			location = file.ParentFolder
			break
		}
	}

	if len(location) == 0 {
		u.env.Debug("No umbraco folder found in parent directories")
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
			return u.TryFindLegacyUmbraco(filepath.Join(location, file.Name()))
		}

		if strings.EqualFold(filepath.Ext(file.Name()), ".csproj") {
			return u.TryFindModernUmbraco(filepath.Join(location, file.Name()))
		}
	}

	return false
}

func (u *Umbraco) Template() string {
	return "{{.Version}} "
}

func (u *Umbraco) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env
}

func (u *Umbraco) TryFindModernUmbraco(configPath string) bool {
	// Check the passed in filepath is not empty
	if len(configPath) == 0 {
		u.env.Debug("UMBRACO: No .CSProj file path passed in")
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
		u.env.Debug("UMBRACO: Error while trying to parse XML of .csproj file")
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
		u.env.Debug("UMBRACO: No web.config file path passed in")
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
		u.env.Debug("UMBRACO: Error while trying to parse XML of web.config file")
		u.env.Debug(err.Error())
	}

	// Loop over all the package references
	for _, appSetting := range webConfigAppSettings.AppSettings {
		if strings.EqualFold(appSetting.Key, "umbraco.core.configurationstatus") {
			u.Modern = false

			if len(appSetting.Value) == 0 {
				u.Version = "Unknown"
			} else {
				u.Version = appSetting.Value
			}

			return true
		}
	}

	return false
}
