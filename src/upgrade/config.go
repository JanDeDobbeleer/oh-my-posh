package upgrade

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

type Config struct {
	Cache         cache.Cache    `json:"-" toml:"-"`
	Source        Source         `json:"source" toml:"source"`
	Interval      cache.Duration `json:"interval" toml:"interval"`
	Version       string         `json:"-" toml:"-"`
	Auto          bool           `json:"auto" toml:"auto"`
	DisplayNotice bool           `json:"notice" toml:"notice"`
	Force         bool           `json:"-" toml:"-"`
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

func (cfg *Config) Latest() (string, error) {
	cfg.Version = "latest"
	v, err := cfg.DownloadAsset("version.txt")
	version := strings.TrimSpace(string(v))
	return strings.TrimPrefix(version, "v"), err
}

func (cfg *Config) DownloadAsset(asset string) ([]byte, error) {
	if len(cfg.Source) == 0 {
		cfg.Source = GitHub
	}

	switch cfg.Source {
	case GitHub:
		var url string

		switch cfg.Version {
		case "latest":
			url = fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/%s", asset)
		default:
			url = fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/download/%s/%s", cfg.Version, asset)
		}

		return cfg.Download(url)
	case CDN:
		fallthrough
	default:
		url := fmt.Sprintf("https://cdn.ohmyposh.dev/releases/%s/%s", cfg.Version, asset)
		return cfg.Download(url)
	}
}

func (cfg *Config) Download(url string) ([]byte, error) {
	req, err := httplib.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "oh-my-posh")
	req.Header.Add("Cache-Control", "max-age=0")

	resp, err := http.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != httplib.StatusOK {
		return nil, fmt.Errorf("failed to download asset: %s", url)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
