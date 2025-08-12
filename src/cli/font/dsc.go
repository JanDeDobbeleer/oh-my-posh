package font

import (
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Resource struct {
	dsc.Resource[*Font]
}

func DSC() *Resource {
	return &Resource{
		Resource: dsc.Resource[*Font]{
			JSONSchemaURL: "https://ohmyposh.dev/dsc.font.schema.json",
		},
	}
}

func (s *Resource) Apply(schema string) error {
	SetCache(s.Cache)
	return s.Resource.Apply(schema)
}

func (s *Resource) Add(name string) {
	if IsLocalZipFile(name) {
		log.Debug("Skipping local zip file font:", name)
		return
	}

	s.Resource.Add(&Font{
		Name: name,
	})
}

func (s *Resource) Manifest() string {
	manifest := `{
  "$schema": "https://aka.ms/dsc/schemas/v3/bundled/resource/manifest.json",
  "description": "Allows configuring the Oh My Posh font installs.",
  "export": {
    "executable": "oh-my-posh",
    "input": "stdin",
    "args": [
      "font",
      "dsc",
      "export"
    ]
  },
  "get": {
    "executable": "oh-my-posh",
    "input": "stdin",
    "args": [
      "font",
      "dsc",
      "get"
    ]
  },
  "schema": {
    "command": {
      "executable": "oh-my-posh",
      "args": [
        "font",
        "dsc",
        "schema"
      ]
    }
  },
  "set": {
    "executable": "oh-my-posh",
    "implementsPretest": true,
    "return": "stateAndDiff",
    "args": [
      "font",
      "dsc",
      "set",
      {
        "jsonInputArg": "--schema",
        "mandatory": true
      }
    ]
  },
  "tags": [
    "OhMyPosh",
    "linux",
    "macos",
    "windows",
    "powershell",
    "terminal",
    "theming",
    "fonts"
  ],
  "type": "OhMyPosh/Font",
  "version": "0.1.0"
}`
	return dsc.CompressJSON(manifest)
}
