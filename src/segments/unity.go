package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Unity struct {
	base

	UnityVersion  string
	CSharpVersion string
}

func (u *Unity) Template() string {
	return " \ue721 {{ .UnityVersion }}{{ if .CSharpVersion }} {{ .CSharpVersion }}{{ end }} "
}

func (u *Unity) Enabled() bool {
	unityVersion, err := u.GetUnityVersion()
	if err != nil {
		log.Error(err)
		return false
	}
	if len(unityVersion) == 0 {
		return false
	}
	u.UnityVersion = unityVersion

	csharpVersion, err := u.GetCSharpVersion()
	if err != nil {
		log.Error(err)
	}
	u.CSharpVersion = csharpVersion

	return true
}

func (u *Unity) GetUnityVersion() (string, error) {
	projectDir, err := u.env.HasParentFilePath("ProjectSettings", false)
	if err != nil {
		log.Debug("no ProjectSettings parent folder found")
		return "", err
	}

	if !u.env.HasFilesInDir(projectDir.Path, "ProjectVersion.txt") {
		log.Debug("no ProjectVersion.txt file found")
		return "", err
	}

	versionFilePath := filepath.Join(projectDir.Path, "ProjectVersion.txt")
	versionFileText := u.env.FileContent(versionFilePath)

	lines := strings.SplitSeq(versionFileText, "\n")
	versionPrefix := "m_EditorVersion: "
	for line := range lines {
		if !strings.HasPrefix(line, versionPrefix) {
			continue
		}
		version := strings.TrimPrefix(line, versionPrefix)
		version = strings.TrimSpace(version)
		if len(version) == 0 {
			return "", errors.New("empty m_EditorVersion")
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

	log.Debug(fmt.Sprintf("Unity version %s doesn't exist in the map", shortUnityVersion))
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
