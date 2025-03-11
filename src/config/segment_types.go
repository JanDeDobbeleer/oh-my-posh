package config

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
)

// SegmentType the type of segment, for more information, see the constants
type SegmentType string

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	Enabled() bool
	Template() string
	SetText(text string)
	SetIndex(index int)
	Text() string
	Init(props properties.Properties, env runtime.Environment)
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
	// ARGOCD writes the current argocd context
	ARGOCD SegmentType = "argocd"
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
	// CMAKE writes the active cmake version
	CMAKE SegmentType = "cmake"
	// CMD writes the output of a shell command
	CMD SegmentType = "command"
	// CONNECTION writes a connection's information
	CONNECTION SegmentType = "connection"
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
	// HASKELL segment
	HASKELL SegmentType = "haskell"
	// HELM segment
	HELM SegmentType = "helm"
	// IPIFY segment
	IPIFY SegmentType = "ipify"
	// JAVA writes the active java version
	JAVA SegmentType = "java"
	// JULIA writes which julia version is currently active
	JULIA SegmentType = "julia"
	// KOTLIN writes the active kotlin version
	KOTLIN SegmentType = "kotlin"
	// KUBECTL writes the Kubernetes context we're currently in
	KUBECTL SegmentType = "kubectl"
	// LASTFM writes the lastfm status
	LASTFM SegmentType = "lastfm"
	// LUA writes the active lua version
	LUA SegmentType = "lua"
	// MERCURIAL writes the Mercurial source control information
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
	// Tauri Segment
	TAURI SegmentType = "tauri"
	// TERRAFORM writes the terraform workspace we're currently in
	TERRAFORM SegmentType = "terraform"
	// TEXT writes a text
	TEXT SegmentType = "text"
	// TIME writes the current timestamp
	TIME SegmentType = "time"
	// UI5 Tooling segment
	UI5TOOLING SegmentType = "ui5tooling"
	// UMBRACO writes the Umbraco version if Umbraco is present
	UMBRACO SegmentType = "umbraco"
	// UNITY writes which Unity version is currently active
	UNITY SegmentType = "unity"
	// UPGRADE lets you know if you can upgrade Oh My Posh
	UPGRADE SegmentType = "upgrade"
	// V writes the active vlang version
	V SegmentType = "v"
	// VALA writes the active vala version
	VALA SegmentType = "vala"
	// WAKATIME writes tracked time spend in dev editors
	WAKATIME SegmentType = "wakatime"
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
)

// Segments contains all available prompt segment writers.
// Consumers of the library can also add their own segment writer.
var Segments = map[SegmentType]func() SegmentWriter{
	ANGULAR:         func() SegmentWriter { return &segments.Angular{} },
	ARGOCD:          func() SegmentWriter { return &segments.Argocd{} },
	AURELIA:         func() SegmentWriter { return &segments.Aurelia{} },
	AWS:             func() SegmentWriter { return &segments.Aws{} },
	AZ:              func() SegmentWriter { return &segments.Az{} },
	AZD:             func() SegmentWriter { return &segments.Azd{} },
	AZFUNC:          func() SegmentWriter { return &segments.AzFunc{} },
	BATTERY:         func() SegmentWriter { return &segments.Battery{} },
	BAZEL:           func() SegmentWriter { return &segments.Bazel{} },
	BREWFATHER:      func() SegmentWriter { return &segments.Brewfather{} },
	BUF:             func() SegmentWriter { return &segments.Buf{} },
	BUN:             func() SegmentWriter { return &segments.Bun{} },
	CARBONINTENSITY: func() SegmentWriter { return &segments.CarbonIntensity{} },
	CDS:             func() SegmentWriter { return &segments.Cds{} },
	CF:              func() SegmentWriter { return &segments.Cf{} },
	CFTARGET:        func() SegmentWriter { return &segments.CfTarget{} },
	CMAKE:           func() SegmentWriter { return &segments.Cmake{} },
	CMD:             func() SegmentWriter { return &segments.Cmd{} },
	CONNECTION:      func() SegmentWriter { return &segments.Connection{} },
	CRYSTAL:         func() SegmentWriter { return &segments.Crystal{} },
	DART:            func() SegmentWriter { return &segments.Dart{} },
	DENO:            func() SegmentWriter { return &segments.Deno{} },
	DOCKER:          func() SegmentWriter { return &segments.Docker{} },
	DOTNET:          func() SegmentWriter { return &segments.Dotnet{} },
	ELIXIR:          func() SegmentWriter { return &segments.Elixir{} },
	EXECUTIONTIME:   func() SegmentWriter { return &segments.Executiontime{} },
	EXIT:            func() SegmentWriter { return &segments.Status{} },
	FIREBASE:        func() SegmentWriter { return &segments.Firebase{} },
	FLUTTER:         func() SegmentWriter { return &segments.Flutter{} },
	FORTRAN:         func() SegmentWriter { return &segments.Fortran{} },
	FOSSIL:          func() SegmentWriter { return &segments.Fossil{} },
	GCP:             func() SegmentWriter { return &segments.Gcp{} },
	GIT:             func() SegmentWriter { return &segments.Git{} },
	GITVERSION:      func() SegmentWriter { return &segments.GitVersion{} },
	GOLANG:          func() SegmentWriter { return &segments.Golang{} },
	HASKELL:         func() SegmentWriter { return &segments.Haskell{} },
	HELM:            func() SegmentWriter { return &segments.Helm{} },
	IPIFY:           func() SegmentWriter { return &segments.IPify{} },
	JAVA:            func() SegmentWriter { return &segments.Java{} },
	JULIA:           func() SegmentWriter { return &segments.Julia{} },
	KOTLIN:          func() SegmentWriter { return &segments.Kotlin{} },
	KUBECTL:         func() SegmentWriter { return &segments.Kubectl{} },
	LASTFM:          func() SegmentWriter { return &segments.LastFM{} },
	LUA:             func() SegmentWriter { return &segments.Lua{} },
	MERCURIAL:       func() SegmentWriter { return &segments.Mercurial{} },
	MOJO:            func() SegmentWriter { return &segments.Mojo{} },
	MVN:             func() SegmentWriter { return &segments.Mvn{} },
	NBA:             func() SegmentWriter { return &segments.Nba{} },
	NBGV:            func() SegmentWriter { return &segments.Nbgv{} },
	NIGHTSCOUT:      func() SegmentWriter { return &segments.Nightscout{} },
	NIXSHELL:        func() SegmentWriter { return &segments.NixShell{} },
	NIM:             func() SegmentWriter { return &segments.Nim{} },
	NODE:            func() SegmentWriter { return &segments.Node{} },
	NPM:             func() SegmentWriter { return &segments.Npm{} },
	NX:              func() SegmentWriter { return &segments.Nx{} },
	OCAML:           func() SegmentWriter { return &segments.OCaml{} },
	OS:              func() SegmentWriter { return &segments.Os{} },
	OWM:             func() SegmentWriter { return &segments.Owm{} },
	PATH:            func() SegmentWriter { return &segments.Path{} },
	PERL:            func() SegmentWriter { return &segments.Perl{} },
	PHP:             func() SegmentWriter { return &segments.Php{} },
	PLASTIC:         func() SegmentWriter { return &segments.Plastic{} },
	PNPM:            func() SegmentWriter { return &segments.Pnpm{} },
	PROJECT:         func() SegmentWriter { return &segments.Project{} },
	PULUMI:          func() SegmentWriter { return &segments.Pulumi{} },
	PYTHON:          func() SegmentWriter { return &segments.Python{} },
	QUASAR:          func() SegmentWriter { return &segments.Quasar{} },
	R:               func() SegmentWriter { return &segments.R{} },
	REACT:           func() SegmentWriter { return &segments.React{} },
	ROOT:            func() SegmentWriter { return &segments.Root{} },
	RUBY:            func() SegmentWriter { return &segments.Ruby{} },
	RUST:            func() SegmentWriter { return &segments.Rust{} },
	SAPLING:         func() SegmentWriter { return &segments.Sapling{} },
	SESSION:         func() SegmentWriter { return &segments.Session{} },
	SHELL:           func() SegmentWriter { return &segments.Shell{} },
	SITECORE:        func() SegmentWriter { return &segments.Sitecore{} },
	SPOTIFY:         func() SegmentWriter { return &segments.Spotify{} },
	STATUS:          func() SegmentWriter { return &segments.Status{} },
	STRAVA:          func() SegmentWriter { return &segments.Strava{} },
	SVELTE:          func() SegmentWriter { return &segments.Svelte{} },
	SVN:             func() SegmentWriter { return &segments.Svn{} },
	SWIFT:           func() SegmentWriter { return &segments.Swift{} },
	SYSTEMINFO:      func() SegmentWriter { return &segments.SystemInfo{} },
	TALOSCTL:        func() SegmentWriter { return &segments.TalosCTL{} },
	TAURI:           func() SegmentWriter { return &segments.Tauri{} },
	TERRAFORM:       func() SegmentWriter { return &segments.Terraform{} },
	TEXT:            func() SegmentWriter { return &segments.Text{} },
	TIME:            func() SegmentWriter { return &segments.Time{} },
	UI5TOOLING:      func() SegmentWriter { return &segments.UI5Tooling{} },
	UMBRACO:         func() SegmentWriter { return &segments.Umbraco{} },
	UNITY:           func() SegmentWriter { return &segments.Unity{} },
	UPGRADE:         func() SegmentWriter { return &segments.Upgrade{} },
	V:               func() SegmentWriter { return &segments.V{} },
	VALA:            func() SegmentWriter { return &segments.Vala{} },
	WAKATIME:        func() SegmentWriter { return &segments.Wakatime{} },
	WINREG:          func() SegmentWriter { return &segments.WindowsRegistry{} },
	WITHINGS:        func() SegmentWriter { return &segments.Withings{} },
	XMAKE:           func() SegmentWriter { return &segments.XMake{} },
	YARN:            func() SegmentWriter { return &segments.Yarn{} },
	YTM:             func() SegmentWriter { return &segments.Ytm{} },
	ZIG:             func() SegmentWriter { return &segments.Zig{} },
}

func (segment *Segment) MapSegmentWithWriter(env runtime.Environment) error {
	segment.env = env

	if segment.Properties == nil {
		segment.Properties = make(properties.Map)
	}

	f, ok := Segments[segment.Type]
	if !ok {
		return errors.New("unable to map writer")
	}

	writer := f()
	wrapper := &properties.Wrapper{
		Properties: segment.Properties,
	}

	writer.Init(wrapper, env)
	segment.writer = writer

	return nil
}
