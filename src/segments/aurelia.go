package segments

import (
	"encoding/json"
)

type Aurelia struct {
	language
}

func (a *Aurelia) Template() string {
	return languageTemplate
}

func (a *Aurelia) Enabled() bool {
	a.extensions = []string{"package.json"}
	a.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(-(?P<prerelease>[a-z]+).(?P<buildmetadata>[0-9]+))?)))`,
			getVersion: a.getVersion,
		},
	}

	a.versionURLTemplate = "https://github.com/aurelia/aurelia/releases/tag/v{{ .Full }}"

	if !a.hasNodePackage("aurelia") {
		return false
	}

	packageVersion, err := a.getVersion()
	if err != nil {
		return false
	}

	version, err := a.commands[0].parse(packageVersion)
	if err != nil {
		return false
	}

	a.language.version = *version

	return true
}

func (a *Aurelia) hasNodePackage(name string) bool {
	packageJSON := a.language.env.FileContent("package.json")

	var packageData map[string]interface{}
	if err := json.Unmarshal([]byte(packageJSON), &packageData); err != nil {
		return false
	}

	dependencies, ok := packageData["dependencies"].(map[string]interface{})
	if !ok {
		return false
	}

	if _, exists := dependencies[name]; !exists {
		return false
	}

	return true
}

func (a *Aurelia) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(a.language.env, "aurelia")
}
