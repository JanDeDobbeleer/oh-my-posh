package template

import (
	"errors"
	"fmt"
	link "net/url"
	"slices"
	"strings"
)

// labelMarkup keeps a Markup label's anchors intact (e.g. an icon field)
// while escaping plain data strings.
func labelMarkup(v any) Markup {
	switch l := v.(type) {
	case Markup:
		return l
	case string:
		return EscapeMarkup(l)
	default:
		return EscapeMarkup(fmt.Sprint(v))
	}
}

func textValue(v any) string {
	switch s := v.(type) {
	case Markup:
		return string(s)
	case string:
		return s
	default:
		return fmt.Sprint(v)
	}
}

func url(label, rawURL any) (Markup, error) {
	unsupported := []string{elvish, xonsh}
	if slices.Contains(unsupported, shell) {
		return labelMarkup(label), nil
	}

	url := textValue(rawURL)
	if url == "" {
		return labelMarkup(label), nil
	}

	if strings.ContainsAny(url, "<>") {
		return "", errors.New("url contains chevrons")
	}

	_, err := link.ParseRequestURI(url)
	if err != nil {
		return "", err
	}

	return RawMarkup(fmt.Sprintf("<LINK>%s<TEXT>%s</TEXT></LINK>", url, labelMarkup(label))), nil
}

func filePath(label, path any) (Markup, error) {
	unsupported := []string{elvish, xonsh}
	if slices.Contains(unsupported, shell) {
		return labelMarkup(label), nil
	}

	encodedPath := (&link.URL{Path: textValue(path)}).EscapedPath()

	return RawMarkup(fmt.Sprintf("<LINK>file:%s<TEXT>%s</TEXT></LINK>", encodedPath, labelMarkup(label))), nil
}
