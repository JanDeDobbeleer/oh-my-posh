package upgrade

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/progress"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

type Config struct {
	Cache         cache.Cache    `json:"-" toml:"-" yaml:"-"`
	Source        Source         `json:"source" toml:"source" yaml:"source"`
	Interval      cache.Duration `json:"interval" toml:"interval" yaml:"interval"`
	Latest        string         `json:"-" toml:"-" yaml:"-"`
	Auto          bool           `json:"auto" toml:"auto" yaml:"auto"`
	DisplayNotice bool           `json:"notice" toml:"notice" yaml:"notice"`
	Force         bool           `json:"-" toml:"-" yaml:"-"`
}

type Source string

const (
	GitHub Source = "github"
	CDN    Source = "cdn"
)

func (s Source) String() string {
	switch s {
	case GitHub:
		return "github.com"
	case CDN:
		return "cdn.ohmyposh.dev"
	default:
		return "Unknown"
	}
}

func (cfg *Config) FetchLatest() (string, error) {
	cfg.Latest = "latest"
	v, err := cfg.DownloadAsset("version.txt")
	if err != nil {
		log.Debugf("failed to get latest version for source: %s", cfg.Source)
		return "", err
	}

	version := strings.TrimSpace(string(v))
	cfg.Latest = version

	version = strings.TrimPrefix(version, "v")
	log.Debugf("latest version: %s", version)

	return version, err
}

func (cfg *Config) DownloadAsset(asset string) ([]byte, error) {
	if cfg.Source == "" {
		log.Debug("no source specified, defaulting to github")
		cfg.Source = GitHub
	}

	switch cfg.Source {
	case GitHub:
		var url string

		switch cfg.Latest {
		case "latest":
			url = fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/%s", asset)
		default:
			url = fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/download/%s/%s", cfg.Latest, asset)
		}

		return cfg.Download(url)
	case CDN:
		fallthrough
	default:
		url := fmt.Sprintf("https://cdn.ohmyposh.dev/releases/%s/%s", cfg.Latest, asset)
		return cfg.Download(url)
	}
}

func (cfg *Config) Download(url string) ([]byte, error) {
	req, err := httplib.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		log.Debugf("failed to create request for url: %s", url)
		return nil, err
	}

	req.Header.Add("User-Agent", "oh-my-posh")
	req.Header.Add("Cache-Control", "max-age=0")

	resp, err := http.HTTPClient.Do(req)
	if err != nil {
		log.Debugf("failed to execute HTTP request: %s", url)
		return nil, err
	}

	if resp.StatusCode != httplib.StatusOK {
		return nil, fmt.Errorf("failed to download asset: %s", url)
	}

	defer resp.Body.Close()

	reader := progress.NewReader(resp.Body, resp.ContentLength, program)

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Debugf("failed to read response body: %s", url)
		return nil, err
	}

	return data, nil
}
