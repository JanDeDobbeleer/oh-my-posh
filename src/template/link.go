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
	return fmt.Sprintf("<LINK>%s<TEXT>%s</TEXT></LINK>", url, text), nil
}

func path(text, path string) (string, error) {
	return fmt.Sprintf("<LINK>file:%s<TEXT>%s</TEXT></LINK>", path, text), nil
}
