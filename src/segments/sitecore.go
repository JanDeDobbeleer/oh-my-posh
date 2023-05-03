package segments

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Sitecore struct {
	props properties.Properties
	env   platform.Environment

	Environment string
	Cloud       bool
}

type SitecoreConfig struct {
	Endpoints struct {
		Default struct {
			Ref        string `json:"ref"`
			AllowWrite bool   `json:"allowWrite"`
			Host       string `json:"host"`
			Variables  struct {
			} `json:"variables"`
		} `json:"default"`
	} `json:"endpoints"`
}

func (s *Sitecore) Enabled() bool {
	return s.shouldDisplay()
}

func (s *Sitecore) Template() string {
	return " {{ if .Cloud }}\uf65e{{ else }}\uf98a{{ end }} {{ .Environment }} "
}

func (s *Sitecore) Init(props properties.Properties, env platform.Environment) {
	s.props = props
	s.env = env
}

func (s *Sitecore) shouldDisplay() bool {
	sitecoreDir, err := s.env.HasParentFilePath(".sitecore")
	if err != nil {
		return false
	}

	if !sitecoreDir.IsDir {
		return false
	}

	if !s.env.HasFilesInDir(sitecoreDir.Path, "user.json") {
		return false
	}

	sitecoreConfigFile := filepath.Join(sitecoreDir.Path, "user.json")

	if len(sitecoreConfigFile) == 0 {
		return false
	}

	content := s.env.FileContent(sitecoreConfigFile)

	var config SitecoreConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return false
	}

	// sitecore xm cloud always has sitecorecloud.io as domain
	s.Cloud = strings.Contains(config.Endpoints.Default.Host, "sitecorecloud.io")

	s.Environment = config.Endpoints.Default.Host

	return true
}
