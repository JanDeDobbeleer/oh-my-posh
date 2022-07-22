package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLua(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		HasLua         bool
		HasLuaJit      bool
		Prefer         string
		ExpectedURL    string
	}{
		{
			Case:           "Lua 5.4.4 - Prefer Lua",
			ExpectedString: "5.4.4",
			Version:        "Lua 5.4.4  Copyright (C) 1994-2022 Lua.org, PUC-Rio",
			ExpectedURL:    "https://www.lua.org/manual/5.4/readme.html#changes",
			HasLua:         true,
			HasLuaJit:      true,
			Prefer:         "lua",
		},
		{
			Case:           "Lua 5.0 - Prefer luajit but missing so fallback to lua",
			ExpectedString: "5.0",
			Version:        "Lua 5.0  Copyright (C) 1994-2003 Tecgraf, PUC-Rio",
			ExpectedURL:    "https://www.lua.org/manual/5.0/readme.html#changes",
			HasLua:         true,
			Prefer:         "luajit",
		},
		{
			Case:           "LuaJIT 2.0.5 - Prefer LuaJIT",
			ExpectedString: "2.0.5",
			Version:        "LuaJIT 2.0.5 -- Copyright (C) 2005-2017 Mike Pall. http://luajit.org/",
			HasLuaJit:      true,
			HasLua:         true,
			ExpectedURL:    "https://github.com/LuaJIT/LuaJIT/tree/v2.0",
			Prefer:         "luajit",
		},
		{
			Case:           "LuaJIT 2.1.0-beta3 - Prefer Lua but missing so try luajit",
			ExpectedString: "2.1.0",
			Version:        "LuaJIT 2.1.0-beta3 -- Copyright (C) 2005-2017 Mike Pall. http://luajit.org/",
			HasLuaJit:      true,
			ExpectedURL:    "https://github.com/LuaJIT/LuaJIT/tree/v2.1",
			Prefer:         "lua",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "lua").Return(tc.HasLua)
		env.On("RunCommand", "lua", []string{"-v"}).Return(tc.Version, nil)
		env.On("HasCommand", "luajit").Return(tc.HasLuaJit)
		env.On("RunCommand", "luajit", []string{"-v"}).Return(tc.Version, nil)
		env.On("HasFiles", "*.lua").Return(true)
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})
		props := properties.Map{
			properties.FetchVersion: true,
		}
		props[PreferredExecutable] = tc.Prefer
		l := &Lua{}
		l.Init(props, env)
		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, l.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, l.Template(), l), failMsg)
		assert.Equal(t, tc.ExpectedURL, l.version.URL, failMsg)
		assert.Equal(t, strings.ToLower(strings.Split(tc.Case, " ")[0]), l.version.Executable, failMsg)
	}
}
