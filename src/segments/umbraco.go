package segments

import (
	"encoding/xml"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Umbraco struct {
	props properties.Properties
	env   platform.Environment

	Found   bool
	Modern  bool
	Version string
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

type FindUmbracoResult struct {
	FoundUmbracoFolder bool
	FoundWebConfig     bool
	FoundCSProj        bool
	FilePath           string
}

func (u *Umbraco) Enabled() bool {
	u.env.Debug("UMBRACO: Checking if we enable segment")

	// Checks if the current folder contains an umbraco folder
	// If not it then checks each parent until the root
	findUmbracoResults, err := u.TryFindUmbracoInParentDirsOrSelf()

	if err != nil {
		u.env.Debug("UMBRACO: Error while searching for Umbraco folder and files")
		u.env.Debug(err.Error())
		return false
	}

	if !findUmbracoResults.FoundUmbracoFolder {
		u.env.Debug("UMBRACO: No Umbraco folder found")
		u.Found = false
		return false
	}

	// We have found an Umbraco folder between CWD and root
	// Along with finding a web.config file at the same level as the Umbraco folder
	// OR one or more .csproj files at the same level as the Umbraco folder

	// Modern .NET Core based Umbraco
	if findUmbracoResults.FoundCSProj {
		return u.TryFindModernUmbraco(findUmbracoResults.FilePath)
	}

	// Legacy .NET Framework based Umbraco
	if findUmbracoResults.FoundWebConfig {
		return u.TryFindLegacyUmbraco(findUmbracoResults.FilePath)
	}

	// If we have got here then neither modern or legacy Umbraco was NOT found
	u.Found = false
	return false
}

func (u *Umbraco) Template() string {
	return "{{.Version}} "
}

func (u *Umbraco) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env
}

func (u *Umbraco) TryFindUmbracoInParentDirsOrSelf() (*FindUmbracoResult, error) {
	defer u.env.Trace(time.Now(), "UMBRACO: Checking for an umbraco folder & files in current or any parent folders until root")
	currentFolder := u.env.Pwd()
	results := FindUmbracoResult{}

	for {
		// Check if a directory named "Umbraco" exists in the current directory
		files := u.env.LsDir(currentFolder)

		// Have to loop over each item found in the current folder
		for _, file := range files {
			// Check if the item is a folder AND it matches 'Umbraco' regardless of casing
			if file.IsDir() && strings.EqualFold(file.Name(), umbracoFolderName) {
				u.env.Debug("UMBRACO: Found an Umbraco folder in " + currentFolder)
				results.FoundUmbracoFolder = true
			}

			// Check if the item is a file AND it matches 'web.config' regardless of casing
			if !file.IsDir() && strings.EqualFold(file.Name(), umbracoWebConfig) {
				u.env.Debug("UMBRACO: Found a web.config file in " + currentFolder)
				results.FoundWebConfig = true
				results.FilePath = filepath.Join(currentFolder, file.Name())
			}

			// Check if the item is a file AND has file extension of .csproj regardless of casing
			if !file.IsDir() && strings.EqualFold(filepath.Ext(file.Name()), ".csproj") {
				u.env.Debug("UMBRACO: Found a .csproj file in " + currentFolder)
				results.FoundCSProj = true
				results.FilePath = filepath.Join(currentFolder, file.Name())
			}

			// If we have found an Umbraco folder AND a .csproj OR  an Umbraco folder AND a web.config file
			// Then we can break the for loop as we have found what we need
			if (results.FoundUmbracoFolder && results.FoundCSProj) || (results.FoundUmbracoFolder && results.FoundWebConfig) {
				// Break out the loop for checking the collection of files in the current folder
				break
			}
		}

		// If we have found an Umbraco folder AND a .csproj OR  an Umbraco folder AND a web.config file
		// Then we can break the for loop as we have found what we need
		if (results.FoundUmbracoFolder && results.FoundCSProj) || (results.FoundUmbracoFolder && results.FoundWebConfig) {
			// Same logic in the inner for loop
			// But we will BREAK out and stop checking for parent folders
			break
		}

		// If we've reached the root directory, stop
		// Otherwise this loop will run forever - EEEK
		if currentFolder == "/" || currentFolder == "\\" {
			break
		}

		// Move up to the parent directory
		currentFolder = filepath.Dir(currentFolder)
	}

	return &results, nil
}

func (u *Umbraco) TryFindModernUmbraco(configPath string) bool {
	// Check the passed in filepath is not empty
	if len(configPath) == 0 {
		u.env.Debug("UMBRACO: No .CSProj file path passed in")
		return false
	}

	// Read the file contents of the csproj file
	contents := u.env.FileContent(configPath)

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
			u.Found = true
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
			u.Found = true

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
