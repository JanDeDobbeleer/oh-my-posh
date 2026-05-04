package segments

type Lua struct {
	Language
}

func (l *Lua) Template() string {
	return languageTemplate
}

func (l *Lua) Enabled() bool {
	const (
		luaToolName    = "lua"
		luajitToolName = "luajit"
	)

	l.extensions = []string{"*.lua", "*.rockspec"}
	l.folders = []string{"lua"}
	l.tooling = map[string]*cmd{
		luaToolName: {
			executable:         luaToolName,
			args:               []string{"-v"},
			regex:              `Lua (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			versionURLTemplate: "https://www.lua.org/manual/{{ .Major }}.{{ .Minor }}/readme.html#changes",
		},
		luajitToolName: {
			executable:         luajitToolName,
			args:               []string{"-v"},
			regex:              `LuaJIT (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?))`,
			versionURLTemplate: "https://github.com/LuaJIT/LuaJIT/tree/v{{ .Major}}.{{ .Minor}}",
		},
	}
	l.defaultTooling = []string{luaToolName, luajitToolName}

	return l.Language.Enabled()
}
