package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestUrl(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "valid url", Expected: "<LINK>https://ohmyposh.dev<TEXT>link</TEXT></LINK>", Template: `{{ url "link" "https://ohmyposh.dev" }}`},
		{Case: "invalid url", Expected: "", Template: `{{ url "link" "Foo" }}`, ShouldError: true},
	}

	env := &mock.Environment{}
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)

	Init(env, nil)

	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
		}

		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}

		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestPath(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{Case: "valid path", Expected: "<LINK>file:/test/test<TEXT>link</TEXT></LINK>", Template: `{{ path "link" "/test/test" }}`},
	}

	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
		}

		text, _ := tmpl.Render()
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
