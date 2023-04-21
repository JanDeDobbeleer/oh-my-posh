package font

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
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
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", location, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
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
