package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

func TestLua(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		ExpectedURL    string
		Tooling        []string
		HasLua         bool
		HasLuaJit      bool
	}{
		{
			Case:           "Lua 5.4.4 - default tooling prefers lua",
			ExpectedString: "5.4.4",
			Version:        "Lua 5.4.4  Copyright (C) 1994-2022 Lua.org, PUC-Rio",
			ExpectedURL:    "https://www.lua.org/manual/5.4/readme.html#changes",
			HasLua:         true,
			HasLuaJit:      true,
		},
		{
			Case:           "Lua 5.0 - tooling set to luajit but missing so fallback to lua",
			ExpectedString: "5.0",
			Version:        "Lua 5.0  Copyright (C) 1994-2003 Tecgraf, PUC-Rio",
			ExpectedURL:    "https://www.lua.org/manual/5.0/readme.html#changes",
			HasLua:         true,
			Tooling:        []string{"luajit", "lua"},
		},
		{
			Case:           "LuaJIT 2.0.5 - tooling set to luajit first",
			ExpectedString: "2.0.5",
			Version:        "LuaJIT 2.0.5 -- Copyright (C) 2005-2017 Mike Pall. http://luajit.org/",
			HasLuaJit:      true,
			HasLua:         true,
			ExpectedURL:    "https://github.com/LuaJIT/LuaJIT/tree/v2.0",
			Tooling:        []string{"luajit"},
		},
		{
			Case:           "LuaJIT 2.1.0-beta3 - tooling set to lua first but missing so try luajit",
			ExpectedString: "2.1.0",
			Version:        "LuaJIT 2.1.0-beta3 -- Copyright (C) 2005-2017 Mike Pall. http://luajit.org/",
			HasLuaJit:      true,
			ExpectedURL:    "https://github.com/LuaJIT/LuaJIT/tree/v2.1",
			Tooling:        []string{"lua", "luajit"},
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "lua",
			versionParam:  "-v",
			versionOutput: tc.Version,
			extension:     "*.lua",
		}
		env, props := getMockedLanguageEnv(params)

		if !tc.HasLua {
			env.Unset("HasCommand")
			env.On("HasCommand", "lua").Return(false)
		}

		env.On("HasCommand", "luajit").Return(tc.HasLuaJit)
		env.On("RunCommand", "luajit", []string{"-v"}).Return(tc.Version, nil)
		env.On("Shell").Return("bash")

		// Initialize template system for version URL rendering
		if template.Cache == nil {
			template.Cache = &cache.Template{}
		}
		template.Init(env, nil, nil)

		if len(tc.Tooling) > 0 {
			props[Tooling] = tc.Tooling
		}

		l := &Lua{}
		l.Init(props, env)

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, l.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, l.Template(), l), failMsg)
		assert.Equal(t, tc.ExpectedURL, l.URL, failMsg)
	}
}
