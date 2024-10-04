package config

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

func Download(cachePath, url string) (string, error) {
	defer log.Trace(time.Now(), cachePath, url)

	// some users use the blob url, we need to convert it to the raw url
	themeBlob := "https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/"
	url = strings.Replace(url, themeBlob, "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/", 1)

	configPath, shouldUpdate := shouldUpdate(cachePath, url)
	if !shouldUpdate {
		return configPath, nil
	}

	log.Debug("downloading config from ", url, " to ", configPath)

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return "", err
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != httplib.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", response.StatusCode)
		log.Error(err)
		return "", err
	}

	if len(configPath) == 0 {
		configPath = formatConfigPath(url, response.Header.Get("Etag"), cachePath)
		log.Debug("config path not set yet, using ", configPath)
	}

	out, err := os.Create(configPath)
	if err != nil {
		log.Error(err)
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}

	log.Debug("config updated to ", configPath)

	return configPath, nil
}

func shouldUpdate(cachePath, url string) (string, bool) {
	defer log.Trace(time.Now(), cachePath, url)

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodHead, url, nil)
	if err != nil {
		log.Error(err)
		return "", true
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return "", true
	}

	defer response.Body.Close()

	etag := response.Header.Get("Etag")
	if len(etag) == 0 {
		log.Debug("no etag found, updating config")
		return "", true
	}

	configPath := formatConfigPath(url, etag, cachePath)

	_, err = os.Stat(configPath)
	if err != nil {
		log.Debug("configfile ", configPath, " doest not exist, updating config")
		return configPath, true
	}

	log.Debug("config found at", configPath, " skipping update")
	return configPath, false
}

func formatConfigPath(url, etag, cachePath string) string {
	ext := filepath.Ext(url)
	etag = strings.TrimLeft(etag, `W/`)
	etag = strings.Trim(etag, `"`)
	filename := fmt.Sprintf("config.%s.omp%s", etag, ext)
	return filepath.Join(cachePath, filename)
}
