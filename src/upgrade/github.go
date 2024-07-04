package upgrade

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

func downloadReleaseAsset(tag, asset string) ([]byte, error) {
	url := fmt.Sprintf("https://github.com/JanDeDobbeleer/oh-my-posh/releases/download/%s/%s", tag, asset)

	req, err := httplib.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "oh-my-posh")

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
