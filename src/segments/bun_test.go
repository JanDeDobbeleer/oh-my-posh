package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

func TestBun(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Extension      string
	}{
		{Case: "Bun 1.1.8 with bun.lockb", ExpectedString: "1.1.8", Version: "1.1.8", Extension: "bun.lockb"},
		{Case: "Bun 1.3.10 with bun.lock", ExpectedString: "1.3.10", Version: "1.3.10", Extension: "bun.lock"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "bun").Return(true)
		env.On("RunCommand", "bun", []string{"--version"}).Return(tc.Version, nil)
		env.On("HasFiles", "bun.lockb").Return(tc.Extension == "bun.lockb")
		env.On("HasFiles", "bun.lock").Return(tc.Extension == "bun.lock")
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("Shell").Return("foo")

		if template.Cache == nil {
			template.Cache = &cache.Template{}
		}
		template.Init(env, nil, nil)

		props := options.Map{
			options.FetchVersion: true,
		}

		b := &Bun{}
		b.Init(props, env)

		assert.True(t, b.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
