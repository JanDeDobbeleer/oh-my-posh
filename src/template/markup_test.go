package template

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderEscapesActionOutput(t *testing.T) {
	type ctx struct {
		Data   string
		Markup Markup
	}

	cases := []struct {
		Case     string
		Template string
		Context  any
		Expected string
	}{
		{
			Case:     "chevrons in data are neutralized",
			Template: `{{ .Data }}`,
			Context:  ctx{Data: "<red>pwned</>"},
			Expected: "<<>red<>>pwned<<>/<>>",
		},
		{
			Case:     "hyperlink injection from data is neutralized",
			Template: `[{{ .Data }}] 100% done`,
			Context:  ctx{Data: "<LINK>https://attacker.example/pwn<TEXT>label</TEXT>"},
			Expected: "[<<>LINK<>>https://attacker.example/pwn<<>TEXT<>>label<<>/TEXT<>>] 100% done",
		},
		{
			Case:     "literal anchors in template text pass through",
			Template: `<red>{{ .Data }}</>`,
			Context:  ctx{Data: "branch"},
			Expected: "<red>branch</>",
		},
		{
			Case:     "markup values pass through unescaped",
			Template: `{{ .Markup }}`,
			Context:  ctx{Markup: RawMarkup("<#ffffff>icon </>")},
			Expected: "<#ffffff>icon </>",
		},
		{
			Case:     "markup and data compose",
			Template: `{{ .Markup }}{{ .Data }}`,
			Context:  ctx{Markup: RawMarkup("<b>"), Data: "<red>x"},
			Expected: "<b><<>red<>>x",
		},
		{
			Case:     "assigned variables are escaped when printed",
			Template: `{{ $x := .Data }}{{ $x }}`,
			Context:  ctx{Data: "<red>"},
			Expected: "<<>red<>>",
		},
		{
			Case:     "actions inside if are escaped",
			Template: `{{ if .Data }}{{ .Data }}{{ end }}`,
			Context:  ctx{Data: "<red>"},
			Expected: "<<>red<>>",
		},
		{
			Case:     "actions inside range are escaped",
			Template: `{{ range .Items }}{{ . }}{{ end }}`,
			Context:  map[string]any{"Items": []string{"<a>", "<b>"}},
			Expected: "<<>a<>><<>b<>>",
		},
		{
			Case:     "actions inside with are escaped",
			Template: `{{ with .Data }}{{ . }}{{ end }}`,
			Context:  ctx{Data: "<red>"},
			Expected: "<<>red<>>",
		},
		{
			Case:     "pipelines through sprig functions are escaped",
			Template: `{{ .Data | upper }}`,
			Context:  ctx{Data: "<red>"},
			Expected: "<<>RED<>>",
		},
		{
			Case:     "present nil map values render empty",
			Template: `[{{ .Data }}]`,
			Context:  map[string]any{"Data": nil},
			Expected: "[]",
		},
		{
			Case:     "plain scalars are untouched",
			Template: `{{ .Code }}`,
			Context:  struct{ Code int }{Code: 42},
			Expected: "42",
		},
		{
			Case:     "empty markup is false, like an empty string",
			Template: `{{ if .Icon }}[{{ .Icon }}]{{ end }}{{ if .Text }}[{{ .Text }}]{{ end }}`,
			Context: struct {
				Icon Markup
				Text string
			}{Icon: RawMarkup("")},
			Expected: "",
		},
		{
			Case:     "markup compares by value",
			Template: `{{ if eq .HEAD "main" }}yes{{ end }}{{ if ne .HEAD "main" }}no{{ end }}`,
			Context:  struct{ HEAD Markup }{HEAD: RawMarkup("main")},
			Expected: "yes",
		},
	}

	origCache := Cache
	t.Cleanup(func() { Cache = origCache })

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return("foo")
		Cache = new(cache.Template)
		Init(env, nil, nil)

		text, err := RenderTrusted(tc.Template, tc.Context)
		require.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestRenderUntrustedEscapesActionOutput(t *testing.T) {
	origCache := Cache
	t.Cleanup(func() { Cache = origCache })

	env := new(mock.Environment)
	env.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(env, nil, nil)

	text, err := RenderUntrusted(`{{ .Data }}`, map[string]any{"Data": "<red>"})
	require.NoError(t, err)
	assert.Equal(t, "<<>red<>>", text)
}

func TestURLMarkup(t *testing.T) {
	cases := []struct {
		Case        string
		Template    string
		Expected    string
		ShouldError bool
	}{
		{
			Case:     "label is escaped, anchor structure intact",
			Template: `{{ url "a<b" "https://ohmyposh.dev" }}`,
			Expected: "<LINK>https://ohmyposh.dev<TEXT>a<<>b</TEXT></LINK>",
		},
		{
			Case:        "chevrons in the URL are rejected",
			Template:    `{{ url "link" "https://ohmyposh.dev/<TEXT>" }}`,
			ShouldError: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return("foo")
		Cache = new(cache.Template)
		Init(env, nil, nil)

		text, err := RenderTrusted(tc.Template, nil)
		if tc.ShouldError {
			assert.Error(t, err, tc.Case)
			continue
		}

		require.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestMarkupGobRoundtrip(t *testing.T) {
	original := RawMarkup("<red>safe</>")

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(original))

	var restored Markup
	require.NoError(t, gob.NewDecoder(&buf).Decode(&restored))
	assert.Equal(t, original.String(), restored.String())
}

// The session cache stores segment data as map[string]any and gob-encodes it
// per process: an unregistered Markup inside an interface value fails the
// whole encode (type not registered for interface), silently dropping the
// segment's cache entry for the next render.
func TestMarkupGobInterfaceRoundtrip(t *testing.T) {
	m := map[string]any{"HEAD": RawMarkup("<red>safe</>")}

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(m))

	var out map[string]any
	require.NoError(t, gob.NewDecoder(&buf).Decode(&out))

	got, ok := out["HEAD"].(Markup)
	require.True(t, ok, "HEAD decoded as %T", out["HEAD"])
	assert.Equal(t, "<red>safe</>", got.String())
}

func TestEscapeText(t *testing.T) {
	assert.Equal(t, "<<>red<>>x<<>/<>>", EscapeText("<red>x</>"))
	assert.Equal(t, "plain", EscapeText("plain"))
}
