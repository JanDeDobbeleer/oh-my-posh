package font

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	httplib "net/http"
	"sort"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const (
	CascadiaCodeMS = "CascadiaCode (MS)"
)

type release struct {
	Assets []*Asset `json:"assets"`
}

type Asset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	State  string `json:"state"`
	Folder string `json:"folder"`
}

func (a Asset) FilterValue() string { return a.Name }

func IsLocalZipFile(name string) bool {
	return !strings.HasPrefix(name, "https") && strings.HasSuffix(name, ".zip")
}

func ResolveFontAsset(font string) (*Asset, error) {
	if strings.HasPrefix(font, "https") {
		return &Asset{URL: font}, nil
	}

	fonts, err := fonts()
	if err != nil {
		return nil, err
	}

	var asset *Asset
	for _, f := range fonts {
		if !strings.EqualFold(font, f.Name) {
			continue
		}

		asset = f
		break
	}

	if asset == nil {
		return nil, fmt.Errorf("no matching font found")
	}

	return asset, nil
}

func fonts() ([]*Asset, error) {
	if assets, err := getCachedFontData(); err == nil {
		return assets, nil
	}

	assets, err := fetchFontAssets("ryanoasis/nerd-fonts")
	if err != nil {
		return nil, err
	}

	cascadiaCode, err := CascadiaCode()
	if err == nil {
		assets = append(assets, cascadiaCode)
	}

	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })

	cache.Set(cache.Device, cache.FONTLISTCACHE, assets, cache.ONEDAY)

	return assets, nil
}

func getCachedFontData() ([]*Asset, error) {
	list, OK := cache.Get[[]*Asset](cache.Device, cache.FONTLISTCACHE)
	if !OK {
		return nil, errors.New("cache not found")
	}

	return list, nil
}

func CascadiaCode() (*Asset, error) {
	assets, err := fetchFontAssets("microsoft/cascadia-code")
	if err != nil || len(assets) != 1 {
		return nil, errors.New("no assets found")
	}

	return &Asset{
		Name:   CascadiaCodeMS,
		URL:    assets[0].URL,
		Folder: "ttf/",
	}, nil
}

func fetchFontAssets(repo string) ([]*Asset, error) {
	ctx, cancelF := context.WithTimeout(context.Background(), time.Second*time.Duration(20))
	defer cancelF()

	repoURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	req, err := httplib.NewRequestWithContext(ctx, "GET", repoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := http.HTTPClient.Do(req)
	if err != nil || response.StatusCode != httplib.StatusOK {
		return nil, fmt.Errorf("failed to get %s release", repo)
	}

	defer response.Body.Close()

	var release release
	err = json.NewDecoder(response.Body).Decode(&release)
	if err != nil {
		return nil, errors.New("failed to parse nerd fonts release")
	}

	var fonts []*Asset
	for _, asset := range release.Assets {
		if asset.State == "uploaded" && strings.HasSuffix(asset.Name, ".zip") {
			asset.Name = strings.TrimSuffix(asset.Name, ".zip")
			fonts = append(fonts, asset)
		}
	}

	return fonts, nil
}
