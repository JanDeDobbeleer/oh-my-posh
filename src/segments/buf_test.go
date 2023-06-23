package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuf(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Buf 1.12.0", ExpectedString: "1.12.0", Version: "1.12.0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "buf",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "buf.yaml",
		}
		env, props := getMockedLanguageEnv(params)
		b := &Buf{}
		b.Init(props, env)
		assert.True(t, b.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
