package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

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
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "perl").Return(true)
		env.On("RunCommand", "perl", []string{"-version"}).Return(tc.Version, nil)
		env.On("HasFiles", ".perl-version").Return(true)
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		props := properties.Map{
			properties.FetchVersion: true,
		}
		p := &Perl{}
		p.Init(props, env)
		assert.True(t, p.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, p.Template(), p), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
