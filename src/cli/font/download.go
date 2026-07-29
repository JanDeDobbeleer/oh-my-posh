package font

import (
	"context"
	"errors"
	"fmt"
	"io"
	httplib "net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

// download fetches a font zip, reporting how much has arrived so a caller can draw a bar. report
// may be nil, which is what the DSC path and the tests pass.
func download(fontURL string, report func(fraction float64)) ([]byte, error) {
	if zipPath, OK := cache.Get[string](cache.Device, fontURL); OK {
		if b, err := os.ReadFile(zipPath); err == nil {
			return b, nil
		}
	}

	// validate if we have a local file
	u, err := url.Parse(fontURL)
	if err != nil || u.Scheme != "https" {
		return nil, errors.New("font path must be a valid URL")
	}

	var b []byte
	if b, err = getRemoteFile(fontURL, report); err != nil {
		return nil, err
	}

	if !isZipFile(b) {
		return nil, fmt.Errorf("%s is not a valid zip file", fontURL)
	}

	fileName := path.Base(fontURL)

	zipPath := filepath.Join(os.TempDir(), fileName)
	tempFile, err := os.Create(zipPath)
	defer func() {
		_ = tempFile.Close()
	}()

	if err != nil {
		return b, nil
	}

	_, err = tempFile.Write(b)
	if err != nil {
		return b, nil
	}

	cache.Set(cache.Device, fontURL, zipPath, cache.ONEDAY)

	return b, nil
}

func isZipFile(data []byte) bool {
	contentType := httplib.DetectContentType(data)
	return contentType == "application/zip"
}

func getRemoteFile(location string, report func(fraction float64)) (data []byte, err error) {
	req, err := httplib.NewRequestWithContext(context.Background(), "GET", location, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.HTTPClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != httplib.StatusOK {
		return data, fmt.Errorf("failed to download zip file: %s\n→ %s", resp.Status, location)
	}

	reader := ui.NewReader(resp.Body, resp.ContentLength, report)

	data, err = io.ReadAll(reader)
	if err != nil {
		return
	}

	return
}

// Download keeps the old name for callers that need no progress reporting (see font.Apply, which
// runs under DSC with nothing drawing).
func Download(fontURL string) ([]byte, error) {
	return download(fontURL, nil)
}
