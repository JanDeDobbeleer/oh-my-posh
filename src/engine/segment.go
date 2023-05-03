package engine

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	c "golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	MaxWidth            int            `json:"max_width,omitempty"`
	MinWidth            int            `json:"min_width,omitempty"`
	Filler              string         `json:"filler,omitempty"`

	Enabled bool `json:"-"`

	colors     *ansi.Colors
	env        platform.Environment
	writer     SegmentWriter
	text       string
	styleCache SegmentStyle
	name       string
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
	Init(props properties.Properties, env platform.Environment)
}

// SegmentStyle the style of segment, for more information, see the constants
type SegmentStyle string

func (s *SegmentStyle) Resolve(env platform.Environment, context interface{}) SegmentStyle {
	txtTemplate := &template.Text{
		Context: context,
		Env:     env,
	}
	txtTemplate.Template = string(*s)
	value, err := txtTemplate.Render()
	// default to Plain
	if err != nil || len(value) == 0 {
		return Plain
	}
	return SegmentStyle(value)
}

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
	// ARGOCD writes the current argocd context
	ARGOCD SegmentType = "argocd"
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
	// Buf segment writes the active buf version
	BUF SegmentType = "buf"
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
	// FLUTTER writes the flutter version
	FLUTTER SegmentType = "flutter"
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
	// MERCURIAL writes the Mercurial source control information
	MERCURIAL SegmentType = "mercurial"
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
	// UNITY writes which Unity version is currently active
	UNITY SegmentType = "unity"
	// UPGRADE lets you know if you can upgrade Oh My Posh
	UPGRADE SegmentType = "upgrade"
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
	// YTM writes YouTube Music information and status
	YTM SegmentType = "ytm"
)

// Segments contains all available prompt segment writers.
// Consumers of the library can also add their own segment writer.
var Segments = map[SegmentType]func() SegmentWriter{
	ANGULAR:       func() SegmentWriter { return &segments.Angular{} },
	ARGOCD:        func() SegmentWriter { return &segments.Argocd{} },
	AWS:           func() SegmentWriter { return &segments.Aws{} },
	AZ:            func() SegmentWriter { return &segments.Az{} },
	AZFUNC:        func() SegmentWriter { return &segments.AzFunc{} },
	BATTERY:       func() SegmentWriter { return &segments.Battery{} },
	BREWFATHER:    func() SegmentWriter { return &segments.Brewfather{} },
	BUF:           func() SegmentWriter { return &segments.Buf{} },
	CDS:           func() SegmentWriter { return &segments.Cds{} },
	CF:            func() SegmentWriter { return &segments.Cf{} },
	CFTARGET:      func() SegmentWriter { return &segments.CfTarget{} },
	CMD:           func() SegmentWriter { return &segments.Cmd{} },
	CONNECTION:    func() SegmentWriter { return &segments.Connection{} },
	CRYSTAL:       func() SegmentWriter { return &segments.Crystal{} },
	CMAKE:         func() SegmentWriter { return &segments.Cmake{} },
	DART:          func() SegmentWriter { return &segments.Dart{} },
	DENO:          func() SegmentWriter { return &segments.Deno{} },
	DOCKER:        func() SegmentWriter { return &segments.Docker{} },
	DOTNET:        func() SegmentWriter { return &segments.Dotnet{} },
	EXECUTIONTIME: func() SegmentWriter { return &segments.Executiontime{} },
	ELIXIR:        func() SegmentWriter { return &segments.Elixir{} },
	EXIT:          func() SegmentWriter { return &segments.Exit{} },
	FLUTTER:       func() SegmentWriter { return &segments.Flutter{} },
	FOSSIL:        func() SegmentWriter { return &segments.Fossil{} },
	GCP:           func() SegmentWriter { return &segments.Gcp{} },
	GIT:           func() SegmentWriter { return &segments.Git{} },
	GITVERSION:    func() SegmentWriter { return &segments.GitVersion{} },
	GOLANG:        func() SegmentWriter { return &segments.Golang{} },
	HASKELL:       func() SegmentWriter { return &segments.Haskell{} },
	IPIFY:         func() SegmentWriter { return &segments.IPify{} },
	ITERM:         func() SegmentWriter { return &segments.ITerm{} },
	JAVA:          func() SegmentWriter { return &segments.Java{} },
	JULIA:         func() SegmentWriter { return &segments.Julia{} },
	KOTLIN:        func() SegmentWriter { return &segments.Kotlin{} },
	KUBECTL:       func() SegmentWriter { return &segments.Kubectl{} },
	LUA:           func() SegmentWriter { return &segments.Lua{} },
	MERCURIAL:     func() SegmentWriter { return &segments.Mercurial{} },
	NBGV:          func() SegmentWriter { return &segments.Nbgv{} },
	NIGHTSCOUT:    func() SegmentWriter { return &segments.Nightscout{} },
	NODE:          func() SegmentWriter { return &segments.Node{} },
	NPM:           func() SegmentWriter { return &segments.Npm{} },
	NX:            func() SegmentWriter { return &segments.Nx{} },
	OS:            func() SegmentWriter { return &segments.Os{} },
	OWM:           func() SegmentWriter { return &segments.Owm{} },
	PATH:          func() SegmentWriter { return &segments.Path{} },
	PERL:          func() SegmentWriter { return &segments.Perl{} },
	PHP:           func() SegmentWriter { return &segments.Php{} },
	PLASTIC:       func() SegmentWriter { return &segments.Plastic{} },
	PROJECT:       func() SegmentWriter { return &segments.Project{} },
	PYTHON:        func() SegmentWriter { return &segments.Python{} },
	R:             func() SegmentWriter { return &segments.R{} },
	ROOT:          func() SegmentWriter { return &segments.Root{} },
	RUBY:          func() SegmentWriter { return &segments.Ruby{} },
	RUST:          func() SegmentWriter { return &segments.Rust{} },
	SAPLING:       func() SegmentWriter { return &segments.Sapling{} },
	SESSION:       func() SegmentWriter { return &segments.Session{} },
	SHELL:         func() SegmentWriter { return &segments.Shell{} },
	SITECORE:      func() SegmentWriter { return &segments.Sitecore{} },
	SPOTIFY:       func() SegmentWriter { return &segments.Spotify{} },
	STRAVA:        func() SegmentWriter { return &segments.Strava{} },
	SVN:           func() SegmentWriter { return &segments.Svn{} },
	SWIFT:         func() SegmentWriter { return &segments.Swift{} },
	SYSTEMINFO:    func() SegmentWriter { return &segments.SystemInfo{} },
	TERRAFORM:     func() SegmentWriter { return &segments.Terraform{} },
	TEXT:          func() SegmentWriter { return &segments.Text{} },
	TIME:          func() SegmentWriter { return &segments.Time{} },
	UI5TOOLING:    func() SegmentWriter { return &segments.UI5Tooling{} },
	UNITY:         func() SegmentWriter { return &segments.Unity{} },
	UPGRADE:       func() SegmentWriter { return &segments.Upgrade{} },
	VALA:          func() SegmentWriter { return &segments.Vala{} },
	WAKATIME:      func() SegmentWriter { return &segments.Wakatime{} },
	WINREG:        func() SegmentWriter { return &segments.WindowsRegistry{} },
	WITHINGS:      func() SegmentWriter { return &segments.Withings{} },
	XMAKE:         func() SegmentWriter { return &segments.XMake{} },
	YTM:           func() SegmentWriter { return &segments.Ytm{} },
}

func (segment *Segment) style() SegmentStyle {
	if len(segment.styleCache) != 0 {
		return segment.styleCache
	}
	segment.styleCache = segment.Style.Resolve(segment.env, segment.writer)
	return segment.styleCache
}

func (segment *Segment) shouldIncludeFolder() bool {
	if segment.env == nil {
		return true
	}
	cwdIncluded := segment.cwdIncluded()
	cwdExcluded := segment.cwdExcluded()
	return cwdIncluded && !cwdExcluded
}

func (segment *Segment) isPowerline() bool {
	style := segment.style()
	return style == Powerline || style == Accordion
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
	if segment.colors == nil {
		segment.colors = &ansi.Colors{}
	}
	if len(segment.colors.Foreground) == 0 {
		segment.colors.Foreground = segment.ForegroundTemplates.FirstMatch(segment.writer, segment.env, segment.Foreground)
	}
	return segment.colors.Foreground
}

func (segment *Segment) background() string {
	if segment.colors == nil {
		segment.colors = &ansi.Colors{}
	}
	if len(segment.colors.Background) == 0 {
		segment.colors.Background = segment.BackgroundTemplates.FirstMatch(segment.writer, segment.env, segment.Background)
	}
	return segment.colors.Background
}

func (segment *Segment) mapSegmentWithWriter(env platform.Environment) error {
	segment.env = env

	if segment.Properties == nil {
		segment.Properties = make(properties.Map)
	}

	if f, ok := Segments[segment.Type]; ok {
		writer := f()
		wrapper := &properties.Wrapper{
			Properties: segment.Properties,
			Env:        env,
		}
		writer.Init(wrapper, env)
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

func (segment *Segment) Name() string {
	if len(segment.name) != 0 {
		return segment.name
	}
	name := segment.Alias
	if len(name) == 0 {
		name = c.Title(language.English).String(string(segment.Type))
	}
	segment.name = name
	return name
}

func (segment *Segment) SetEnabled(env platform.Environment) {
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
	// validate toggles
	if toggles, OK := segment.env.Cache().Get(platform.TOGGLECACHE); OK && len(toggles) > 0 {
		list := strings.Split(toggles, ",")
		for _, toggle := range list {
			if SegmentType(toggle) == segment.Type || toggle == segment.Alias {
				return
			}
		}
	}
	if shouldHideForWidth(segment.env, segment.MinWidth, segment.MaxWidth) {
		return
	}
	if segment.writer.Enabled() {
		segment.Enabled = true
		env.TemplateCache().AddSegmentData(segment.Name(), segment.writer)
	}
}

func (segment *Segment) SetText() {
	if !segment.Enabled {
		return
	}
	segment.text = segment.string()
	segment.Enabled = len(strings.ReplaceAll(segment.text, " ", "")) > 0
	if !segment.Enabled {
		segment.env.TemplateCache().RemoveSegmentData(segment.Name())
	}

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
