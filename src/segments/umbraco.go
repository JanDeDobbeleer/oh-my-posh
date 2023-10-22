package segments

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type FindUmbracoResult struct {
	FoundUmbracoFolder bool
	FoundWebConfig     bool
	FoundCSProj        bool
	FolderPath         string
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
		u.FoundUmbraco = false
		return false
	}

	// We have found an Umbraco folder between CWD and root
	// Along with finding a web.config file at the same level as the Umbraco folder
	// OR one or more .csproj files at the same level as the Umbraco folder

	// Modern .NET Core based Umbraco
	if findUmbracoResults.FoundCSProj {
		u.env.Debug("UMBRACO: Checking for modern Umbraco as we have found .csproj file")
		return u.TryFindModernUmbraco(findUmbracoResults.FolderPath)
	}

	// Legacy .NET Framework based Umbraco
	if findUmbracoResults.FoundWebConfig {
		u.env.Debug("UMBRACO: Checking for legacy Umbraco as we have found web.config file")
		return u.TryFindLegacyUmbraco(findUmbracoResults.FolderPath)
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
}

func (u *Umbraco) TryFindUmbracoInParentDirsOrSelf() (*FindUmbracoResult, error) {
	defer u.env.Trace(time.Now(), "UMBRACO: Checking for an umbraco folder & files in current or any parent folders until root")
	currentFolder := u.env.Pwd()
	results := FindUmbracoResult{}

	for {
		fmt.Println("Checking " + currentFolder)

		// Check if a directory named "Umbraco" exists in the current directory
		files, err := os.ReadDir(currentFolder)
		if err != nil {
			return nil, err
		}

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
			}

			// Check if the item is a file AND has file extension of .csproj regardless of casing
			if !file.IsDir() && strings.EqualFold(filepath.Ext(file.Name()), ".csproj") {
				u.env.Debug("UMBRACO: Found a .csproj file in " + currentFolder)
				results.FoundCSProj = true
			}

			// If we have found an Umbraco folder AND a .csproj OR  an Umbraco folder AND a web.config file
			// Then we can break the for loop as we have found what we need
			if (results.FoundUmbracoFolder && results.FoundCSProj) || (results.FoundUmbracoFolder && results.FoundWebConfig) {
				results.FolderPath = currentFolder
				break
			}
		}

		// If we've reached the root directory, stop
		// Otherwise this loop will run forever - EEEK
		if currentFolder == "/" {
			break
		}

		// Move up to the parent directory
		currentFolder = filepath.Dir(currentFolder)
	}

	return &results, nil
}

func (u *Umbraco) TryFindModernUmbraco(foundCSProjPath string) bool {
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

func (u *Umbraco) TryFindLegacyUmbraco(foundWebConfigPath string) bool {
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
