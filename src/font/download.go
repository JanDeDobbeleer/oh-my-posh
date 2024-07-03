package font

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"context"
	"errors"
	"fmt"
	"io"
	httplib "net/http"
	"net/url"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

func Download(fontPath string) ([]byte, error) {
	u, err := url.Parse(fontPath)
	if err != nil || u.Scheme != "https" {
		return nil, errors.New("font path must be a valid URL")
	}

	var b []byte
	if b, err = getRemoteFile(fontPath); err != nil {
		return nil, err
	}

	if !isZipFile(b) {
		return nil, fmt.Errorf("%s is not a valid zip file", fontPath)
	}

	return b, nil
}

func isZipFile(data []byte) bool {
	contentType := httplib.DetectContentType(data)
	return contentType == "application/zip"
}

func getRemoteFile(location string) (data []byte, err error) {
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
		return data, fmt.Errorf("Failed to download zip file: %s\n→ %s", resp.Status, location)
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}
