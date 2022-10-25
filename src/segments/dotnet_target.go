package segments

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
	"strings"
)

type DotnetTarget struct {
	props      properties.Properties
	env        environment.Environment
	extensions []string

	Text  string
	Error string
}

const (
	tfmRegex = "(?P<tag><.*TargetFramework.*>(?P<tfm>.*)</.*TargetFramework.*>)"
)

func (d *DotnetTarget) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ .Text }}{{ end }} "
}

func (d *DotnetTarget) Init(props properties.Properties, env environment.Environment) {
	d.props = props
	d.env = env
	d.extensions = []string{".csproj", ".vbproj", ".fsproj"}
}

func (d *DotnetTarget) Enabled() bool {
	files := d.getProjectFiles()
	if len(files) < 1 {
		return false
	}
	tfms := make([]string, len(files))
	displayError := d.props.GetBool(properties.DisplayError, false)
	for i, file := range files {
		tfm, err := d.extractTFM(file)
		if err != nil {
			d.Error = err.Error()
			return displayError
		}
		tfms[i] = tfm
	}
	d.Text = d.unifyTFMs(tfms)
	return true
}

func (d *DotnetTarget) getProjectFiles() []string {
	var files []string
	for _, extension := range d.extensions {
		cwd := d.env.Pwd()
		entries := d.env.LsDir(cwd)
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, extension) && !entry.IsDir() {
				files = append(files, name)
			}
		}
	}
	return files
}

func (d *DotnetTarget) extractTFM(file string) (string, error) {
	content := d.env.FileContent(file)
	values := regex.FindNamedRegexMatch(tfmRegex, content)
	if len(values) == 0 {
		return "", errors.New("cannot extract TFM from " + file)
	}
	return values["tfm"], nil
}

// If (somehow!!!) there will be several TFMs,
// this command will combine them and remove duplicates.
func (d *DotnetTarget) unifyTFMs(tfms []string) string {
	// The monikers inside <TargetFrameworks> tag are separated by a semicolon,
	// so we join them in a single string first
	tfmsStr := strings.Join(tfms, ";")
	// and then split to select the unique instances
	tfmsArr := strings.Split(tfmsStr, ";")
	uniqueArr := unique(tfmsArr)
	return strings.Join(uniqueArr, ";")
}

func unique(arr []string) []string {
	occurred := map[string]bool{}
	result := []string{}
	for e := range arr {
		if !occurred[arr[e]] {
			occurred[arr[e]] = true
			result = append(result, arr[e])
		}
	}

	return result
}
