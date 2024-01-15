package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCaml(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "OCaml 4.12.0", ExpectedString: "4.12.0", Version: "The OCaml toplevel, version 4.12.0"},
		{Case: "OCaml 4.11.0", ExpectedString: "4.11.0", Version: "The OCaml toplevel, version 4.11.0"},
		{Case: "OCaml 4.13.0", ExpectedString: "4.13.0", Version: "The OCaml toplevel, version 4.13.0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "ocaml",
			versionParam:  "-version",
			versionOutput: tc.Version,
			extension:     "*.ml",
		}
		env, props := getMockedLanguageEnv(params)
		o := &OCaml{}
		o.Init(props, env)
		assert.True(t, o.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, o.Template(), o), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
