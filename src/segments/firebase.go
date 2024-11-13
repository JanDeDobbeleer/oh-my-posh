package segments

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const (
	FIREBASENOACTIVECONFIG = "NO ACTIVE CONFIG FOUND"
)

type Firebase struct {
	base

	Project string
}

type FirebaseData struct {
	ActiveProject map[string]string `json:"activeProjects"`
}

func (f *Firebase) Template() string {
	return " {{ .Project}} "
}

func (f *Firebase) Enabled() bool {
	cfgDir := filepath.Join(f.env.Home(), ".config", "configstore")
	configFile, err := f.getActiveConfig(cfgDir)
	if err != nil {
		log.Error(err)
		return false
	}

	data, err := f.getFirebaseData(configFile)
	if err != nil {
		log.Error(err)
		return false
	}

	// Within the activeProjects is a key value pair
	// of the path to the project and the project name

	// Test if the current directory is a project path
	// and if it is, return the project name
	for key, value := range data.ActiveProject {
		if strings.HasPrefix(f.env.Pwd(), key) {
			f.Project = value
			return true
		}
	}

	return false
}

func (f *Firebase) getActiveConfig(cfgDir string) (string, error) {
	activeConfigFile := filepath.Join(cfgDir, "firebase-tools.json")
	activeConfigData := f.env.FileContent(activeConfigFile)
	if len(activeConfigData) == 0 {
		return "", errors.New(FIREBASENOACTIVECONFIG)
	}
	return activeConfigData, nil
}

func (f *Firebase) getFirebaseData(configFile string) (*FirebaseData, error) {
	var data FirebaseData

	err := json.Unmarshal([]byte(configFile), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
