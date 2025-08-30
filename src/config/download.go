package config

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

func download(url string) ([]byte, error) {
	defer log.Trace(time.Now(), url)

	// some users use the blob url, we need to convert it to the raw url
	themeBlob := "https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/"
	url = strings.Replace(url, themeBlob, "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/", 1)

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != httplib.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", response.StatusCode)
		log.Error(err)
		return nil, err
	}

	return io.ReadAll(response.Body)
}
