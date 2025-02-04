package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBun(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Extension      string
	}{
		{Case: "Bun 1.1.8", ExpectedString: "1.1.8", Version: "1.1.8", Extension: "bun.lockb"},
		{Case: "Bun 1.2.2 with bun.lockb", ExpectedString: "1.2.2", Version: "1.2.2", Extension: "bun.lockb"},
		{Case: "Bun 1.2.2 with bun.lock", ExpectedString: "1.2.2", Version: "1.2.2", Extension: "bun.lock"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "bun",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     tc.Extension,
		}
		env, props := getMockedLanguageEnv(params)
		b := &Bun{}
		b.Init(props, env)
		assert.True(t, b.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
