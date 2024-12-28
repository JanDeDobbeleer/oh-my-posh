package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNim(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "Nim 2.2.0",
			ExpectedString: "2.2.0",
			Version:        "Nim Compiler Version 2.2.0 [MacOSX: arm64]\nCompiled at 2024-11-30\nCopyright (c) 2006-2024 by Andreas Rumpf",
		},
		{
			Case:           "Nim 1.6.12",
			ExpectedString: "1.6.12",
			Version:        "Nim Compiler Version 1.6.12 [Linux: amd64]\nCompiled at 2023-06-15\nCopyright (c) 2006-2023 by Andreas Rumpf",
		},
		{
			Case:           "Nim 2.0.0",
			ExpectedString: "2.0.0",
			Version:        "Nim Compiler Version 2.0.0 [Windows: amd64]\nCompiled at 2023-12-25\nCopyright (c) 2006-2023 by Andreas Rumpf",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "nim",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.nim",
		}
		env, props := getMockedLanguageEnv(params)
		n := &Nim{}
		n.Init(props, env)
		assert.True(t, n.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, n.Template(), n), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
