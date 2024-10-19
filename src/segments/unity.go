package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Unity struct {
	props properties.Properties
	env   runtime.Environment

	UnityVersion  string
	CSharpVersion string
}

func (u *Unity) Enabled() bool {
	unityVersion, err := u.GetUnityVersion()
	if err != nil {
		u.env.Error(err)
		return false
	}
	if len(unityVersion) == 0 {
		return false
	}
	u.UnityVersion = unityVersion

	csharpVersion, err := u.GetCSharpVersion()
	if err != nil {
		u.env.Error(err)
	}
	u.CSharpVersion = csharpVersion

	return true
}

func (u *Unity) GetUnityVersion() (string, error) {
	projectDir, err := u.env.HasParentFilePath("ProjectSettings", false)
	if err != nil {
		u.env.Debug("no ProjectSettings parent folder found")
		return "", err
	}

	if !u.env.HasFilesInDir(projectDir.Path, "ProjectVersion.txt") {
		u.env.Debug("no ProjectVersion.txt file found")
		return "", err
	}

	versionFilePath := filepath.Join(projectDir.Path, "ProjectVersion.txt")
	versionFileText := u.env.FileContent(versionFilePath)

	lines := strings.Split(versionFileText, "\n")
	versionPrefix := "m_EditorVersion: "
	for _, line := range lines {
		if !strings.HasPrefix(line, versionPrefix) {
			continue
		}
		version := strings.TrimPrefix(line, versionPrefix)
		version = strings.TrimSpace(version)
		if len(version) == 0 {
			return "", errors.New("Empty m_EditorVersion")
		}
		fIndex := strings.Index(version, "f")
		if fIndex > 0 {
			return version[:fIndex], nil
		}
		return version, nil
	}

	return "", errors.New("ProjectSettings/ProjectVersion.txt is missing m_EditorVersion")
}

func (u *Unity) GetCSharpVersion() (version string, err error) {
	lastDotIndex := strings.LastIndex(u.UnityVersion, ".")
	if lastDotIndex == -1 {
		return "", errors.New("lastDotIndex")
	}
	shortUnityVersion := u.UnityVersion[0:lastDotIndex]

	var csharpVersionsByUnityVersion = map[string]string{
		"2017.1": "C# 6",
		"2017.2": "C# 6",
		"2017.3": "C# 6",
		"2017.4": "C# 6",
		"2018.1": "C# 6",
		"2018.2": "C# 6",
		"2018.3": "C# 7.3",
		"2018.4": "C# 7.3",
		"2019.1": "C# 7.3",
		"2019.2": "C# 7.3",
		"2019.3": "C# 7.3",
		"2019.4": "C# 7.3",
		"2020.1": "C# 7.3",
		"2020.2": "C# 8",
		"2020.3": "C# 8",
		"2021.1": "C# 8",
		"2021.2": "C# 9",
		"2021.3": "C# 9",
		"2022.1": "C# 9",
		"2022.2": "C# 9",
		"2023.1": "C# 9",
		"2023.2": "C# 9",
	}

	csharpVersion, found := csharpVersionsByUnityVersion[shortUnityVersion]
	if found {
		return csharpVersion, nil
	}

	u.env.Debug(fmt.Sprintf("Unity version %s doesn't exist in the map", shortUnityVersion))
	return u.GetCSharpVersionFromWeb(shortUnityVersion)
}

func (u *Unity) GetCSharpVersionFromWeb(shortUnityVersion string) (version string, err error) {
	url := fmt.Sprintf("https://docs.unity3d.com/%s/Documentation/Manual/CSharpCompiler.html", shortUnityVersion)
	httpTimeout := u.props.GetInt(properties.HTTPTimeout, 2000)

	body, err := u.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return "", err
	}

	pageContent := string(body)

	pattern := `<a href="https://(?:docs|learn)\.microsoft\.com/en-us/dotnet/csharp/whats-new/csharp-[0-9]+-?[0-9]*">(?P<csharpVersion>.*)</a>`
	matches := regex.FindNamedRegexMatch(pattern, pageContent)
	if matches != nil && matches["csharpVersion"] != "" {
		csharpVersion := strings.TrimSuffix(matches["csharpVersion"], ".0")
		return csharpVersion, nil
	}

	return "", nil
}

func (u *Unity) Template() string {
	return " \ue721 {{ .UnityVersion }}{{ if .CSharpVersion }} {{ .CSharpVersion }}{{ end }} "
}

func (u *Unity) Init(props properties.Properties, env runtime.Environment) {
	u.props = props
	u.env = env
}
