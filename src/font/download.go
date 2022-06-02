package font

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	contentType := http.DetectContentType(data)
	return contentType == "application/zip"
}

func getRemoteFile(location string) (data []byte, err error) {
	print("Downloading %s", location)
	var client = http.Client{}

	resp, err := client.Get(location)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}
