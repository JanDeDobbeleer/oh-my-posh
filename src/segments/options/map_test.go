package options

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

func setupTemplateTestEnv() *mock.Environment {
	env := &mock.Environment{}
	env.On("Getenv", "SHLVL").Return("1")
	env.On("Shell").Return("bash")
	env.On("Flags").Return(&runtime.Flags{
		IsPrimary:    true,
		ShellVersion: "1.0.0",
		PromptCount:  1,
		JobCount:     0,
		PSWD:         "/home/test",
		AbsolutePWD:  "/home/test",
	})
	env.On("Root").Return(false)
	env.On("StatusCodes").Return(0, "0")
	env.On("IsWsl").Return(false)
	env.On("Pwd").Return("/home/test")
	env.On("GOOS").Return(runtime.LINUX)
	env.On("Platform").Return("ubuntu")
	env.On("User").Return("testuser")
	env.On("Host").Return("testhost", nil)

	template.Cache = new(cache.Template)
	template.Cache.TerminalWidth = 120
	template.Init(env, nil, nil)

	return env
}

const (
	expected      = "expected"
	expectedColor = color.Ansi("#768954")

	Foo Option = "color"
)

func TestGetString(t *testing.T) {
	var options = Map{Foo: expected}
	value := options.String(Foo, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	var options = Map{}
	value := options.String(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	var options = Map{Foo: true}
	value := options.String(Foo, expected)
	assert.Equal(t, "true", value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	var options = Map{Foo: expected}
	value := options.Color(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := color.Ansi("yellow")
	var options = Map{Foo: expected}
	value := options.Color(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	var options = Map{Foo: "invalid"}
	value := options.Color(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	var options = Map{}
	value := options.Color(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := color.Ansi("p:red")
	var options = Map{Foo: expected}
	value := options.Color(Foo, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	var options = Map{Foo: expected}
	value := options.Bool(Foo, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	var options = Map{}
	value := options.Bool(Foo, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	var options = Map{Foo: "borked"}
	value := options.Bool(Foo, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	cases := []struct {
		Input    any
		Case     string
		Expected float64
	}{
		{Case: "int", Expected: 1337, Input: 1337},
		{Case: "float64", Expected: 1337, Input: float64(1337)},
		{Case: "uint64", Expected: 1337, Input: uint64(1337)},
		{Case: "int64", Expected: 1337, Input: int64(1337)},
		{Case: "string", Expected: 9001, Input: "invalid"},
		{Case: "bool", Expected: 9001, Input: true},
	}
	for _, tc := range cases {
		options := Map{Foo: tc.Input}
		value := options.Float64(Foo, 9001)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	var options = Map{}
	value := options.Float64(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestOneOf(t *testing.T) {
	cases := []struct {
		Expected     any
		Map          Map
		Case         string
		DefaultValue string
		Options      []Option
	}{
		{
			Case:     "one element",
			Expected: "1337",
			Options:  []Option{Foo},
			Map: Map{
				Foo: "1337",
			},
			DefaultValue: "2000",
		},
		{
			Case:     "two elements",
			Expected: "1337",
			Options:  []Option{Foo},
			Map: Map{
				Foo:   "1337",
				"Bar": "9001",
			},
			DefaultValue: "2000",
		},
		{
			Case:     "no match",
			Expected: "2000",
			Options:  []Option{"Moo"},
			Map: Map{
				Foo:   "1337",
				"Bar": "9001",
			},
			DefaultValue: "2000",
		},
		{
			Case:     "incorrect type",
			Expected: "2000",
			Options:  []Option{Foo},
			Map: Map{
				Foo:   1337,
				"Bar": "9001",
			},
			DefaultValue: "2000",
		},
	}
	for _, tc := range cases {
		value := OneOf(tc.Map, tc.DefaultValue, tc.Options...)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestResolveTemplate(t *testing.T) {
	env := setupTemplateTestEnv()
	env.On("Getenv", "MY_VAR").Return("42")

	cases := []struct {
		Case        string
		Raw         string
		Expected    string
		WasTemplate bool
	}{
		{
			Case:        "plain string, no template",
			Raw:         "hello",
			Expected:    "hello",
			WasTemplate: false,
		},
		{
			Case:        "env var template",
			Raw:         "{{ .Env.MY_VAR }}",
			Expected:    "42",
			WasTemplate: true,
		},
		{
			Case:        "terminal width from global context",
			Raw:         "{{ sub .TerminalWidth 30 }}",
			Expected:    "90",
			WasTemplate: true,
		},
		{
			Case:        "unknown field returns raw",
			Raw:         "{{ .Unknown }}",
			Expected:    "{{ .Unknown }}",
			WasTemplate: false,
		},
	}

	for _, tc := range cases {
		resolved, wasTemplate := resolveTemplate("test_option", tc.Raw)
		assert.Equal(t, tc.Expected, resolved, tc.Case)
		assert.Equal(t, tc.WasTemplate, wasTemplate, tc.Case)
	}
}

func TestTemplate(t *testing.T) {
	env := setupTemplateTestEnv()
	env.On("Getenv", "MY_API_KEY").Return("secret-key-123")
	env.On("Getenv", "MY_USER").Return("testuser")

	cases := []struct {
		Case         string
		Options      Map
		Option       Option
		DefaultValue string
		Context      any
		Expected     string
	}{
		{
			Case:         "plain string no template",
			Options:      Map{"key": "plain-value"},
			Option:       "key",
			DefaultValue: "",
			Context:      nil,
			Expected:     "plain-value",
		},
		{
			Case:         "template with env var",
			Options:      Map{"key": "{{ .Env.MY_API_KEY }}"},
			Option:       "key",
			DefaultValue: "",
			Context:      nil,
			Expected:     "secret-key-123",
		},
		{
			Case:         "template with multiple env vars",
			Options:      Map{"key": "{{ .Env.MY_USER }}/{{ .Env.MY_API_KEY }}"},
			Option:       "key",
			DefaultValue: "",
			Context:      nil,
			Expected:     "testuser/secret-key-123",
		},
		{
			Case:         "empty value returns default",
			Options:      Map{},
			Option:       "key",
			DefaultValue: "default-value",
			Context:      nil,
			Expected:     "default-value",
		},
		{
			Case:         "invalid template returns raw value",
			Options:      Map{"key": "{{ .Invalid }}"},
			Option:       "key",
			DefaultValue: "",
			Context:      nil,
			Expected:     "{{ .Invalid }}",
		},
	}

	for _, tc := range cases {
		value := tc.Options.Template(tc.Option, tc.DefaultValue, tc.Context)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestIntWithTemplate(t *testing.T) {
	setupTemplateTestEnv()

	cases := []struct {
		Input    any
		Case     string
		Expected int
		Default  int
	}{
		{
			Case:     "plain int unchanged",
			Input:    42,
			Expected: 42,
			Default:  0,
		},
		{
			Case:     "template resolves to int",
			Input:    "{{ add 10 20 }}",
			Expected: 30,
			Default:  0,
		},
		{
			Case:     "template with terminal width",
			Input:    "{{ sub .TerminalWidth 30 }}",
			Expected: 90,
			Default:  0,
		},
		{
			Case:     "template resolves to non-numeric",
			Input:    `{{ "hello" }}`,
			Expected: 99,
			Default:  99,
		},
		{
			Case:     "template with unknown field returns default",
			Input:    "{{ .Unknown }}",
			Expected: 99,
			Default:  99,
		},
		{
			Case:     "plain string not a template returns default",
			Input:    "not-a-number",
			Expected: 99,
			Default:  99,
		},
	}

	for _, tc := range cases {
		m := Map{Foo: tc.Input}

		value := m.Int(Foo, tc.Default)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestFloat64WithTemplate(t *testing.T) {
	env := setupTemplateTestEnv()
	env.On("Getenv", "RATE").Return("3.14")

	cases := []struct {
		Input    any
		Case     string
		Expected float64
		Default  float64
	}{
		{
			Case:     "plain float64 unchanged",
			Input:    3.14,
			Expected: 3.14,
			Default:  0,
		},
		{
			Case:     "template resolves to float",
			Input:    "{{ .Env.RATE }}",
			Expected: 3.14,
			Default:  0,
		},
		{
			Case:     "template resolves to non-numeric",
			Input:    `{{ "hello" }}`,
			Expected: 99.0,
			Default:  99.0,
		},
		{
			Case:     "template with unknown field returns default",
			Input:    "{{ .Unknown }}",
			Expected: 99.0,
			Default:  99.0,
		},
	}

	for _, tc := range cases {
		m := Map{Foo: tc.Input}

		value := m.Float64(Foo, tc.Default)
		assert.InDelta(t, tc.Expected, value, 0.001, tc.Case)
	}
}

func TestBoolWithTemplate(t *testing.T) {
	setupTemplateTestEnv()

	cases := []struct {
		Input    any
		Case     string
		Expected bool
		Default  bool
	}{
		{
			Case:     "plain bool unchanged",
			Input:    true,
			Expected: true,
			Default:  false,
		},
		{
			Case:     "template resolves to true",
			Input:    "{{ gt .TerminalWidth 100 }}",
			Expected: true,
			Default:  false,
		},
		{
			Case:     "template resolves to false",
			Input:    "{{ lt .TerminalWidth 100 }}",
			Expected: false,
			Default:  true,
		},
		{
			Case:     "template resolves to non-bool",
			Input:    `{{ "hello" }}`,
			Expected: false,
			Default:  false,
		},
		{
			Case:     "template with unknown field returns default",
			Input:    "{{ .Unknown }}",
			Expected: true,
			Default:  true,
		},
	}

	for _, tc := range cases {
		m := Map{Foo: tc.Input}

		value := m.Bool(Foo, tc.Default)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestColorWithTemplate(t *testing.T) {
	env := setupTemplateTestEnv()
	env.On("Getenv", "MY_COLOR").Return("#AABBCC")

	cases := []struct {
		Case     string
		Input    any
		Expected color.Ansi
		Default  color.Ansi
	}{
		{
			Case:     "template resolves to hex color",
			Input:    `{{ "#FF0000" }}`,
			Expected: color.Ansi("#FF0000"),
			Default:  color.Ansi("#000000"),
		},
		{
			Case:     "template resolves to named color",
			Input:    `{{ "yellow" }}`,
			Expected: color.Ansi("yellow"),
			Default:  color.Ansi("#000000"),
		},
		{
			Case:     "template resolves to palette color",
			Input:    `{{ "p:red" }}`,
			Expected: color.Ansi("p:red"),
			Default:  color.Ansi("#000000"),
		},
		{
			Case:     "template with env var",
			Input:    "{{ .Env.MY_COLOR }}",
			Expected: color.Ansi("#AABBCC"),
			Default:  color.Ansi("#000000"),
		},
		{
			Case:     "template with unknown field returns default",
			Input:    "{{ .Unknown }}",
			Expected: color.Ansi("#000000"),
			Default:  color.Ansi("#000000"),
		},
	}

	for _, tc := range cases {
		m := Map{Foo: tc.Input}

		value := m.Color(Foo, tc.Default)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}
