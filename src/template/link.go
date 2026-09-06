package template

import (
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
	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprint(v)
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

	// A URL a repository controls (a remote) can be anything; a bad one
	// drops the link rather than the whole segment.
	if strings.ContainsAny(url, "<>") {
		return labelMarkup(label), nil
	}

	if _, err := link.ParseRequestURI(url); err != nil {
		return labelMarkup(label), nil
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
