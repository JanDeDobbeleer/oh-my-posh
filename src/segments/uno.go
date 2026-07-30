package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type unoGlobalJSON struct {
	MSBuildSDKs map[string]string `json:"msbuild-sdks"`
}

type Uno struct {
	Base

	Version string
}

func (u *Uno) Template() string {
	return " {{ .Version }} "
}

func (u *Uno) Enabled() bool {
	file, err := u.env.HasParentFilePath("global.json", false)
	if err != nil {
		return false
	}

	content := text.StripJSONComments(u.env.FileContent(file.Path))

	var globalJSON unoGlobalJSON
	err = json.Unmarshal([]byte(content), &globalJSON)
	if err != nil {
		return false
	}

	u.Version = globalJSON.MSBuildSDKs["Uno.Sdk"]
	return len(u.Version) > 0
}
