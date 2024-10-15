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

type release struct {
	Assets []*Asset `json:"assets"`
}

type Asset struct {
	Name  string `json:"name"`
	URL   string `json:"browser_download_url"`
	State string `json:"state"`
}

func (a Asset) FilterValue() string { return a.Name }

func Fonts() ([]*Asset, error) {
	if assets, err := getCachedFontData(); err == nil {
		return assets, nil
	}

	assets, err := fetchFontAssets("ryanoasis/nerd-fonts")
	if err != nil {
		return nil, err
	}

	cascadiaCode, err := fetchFontAssets("microsoft/cascadia-code")
	if err != nil {
		return assets, nil
	}

	assets = append(assets, cascadiaCode...)
	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })

	setCachedFontData(assets)

	return assets, nil
}

func getCachedFontData() ([]*Asset, error) {
	if environment == nil {
		return nil, errors.New("environment not set")
	}

	list, OK := environment.Cache().Get(cache.FONTLISTCACHE)
	if !OK {
		return nil, errors.New("cache not found")
	}

	assets := make([]*Asset, 0)
	err := json.Unmarshal([]byte(list), &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func setCachedFontData(assets []*Asset) {
	if environment == nil {
		return
	}

	data, err := json.Marshal(assets)
	if err != nil {
		return
	}

	environment.Cache().Set(cache.FONTLISTCACHE, string(data), "1day")
}

func CascadiaCode() ([]*Asset, error) {
	return fetchFontAssets("microsoft/cascadia-code")
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
