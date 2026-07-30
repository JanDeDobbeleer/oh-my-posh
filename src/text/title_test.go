package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitle(t *testing.T) {
	// expected values are what x/text's cases.Title(language.English)
	// produced for the same input before it was replaced
	cases := map[string]string{
		"git":         "Git",
		"nix-shell":   "Nix-Shell",
		"ui5tooling":  "Ui5tooling",
		"GIT":         "Git",
		"foo_bar":     "Foo_bar",
		"os":          "Os",
		"context":     "Context",
		"Nice2see":    "Nice2see",
		"hello world": "Hello World",
		"a-1b":        "A-1B",
		"":            "",
	}

	for input, expected := range cases {
		assert.Equal(t, expected, Title(input), input)
	}
}
