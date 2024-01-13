package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerl(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		Version         string
		PerlHomeVersion string
		PerlHomeEnabled bool
	}{
		{
			Case:           "v5.12+",
			ExpectedString: "5.32.1",
			Version:        "This is perl 5, version 32, subversion 1 (v5.32.1) built for MSWin32-x64-multi-thread",
		},
		{
			Case:           "v5.6 - v5.10",
			ExpectedString: "5.6.1",
			Version:        "This is perl, v5.6.1 built for MSWin32-x86-multi-thread",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "perl",
			versionParam:  "-version",
			versionOutput: tc.Version,
			extension:     ".perl-version",
		}
		env, props := getMockedLanguageEnv(params)

		p := &Perl{}
		p.Init(props, env)
		assert.True(t, p.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, p.Template(), p), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
