package engine

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"oh-my-posh/shell"
	"oh-my-posh/template"
)

// Segment represent a single segment and it's configuration
type Segment struct {
	Type                SegmentType    `json:"type,omitempty"`
	Tips                []string       `json:"tips,omitempty"`
	Style               SegmentStyle   `json:"style,omitempty"`
	PowerlineSymbol     string         `json:"powerline_symbol,omitempty"`
	InvertPowerline     bool           `json:"invert_powerline,omitempty"`
	Foreground          string         `json:"foreground,omitempty"`
	ForegroundTemplates template.List  `json:"foreground_templates,omitempty"`
	Background          string         `json:"background,omitempty"`
	BackgroundTemplates template.List  `json:"background_templates,omitempty"`
	LeadingDiamond      string         `json:"leading_diamond,omitempty"`
	TrailingDiamond     string         `json:"trailing_diamond,omitempty"`
	Template            string         `json:"template,omitempty"`
	Templates           template.List  `json:"templates,omitempty"`
	TemplatesLogic      template.Logic `json:"templates_logic,omitempty"`
	Properties          properties.Map `json:"properties,omitempty"`
	Interactive         bool           `json:"interactive,omitempty"`
	Alias               string         `json:"alias,omitempty"`

	writer          SegmentWriter
	Enabled         bool `json:"-"`
	text            string
	env             environment.Environment
	backgroundCache string
	foregroundCache string
}

// SegmentTiming holds the timing context for a segment
type SegmentTiming struct {
	name       string
	nameLength int
	active     bool
	text       string
	duration   time.Duration
}

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	Enabled() bool
	Template() string
	Init(props properties.Properties, env environment.Environment)
}

// SegmentStyle the style of segment, for more information, see the constants
type SegmentStyle string

// SegmentType the type of segment, for more information, see the constants
type SegmentType string

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
	// AWS writes the active aws context
	AWS SegmentType = "aws"
	// AZ writes the Azure subscription info we're currently in
	AZ SegmentType = "az"
	// AZFUNC writes current AZ func version
	AZFUNC SegmentType = "azfunc"
	// BATTERY writes the battery percentage
	BATTERY SegmentType = "battery"
	// Brewfather segment
	BREWFATHER SegmentType = "brewfather"
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
	// DOTNET writes which dotnet version is currently active
	DOTNET SegmentType = "dotnet"
	// EXECUTIONTIME writes the execution time of the last run command
	EXECUTIONTIME SegmentType = "executiontime"
	// EXIT writes the last exit code
	EXIT SegmentType = "exit"
	// FLUTTER writes the flutter version
	FLUTTER SegmentType = "flutter"
	// FOSSIL writes the fossil status
	FOSSIL SegmentType = "fossil"
	// GCP writes the active GCP context
	GCP SegmentType = "gcp"
	// GIT represents the git status and information
	GIT SegmentType = "git"
	// GOLANG writes which go version is currently active
	GOLANG SegmentType = "go"
	// HASKELL segment
	HASKELL SegmentType = "haskell"
	// IPIFY segment
	IPIFY SegmentType = "ipify"
	// ITERM inserts the Shell Integration prompt mark on iTerm zsh/bash/fish
	ITERM SegmentType = "iterm"
	// JAVA writes the active java version
	JAVA SegmentType = "java"
	// JULIA writes which julia version is currently active
	JULIA SegmentType = "julia"
	// KOTLIN writes the active kotlin version
	KOTLIN SegmentType = "kotlin"
	// KUBECTL writes the Kubernetes context we're currently in
	KUBECTL SegmentType = "kubectl"
	// LUA writes the active lua version
	LUA SegmentType = "lua"
	// NBGV writes the nbgv version information
	NBGV SegmentType = "nbgv"
	// NIGHTSCOUT is an open source diabetes system
	NIGHTSCOUT SegmentType = "nightscout"
	// NODE writes which node version is currently active
	NODE SegmentType = "node"
	// npm version
	NPM SegmentType = "npm"
	// NX writes which Nx version us currently active
	NX SegmentType = "nx"
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
	// Project version
	PROJECT SegmentType = "project"
	// PYTHON writes the virtual env name
	PYTHON SegmentType = "python"
	// R version
	R SegmentType = "r"
	// ROOT writes root symbol
	ROOT SegmentType = "root"
	// RUBY writes which ruby version is currently active
	RUBY SegmentType = "ruby"
	// RUST writes the cargo version information if cargo.toml is present
	RUST SegmentType = "rust"
	// SESSION represents the user info segment
	SESSION SegmentType = "session"
	// SHELL writes which shell we're currently in
	SHELL SegmentType = "shell"
	// SPOTIFY writes the SPOTIFY status for Mac
	SPOTIFY SegmentType = "spotify"
	// STRAVA is a sports activity tracker
	STRAVA SegmentType = "strava"
	// Subversion segment
	SVN SegmentType = "svn"
	// SWIFT writes the active swift version
	SWIFT SegmentType = "swift"
	// SYSTEMINFO writes system information (memory, cpu, load)
	SYSTEMINFO SegmentType = "sysinfo"
	// TERRAFORM writes the terraform workspace we're currently in
	TERRAFORM SegmentType = "terraform"
	// TEXT writes a text
	TEXT SegmentType = "text"
	// TIME writes the current timestamp
	TIME SegmentType = "time"
	// UI5 Tooling segment
	UI5TOOLING SegmentType = "ui5tooling"
	// WAKATIME writes tracked time spend in dev editors
	WAKATIME SegmentType = "wakatime"
	// WINREG queries the Windows registry.
	WINREG SegmentType = "winreg"
	// WITHINGS queries the Withings API.
	WITHINGS SegmentType = "withings"
	// XMAKE write the xmake version if xmake.lua is present
	XMAKE SegmentType = "xmake"
	// YTM writes YouTube Music information and status
	YTM SegmentType = "ytm"
)

func (segment *Segment) shouldIncludeFolder() bool {
	if segment.env == nil {
		return true
	}
	cwdIncluded := segment.cwdIncluded()
	cwdExcluded := segment.cwdExcluded()
	return cwdIncluded && !cwdExcluded
}

func (segment *Segment) isPowerline() bool {
	return segment.Style == Powerline || segment.Style == Accordion
}

func (segment *Segment) cwdIncluded() bool {
	value, ok := segment.Properties[properties.IncludeFolders]
	if !ok {
		// IncludeFolders isn't specified, everything is included
		return true
	}

	list := properties.ParseStringArray(value)

	if len(list) == 0 {
		// IncludeFolders is an empty array, everything is included
		return true
	}

	return segment.env.DirMatchesOneOf(segment.env.Pwd(), list)
}

func (segment *Segment) cwdExcluded() bool {
	value, ok := segment.Properties[properties.ExcludeFolders]
	if !ok {
		value = segment.Properties[properties.IgnoreFolders]
	}
	list := properties.ParseStringArray(value)
	return segment.env.DirMatchesOneOf(segment.env.Pwd(), list)
}

func (segment *Segment) shouldInvokeWithTip(tip string) bool {
	for _, t := range segment.Tips {
		if t == tip {
			return true
		}
	}
	return false
}

func (segment *Segment) foreground() string {
	if len(segment.foregroundCache) == 0 {
		segment.foregroundCache = segment.ForegroundTemplates.FirstMatch(segment.writer, segment.env, segment.Foreground)
	}
	return segment.foregroundCache
}

func (segment *Segment) background() string {
	if len(segment.backgroundCache) == 0 {
		segment.backgroundCache = segment.BackgroundTemplates.FirstMatch(segment.writer, segment.env, segment.Background)
	}
	return segment.backgroundCache
}

func (segment *Segment) mapSegmentWithWriter(env environment.Environment) error {
	segment.env = env
	functions := map[SegmentType]SegmentWriter{
		ANGULAR:       &segments.Angular{},
		AWS:           &segments.Aws{},
		AZ:            &segments.Az{},
		AZFUNC:        &segments.AzFunc{},
		BATTERY:       &segments.Battery{},
		BREWFATHER:    &segments.Brewfather{},
		CDS:           &segments.Cds{},
		CF:            &segments.Cf{},
		CFTARGET:      &segments.CfTarget{},
		CMD:           &segments.Cmd{},
		CONNECTION:    &segments.Connection{},
		CRYSTAL:       &segments.Crystal{},
		CMAKE:         &segments.Cmake{},
		DART:          &segments.Dart{},
		DENO:          &segments.Deno{},
		DOTNET:        &segments.Dotnet{},
		EXECUTIONTIME: &segments.Executiontime{},
		EXIT:          &segments.Exit{},
		FLUTTER:       &segments.Flutter{},
		FOSSIL:        &segments.Fossil{},
		GCP:           &segments.Gcp{},
		GIT:           &segments.Git{},
		GOLANG:        &segments.Golang{},
		HASKELL:       &segments.Haskell{},
		IPIFY:         &segments.IPify{},
		ITERM:         &segments.ITerm{},
		JAVA:          &segments.Java{},
		JULIA:         &segments.Julia{},
		KOTLIN:        &segments.Kotlin{},
		KUBECTL:       &segments.Kubectl{},
		LUA:           &segments.Lua{},
		NBGV:          &segments.Nbgv{},
		NIGHTSCOUT:    &segments.Nightscout{},
		NODE:          &segments.Node{},
		NPM:           &segments.Npm{},
		NX:            &segments.Nx{},
		OS:            &segments.Os{},
		OWM:           &segments.Owm{},
		PATH:          &segments.Path{},
		PERL:          &segments.Perl{},
		PHP:           &segments.Php{},
		PLASTIC:       &segments.Plastic{},
		PROJECT:       &segments.Project{},
		PYTHON:        &segments.Python{},
		R:             &segments.R{},
		ROOT:          &segments.Root{},
		RUBY:          &segments.Ruby{},
		RUST:          &segments.Rust{},
		SESSION:       &segments.Session{},
		SHELL:         &segments.Shell{},
		SPOTIFY:       &segments.Spotify{},
		STRAVA:        &segments.Strava{},
		SVN:           &segments.Svn{},
		SWIFT:         &segments.Swift{},
		SYSTEMINFO:    &segments.SystemInfo{},
		TERRAFORM:     &segments.Terraform{},
		TEXT:          &segments.Text{},
		TIME:          &segments.Time{},
		UI5TOOLING:    &segments.UI5Tooling{},
		WAKATIME:      &segments.Wakatime{},
		WINREG:        &segments.WindowsRegistry{},
		WITHINGS:      &segments.Withings{},
		XMAKE:         &segments.XMake{},
		YTM:           &segments.Ytm{},
	}
	if segment.Properties == nil {
		segment.Properties = make(properties.Map)
	}
	if writer, ok := functions[segment.Type]; ok {
		writer.Init(segment.Properties, env)
		segment.writer = writer
		return nil
	}
	return errors.New("unable to map writer")
}

func (segment *Segment) string() string {
	var templatesResult string
	if !segment.Templates.Empty() {
		templatesResult = segment.Templates.Resolve(segment.writer, segment.env, "", segment.TemplatesLogic)
		if len(segment.Template) == 0 {
			return templatesResult
		}
	}
	if len(segment.Template) == 0 {
		segment.Template = segment.writer.Template()
	}
	tmpl := &template.Text{
		Template:        segment.Template,
		Context:         segment.writer,
		Env:             segment.env,
		TemplatesResult: templatesResult,
	}
	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (segment *Segment) SetEnabled(env environment.Environment) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		// display a message explaining omp failed(with the err)
		message := fmt.Sprintf("\noh-my-posh fatal error rendering %s segment:%s\n\n%s\n", segment.Type, err, debug.Stack())
		fmt.Println(message)
		segment.Enabled = true
	}()
	err := segment.mapSegmentWithWriter(env)
	if err != nil || !segment.shouldIncludeFolder() {
		return
	}
	if segment.writer.Enabled() {
		segment.Enabled = true
		name := segment.Alias
		if len(name) == 0 {
			name = string(segment.Type)
		}
		env.TemplateCache().AddSegmentData(name, segment.writer)
	}
}

func (segment *Segment) SetText() {
	if !segment.Enabled {
		return
	}
	segment.text = segment.string()
	segment.Enabled = len(strings.ReplaceAll(segment.text, " ", "")) > 0
	if segment.Interactive {
		return
	}
	// we have to do this to prevent bash/zsh from misidentifying escape sequences
	switch segment.env.Shell() {
	case shell.BASH:
		segment.text = strings.NewReplacer("`", "\\`", `\`, `\\`).Replace(segment.text)
	case shell.ZSH:
		segment.text = strings.NewReplacer("`", "\\`", `%`, `%%`).Replace(segment.text)
	}
}
