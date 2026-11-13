package segments

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const (
	uprojectExtension = ".uproject"
	// Launcher builds register under HKLM keyed by version, source builds under HKCU keyed by GUID.
	unrealLauncherKey = `HKLM\SOFTWARE\EpicGames\Unreal Engine\%s\InstalledDirectory`
	unrealBuildsKey   = `HKCU\Software\Epic Games\Unreal Engine\Builds\%s`
)

type Unreal struct {
	Base

	Version string
}

type unrealProject struct {
	EngineAssociation string `json:"EngineAssociation"`
}

type unrealBuildVersion struct {
	MajorVersion int `json:"MajorVersion"`
	MinorVersion int `json:"MinorVersion"`
	PatchVersion int `json:"PatchVersion"`
}

func (u *Unreal) Template() string {
	return "  {{ .Version }} "
}

func (u *Unreal) Enabled() bool {
	dir := u.env.Pwd()
	for {
		if version := u.versionIn(dir); version != "" {
			u.Version = version
			return true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

func (u *Unreal) versionIn(dir string) string {
	if u.env.HasFilesInDir(dir, "*"+uprojectExtension) {
		return u.projectVersion(dir)
	}

	return u.engineVersion(dir)
}

func (u *Unreal) projectVersion(dir string) string {
	for _, entry := range u.env.LsDir(dir) {
		name := entry.Name()
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(name), uprojectExtension) {
			continue
		}

		var project unrealProject
		if err := json.Unmarshal([]byte(u.env.FileContent(filepath.Join(dir, name))), &project); err != nil {
			log.Error(err)
			return ""
		}

		association := project.EngineAssociation
		if version := u.engineVersion(u.engineRoot(association)); version != "" {
			return version
		}

		return association
	}

	return ""
}

func (u *Unreal) engineRoot(association string) string {
	if association == "" || u.env.GOOS() != runtime.WINDOWS {
		return ""
	}

	key := unrealLauncherKey
	if strings.HasPrefix(association, "{") {
		key = unrealBuildsKey
	}

	value, err := u.env.WindowsRegistryKeyValue(fmt.Sprintf(key, association))
	if err != nil {
		log.Error(err)
		return ""
	}

	return value.String
}

func (u *Unreal) engineVersion(root string) string {
	if root == "" {
		return ""
	}

	content := u.env.FileContent(filepath.Join(root, "Engine", "Build", "Build.version"))
	if content == "" {
		return ""
	}

	var build unrealBuildVersion
	if err := json.Unmarshal([]byte(content), &build); err != nil {
		log.Error(err)
		return ""
	}

	return fmt.Sprintf("%d.%d.%d", build.MajorVersion, build.MinorVersion, build.PatchVersion)
}
