package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
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
		// Both cases resolve "bun" to the same mocked path/stat, so without
		// clearing between them the second case would be served the first
		// case's cached output instead of exercising its own.
		cache.Device.DeleteAll()

		env := new(mock.Environment)
		env.On("HasCommand", "bun").Return(true)
		env.On("RunCommandWithEnv", "bun", []string(nil), []string{"--version"}).Return(tc.Version, nil)
		env.On("HasFiles", "bun.lockb").Return(tc.Extension == "bun.lockb")
		env.On("HasFiles", "bun.lock").Return(tc.Extension == "bun.lock")
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("Shell").Return("foo")
		env.On("Flags").Return(&runtime.Flags{})
		env.On("CommandPath", "bun").Return("/usr/bin/bun")
		env.On("StatFile", "/usr/bin/bun").Return(runtime.FileStat{ModTime: 1, Size: 1}, nil)

		if template.Cache == nil {
			template.Cache = &cache.Template{}
		}
		template.Init(env, nil, nil)

		props := options.Map{}

		b := &Bun{}
		b.Init(props, env)

		assert.True(t, b.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
