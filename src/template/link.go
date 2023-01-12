package template

import (
	"fmt"
	link "net/url"
)

// url builds an hyperlink if url is not empty, otherwise returns the text only
func url(text, url string) (string, error) {
	if url == "" {
		return text, nil
	}
	_, err := link.ParseRequestURI(url)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("«%s»(%s)", text, url), nil
}

func path(text, path string) (string, error) {
	return fmt.Sprintf("«%s»(file:%s)", text, path), nil
}
