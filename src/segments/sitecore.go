package segments

import (
	"encoding/json"
	"path"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	sitecoreFileName   = "sitecore.json"
	sitecoreFolderName = ".sitecore"
	userFileName       = "user.json"
	defaultEnpointName = "default"
)

type Sitecore struct {
	base

	EndpointName string
	CmHost       string
}

type EndpointConfig struct {
	Host string `json:"host"`
}

type UserConfig struct {
	Endpoints       map[string]EndpointConfig `json:"endpoints"`
	DefaultEndpoint string                    `json:"defaultEndpoint"`
}

func (s *Sitecore) Enabled() bool {
	if !s.env.HasFiles(sitecoreFileName) || !s.env.HasFilesInDir(sitecoreFolderName, userFileName) {
		log.Debug("sitecore cli configuration files were not found")
		return false
	}

	var userConfig, err = getUserConfig(s)

	if err != nil {
		log.Error(err)
		return false
	}

	s.EndpointName = userConfig.getDefaultEndpoint()

	displayDefault := s.props.GetBool(properties.DisplayDefault, true)

	if !displayDefault && s.EndpointName == defaultEnpointName {
		log.Debug("displaying of the default environment is turned off")
		return false
	}

	if endpoint := userConfig.getEndpoint(s.EndpointName); endpoint != nil && len(endpoint.Host) > 0 {
		s.CmHost = endpoint.Host
	}

	return true
}

func (s *Sitecore) Template() string {
	return "{{ .EndpointName }} {{ if .CmHost }}({{ .CmHost }}){{ end }}"
}

func getUserConfig(s *Sitecore) (*UserConfig, error) {
	userJSON := s.env.FileContent(path.Join(sitecoreFolderName, userFileName))
	var userConfig UserConfig

	if err := json.Unmarshal([]byte(userJSON), &userConfig); err != nil {
		return nil, err
	}

	return &userConfig, nil
}

func (u *UserConfig) getDefaultEndpoint() string {
	if len(u.DefaultEndpoint) > 0 {
		return u.DefaultEndpoint
	}

	return defaultEnpointName
}

func (u *UserConfig) getEndpoint(name string) *EndpointConfig {
	endpoint, exists := u.Endpoints[name]

	if exists {
		return &endpoint
	}

	return nil
}
