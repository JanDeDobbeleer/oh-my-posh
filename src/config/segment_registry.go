//go:build !(js && wasm)

package config

import (
	"encoding/gob"
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/segments"
)

// newSegmentWriter constructs the writer a segment renders through. Split out per platform: the
// js/wasm build renders only from recorded data and has no writers at all (see
// segment_registry_js.go), and keeping the registry itself out of that build is what lets the
// linker drop every segment package with it - measured at 4.2MB raw, 0.5MB brotli.
func newSegmentWriter(segmentType SegmentType) (SegmentWriter, error) {
	f, ok := Segments[segmentType]
	if !ok {
		return nil, errors.New("unable to map writer")
	}

	return f(), nil
}

func init() {
	gob.Register(&segments.Angular{})
	gob.Register(&segments.Version{})
	gob.Register(&segments.Antigravity{})
	gob.Register(&segments.AntigravityData{})
	gob.Register(&segments.Argocd{})
	gob.Register(&segments.Aspire{})
	gob.Register(&segments.Aurelia{})
	gob.Register(&segments.Aws{})
	gob.Register(&segments.Az{})
	gob.Register(&segments.Azd{})
	gob.Register(&segments.AzFunc{})
	gob.Register(&segments.Battery{})
	gob.Register(&segments.Bazel{})
	gob.Register(&segments.Brewfather{})
	gob.Register(&segments.Buf{})
	gob.Register(&segments.Bun{})
	gob.Register(&segments.CarbonIntensity{})
	gob.Register(&segments.Cds{})
	gob.Register(&segments.Copilot{})
	gob.Register(&segments.CopilotCLI{})
	gob.Register(&segments.CopilotCLIData{})
	gob.Register(&segments.Cf{})
	gob.Register(&segments.CfTarget{})
	gob.Register(&segments.Claude{})
	gob.Register(&segments.ClaudeData{})
	gob.Register(&segments.Cmake{})
	gob.Register(&segments.ConfiguredLanguage{})
	gob.Register(&segments.Connection{})
	gob.Register(&segments.Deno{})
	gob.Register(&segments.Docker{})
	gob.Register(&segments.Dotnet{})
	gob.Register(&segments.Dvc{})
	gob.Register(&segments.DvcStatus{})
	gob.Register(&segments.Executiontime{})
	gob.Register(&segments.Status{})
	gob.Register(&segments.Firebase{})
	gob.Register(&segments.Flutter{})
	gob.Register(&segments.Fossil{})
	gob.Register(&segments.FossilStatus{})
	gob.Register(&segments.Gcp{})
	gob.Register(&segments.Git{})
	gob.Register(&segments.GitStatus{})
	gob.Register(&segments.Rebase{})
	gob.Register(&segments.User{})
	gob.Register(&segments.Commit{})
	gob.Register(&segments.GitVersion{})
	gob.Register(&segments.Golang{})
	gob.Register(&segments.Gradle{})
	gob.Register(&segments.Haskell{})
	gob.Register(&segments.Helm{})
	gob.Register(&segments.IPify{})
	gob.Register(&segments.Java{})
	gob.Register(&segments.HTTP{})
	gob.Register(&segments.Jujutsu{})
	gob.Register(&segments.JujutsuStatus{})
	gob.Register(&segments.Kubectl{})
	gob.Register(&segments.LastFM{})
	gob.Register(&segments.Mercurial{})
	gob.Register(&segments.MercurialStatus{})
	gob.Register(&segments.Mojo{})
	gob.Register(&segments.Mvn{})
	gob.Register(&segments.Nba{})
	gob.Register(&segments.Nbgv{})
	gob.Register(&segments.Nightscout{})
	gob.Register(&segments.NixShell{})
	gob.Register(&segments.Node{})
	gob.Register(&segments.Npm{})
	gob.Register(&segments.Nx{})
	gob.Register(&segments.OrthodoxCal{})
	gob.Register(&segments.Os{})
	gob.Register(&segments.Owm{})
	gob.Register(&segments.Path{})
	gob.Register(&segments.Folders{})
	gob.Register(&segments.Plastic{})
	gob.Register(&segments.PlasticStatus{})
	gob.Register(&segments.Pnpm{})
	gob.Register(&segments.Project{})
	gob.Register(&segments.Pulumi{})
	gob.Register(&segments.Python{})
	gob.Register(&segments.Quasar{})
	gob.Register(&segments.Package{})
	gob.Register(&segments.Ramadan{})
	gob.Register(&segments.React{})
	gob.Register(&segments.Root{})
	gob.Register(&segments.Sapling{})
	gob.Register(&segments.SaplingStatus{})
	gob.Register(&segments.Session{})
	gob.Register(&segments.Shell{})
	gob.Register(&segments.Sitecore{})
	gob.Register(&segments.Spotify{})
	gob.Register(&segments.Status{})
	gob.Register(&segments.Strava{})
	gob.Register(&segments.Svelte{})
	gob.Register(&segments.Svn{})
	gob.Register(&segments.SvnStatus{})
	gob.Register(&segments.SystemInfo{})
	gob.Register(&segments.TalosCTL{})
	gob.Register(&segments.Taskwarrior{})
	gob.Register(&segments.Tauri{})
	gob.Register(&segments.Terraform{})
	gob.Register(&segments.Text{})
	gob.Register(&segments.Time{})
	gob.Register(&segments.Todoist{})
	gob.Register(&segments.UI5Tooling{})
	gob.Register(&segments.Umbraco{})
	gob.Register(&segments.Uno{})
	gob.Register(&segments.Unity{})
	gob.Register(&segments.Upgrade{})
	gob.Register(&segments.UpgradeCache{})
	gob.Register(&segments.VIMode{})
	gob.Register(&segments.Wakatime{})
	gob.Register(&segments.WinGet{})
	gob.Register(&segments.WinGetPackage{})
	gob.Register(&segments.WindowsRegistry{})
	gob.Register(&segments.Withings{})
	gob.Register(&segments.XMake{})
	gob.Register(&segments.Yarn{})
	gob.Register(&segments.Ytm{})
	gob.Register(&segments.Segment{})
}

var Segments = map[SegmentType]func() SegmentWriter{
	ANGULAR:         func() SegmentWriter { return &segments.Angular{} },
	ANTIGRAVITY:     func() SegmentWriter { return &segments.Antigravity{} },
	ARGOCD:          func() SegmentWriter { return &segments.Argocd{} },
	ASPIRE:          func() SegmentWriter { return &segments.Aspire{} },
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
	CLAUDE:          func() SegmentWriter { return &segments.Claude{} },
	CLOJURE:         func() SegmentWriter { return segments.NewLanguage(string(CLOJURE)) },
	CMAKE:           func() SegmentWriter { return &segments.Cmake{} },
	CONNECTION:      func() SegmentWriter { return &segments.Connection{} },
	COPILOT:         func() SegmentWriter { return &segments.Copilot{} },
	COPILOTCLI:      func() SegmentWriter { return &segments.CopilotCLI{} },
	CRYSTAL:         func() SegmentWriter { return segments.NewLanguage(string(CRYSTAL)) },
	DART:            func() SegmentWriter { return segments.NewLanguage(string(DART)) },
	DENO:            func() SegmentWriter { return &segments.Deno{} },
	DOCKER:          func() SegmentWriter { return &segments.Docker{} },
	DOTNET:          func() SegmentWriter { return &segments.Dotnet{} },
	DVC:             func() SegmentWriter { return &segments.Dvc{} },
	ELIXIR:          func() SegmentWriter { return segments.NewLanguage(string(ELIXIR)) },
	EXECUTIONTIME:   func() SegmentWriter { return &segments.Executiontime{} },
	EXIT:            func() SegmentWriter { return &segments.Status{} },
	FIREBASE:        func() SegmentWriter { return &segments.Firebase{} },
	FLUTTER:         func() SegmentWriter { return &segments.Flutter{} },
	FORTRAN:         func() SegmentWriter { return segments.NewLanguage(string(FORTRAN)) },
	FOSSIL:          func() SegmentWriter { return &segments.Fossil{} },
	GCP:             func() SegmentWriter { return &segments.Gcp{} },
	GIT:             func() SegmentWriter { return &segments.Git{} },
	GITVERSION:      func() SegmentWriter { return &segments.GitVersion{} },
	GOLANG:          func() SegmentWriter { return &segments.Golang{} },
	GRADLE:          func() SegmentWriter { return &segments.Gradle{} },
	HASKELL:         func() SegmentWriter { return &segments.Haskell{} },
	HELM:            func() SegmentWriter { return &segments.Helm{} },
	IPIFY:           func() SegmentWriter { return &segments.IPify{} },
	JAVA:            func() SegmentWriter { return &segments.Java{} },
	HTTP:            func() SegmentWriter { return &segments.HTTP{} },
	JUJUTSU:         func() SegmentWriter { return &segments.Jujutsu{} },
	JULIA:           func() SegmentWriter { return segments.NewLanguage(string(JULIA)) },
	KOTLIN:          func() SegmentWriter { return segments.NewLanguage(string(KOTLIN)) },
	KUBECTL:         func() SegmentWriter { return &segments.Kubectl{} },
	LANGUAGE:        func() SegmentWriter { return &segments.ConfiguredLanguage{} },
	LASTFM:          func() SegmentWriter { return &segments.LastFM{} },
	LUA:             func() SegmentWriter { return segments.NewLanguage(string(LUA)) },
	MERCURIAL:       func() SegmentWriter { return &segments.Mercurial{} },
	MOJO:            func() SegmentWriter { return &segments.Mojo{} },
	MVN:             func() SegmentWriter { return &segments.Mvn{} },
	NBA:             func() SegmentWriter { return &segments.Nba{} },
	NBGV:            func() SegmentWriter { return &segments.Nbgv{} },
	NIGHTSCOUT:      func() SegmentWriter { return &segments.Nightscout{} },
	NIXSHELL:        func() SegmentWriter { return &segments.NixShell{} },
	NIM:             func() SegmentWriter { return segments.NewLanguage(string(NIM)) },
	NODE:            func() SegmentWriter { return &segments.Node{} },
	NPM:             func() SegmentWriter { return &segments.Npm{} },
	NX:              func() SegmentWriter { return &segments.Nx{} },
	OCAML:           func() SegmentWriter { return segments.NewLanguage(string(OCAML)) },
	ORTHODOXCAL:     func() SegmentWriter { return &segments.OrthodoxCal{} },
	OS:              func() SegmentWriter { return &segments.Os{} },
	OWM:             func() SegmentWriter { return &segments.Owm{} },
	PATH:            func() SegmentWriter { return &segments.Path{} },
	PERL:            func() SegmentWriter { return segments.NewLanguage(string(PERL)) },
	PHP:             func() SegmentWriter { return segments.NewLanguage(string(PHP)) },
	PLASTIC:         func() SegmentWriter { return &segments.Plastic{} },
	PNPM:            func() SegmentWriter { return &segments.Pnpm{} },
	PROJECT:         func() SegmentWriter { return &segments.Project{} },
	PULUMI:          func() SegmentWriter { return &segments.Pulumi{} },
	PYTHON:          func() SegmentWriter { return &segments.Python{} },
	QUASAR:          func() SegmentWriter { return &segments.Quasar{} },
	R:               func() SegmentWriter { return segments.NewLanguage(string(R)) },
	RAMADAN:         func() SegmentWriter { return &segments.Ramadan{} },
	REACT:           func() SegmentWriter { return &segments.React{} },
	ROOT:            func() SegmentWriter { return &segments.Root{} },
	RUBY:            func() SegmentWriter { return segments.NewLanguage(string(RUBY)) },
	RUST:            func() SegmentWriter { return segments.NewLanguage(string(RUST)) },
	SAPLING:         func() SegmentWriter { return &segments.Sapling{} },
	SESSION:         func() SegmentWriter { return &segments.Session{} },
	SHELL:           func() SegmentWriter { return &segments.Shell{} },
	SITECORE:        func() SegmentWriter { return &segments.Sitecore{} },
	SPOTIFY:         func() SegmentWriter { return &segments.Spotify{} },
	STATUS:          func() SegmentWriter { return &segments.Status{} },
	STRAVA:          func() SegmentWriter { return &segments.Strava{} },
	SVELTE:          func() SegmentWriter { return &segments.Svelte{} },
	SVN:             func() SegmentWriter { return &segments.Svn{} },
	SWIFT:           func() SegmentWriter { return segments.NewLanguage(string(SWIFT)) },
	SYSTEMINFO:      func() SegmentWriter { return &segments.SystemInfo{} },
	TALOSCTL:        func() SegmentWriter { return &segments.TalosCTL{} },
	TASKWARRIOR:     func() SegmentWriter { return &segments.Taskwarrior{} },
	TAURI:           func() SegmentWriter { return &segments.Tauri{} },
	TERRAFORM:       func() SegmentWriter { return &segments.Terraform{} },
	TEXT:            func() SegmentWriter { return &segments.Text{} },
	TIME:            func() SegmentWriter { return &segments.Time{} },
	TODOIST:         func() SegmentWriter { return &segments.Todoist{} },
	UI5TOOLING:      func() SegmentWriter { return &segments.UI5Tooling{} },
	UMBRACO:         func() SegmentWriter { return &segments.Umbraco{} },
	UNO:             func() SegmentWriter { return &segments.Uno{} },
	UNITY:           func() SegmentWriter { return &segments.Unity{} },
	UPGRADE:         func() SegmentWriter { return &segments.Upgrade{} },
	V:               func() SegmentWriter { return segments.NewLanguage(string(V)) },
	VALA:            func() SegmentWriter { return segments.NewLanguage(string(VALA)) },
	VIMODE:          func() SegmentWriter { return &segments.VIMode{} },
	WAKATIME:        func() SegmentWriter { return &segments.Wakatime{} },
	WINGET:          func() SegmentWriter { return &segments.WinGet{} },
	WINREG:          func() SegmentWriter { return &segments.WindowsRegistry{} },
	WITHINGS:        func() SegmentWriter { return &segments.Withings{} },
	XMAKE:           func() SegmentWriter { return &segments.XMake{} },
	YARN:            func() SegmentWriter { return &segments.Yarn{} },
	YTM:             func() SegmentWriter { return &segments.Ytm{} },
	ZIG:             func() SegmentWriter { return segments.NewLanguage(string(ZIG)) },
	ZVM:             func() SegmentWriter { return &segments.Zvm{} },
}
