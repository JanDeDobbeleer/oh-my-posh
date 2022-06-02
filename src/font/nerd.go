package font

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

func Nerds() ([]*Asset, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/ryanoasis/nerd-fonts/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := client.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get nerd fonts release")
	}
	defer response.Body.Close()
	var release release
	err = json.NewDecoder(response.Body).Decode(&release)
	if err != nil {
		return nil, errors.New("failed to parse nerd fonts release")
	}
	var nerdFonts []*Asset
	for _, asset := range release.Assets {
		if asset.State == "uploaded" && strings.HasSuffix(asset.Name, ".zip") {
			asset.Name = strings.TrimSuffix(asset.Name, ".zip")
			nerdFonts = append(nerdFonts, asset)
		}
	}
	return nerdFonts, nil
}
