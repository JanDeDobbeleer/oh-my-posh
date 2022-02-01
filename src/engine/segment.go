package engine

import (
	"errors"
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/segments"
	"oh-my-posh/template"
	"runtime/debug"
	"time"
)

// Segment represent a single segment and it's configuration
type Segment struct {
	Type                SegmentType    `json:"type,omitempty"`
	Tips                []string       `json:"tips,omitempty"`
	Style               SegmentStyle   `json:"style,omitempty"`
	PowerlineSymbol     string         `json:"powerline_symbol,omitempty"`
	InvertPowerline     bool           `json:"invert_powerline,omitempty"`
	Foreground          string         `json:"foreground,omitempty"`
	ForegroundTemplates []string       `json:"foreground_templates,omitempty"`
	Background          string         `json:"background,omitempty"`
	BackgroundTemplates []string       `json:"background_templates,omitempty"`
	LeadingDiamond      string         `json:"leading_diamond,omitempty"`
	TrailingDiamond     string         `json:"trailing_diamond,omitempty"`
	Properties          properties.Map `json:"properties,omitempty"`
	writer              SegmentWriter
	stringValue         string
	active              bool
	env                 environment.Environment
}

// SegmentTiming holds the timing context for a segment
type SegmentTiming struct {
	name            string
	nameLength      int
	enabled         bool
	stringValue     string
	enabledDuration time.Duration
	stringDuration  time.Duration
}

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	Enabled() bool
	Template() string
	Init(props properties.Properties, env environment.Environment)
}

// SegmentStyle the syle of segment, for more information, see the constants
type SegmentStyle string

// SegmentType the type of segment, for more information, see the constants
type SegmentType string

const (
	// SESSION represents the user info segment
	SESSION SegmentType = "session"
	// PATH represents the current path segment
	PATH SegmentType = "path"
	// GIT represents the git status and information
	GIT SegmentType = "git"
	// PLASTIC represents the plastic scm status and information
	PLASTIC SegmentType = "plastic"
	// EXIT writes the last exit code
	EXIT SegmentType = "exit"
	// PYTHON writes the virtual env name
	PYTHON SegmentType = "python"
	// ROOT writes root symbol
	ROOT SegmentType = "root"
	// TIME writes the current timestamp
	TIME SegmentType = "time"
	// TEXT writes a text
	TEXT SegmentType = "text"
	// CMD writes the output of a shell command
	CMD SegmentType = "command"
	// BATTERY writes the battery percentage
	BATTERY SegmentType = "battery"
	// SPOTIFY writes the SPOTIFY status for Mac
	SPOTIFY SegmentType = "spotify"
	// SHELL writes which shell we're currently in
	SHELL SegmentType = "shell"
	// NODE writes which node version is currently active
	NODE SegmentType = "node"
	// OS write os specific icon
	OS SegmentType = "os"
	// AZ writes the Azure subscription info we're currently in
	AZ SegmentType = "az"
	// KUBECTL writes the Kubernetes context we're currently in
	KUBECTL SegmentType = "kubectl"
	// DOTNET writes which dotnet version is currently active
	DOTNET SegmentType = "dotnet"
	// TERRAFORM writes the terraform workspace we're currently in
	TERRAFORM SegmentType = "terraform"
	// GOLANG writes which go version is currently active
	GOLANG SegmentType = "go"
	// JULIA writes which julia version is currently active
	JULIA SegmentType = "julia"
	// Powerline writes it Powerline style
	Powerline SegmentStyle = "powerline"
	// Plain writes it without ornaments
	Plain SegmentStyle = "plain"
	// Diamond writes the prompt shaped with a leading and trailing symbol
	Diamond SegmentStyle = "diamond"
	// YTM writes YouTube Music information and status
	YTM SegmentType = "ytm"
	// EXECUTIONTIME writes the execution time of the last run command
	EXECUTIONTIME SegmentType = "executiontime"
	// RUBY writes which ruby version is currently active
	RUBY SegmentType = "ruby"
	// AWS writes the active aws context
	AWS SegmentType = "aws"
	// JAVA writes the active java version
	JAVA SegmentType = "java"
	// POSHGIT writes the posh git prompt
	POSHGIT SegmentType = "poshgit"
	// AZFUNC writes current AZ func version
	AZFUNC SegmentType = "azfunc"
	// CRYSTAL writes the active crystal version
	CRYSTAL SegmentType = "crystal"
	// DART writes the active dart version
	DART SegmentType = "dart"
	// NBGV writes the nbgv version information
	NBGV SegmentType = "nbgv"
	// RUST writes the cargo version information if cargo.toml is present
	RUST SegmentType = "rust"
	// OWM writes the weather coming from openweatherdata
	OWM SegmentType = "owm"
	// SYSTEMINFO writes system information (memory, cpu, load)
	SYSTEMINFO SegmentType = "sysinfo"
	// ANGULAR writes which angular cli version us currently active
	ANGULAR SegmentType = "angular"
	// PHP writes which php version is currently active
	PHP SegmentType = "php"
	// NIGHTSCOUT is an open source diabetes system
	NIGHTSCOUT SegmentType = "nightscout"
	// STRAVA is a sports activity tracker
	STRAVA SegmentType = "strava"
	// WAKATIME writes tracked time spend in dev editors
	WAKATIME SegmentType = "wakatime"
	// WIFI writes details about the current WIFI connection
	WIFI SegmentType = "wifi"
	// WINREG queries the Windows registry.
	WINREG SegmentType = "winreg"
	// Brewfather segment
	BREWFATHER SegmentType = "brewfather"
	// IPIFY segment
	IPIFY SegmentType = "ipify"
)

func (segment *Segment) string() string {
	segmentTemplate := segment.Properties.GetString(properties.SegmentTemplate, segment.writer.Template())
	tmpl := &template.Text{
		Template: segmentTemplate,
		Context:  segment.writer,
		Env:      segment.env,
	}
	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}
	segment.active = len(text) > 0
	return text
}

func (segment *Segment) enabled() bool {
	segment.active = segment.writer.Enabled()
	return segment.active
}

func (segment *Segment) shouldIncludeFolder() bool {
	cwdIncluded := segment.cwdIncluded()
	cwdExcluded := segment.cwdExcluded()
	return (cwdIncluded && !cwdExcluded)
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

	return environment.DirMatchesOneOf(segment.env, segment.env.Pwd(), list)
}

func (segment *Segment) cwdExcluded() bool {
	value, ok := segment.Properties[properties.ExcludeFolders]
	if !ok {
		value = segment.Properties[properties.IgnoreFolders]
	}
	list := properties.ParseStringArray(value)
	return environment.DirMatchesOneOf(segment.env, segment.env.Pwd(), list)
}

func (segment *Segment) getColor(templates []string, defaultColor string) string {
	if len(templates) == 0 {
		return defaultColor
	}
	txtTemplate := &template.Text{
		Context: segment.writer,
		Env:     segment.env,
	}
	for _, tmpl := range templates {
		txtTemplate.Template = tmpl
		value, err := txtTemplate.Render()
		if err != nil || value == "" {
			continue
		}
		return value
	}
	return defaultColor
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
	return segment.getColor(segment.ForegroundTemplates, segment.Foreground)
}

func (segment *Segment) background() string {
	return segment.getColor(segment.BackgroundTemplates, segment.Background)
}

func (segment *Segment) mapSegmentWithWriter(env environment.Environment) error {
	segment.env = env
	functions := map[SegmentType]SegmentWriter{
		OWM:           &segments.Owm{},
		SESSION:       &segments.Session{},
		PATH:          &segments.Path{},
		GIT:           &segments.Git{},
		PLASTIC:       &segments.Plastic{},
		EXIT:          &segments.Exit{},
		PYTHON:        &segments.Python{},
		ROOT:          &segments.Root{},
		TEXT:          &segments.Text{},
		TIME:          &segments.Time{},
		CMD:           &segments.Cmd{},
		BATTERY:       &segments.Battery{},
		SPOTIFY:       &segments.Spotify{},
		SHELL:         &segments.Shell{},
		NODE:          &segments.Node{},
		OS:            &segments.Os{},
		AZ:            &segments.Az{},
		KUBECTL:       &segments.Kubectl{},
		DOTNET:        &segments.Dotnet{},
		TERRAFORM:     &segments.Terraform{},
		GOLANG:        &segments.Golang{},
		JULIA:         &segments.Julia{},
		YTM:           &segments.Ytm{},
		EXECUTIONTIME: &segments.Executiontime{},
		RUBY:          &segments.Ruby{},
		AWS:           &segments.Aws{},
		JAVA:          &segments.Java{},
		POSHGIT:       &segments.PoshGit{},
		AZFUNC:        &segments.AzFunc{},
		CRYSTAL:       &segments.Crystal{},
		DART:          &segments.Dart{},
		NBGV:          &segments.Nbgv{},
		RUST:          &segments.Rust{},
		SYSTEMINFO:    &segments.SystemInfo{},
		ANGULAR:       &segments.Angular{},
		PHP:           &segments.Php{},
		NIGHTSCOUT:    &segments.Nightscout{},
		STRAVA:        &segments.Strava{},
		WAKATIME:      &segments.Wakatime{},
		WIFI:          &segments.Wifi{},
		WINREG:        &segments.WindowsRegistry{},
		BREWFATHER:    &segments.Brewfather{},
		IPIFY:         &segments.IPify{},
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

func (segment *Segment) setStringValue(env environment.Environment) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		// display a message explaining omp failed(with the err)
		message := fmt.Sprintf("\noh-my-posh fatal error rendering %s segment:%s\n\n%s\n", segment.Type, err, debug.Stack())
		fmt.Println(message)
		segment.stringValue = "error"
		segment.active = true
	}()
	err := segment.mapSegmentWithWriter(env)
	if err != nil || !segment.shouldIncludeFolder() {
		return
	}
	if segment.enabled() {
		segment.stringValue = segment.string()
	}
}
