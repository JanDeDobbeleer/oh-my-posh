package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestConfiguredLanguageFortranPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "GNU Fortran 10.2.1 Debian",
			ExpectedString: "10.2.1",
			Version: `GNU Fortran (Debian 10.2.1-6) 10.2.1 20210110
			Copyright (C) 2020 Free Software Foundation, Inc.
			This is free software; see the source for copying conditions.  There is NO
			warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.`,
		},
		{
			Case:           "GNU Fortran 11.4.0 Ubuntu",
			ExpectedString: "11.4.0",
			Version: `GNU Fortran (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0
			Copyright (C) 2021 Free Software Foundation, Inc.
			This is free software; see the source for copying conditions.  There is NO
			warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.`,
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "gfortran",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.f",
		}
		env, props := getMockedLanguageEnv(params)

		f := NewLanguage("fortran")
		f.Init(props, env)

		assert.True(t, f.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, f.Template(), f), tc.Case)
	}
}

func TestConfiguredLanguageRubyPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		HasRbenv       bool
		HasRvmprompt   bool
		HasChruby      bool
		HasAsdf        bool
		HasRuby        bool
		HasRubyFiles   bool
		HasRakeFile    bool
		HasGemFile     bool
		FetchVersion   bool
	}{
		{Case: "Ruby files", ExpectedString: "", FetchVersion: false, HasRubyFiles: true},
		{Case: "Rakefile", ExpectedString: "", FetchVersion: false, HasRakeFile: true},
		{Case: "Gemfile", ExpectedString: "", FetchVersion: false, HasGemFile: true},
		{Case: "Gemfile with version", ExpectedString: "err parsing info from ruby with", FetchVersion: true, HasGemFile: true},
		{
			Case:           "Version with chruby",
			ExpectedString: "ruby-2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasChruby:      true,
			Version: ` * ruby-2.6.3
			ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:           "Version with chruby line 2",
			ExpectedString: "ruby-1.9.3-p392",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasChruby:      true,
			Version: ` ruby-2.6.3
			* ruby-1.9.3-p392
			jruby-1.7.0
			rubinius-2.0.0-rc1`,
		},
		{
			Case:           "Version with asdf",
			ExpectedString: "2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasAsdf:        true,
			Version:        "ruby            2.6.3           /Users/jan/Projects/oh-my-posh/.tool-versions",
		},
		{
			Case:           "Version with asdf not set",
			ExpectedString: "",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasAsdf:        true,
			Version:        "ruby            ______          No version set. Run \"asdf <global|shell|local> ruby <version>\"",
		},
		{
			Case:           "Version with ruby",
			ExpectedString: "2.6.3",
			FetchVersion:   true,
			HasRubyFiles:   true,
			HasRuby:        true,
			Version:        "ruby  2.6.3 (2019-04-16 revision 67580) [universal.x86_64-darwin20]",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "ruby",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.rb",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasCommand", "rbenv").Return(tc.HasRbenv)
		env.On("RunCommandWithEnv", "rbenv", []string(nil), []string{"version-name"}).Return(tc.Version, nil)
		env.On("HasCommand", "rvm-prompt").Return(tc.HasRvmprompt)
		env.On("RunCommandWithEnv", "rvm-prompt", []string(nil), []string{"i", "v", "g"}).Return(tc.Version, nil)
		env.On("HasCommand", "chruby").Return(tc.HasChruby)
		env.On("RunCommandWithEnv", "chruby", []string(nil), []string(nil)).Return(tc.Version, nil)
		env.On("HasCommand", "asdf").Return(tc.HasAsdf)
		env.On("RunCommandWithEnv", "asdf", []string(nil), []string{"current", "ruby"}).Return(tc.Version, nil)
		env.On("HasFiles", "Rakefile").Return(tc.HasRakeFile)
		env.On("HasFiles", "Gemfile").Return(tc.HasGemFile)

		props[options.FetchVersion] = tc.FetchVersion

		ruby := NewLanguage("ruby")
		ruby.Init(props, env)

		assert.True(t, ruby.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, ruby.Template(), ruby), tc.Case)
	}
}

func TestConfiguredLanguageClojurePreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Cmd            string
	}{
		{
			Case:           "Clojure CLI 1.11.1.1113",
			ExpectedString: "1.11.1.1113",
			Version:        "Clojure CLI version 1.11.1.1113",
			Cmd:            "clojure",
		},
		{
			Case:           "Leiningen 2.9.8",
			ExpectedString: "2.9.8",
			Version:        "Leiningen 2.9.8 on Java 11.0.11 OpenJDK 64-Bit Server VM",
			Cmd:            "lein",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           tc.Cmd,
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.clj",
		}
		env, props := getMockedLanguageEnv(params)
		props[LanguageExtensions] = []string{params.extension}
		if tc.Cmd != "clojure" {
			env.On("HasCommand", "clojure").Return(false)
		}
		c := NewLanguage("clojure")
		c.Init(props, env)
		assert.True(t, c.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c), tc.Case)
	}
}

func TestConfiguredLanguageCrystalPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Crystal 1.0.0", ExpectedString: "1.0.0", Version: "Crystal 1.0.0 (2021-03-22)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "crystal",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.cr",
		}
		env, props := getMockedLanguageEnv(params)
		c := NewLanguage("crystal")
		c.Init(props, env)
		assert.True(t, c.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c), tc.Case)
	}
}

func TestConfiguredLanguageElixirPreset(t *testing.T) {
	cases := []struct {
		Case                string
		ExpectedString      string
		ElixirVersionOutput string
		AsdfVersionOutput   string
		HasAsdf             bool
		AsdfExitCode        int
	}{
		{
			Case:                "Version without asdf",
			ExpectedString:      "1.14.2",
			ElixirVersionOutput: "Erlang/OTP 25 [erts-13.1.3] [source] [64-bit] [smp:8:8] [ds:8:8:10] [async-threads:1] [jit] [dtrace]\n\nElixir 1.14.2 (compiled with Erlang/OTP 25)",
		},
		{
			Case:                "Version with asdf",
			ExpectedString:      "1.14.2",
			HasAsdf:             true,
			AsdfVersionOutput:   "elixir          1.14.2-otp-25   /path/to/.tool-versions",
			ElixirVersionOutput: "Should not be used",
		},
		{
			Case:                "Version with asdf not set: should fall back to elixir --version",
			ExpectedString:      "1.14.2",
			HasAsdf:             true,
			AsdfVersionOutput:   "elixir             ______          No version is set. Run \"asdf <global|shell|local> elixir <version>\"",
			AsdfExitCode:        126,
			ElixirVersionOutput: "Erlang/OTP 25 [erts-13.1.3] [source] [64-bit] [smp:8:8] [ds:8:8:10] [async-threads:1] [jit] [dtrace]\n\nElixir 1.14.2 (compiled with Erlang/OTP 25)",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "elixir",
			versionParam:  "--version",
			versionOutput: tc.ElixirVersionOutput,
			extension:     "*.ex",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasCommand", "asdf").Return(tc.HasAsdf)
		var asdfErr error
		if tc.AsdfExitCode != 0 {
			asdfErr = &runtime.CommandError{ExitCode: tc.AsdfExitCode}
		}
		env.On("RunCommandWithEnv", "asdf", []string(nil), []string{"current", "elixir"}).Return(tc.AsdfVersionOutput, asdfErr)

		r := NewLanguage("elixir")
		r.Init(props, env)
		assert.True(t, r.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), tc.Case)
	}
}

func TestConfiguredLanguageJuliaPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Julia 1.6.0", ExpectedString: "1.6.0", Version: "julia version 1.6.0"},
		{Case: "Julia 1.6.1", ExpectedString: "1.6.1", Version: "julia version 1.6.1"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "julia",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.jl",
		}
		env, props := getMockedLanguageEnv(params)
		j := NewLanguage("julia")
		j.Init(props, env)
		assert.True(t, j.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, j.Template(), j), tc.Case)
	}
}

func TestConfiguredLanguageKotlinPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Kotlin 1.6.10", ExpectedString: "1.6.10", Version: "Kotlin version 1.6.10-release-923 (JRE 17.0.2+0)"},
		{Case: "Kotlin 1.6.0", ExpectedString: "1.6.0", Version: "Kotlin version 1.6.0-release-915 (JRE 17.0.2+0)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "kotlin",
			versionParam:  "-version",
			versionOutput: tc.Version,
			extension:     "*.kt",
		}
		env, props := getMockedLanguageEnv(params)
		k := NewLanguage("kotlin")
		k.Init(props, env)
		assert.True(t, k.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, k.Template(), k), tc.Case)
	}
}

func TestConfiguredLanguageLuaPreset(t *testing.T) {
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
		env.On("RunCommandWithEnv", "luajit", []string(nil), []string{"-v"}).Return(tc.Version, nil)
		env.On("Shell").Return("bash")

		// Initialize template system for version URL rendering
		if template.Cache == nil {
			template.Cache = &cache.Template{}
		}
		template.Init(env, nil, nil)

		if len(tc.Tooling) > 0 {
			props[Tooling] = tc.Tooling
		}

		l := NewLanguage("lua")
		l.Init(props, env)

		failMsg := tc.Case
		assert.True(t, l.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, l.Template(), l), failMsg)
		assert.Equal(t, tc.ExpectedURL, l.URL, failMsg)
	}
}

func TestConfiguredLanguageNimPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "Nim 2.2.0",
			ExpectedString: "2.2.0",
			Version:        "Nim Compiler Version 2.2.0 [MacOSX: arm64]\nCompiled at 2024-11-30\nCopyright (c) 2006-2024 by Andreas Rumpf",
		},
		{
			Case:           "Nim 1.6.12",
			ExpectedString: "1.6.12",
			Version:        "Nim Compiler Version 1.6.12 [Linux: amd64]\nCompiled at 2023-06-15\nCopyright (c) 2006-2023 by Andreas Rumpf",
		},
		{
			Case:           "Nim 2.0.0",
			ExpectedString: "2.0.0",
			Version:        "Nim Compiler Version 2.0.0 [Windows: amd64]\nCompiled at 2023-12-25\nCopyright (c) 2006-2023 by Andreas Rumpf",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "nim",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.nim",
		}
		env, props := getMockedLanguageEnv(params)

		n := NewLanguage("nim")
		n.Init(props, env)

		assert.True(t, n.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, n.Template(), n), tc.Case)
	}
}

func TestConfiguredLanguageOCamlPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "OCaml 4.12.0", ExpectedString: "4.12.0", Version: "The OCaml toplevel, version 4.12.0"},
		{Case: "OCaml 4.11.0", ExpectedString: "4.11.0", Version: "The OCaml toplevel, version 4.11.0"},
		{Case: "OCaml 4.13.0", ExpectedString: "4.13.0", Version: "The OCaml toplevel, version 4.13.0"},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "ocaml",
			versionParam:  "-version",
			versionOutput: tc.Version,
			extension:     "*.ml",
		}
		env, props := getMockedLanguageEnv(params)

		o := NewLanguage("ocaml")
		o.Init(props, env)

		assert.True(t, o.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, o.Template(), o), tc.Case)
	}
}

func TestConfiguredLanguagePerlPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
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

		p := NewLanguage("perl")
		p.Init(props, env)

		assert.True(t, p.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, p.Template(), p), tc.Case)
	}
}

func TestConfiguredLanguagePhpPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "PHP 6.1.0", ExpectedString: "6.1.0", Version: "PHP 6.1.0(cli) (built: Jul  2 2021 03:59:48) ( NTS )"},
		{Case: "php 7.4.21", ExpectedString: "7.4.21", Version: "PHP 7.4.21 (cli) (built: Jul  2 2021 03:59:48) ( NTS )"},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "php",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.php",
		}
		env, props := getMockedLanguageEnv(params)

		j := NewLanguage("php")
		j.Init(props, env)

		assert.True(t, j.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, j.Template(), j), tc.Case)
	}
}

func TestConfiguredLanguageRPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		HasRscript     bool
		HasR           bool
		HasRexe        bool
	}{
		{Case: "Rscript 4.2.0", ExpectedString: "4.2.0", HasRscript: true, Version: "Rscript (R) version 4.2.0 (2022-04-22)"},
		{Case: "Rscript 4.1.3", ExpectedString: "4.1.3", HasRscript: true, Version: "R scripting front-end version 4.1.3 (2022-03-10)"},
		{Case: "Rscript 4.1.3 patched", ExpectedString: "4.1.3", HasRscript: true, Version: "R scripting front-end version 4.1.3 Patched (2022-03-10 r81896)"},
		{Case: "Rscript 4.0.0", ExpectedString: "4.0.0", HasRscript: true, Version: "R scripting front-end version 4.0.0 (2020-04-24)"},
		{Case: "Rscript devel", ExpectedString: "4.2.0", HasRscript: true, Version: "R scripting front-end version 4.2.0 Under development (unstable) (2022-03-14 r81896)"},

		{Case: "R 4.1.2", ExpectedString: "4.1.2", HasR: true, Version: "R version 4.1.2 (2021-11-01) -- \"Bird Hippie\""},
		{Case: "R 4.1.3 patched", ExpectedString: "4.1.3", HasR: true, Version: "R version 4.1.3 Patched (2022-03-10 r81896) -- \"One Push-Up\""},
		{Case: "R 4.0.0", ExpectedString: "4.0.0", HasR: true, Version: "R version 4.0.0 (2020-04-24) -- \"Arbor Day\""},

		{Case: "R.exe 4.1.2", ExpectedString: "4.1.2", HasRexe: true, Version: "R version 4.1.2 (2021-11-01) -- \"Bird Hippie\""},
		{Case: "R.exe 4.1.3 patched", ExpectedString: "4.1.3", HasRexe: true, Version: "R version 4.1.3 Patched (2022-03-10 r81896) -- \"One Push-Up\""},
		{Case: "R.exe 4.0.0", ExpectedString: "4.0.0", HasRexe: true, Version: "R version 4.0.0 (2020-04-24) -- \"Arbor Day\""},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "R",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.R",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasCommand", "Rscript").Return(tc.HasRscript)
		env.On("RunCommandWithEnv", "Rscript", []string(nil), []string{"--version"}).Return(tc.Version, nil)
		env.On("HasCommand", "R.exe").Return(tc.HasRexe)
		env.On("RunCommandWithEnv", "R.exe", []string(nil), []string{"--version"}).Return(tc.Version, nil)

		r := NewLanguage("r")
		r.Init(props, env)

		assert.True(t, r.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), tc.Case)
	}
}

func TestConfiguredLanguageRustPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Rust 1.64.0", ExpectedString: "1.64.0", Version: "rustc 1.64.0"},
		{Case: "Rust 1.53.0", ExpectedString: "1.53.0", Version: "rustc 1.53.0 (4369396ce 2021-04-27)"},
		{Case: "Rust 1.66.0", ExpectedString: "1.66.0-nightly", Version: "rustc 1.66.0-nightly (01af5040f 2022-10-04)"},
		{
			Case:           "Toolchain not installed",
			ExpectedString: "1.81.0",
			Version: ` info: syncing channel updates for '1.81.0-x86_64-pc-windows-msvc'
	    info: latest update on 2024-09-05, rust version 1.81.0 (eeb90cda1 2024-09-04)
	    info: downloading component 'cargo'
	    info: downloading component 'clippy'
	    info: downloading component 'rust-analyzer'
	    info: downloading component 'rust-src'
	    info: downloading component 'rust-std'
	    info: downloading component 'rustc'
	    info: downloading component 'rustfmt'
	    info: installing component 'cargo'
	    info: installing component 'clippy'
	    info: installing component 'rust-analyzer'
	    info: installing component 'rust-src'
	    info: installing component 'rust-std'
	    info: installing component 'rustc'`,
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "rustc",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.rs",
		}
		env, props := getMockedLanguageEnv(params)
		r := NewLanguage("rust")
		r.Init(props, env)
		assert.True(t, r.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), tc.Case)
	}
}

func TestConfiguredLanguageSwiftPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Swift 5.5.3", ExpectedString: "5.5.3", Version: "Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5.3 on Windows", ExpectedString: "5.5.3", Version: "compnerd.org Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5.3 on Mac", ExpectedString: "5.5.3", Version: "Apple Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5", ExpectedString: "5.5", Version: "Swift version 5.5 (swift-5.5-RELEASE)"},
		{Case: "Swift 5.6-dev", ExpectedString: "5.6-dev", Version: "Swift version 5.6-dev (LLVM 62b900d3d0d5be9, Swift ce64fe8867792d4)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "swift",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.swift",
		}
		env, props := getMockedLanguageEnv(params)
		s := NewLanguage("swift")
		s.Init(props, env)
		assert.True(t, s.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
	}
}

func TestConfiguredLanguageVPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "V 0.4.9",
			ExpectedString: "0.4.9",
			Version:        "V 0.4.9 b487986",
		},
		{
			Case:           "V 0.4.8",
			ExpectedString: "0.4.8",
			Version:        "V 0.4.8 a123456",
		},
		{
			Case:           "V 0.4.7",
			ExpectedString: "0.4.7",
			Version:        "V 0.4.7 f789012",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "v",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.v",
		}
		env, props := getMockedLanguageEnv(params)
		v := NewLanguage("v")
		v.Init(props, env)
		assert.True(t, v.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, v.Template(), v), tc.Case)
	}
}

func TestConfiguredLanguageValaPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "vala 0.48.17",
			ExpectedString: "0.48.17",
			Version:        "Vala 0.48.17",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "vala",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.vala",
		}
		env, props := getMockedLanguageEnv(params)
		v := NewLanguage("vala")
		v.Init(props, env)
		assert.True(t, v.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, v.Template(), v), tc.Case)
	}
}

// TestConfiguredLanguageCustomTools exercises the `tools` option path used by the public
// "language" segment type, which has no built-in preset and relies entirely on user config.
func TestConfiguredLanguageCustomTools(t *testing.T) {
	env := new(mock.Environment)
	env.On("HasCommand", "mytool").Return(true)
	env.On("RunCommandWithEnv", "mytool", []string(nil), []string{"--version"}).Return("mytool version 1.2.3", nil)
	env.On("HasFiles", "*.myl").Return(true)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")

	props := options.Map{
		options.FetchVersion: true,
		LanguageName:         "mylang",
		LanguageExtensions:   []string{"*.myl"},
		Tools: []any{
			map[string]any{
				"name":       "mytool",
				"executable": "mytool",
				"args":       []any{"--version"},
				"regex":      `mytool version (?P<version>(?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))`,
			},
		},
	}

	l := &ConfiguredLanguage{}
	l.Init(props, env)

	assert.True(t, l.Enabled())
	assert.Equal(t, "1.2.3", renderTemplate(env, l.Template(), l))
	assert.Equal(t, []string{"mytool"}, l.defaultTooling)
}

func TestConfiguredLanguageZigPreset(t *testing.T) {
	cases := []struct {
		Case           string
		Version        string
		ExpectedString string
		ExpectedURL    string
		InProjectDir   bool
	}{
		{
			Case:           "zig 0.13.0 - not in project dir",
			Version:        "0.13.0",
			InProjectDir:   false,
			ExpectedString: "0.13.0",
			ExpectedURL:    "https://ziglang.org/download/0.13.0/release-notes.html",
		},
		{
			Case:           "zig 0.12.0-dev.2063+804cee3b9 - not in project dir",
			Version:        "0.12.0-dev.2063+804cee3b9",
			InProjectDir:   false,
			ExpectedString: "0.12.0-dev.2063+804cee3b9",
			ExpectedURL:    "https://ziglang.org/download/0.12.0/release-notes.html",
		},
		{
			Case:           "zig 0.13.0 - in project dir",
			Version:        "0.13.0",
			InProjectDir:   true,
			ExpectedString: "0.13.0",
			ExpectedURL:    "https://ziglang.org/download/0.13.0/release-notes.html",
		},
		{
			Case:           "zig 0.12.0-dev.2063+804cee3b9 - in project dir",
			Version:        "0.12.0-dev.2063+804cee3b9",
			InProjectDir:   true,
			ExpectedString: "0.12.0-dev.2063+804cee3b9",
			ExpectedURL:    "https://ziglang.org/download/0.12.0/release-notes.html",
		},
	}

	// Initialize template system for version URL rendering
	if template.Cache == nil {
		template.Cache = &cache.Template{}
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "zig",
			versionParam:  "version",
			versionOutput: tc.Version,
			extension:     "*.zig",
		}

		env, props := getMockedLanguageEnv(params)
		env.On("Shell").Return("bash")
		template.Init(env, nil, nil)

		dummyDir := &runtime.FileInfo{}

		if tc.InProjectDir {
			env.On("HasParentFilePath", "build.zig", false).Return(dummyDir, nil)
		} else {
			env.On("HasParentFilePath", "build.zig", false).Return(dummyDir, errors.New("build.zig not found"))
		}

		zig := NewLanguage("zig")
		zig.Init(props, env)

		assert.True(t, zig.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, zig.Template(), zig), tc.Case)
		assert.Equal(t, tc.ExpectedURL, renderTemplate(env, zig.URL, zig), tc.Case)
		assert.Equal(t, tc.InProjectDir, zig.InProjectDir(), tc.Case)
	}
}

func TestConfiguredLanguageDartPreset(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Dart 2.12.4", ExpectedString: "2.12.4", Version: "Dart SDK version: 2.12.4 (stable) (Thu Apr 15 12:26:53 2021 +0200) on \"macos_x64\""},
	}

	// Initialize template system for version URL rendering
	if template.Cache == nil {
		template.Cache = &cache.Template{}
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "dart",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.dart",
		}

		env, props := getMockedLanguageEnv(params)
		env.On("Shell").Return("bash")
		template.Init(env, nil, nil)
		env.On("HasCommand", "fvm").Return(false)

		d := NewLanguage("dart")
		d.Init(props, env)

		assert.True(t, d.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, d.Template(), d), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
