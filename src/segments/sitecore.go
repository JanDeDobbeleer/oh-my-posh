package segments

import (
	"encoding/json"
	"path"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	sitecoreFileName   = "sitecore.json"
	sitecoreFolderName = ".sitecore"
	userFileName       = "user.json"
	defaultEnpointName = "default"
)

type Sitecore struct {
	props properties.Properties
	env   platform.Environment

	EndpointName string
	CmHost       string
}

type EndpointConfig struct {
	Host string `json:"host"`
}

type UserConfig struct {
	DefaultEndpoint string                    `json:"defaultEndpoint"`
	Endpoints       map[string]EndpointConfig `json:"endpoints"`
}

func (s *Sitecore) Enabled() bool {
	if !s.env.HasFiles(sitecoreFileName) || !s.env.HasFiles(path.Join(sitecoreFolderName, userFileName)) {
		return false
	}

	var userConfig, err = getUserConfig(s)

	if err != nil {
		return false
	}

	s.EndpointName = userConfig.getDefaultEndpoint()

	displayDefault := s.props.GetBool(properties.DisplayDefault, true)

	if !displayDefault && s.EndpointName == defaultEnpointName {
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

func (s *Sitecore) Init(props properties.Properties, env platform.Environment) {
	s.props = props
	s.env = env
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
