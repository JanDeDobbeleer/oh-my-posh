package template

import (
	"fmt"
	link "net/url"
)

func url(text, url string) (string, error) {
	_, err := link.ParseRequestURI(url)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[%s](%s)", text, url), nil
}

func path(text, path string) (string, error) {
	return fmt.Sprintf("[%s](file:%s)", text, path), nil
}
