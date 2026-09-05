package template

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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
		{
			Case:     "predicates accept markup",
			Template: `{{ if contains "ma" .HEAD }}yes{{ end }}{{ if hasPrefix "x" .HEAD }}no{{ end }}`,
			Context:  struct{ HEAD Markup }{HEAD: RawMarkup("<red>main</>")},
			Expected: "yes",
		},
		{
			Case:     "transforms keep markup anchors",
			Template: `{{ .HEAD | trimSuffix "-wip" | upper }}{{ .HEAD | replace "main" "dev" }}`,
			Context:  struct{ HEAD Markup }{HEAD: RawMarkup("<red>main</>-wip")},
			Expected: "<RED>MAIN</><red>dev</>-wip",
		},
		{
			Case:     "transforms on data stay escaped",
			Template: `{{ .Branch | trimSuffix "-wip" }} {{ .Branch | upper }}`,
			Context:  struct{ Branch string }{Branch: "<b>-wip"},
			Expected: "<<>b<>> <<>B<>>-WIP",
		},
		{
			Case:     "data mixed into a markup call is escaped",
			Template: `{{ printf "%s%s" .Icon .Branch }}|{{ cat .Icon .Branch }}|{{ .Icon | replace "x" .Branch }}`,
			Context: struct {
				Icon   Markup
				Branch string
			}{Icon: RawMarkup("<red>x</>"), Branch: "<b>"},
			Expected: "<red>x</><<>b<>>|<red>x</> <<>b<>>|<red><<>b<>></>",
		},
		{
			Case:     "markup-free calls are unchanged",
			Template: `{{ printf "<%s>" .Branch }}{{ trunc 2 .Branch }}{{ if contains "a" .Branch }}!{{ end }}`,
			Context:  struct{ Branch string }{Branch: "main"},
			Expected: "<<>main<>>ma!",
		},
		{
			Case:     "numeric literals reach typed parameters",
			Template: `{{ repeat 2 .HEAD }}{{ substr 0 3 .HEAD }}{{ add 1 2 }}`,
			Context:  struct{ HEAD Markup }{HEAD: RawMarkup("<red>ab</>")},
			Expected: "<red>ab</><red>ab</><re3",
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

func TestMarkupJSON(t *testing.T) {
	cases := []struct {
		Case     string
		JSON     string
		Expected Markup
		Error    bool
	}{
		{Case: "tagged form", JSON: `{"$markup":"<red>x</>"}`, Expected: RawMarkup("<red>x</>")},
		{Case: "bare string from an older fixture", JSON: `"<red>x</>"`, Expected: RawMarkup("<red>x</>")},
		{Case: "empty tagged form", JSON: `{"$markup":""}`, Expected: RawMarkup("")},
		{Case: "object without the tag", JSON: `{"text":"x"}`, Error: true},
	}

	for _, tc := range cases {
		var got Markup
		err := json.Unmarshal([]byte(tc.JSON), &got)

		if tc.Error {
			assert.Error(t, err, tc.Case)
			continue
		}

		require.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}

	encoded, err := json.Marshal(struct {
		HEAD Markup
		Ref  string
	}{HEAD: RawMarkup("<red>x</>"), Ref: "<b>"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"HEAD":{"$markup":"<red>x</>"},"Ref":"<b>"}`, string(encoded))
}

// A generically decoded document (the writer-less restore path) must carry
// the same trust as the struct it was recorded from: tagged objects become
// Markup, everything else stays data.
func TestReviveMarkup(t *testing.T) {
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{
		"HEAD": {"$markup": "<red>main</>"},
		"Ref": "<b>",
		"Rebase": {"Onto": {"$markup": "<b>x</>"}, "Total": 2},
		"Items": [{"$markup": "<i>y</>"}, "plain"],
		"Shaped": {"$markup": "x", "extra": 1}
	}`), &decoded))

	got := ReviveMarkup(decoded).(map[string]any)

	assert.Equal(t, RawMarkup("<red>main</>"), got["HEAD"])
	assert.Equal(t, "<b>", got["Ref"])
	assert.Equal(t, RawMarkup("<b>x</>"), got["Rebase"].(map[string]any)["Onto"])
	assert.Equal(t, RawMarkup("<i>y</>"), got["Items"].([]any)[0])
	assert.Equal(t, "plain", got["Items"].([]any)[1])
	_, stillObject := got["Shaped"].(map[string]any)
	assert.True(t, stillObject, "an object with more than the tag is not markup")

	text, err := RenderTrusted(`{{ .HEAD }}|{{ .Ref }}`, got)
	require.NoError(t, err)
	assert.Equal(t, "<red>main</>|<<>b<>>", text)
}

func TestEscapeText(t *testing.T) {
	assert.Equal(t, "<<>red<>>x<<>/<>>", EscapeText("<red>x</>"))
	assert.Equal(t, "plain", EscapeText("plain"))
}
