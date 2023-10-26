package segments

import (
	"encoding/xml"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Umbraco struct {
	props properties.Properties
	env   platform.Environment

	Modern  bool
	Version string
}

const (
	umbracoFolderName = "umbraco"
	umbracoWebConfig  = "web.config"
)

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
	u.env.Debug("UMBRACO: Checking if we enable segment")

	// Checks if the current folder contains an umbraco folder
	// If not it then checks each parent until the root
	foundWebConfig, foundCSProj, configFilePath, err := u.TryFindUmbracoInParentDirsOrSelf()

	if err != nil {
		u.env.Debug("UMBRACO: Error while searching for Umbraco folder and files")
		u.env.Debug(err.Error())
		return false
	}

	// We have found an Umbraco folder between CWD and root
	// Along with finding a web.config file at the same level as the Umbraco folder
	// OR one or more .csproj files at the same level as the Umbraco folder

	// Modern .NET Core based Umbraco
	if foundCSProj {
		return u.TryFindModernUmbraco(configFilePath)
	}

	// Legacy .NET Framework based Umbraco
	if foundWebConfig {
		return u.TryFindLegacyUmbraco(configFilePath)
	}

	// If we have got here then neither modern or legacy Umbraco was NOT found
	return false
}

func (u *Umbraco) Template() string {
	return "{{.Version}} "
}

func (u *Umbraco) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env
}

// Return type is bool FoundWebConfig, bool FoundCSProj, string FilePath, error
func (u *Umbraco) TryFindUmbracoInParentDirsOrSelf() (bool, bool, string, error) {
	defer u.env.Trace(time.Now(), "UMBRACO: Checking for an umbraco folder & files in current or any parent folders until root")
	currentFolder := u.env.Pwd()

	foundUmbracoFolder := false
	foundWebConfig := false
	foundCSProj := false
	configFilePath := ""

	for {
		// Check if a directory named "Umbraco" exists in the current directory
		files := u.env.LsDir(currentFolder)

		// Have to loop over each item found in the current folder
		for _, file := range files {
			// Check if the item is a folder AND it matches 'Umbraco' regardless of casing
			if file.IsDir() && strings.EqualFold(file.Name(), umbracoFolderName) {
				u.env.Debug("UMBRACO: Found an Umbraco folder in " + currentFolder)
				foundUmbracoFolder = true
			}

			// Check if the item is a file AND it matches 'web.config' regardless of casing
			if !file.IsDir() && strings.EqualFold(file.Name(), umbracoWebConfig) {
				u.env.Debug("UMBRACO: Found a web.config file in " + currentFolder)
				foundWebConfig = true
				configFilePath = filepath.Join(currentFolder, file.Name())
			}

			// Check if the item is a file AND has file extension of .csproj regardless of casing
			if !file.IsDir() && strings.EqualFold(filepath.Ext(file.Name()), ".csproj") {
				u.env.Debug("UMBRACO: Found a .csproj file in " + currentFolder)
				foundCSProj = true
				configFilePath = filepath.Join(currentFolder, file.Name())
			}

			// If we have found an Umbraco folder AND a .csproj
			// OR an Umbraco folder AND a web.config file
			if (foundUmbracoFolder && foundCSProj) || (foundUmbracoFolder && foundWebConfig) {
				// Break out the loop for checking the collection of files in the current folder
				break
			}
		}

		// If we have found an Umbraco folder AND a .csproj OR  an Umbraco folder AND a web.config file
		// Then we can break the for loop as we have found what we need
		if (foundUmbracoFolder && foundCSProj) || (foundUmbracoFolder && foundWebConfig) {
			// Same logic in the inner for loop
			// But we will BREAK out and stop checking for parent folders
			break
		}

		// If we've reached the root directory, stop
		// Otherwise this loop will run forever - EEEK
		if currentFolder == "/" || currentFolder == "\\" {
			// Still found nothing even at the root
			err := errors.New("scanned all directories to the root and did not find an umbraco folder with a web.config or *.csproj file belongside it")
			return false, false, "", err
		}

		// Move up to the parent directory
		currentFolder = filepath.Dir(currentFolder)
	}

	return foundWebConfig, foundCSProj, configFilePath, nil
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
