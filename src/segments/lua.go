package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Lua struct {
	language
}

const (
	PreferredExecutable properties.Property = "preferred_executable"
)

func (l *Lua) Template() string {
	return languageTemplate
}

func (l *Lua) Enabled() bool {
	l.extensions = []string{"*.lua", "*.rockspec"}
	l.folders = []string{"lua"}
	l.commands = []*cmd{
		{
			executable:         "lua",
			args:               []string{"-v"},
			regex:              `Lua (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			versionURLTemplate: "https://www.lua.org/manual/{{ .Major }}.{{ .Minor }}/readme.html#changes",
		},
		{
			executable:         "luajit",
			args:               []string{"-v"},
			regex:              `LuaJIT (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			versionURLTemplate: "https://github.com/LuaJIT/LuaJIT/tree/v{{ .Major}}.{{ .Minor}}",
		},
	}

	if l.props.GetString(PreferredExecutable, "lua") == "luajit" {
		l.commands = []*cmd{l.commands[1], l.commands[0]}
	}

	return l.language.Enabled()
}
