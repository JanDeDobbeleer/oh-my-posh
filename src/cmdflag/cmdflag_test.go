package cmdflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStyles(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Args     []string
		Rest     []string
		Number   int
		Bool     bool
	}{
		{Case: "long with space", Args: []string{"--name", "x"}, Expected: "x", Rest: []string{}},
		{Case: "long with equals", Args: []string{"--name=x"}, Expected: "x", Rest: []string{}},
		{Case: "shorthand with space", Args: []string{"-n", "x"}, Expected: "x", Rest: []string{}},
		{Case: "shorthand joined", Args: []string{"-nx"}, Expected: "x", Rest: []string{}},
		{Case: "shorthand with equals", Args: []string{"-n=x"}, Expected: "x", Rest: []string{}},
		{Case: "bool does not eat value", Args: []string{"--flag", "positional"}, Bool: true, Rest: []string{"positional"}},
		{Case: "terminator", Args: []string{"--", "--name", "x"}, Rest: []string{"--name", "x"}},
		{Case: "interspersed", Args: []string{"a", "--name", "x", "b"}, Expected: "x", Rest: []string{"a", "b"}},
		{Case: "int value", Args: []string{"--count", "42"}, Number: 42, Rest: []string{}},
		{Case: "combined bool shorthands", Args: []string{"-fv"}, Bool: true, Rest: []string{}},
	}

	for _, tc := range cases {
		var name string
		var flag, verbose bool
		var count int

		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVarP(&name, "name", "n", "", "")
		fs.BoolVarP(&flag, "flag", "f", false, "")
		fs.BoolVarP(&verbose, "verbose", "v", false, "")
		fs.IntVar(&count, "count", 0, "")

		err := fs.Parse(tc.Args)
		assert.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, name, tc.Case)
		assert.Equal(t, tc.Bool, flag, tc.Case)
		assert.Equal(t, tc.Number, count, tc.Case)
		assert.Equal(t, tc.Rest, fs.Args(), tc.Case)
	}
}

func TestParseErrors(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var name string
	fs.StringVar(&name, "name", "", "")

	err := fs.Parse([]string{"--unknown"})
	assert.EqualError(t, err, "unknown flag: --unknown")

	err = fs.Parse([]string{"-x"})
	assert.EqualError(t, err, `unknown shorthand flag: 'x' in -x`)

	err = fs.Parse([]string{"--name"})
	assert.EqualError(t, err, "flag needs an argument: --name")
}

func TestUnknownFlagsAllowlist(t *testing.T) {
	// the argocd segment parses ARGOCD_OPTS this way
	fs := NewFlagSet("test", ContinueOnError)
	fs.ParseErrorsAllowlist.UnknownFlags = true
	fs.String("config", "", "")

	err := fs.Parse([]string{"--grpc-web", "--server", "foo.com", "--config", "x", "--insecure"})
	assert.NoError(t, err)
	assert.Equal(t, "x", fs.Lookup("config").Value.String())
}

func TestChangedAndDefaults(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var escape bool
	fs.BoolVar(&escape, "escape", true, "escape the output")

	assert.False(t, fs.Changed("escape"))
	assert.NoError(t, fs.Parse([]string{"--escape=false"}))
	assert.True(t, fs.Changed("escape"))
	assert.False(t, escape)
	assert.Equal(t, "true", fs.Lookup("escape").DefValue)
}

func TestFlagUsages(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var name, pswd string
	var force, escape bool
	var width int

	fs.StringVarP(&name, "config", "c", "", "config file path")
	fs.BoolVarP(&force, "force", "f", false, "force rendering the segments")
	fs.BoolVar(&escape, "escape", true, "escape the ANSI sequences for the shell")
	fs.IntVarP(&width, "terminal-width", "w", 0, "width of the terminal")
	fs.StringVar(&pswd, "pswd", "", "hidden flag")
	_ = fs.MarkHidden("pswd")

	expected := `  -c, --config string        config file path
      --escape               escape the ANSI sequences for the shell (default true)
  -f, --force                force rendering the segments
  -w, --terminal-width int   width of the terminal
`
	assert.Equal(t, expected, fs.FlagUsages())
}
