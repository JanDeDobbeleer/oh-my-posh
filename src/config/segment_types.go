package config

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// SegmentType the type of segment, for more information, see the constants
type SegmentType string

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	Enabled() bool
	// Activation declares the cheap preconditions under which the writer can
	// possibly be enabled; the engine skips Enabled() entirely when none of
	// them holds (see runtime.Activation for the contract). Called after
	// Init(), so implementations may consult their options. segments.Base
	// provides the default (Always) implementation.
	Activation() runtime.Activation
	Template() string
	SetText(text string)
	SetIndex(index int)
	Text() string
	Init(props options.Provider, env runtime.Environment)
	CacheKey() (string, bool)
}

const (
	// Plain writes it without ornaments
	Plain SegmentStyle = "plain"
	// Powerline writes it Powerline style
	Powerline SegmentStyle = "powerline"
	// Accordion writes it Powerline style but collapses the segment when disabled instead of hiding
	Accordion SegmentStyle = "accordion"
	// Diamond writes the prompt shaped with a leading and trailing symbol
	Diamond SegmentStyle = "diamond"
	// ANGULAR writes which angular cli version us currently active
	ANGULAR SegmentType = "angular"
	// ANTIGRAVITY writes Antigravity CLI session information
	ANTIGRAVITY SegmentType = "antigravity"
	// ARGOCD writes the current argocd context
	ARGOCD SegmentType = "argocd"
	// ASPIRE writes the Aspire apphost status
	ASPIRE SegmentType = "aspire"
	// AURELIA writes which aurelia version is currently referenced in package.json
	AURELIA SegmentType = "aurelia"
	// AWS writes the active aws context
	AWS SegmentType = "aws"
	// AZ writes the Azure subscription info we're currently in
	AZ SegmentType = "az"
	// AZD writes the Azure Developer CLI environment info we're current in
	AZD SegmentType = "azd"
	// AZFUNC writes current AZ func version
	AZFUNC SegmentType = "azfunc"
	// BATTERY writes the battery percentage
	BATTERY SegmentType = "battery"
	// BAZEL writes the bazel version
	BAZEL SegmentType = "bazel"
	// Brewfather segment
	BREWFATHER SegmentType = "brewfather"
	// Buf segment writes the active buf version
	BUF SegmentType = "buf"
	// BUN writes the active bun version
	BUN SegmentType = "bun"
	// CARBONINTENSITY writes the actual and forecast carbon intensity in gCO2/kWh
	CARBONINTENSITY SegmentType = "carbonintensity"
	// cds (SAP CAP) version
	CDS SegmentType = "cds"
	// Cloud Foundry segment
	CF SegmentType = "cf"
	// Cloud Foundry logged in target
	CFTARGET SegmentType = "cftarget"
	// CLAUDE writes Claude Code session information
	CLAUDE SegmentType = "claude"
	// CLOJURE writes the active clojure version
	CLOJURE SegmentType = "clojure"
	// CMAKE writes the active cmake version
	CMAKE SegmentType = "cmake"
	// CONNECTION writes a connection's information
	CONNECTION SegmentType = "connection"
	// COPILOT writes GitHub Copilot usage statistics
	COPILOT SegmentType = "copilot"
	// COPILOTCLI writes GitHub Copilot CLI session information
	COPILOTCLI SegmentType = "copilot_cli"
	// CRYSTAL writes the active crystal version
	CRYSTAL SegmentType = "crystal"
	// DART writes the active dart version
	DART SegmentType = "dart"
	// DENO writes the active deno version
	DENO SegmentType = "deno"
	// DOCKER writes the docker context
	DOCKER SegmentType = "docker"
	// DOTNET writes which dotnet version is currently active
	DOTNET SegmentType = "dotnet"
	// DVC writes the dvc status
	DVC SegmentType = "dvc"
	// ELIXIR writes the elixir version
	ELIXIR SegmentType = "elixir"
	// EXECUTIONTIME writes the execution time of the last run command
	EXECUTIONTIME SegmentType = "executiontime"
	// EXIT writes the last exit code
	EXIT SegmentType = "exit"
	// FIREBASE writes the active firebase project
	FIREBASE SegmentType = "firebase"
	// FLUTTER writes the flutter version
	FLUTTER SegmentType = "flutter"
	// FORTRAN writes the gfortran version
	FORTRAN SegmentType = "fortran"
	// FOSSIL writes the fossil status
	FOSSIL SegmentType = "fossil"
	// GCP writes the active GCP context
	GCP SegmentType = "gcp"
	// GIT represents the git status and information
	GIT SegmentType = "git"
	// GITVERSION represents the gitversion information
	GITVERSION SegmentType = "gitversion"
	// GOLANG writes which go version is currently active
	GOLANG SegmentType = "go"
	// GRADLE writes the active gradle version
	GRADLE SegmentType = "gradle"
	// HASKELL segment
	HASKELL SegmentType = "haskell"
	// HELM segment
	HELM SegmentType = "helm"
	// IPIFY segment
	IPIFY SegmentType = "ipify"
	// JAVA writes the active java version
	JAVA SegmentType = "java"
	// API writes the output of a custom JSON API
	HTTP SegmentType = "http"
	// JUJUTSU writes Jujutsu source control information
	JUJUTSU SegmentType = "jujutsu"
	// JULIA writes which julia version is currently active
	JULIA SegmentType = "julia"
	// KOTLIN writes the active kotlin version
	KOTLIN SegmentType = "kotlin"
	// KUBECTL writes the Kubernetes context we're currently in
	KUBECTL SegmentType = "kubectl"
	// LANGUAGE writes the version of a user-configured language/tool, detected via custom tools
	LANGUAGE SegmentType = "language"
	// LASTFM writes the lastfm status
	LASTFM SegmentType = "lastfm"
	// LUA writes the active lua version
	LUA SegmentType = "lua"
	// MERCURIAL writes Mercurial source control information
	MERCURIAL SegmentType = "mercurial"
	// MOJO writes the active version of Mojo and the name of the Magic virtual env
	MOJO SegmentType = "mojo"
	// MVN writes the active maven version
	MVN SegmentType = "mvn"
	// NBA writes NBA game data
	NBA SegmentType = "nba"
	// NBGV writes the nbgv version information
	NBGV SegmentType = "nbgv"
	// NIGHTSCOUT is an open source diabetes system
	NIGHTSCOUT SegmentType = "nightscout"
	// NIM writes the active nim version
	NIM SegmentType = "nim"
	// NIXSHELL writes the active nix shell details
	NIXSHELL SegmentType = "nix-shell"
	// NODE writes which node version is currently active
	NODE SegmentType = "node"
	// npm version
	NPM SegmentType = "npm"
	// NX writes which Nx version us currently active
	NX SegmentType = "nx"
	// OCAML writes the active Ocaml version
	OCAML SegmentType = "ocaml"
	// ORTHODOXCAL displays Orthodox fasting and feast information
	ORTHODOXCAL SegmentType = "orthodoxcal"
	// OS write os specific icon
	OS SegmentType = "os"
	// OWM writes the weather coming from openweatherdata
	OWM SegmentType = "owm"
	// PATH represents the current path segment
	PATH SegmentType = "path"
	// PERL writes which perl version is currently active
	PERL SegmentType = "perl"
	// PHP writes which php version is currently active
	PHP SegmentType = "php"
	// PLASTIC represents the plastic scm status and information
	PLASTIC SegmentType = "plastic"
	// pnpm version
	PNPM SegmentType = "pnpm"
	// Project version
	PROJECT SegmentType = "project"
	// PULUMI writes the pulumi user, store and stack
	PULUMI SegmentType = "pulumi"
	// PYTHON writes the virtual env name
	PYTHON SegmentType = "python"
	// QUASAR writes the QUASAR version and context
	QUASAR SegmentType = "quasar"
	// R version
	R SegmentType = "r"
	// RAMADAN displays Sehar and Iftar prayer times during Ramadan
	RAMADAN SegmentType = "ramadan"
	// REACT writes the current react version
	REACT SegmentType = "react"
	// ROOT writes root symbol
	ROOT SegmentType = "root"
	// RUBY writes which ruby version is currently active
	RUBY SegmentType = "ruby"
	// RUST writes the cargo version information if cargo.toml is present
	RUST SegmentType = "rust"
	// SAPLING represents the sapling segment
	SAPLING SegmentType = "sapling"
	// SESSION represents the user info segment
	SESSION SegmentType = "session"
	// SHELL writes which shell we're currently in
	SHELL SegmentType = "shell"
	// SITECORE displays the current context for the Sitecore CLI
	SITECORE SegmentType = "sitecore"
	// SPOTIFY writes the SPOTIFY status for Mac
	SPOTIFY SegmentType = "spotify"
	// STATUS writes the last know command status
	STATUS SegmentType = "status"
	// STRAVA is a sports activity tracker
	STRAVA SegmentType = "strava"
	// Svelte segment
	SVELTE SegmentType = "svelte"
	// Subversion segment
	SVN SegmentType = "svn"
	// SWIFT writes the active swift version
	SWIFT SegmentType = "swift"
	// SYSTEMINFO writes system information (memory, cpu, load)
	SYSTEMINFO SegmentType = "sysinfo"
	// TALOSCTL writes the talosctl context
	TALOSCTL SegmentType = "talosctl"
	// TASKWARRIOR writes Taskwarrior task counts and context
	TASKWARRIOR SegmentType = "taskwarrior"
	// Tauri Segment
	TAURI SegmentType = "tauri"
	// TERRAFORM writes the terraform workspace we're currently in
	TERRAFORM SegmentType = "terraform"
	// TEXT writes a text
	TEXT SegmentType = "text"
	// TIME writes the current timestamp
	TIME SegmentType = "time"
	// TODOIST segment
	TODOIST SegmentType = "todoist"
	// UI5 Tooling segment
	UI5TOOLING SegmentType = "ui5tooling"
	// UMBRACO writes the Umbraco version if Umbraco is present
	UMBRACO SegmentType = "umbraco"
	// UNO writes the Uno.Sdk version from global.json
	UNO SegmentType = "uno"
	// UNITY writes which Unity version is currently active
	UNITY SegmentType = "unity"
	// UPGRADE lets you know if you can upgrade Oh My Posh
	UPGRADE SegmentType = "upgrade"
	// V writes the active vlang version
	V SegmentType = "v"
	// VALA writes the active vala version
	VALA SegmentType = "vala"
	// VIMODE writes the active Vi mode (insert/normal/visual)
	VIMODE SegmentType = "vimode"
	// WAKATIME writes tracked time spend in dev editors
	WAKATIME SegmentType = "wakatime"
	// WINGET writes the number of available WinGet package updates
	WINGET SegmentType = "winget"
	// WINREG queries the Windows registry.
	WINREG SegmentType = "winreg"
	// WITHINGS queries the Withings API.
	WITHINGS SegmentType = "withings"
	// XMAKE write the xmake version if xmake.lua is present
	XMAKE SegmentType = "xmake"
	// yarn version
	YARN SegmentType = "yarn"
	// YTM writes YouTube Music information and status
	YTM SegmentType = "ytm"
	// ZIG writes the active zig version
	ZIG SegmentType = "zig"
	// ZVM writes the active zig version used in the zvm environment
	ZVM SegmentType = "zvm"
)

// Consumers of the library can also add their own segment writer.

func (segment *Segment) MapSegmentWithWriter(env runtime.Environment) error {
	segment.env = env

	if segment.Options == nil {
		segment.Options = make(options.Map)
	}

	writer, err := newSegmentWriter(segment.Type)
	if err != nil {
		return err
	}

	// nil on a build with no writers (see segment_registry_js.go); the segment renders from its
	// recorded data instead.
	if writer != nil {
		writer.Init(segment.Options, env)
	}

	segment.writer = writer

	return nil
}
