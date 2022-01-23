package main

import (
	"encoding/json"
	"fmt"
)

type angular struct {
	language
}

func (a *angular) string() string {
	segmentTemplate := a.language.props.getString(SegmentTemplate, "{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}")
	return a.language.string(segmentTemplate, a)
}

func (a *angular) init(props Properties, env Environment) {
	a.language = language{
		env:        env,
		props:      props,
		extensions: []string{"angular.json"},
		commands: []*cmd{
			{
				regex: `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
				getVersion: func() (string, error) {
					const fileName string = "package.json"
					const fileFolder string = "/node_modules/@angular/core"
					angularFilePath := a.language.env.pwd() + fileFolder
					if !a.language.env.hasFilesInDir(angularFilePath, fileName) {
						return "", fmt.Errorf("%s not found in %s", fileName, angularFilePath)
					}
					// parse file
					objmap := map[string]json.RawMessage{}
					content := a.language.env.getFileContent(a.language.env.pwd() + fileFolder + "/" + fileName)
					err := json.Unmarshal([]byte(content), &objmap)
					if err != nil {
						return "", err
					}
					var str string
					err = json.Unmarshal(objmap["version"], &str)
					if err != nil {
						return "", err
					}
					return str, nil
				},
			},
		},
		versionURLTemplate: "https://github.com/angular/angular/releases/tag/{{.Full}}",
	}
}

func (a *angular) enabled() bool {
	return a.language.enabled()
}
