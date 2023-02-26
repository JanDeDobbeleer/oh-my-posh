package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Unity struct {
	props properties.Properties
	env   platform.Environment

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

func (u *Unity) GetUnityVersion() (version string, err error) {
	projectDir, err := u.env.HasParentFilePath("ProjectSettings")
	if err != nil {
		u.env.Debug("No ProjectSettings parent folder found")
		return
	}

	if !u.env.HasFilesInDir(projectDir.Path, "ProjectVersion.txt") {
		u.env.Debug("No ProjectVersion.txt file found")
		return
	}

	versionFilePath := filepath.Join(projectDir.Path, "ProjectVersion.txt")
	versionFileText := u.env.FileContent(versionFilePath)

	firstLine := strings.Split(versionFileText, "\n")[0]
	versionPrefix := "m_EditorVersion: "
	versionPrefixIndex := strings.Index(firstLine, versionPrefix)
	if versionPrefixIndex == -1 {
		err := errors.New("ProjectSettings/ProjectVersion.txt is missing 'm_EditorVersion: ' prefix")
		return "", err
	}

	versionStartIndex := versionPrefixIndex + len(versionPrefix)
	unityVersion := firstLine[versionStartIndex:]

	return strings.TrimSuffix(unityVersion, "f1"), nil
}

func (u *Unity) GetCSharpVersion() (version string, err error) {
	lastDotIndex := strings.LastIndex(u.UnityVersion, ".")
	if lastDotIndex == -1 {
		return "", errors.New("lastDotIndex")
	}
	shortUnityVersion := u.UnityVersion[0:lastDotIndex]

	if val, found := u.env.Cache().Get(shortUnityVersion); found {
		csharpVersion := strings.TrimSuffix(val, ".0")
		return csharpVersion, nil
	}

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
		u.env.Cache().Set(shortUnityVersion, csharpVersion, -1)
		return csharpVersion, nil
	}

	u.env.Cache().Set(shortUnityVersion, "", -1)
	return "", nil
}

func (u *Unity) Template() string {
	return " \ue721 {{ .UnityVersion }}{{ if .CSharpVersion }} {{ .CSharpVersion }}{{ end }} "
}

func (u *Unity) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env
}
