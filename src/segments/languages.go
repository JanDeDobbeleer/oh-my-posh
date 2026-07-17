package segments

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// languageDefinition is a built-in preset for ConfiguredLanguage. Each entry mirrors what a
// dedicated language segment used to hardcode: which files trigger it, which executables can
// report its version, and how to parse that output.
type languageDefinition struct {
	tooling            map[string]*cmd
	sanitizeVersion    func(string) string
	versionURLTemplate string
	extensions         []string
	projectFiles       []string
	folders            []string
	defaultTooling     []string
}

const (
	gfortranToolName  = "gfortran"
	rbenvToolName     = "rbenv"
	rvmPromptToolName = "rvm-prompt"
	chrubyToolName    = "chruby"
	rubyToolName      = "ruby"
	clojureToolName   = "clojure"
	leinToolName      = "lein"
	crystalToolName   = "crystal"
	elixirToolName    = "elixir"
	kotlinToolName    = "kotlin"
	luaToolName       = "lua"
	luajitToolName    = "luajit"
	nimToolName       = "nim"
	ocamlToolName     = "ocaml"
	perlToolName      = "perl"
	rscriptToolName   = "Rscript"
	rExeToolName      = "R.exe"
	rustcToolName     = "rustc"
	swiftToolName     = "swift"
	valaToolName      = "vala"
	zigToolName       = "zig"
)

// languageDefinitions holds the built-in presets, keyed by the same name a migrated segment's
// SegmentType string used (e.g. "fortran"). ConfiguredLanguage looks itself up here by name;
// a name with no entry (the public "language" segment type) relies entirely on the `tools` option.
var languageDefinitions = map[string]languageDefinition{
	"fortran": {
		extensions: []string{
			"*.f", "*.for", "*.fpp",
			"*.f77", "*.f90", "*.f95",
			"*.f03", "*.f08",
			"*.F", "*.FOR", "*.FPP",
			"*.F77", "*.F90", "*.F95",
			"*.F03", "*.F08",
			"fpm.toml",
		},
		tooling: map[string]*cmd{
			gfortranToolName: {
				executable: gfortranToolName,
				args:       []string{versionFlagArg},
				regex:      `GNU Fortran \(.*\) ` + versionRegex,
			},
		},
		defaultTooling: []string{gfortranToolName},
	},
	rubyToolName: {
		extensions: []string{"*.rb", "Rakefile", "Gemfile"},
		tooling: map[string]*cmd{
			rbenvToolName: {
				executable: rbenvToolName,
				args:       []string{"version-name"},
				regex:      `(?P<version>.+)`,
			},
			rvmPromptToolName: {
				executable: rvmPromptToolName,
				args:       []string{"i", "v", "g"},
				regex:      `(?P<version>.+)`,
			},
			chrubyToolName: {
				executable: chrubyToolName,
				args:       []string(nil),
				regex:      `\* (?P<version>.+)\n`,
			},
			asdfToolName: {
				executable: asdfToolName,
				args:       []string{"current", rubyToolName},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
			rubyToolName: {
				executable: rubyToolName,
				args:       []string{versionFlagArg},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
		},
		defaultTooling: []string{rbenvToolName, rvmPromptToolName, chrubyToolName, asdfToolName, rubyToolName},
		// asdf reports "______" when no version is set for the tool; that isn't a version, it's an absence of one.
		sanitizeVersion: func(version string) string {
			if version == "______" {
				return ""
			}
			return version
		},
	},
	"clojure": {
		extensions: []string{
			"project.clj",
			"deps.edn",
			"build.boot",
			"bb.edn",
			"*.clj",
			"*.cljc",
			"*.cljs",
		},
		tooling: map[string]*cmd{
			clojureToolName: {
				executable: clojureToolName,
				args:       []string{versionFlagArg},
				regex:      `Clojure CLI version (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+)(?:\.(?P<build>[0-9]+))?)`,
			},
			leinToolName: {
				executable: leinToolName,
				args:       []string{versionFlagArg},
				regex:      `Leiningen (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+))`,
			},
		},
		defaultTooling: []string{clojureToolName, leinToolName},
	},
	"crystal": {
		extensions: []string{"*.cr", "shard.yml"},
		tooling: map[string]*cmd{
			crystalToolName: {
				executable: crystalToolName,
				args:       []string{versionFlagArg},
				regex:      `Crystal ` + versionRegex,
			},
		},
		defaultTooling:     []string{crystalToolName},
		versionURLTemplate: "https://github.com/crystal-lang/crystal/releases/tag/{{ .Full }}",
	},
	"elixir": {
		extensions: []string{"*.ex", "*.exs"},
		tooling: map[string]*cmd{
			asdfToolName: {
				executable: asdfToolName,
				args:       []string{"current", elixirToolName},
				regex:      `elixir\s+` + versionRegex + `[^\s]*\s+`,
			},
			elixirToolName: {
				executable: elixirToolName,
				args:       []string{versionFlagArg},
				regex:      `Elixir ` + versionRegex,
			},
		},
		defaultTooling:     []string{asdfToolName, elixirToolName},
		versionURLTemplate: "https://github.com/elixir-lang/elixir/releases/tag/v{{ .Full }}",
	},
	"julia": {
		extensions: []string{"*.jl"},
		tooling: map[string]*cmd{
			juliaToolName: {
				executable: juliaToolName,
				args:       []string{versionFlagArg},
				regex:      `julia version ` + versionRegex,
			},
		},
		defaultTooling:     []string{juliaToolName},
		versionURLTemplate: "https://github.com/JuliaLang/julia/releases/tag/v{{ .Full }}",
	},
	"kotlin": {
		extensions: []string{"*.kt", "*.kts", "*.ktm"},
		tooling: map[string]*cmd{
			kotlinToolName: {
				executable: kotlinToolName,
				args:       []string{versionFlagShortArg},
				regex:      `Kotlin version ` + versionRegex,
			},
		},
		defaultTooling:     []string{kotlinToolName},
		versionURLTemplate: "https://github.com/JetBrains/kotlin/releases/tag/v{{ .Full }}",
	},
	"lua": {
		extensions: []string{"*.lua", "*.rockspec"},
		folders:    []string{"lua"},
		tooling: map[string]*cmd{
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
		},
		defaultTooling: []string{luaToolName, luajitToolName},
	},
	"nim": {
		extensions: []string{"*.nim", "*.nims"},
		tooling: map[string]*cmd{
			nimToolName: {
				executable: nimToolName,
				args:       []string{versionFlagArg},
				regex:      `Nim Compiler Version (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
			},
		},
		defaultTooling: []string{nimToolName},
	},
	"ocaml": {
		extensions: []string{"*.ml", "*.mli", "dune", "dune-project", "dune-workspace"},
		tooling: map[string]*cmd{
			ocamlToolName: {
				executable: ocamlToolName,
				args:       []string{versionFlagShortArg},
				regex:      `The OCaml toplevel, version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)`,
			},
		},
		defaultTooling: []string{ocamlToolName},
	},
	"perl": {
		extensions: []string{".perl-version", "*.pl", "*.pm", "*.t"},
		tooling: map[string]*cmd{
			perlToolName: {
				executable: perlToolName,
				args:       []string{versionFlagShortArg},
				regex:      `This is perl.*v(?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))(?:\.(?P<patch>[0-9]+))?).* built for .+`,
			},
		},
		defaultTooling: []string{perlToolName},
	},
	"php": {
		extensions: []string{"*.php", "composer.json", "composer.lock", ".php-version", "blade.php"},
		tooling: map[string]*cmd{
			phpToolName: {
				executable: phpToolName,
				args:       []string{versionFlagArg},
				regex:      `(?:PHP (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		defaultTooling:     []string{phpToolName},
		versionURLTemplate: "https://www.php.net/ChangeLog-{{ .Major }}.php#PHP_{{ .Major }}_{{ .Minor }}",
	},
	"r": {
		extensions: []string{"*.R", "*.Rmd", "*.Rsx", "*.Rda", "*.Rd", "*.Rproj", ".Rproj.user"},
		tooling: map[string]*cmd{
			rscriptToolName: {
				executable: rscriptToolName,
				args:       []string{versionFlagArg},
				regex:      `version ` + versionRegex,
			},
			"R": {
				executable: "R",
				args:       []string{versionFlagArg},
				regex:      `version ` + versionRegex,
			},
			rExeToolName: {
				executable: rExeToolName,
				args:       []string{versionFlagArg},
				regex:      `version ` + versionRegex,
			},
		},
		defaultTooling:     []string{rscriptToolName, "R", rExeToolName},
		versionURLTemplate: "https://www.r-project.org/",
	},
	"rust": {
		extensions: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		tooling: map[string]*cmd{
			rustcToolName: {
				executable: rustcToolName,
				args:       []string{versionFlagArg},
				regex:      `(rust version|rustc) (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))(-(?P<prerelease>[a-z]+))?)(( \((?P<buildmetadata>[0-9a-f]+ [0-9]+-[0-9]+-[0-9]+)\))?)`, //nolint:lll
			},
		},
		defaultTooling: []string{rustcToolName},
	},
	"swift": {
		extensions: []string{"*.swift", "*.SWIFT", "Podfile"},
		tooling: map[string]*cmd{
			swiftToolName: {
				executable: swiftToolName,
				args:       []string{versionFlagArg},
				regex:      `Swift version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)((.|-)(?P<patch>[0-9]+|dev))?))`,
			},
		},
		defaultTooling:     []string{swiftToolName},
		versionURLTemplate: "https://github.com/apple/swift/releases/tag/swift-{{ .Full }}-RELEASE",
	},
	"v": {
		extensions: []string{"*.v"},
		tooling: map[string]*cmd{
			"v": {
				executable: "v",
				args:       []string{"--version"},
				regex:      `V (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)) [a-f0-9]+`,
			},
		},
		defaultTooling: []string{"v"},
	},
	"vala": {
		extensions: []string{"*.vala"},
		tooling: map[string]*cmd{
			valaToolName: {
				executable: valaToolName,
				args:       []string{versionFlagArg},
				regex:      `Vala (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		defaultTooling:     []string{valaToolName},
		versionURLTemplate: "https://gitlab.gnome.org/GNOME/vala/raw/{{ .Major }}.{{ .Minor }}/NEWS",
	},
	"zig": {
		extensions:   []string{"*.zig", "*.zon"},
		projectFiles: []string{"build.zig"},
		tooling: map[string]*cmd{
			zigToolName: {
				executable: zigToolName,
				args:       []string{versionArg},
				regex:      `(?P<version>(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)`, //nolint:lll
			},
		},
		defaultTooling:     []string{zigToolName},
		versionURLTemplate: "https://ziglang.org/download/{{ .Major }}.{{ .Minor }}.{{ .Patch }}/release-notes.html",
	},
	"dart": {
		extensions: []string{"*.dart", pubspecFileName, "pubspec.yml", "pubspec.lock"},
		folders:    []string{".dart_tool"},
		tooling: map[string]*cmd{
			fvmToolName: {
				executable: fvmToolName,
				args:       []string{dartToolName, versionFlagArg},
				regex:      `Dart SDK version: ` + versionRegex,
			},
			dartToolName: {
				executable: dartToolName,
				args:       []string{versionFlagArg},
				regex:      `Dart SDK version: ` + versionRegex,
			},
		},
		defaultTooling:     []string{fvmToolName, dartToolName},
		versionURLTemplate: "https://dart.dev/guides/language/evolution#dart-{{ .Major }}{{ .Minor }}",
	},
}

// ConfiguredLanguage is a generic Language segment driven entirely by data: a built-in preset
// looked up by name, and/or a user-supplied `tools` option. It backs both the migrated
// single-purpose segment types (which set name via NewLanguage) and the public
// "language" segment type (which reads name from the `name` option).
type ConfiguredLanguage struct {
	Language
}

func (c *ConfiguredLanguage) Template() string {
	return languageTemplate
}

// NewLanguage returns a ConfiguredLanguage preset to the built-in definition
// registered under name. Used by segment_types.go to back the migrated segment types.
func NewLanguage(name string) *ConfiguredLanguage {
	c := &ConfiguredLanguage{}
	c.name = name
	return c
}

func (c *ConfiguredLanguage) Enabled() bool {
	if c.name == "" {
		c.name = c.options.String(LanguageName, "")
	}

	if preset, ok := languageDefinitions[c.name]; ok {
		c.extensions = preset.extensions
		c.projectFiles = preset.projectFiles
		c.folders = preset.folders
		c.tooling = preset.tooling
		c.defaultTooling = preset.defaultTooling
		c.versionURLTemplate = preset.versionURLTemplate
	}

	c.applyCustomTools()

	enabled := c.Language.Enabled()

	if preset, ok := languageDefinitions[c.name]; ok && preset.sanitizeVersion != nil {
		c.Full = preset.sanitizeVersion(c.Full)
	}

	return enabled
}

// applyCustomTools reads the `tools` option (a list of {name, executable, args, regex,
// version_url_template}) and merges it into the tooling map. For presets, this lets a user
// add or override a tool. For the public "language" segment type, which has no preset, this
// is the only source of tooling, and its declaration order becomes the default tool order.
func (c *ConfiguredLanguage) applyCustomTools() {
	raw, ok := c.options.Any(Tools, nil).([]any)
	if !ok || len(raw) == 0 {
		return
	}

	if c.tooling == nil {
		c.tooling = make(map[string]*cmd, len(raw))
	}

	order := make([]string, 0, len(raw))

	for _, item := range raw {
		entry, ok := toStringMap(item)
		if !ok {
			continue
		}

		name := stringField(entry, "name")
		if name == "" {
			continue
		}

		c.tooling[name] = &cmd{
			executable:         stringField(entry, "executable"),
			args:               options.ParseStringArray(entry["args"]),
			regex:              stringField(entry, "regex"),
			versionURLTemplate: stringField(entry, "version_url_template"),
		}
		order = append(order, name)
	}

	if _, hasPreset := languageDefinitions[c.name]; !hasPreset {
		c.defaultTooling = order
	}
}

func toStringMap(value any) (map[string]any, bool) {
	switch v := value.(type) {
	case map[string]any:
		return v, true
	case map[any]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[fmt.Sprint(key)] = val
		}
		return out, true
	default:
		return nil, false
	}
}

func stringField(entry map[string]any, key string) string {
	value, ok := entry[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}
