package http

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func Download(url string, isCacheEnabled bool) ([]byte, error) {
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

	request.Header.Add("User-Agent", "oh-my-posh")
	// if we have an etag, add it to the request to check if the file changed
	etag, OK := cache.Get[string](cache.Device, etagKey(url))
	if OK {
		log.Debugf("found etag in cache: %s", etag)
		request.Header.Set("If-None-Match", etag)
	}

	response, err := HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode == httplib.StatusNotModified {
		log.Debug("resource not modified, using cached version")
		cachedData, OK := cache.Get[[]byte](cache.Device, dataKey(url))
		if OK {
			return cachedData, nil
		}

		return nil, fmt.Errorf("resource not modified but no cached data found")
	}

	if response.StatusCode != httplib.StatusOK {
		err := fmt.Errorf("status code: %d", response.StatusCode)
		log.Error(err)
		return nil, err
	}

	etag = response.Header.Get("ETag")
	if etag != "" && isCacheEnabled {
		cache.Set(cache.Device, etagKey(url), etag, cache.INFINITE)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if isCacheEnabled {
		cache.Set(cache.Device, dataKey(url), data, cache.INFINITE)
	}

	return data, nil
}

func etagKey(url string) string {
	return fmt.Sprintf("%s.etag", url)
}

func dataKey(url string) string {
	return fmt.Sprintf("%s.data", url)
}
