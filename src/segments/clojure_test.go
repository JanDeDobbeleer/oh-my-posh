package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClojure(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Cmd            string
	}{
		{
			Case:           "Clojure CLI 1.11.1.1113",
			ExpectedString: "1.11.1.1113",
			Version:        "Clojure CLI version 1.11.1.1113",
			Cmd:            "clojure",
		},
		{
			Case:           "Leiningen 2.9.8",
			ExpectedString: "2.9.8",
			Version:        "Leiningen 2.9.8 on Java 11.0.11 OpenJDK 64-Bit Server VM",
			Cmd:            "lein",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           tc.Cmd,
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.clj",
		}
		env, props := getMockedLanguageEnv(params)
		props[LanguageExtensions] = []string{params.extension}
		if tc.Cmd != "clojure" {
			env.On("HasCommand", "clojure").Return(false)
		}
		c := &Clojure{}
		c.Init(props, env)
		assert.True(t, c.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
