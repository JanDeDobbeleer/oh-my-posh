package main

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

// Segment represent a single segment and it's configuration
type Segment struct {
	Type                SegmentType  `config:"type"`
	Tips                []string     `config:"tips"`
	Style               SegmentStyle `config:"style"`
	PowerlineSymbol     string       `config:"powerline_symbol"`
	InvertPowerline     bool         `config:"invert_powerline"`
	Foreground          string       `config:"foreground"`
	ForegroundTemplates []string     `config:"foreground_templates"`
	Background          string       `config:"background"`
	BackgroundTemplates []string     `config:"background_templates"`
	LeadingDiamond      string       `config:"leading_diamond"`
	TrailingDiamond     string       `config:"trailing_diamond"`
	Properties          properties   `config:"properties"`
	writer              SegmentWriter
	stringValue         string
	active              bool
	env                 Environment
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

type Properties interface {
	getColor(property Property, defaultColor string) string
	getBool(property Property, defaultValue bool) bool
	getString(property Property, defaultValue string) string
	getFloat64(property Property, defaultValue float64) float64
	getInt(property Property, defaultValue int) int
	getKeyValueMap(property Property, defaultValue map[string]string) map[string]string
	getStringArray(property Property, defaultValue []string) []string
	// for legacy purposes
	getOneOfBool(property, legacyProperty Property, defaultValue bool) bool
	getOneOfString(property, legacyProperty Property, defaultValue string) string
	hasOneOf(properties ...Property) bool
	set(property Property, value interface{})
}

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	enabled() bool
	string() string
	init(props Properties, env Environment)
}

// SegmentStyle the syle of segment, for more information, see the constants
type SegmentStyle string

// SegmentType the type of segment, for more information, see the constants
type SegmentType string

const (
	// Session represents the user info segment
	Session SegmentType = "session"
	// Path represents the current path segment
	Path SegmentType = "path"
	// Git represents the git status and information
	Git SegmentType = "git"
	// Plastic represents the plastic scm status and information
	Plastic SegmentType = "plastic"
	// Exit writes the last exit code
	Exit SegmentType = "exit"
	// Python writes the virtual env name
	Python SegmentType = "python"
	// Root writes root symbol
	Root SegmentType = "root"
	// Time writes the current timestamp
	Time SegmentType = "time"
	// Text writes a text
	Text SegmentType = "text"
	// Cmd writes the output of a shell command
	Cmd SegmentType = "command"
	// Battery writes the battery percentage
	Battery SegmentType = "battery"
	// Spotify writes the Spotify status for Mac
	Spotify SegmentType = "spotify"
	// ShellInfo writes which shell we're currently in
	ShellInfo SegmentType = "shell"
	// Node writes which node version is currently active
	Node SegmentType = "node"
	// Os write os specific icon
	Os SegmentType = "os"
	// EnvVar writes the content of an environment variable
	EnvVar SegmentType = "envvar"
	// Az writes the Azure subscription info we're currently in
	Az SegmentType = "az"
	// Kubectl writes the Kubernetes context we're currently in
	Kubectl SegmentType = "kubectl"
	// Dotnet writes which dotnet version is currently active
	Dotnet SegmentType = "dotnet"
	// Terraform writes the terraform workspace we're currently in
	Terraform SegmentType = "terraform"
	// Golang writes which go version is currently active
	Golang SegmentType = "go"
	// Julia writes which julia version is currently active
	Julia SegmentType = "julia"
	// Powerline writes it Powerline style
	Powerline SegmentStyle = "powerline"
	// Plain writes it without ornaments
	Plain SegmentStyle = "plain"
	// Diamond writes the prompt shaped with a leading and trailing symbol
	Diamond SegmentStyle = "diamond"
	// YTM writes YouTube Music information and status
	YTM SegmentType = "ytm"
	// ExecutionTime writes the execution time of the last run command
	ExecutionTime SegmentType = "executiontime"
	// Ruby writes which ruby version is currently active
	Ruby SegmentType = "ruby"
	// Aws writes the active aws context
	Aws SegmentType = "aws"
	// Java writes the active java version
	Java SegmentType = "java"
	// PoshGit writes the posh git prompt
	PoshGit SegmentType = "poshgit"
	// AZFunc writes current AZ func version
	AZFunc SegmentType = "azfunc"
	// Crystal writes the active crystal version
	Crystal SegmentType = "crystal"
	// Dart writes the active dart version
	Dart SegmentType = "dart"
	// Nbgv writes the nbgv version information
	Nbgv SegmentType = "nbgv"
	// Rust writes the cargo version information if cargo.toml is present
	Rust SegmentType = "rust"
	// OWM writes the weather coming from openweatherdata
	OWM SegmentType = "owm"
	// SysInfo writes system information (memory, cpu, load)
	SysInfo SegmentType = "sysinfo"
	// Angular writes which angular cli version us currently active
	Angular SegmentType = "angular"
	// PHP writes which php version is currently active
	PHP SegmentType = "php"
	// Nightscout is an open source diabetes system
	Nightscout SegmentType = "nightscout"
	// Strava is a sports activity tracker
	Strava SegmentType = "strava"
	// Wakatime writes tracked time spend in dev editors
	Wakatime SegmentType = "wakatime"
	// WiFi writes details about the current WiFi connection
	WiFi SegmentType = "wifi"
	// WinReg queries the Windows registry.
	WinReg SegmentType = "winreg"
	// Brewfather segment
	BrewFather SegmentType = "brewfather"
	// Ipify segment
	Ipify SegmentType = "ipify"
)

func (segment *Segment) string() string {
	return segment.writer.string()
}

func (segment *Segment) enabled() bool {
	segment.active = segment.writer.enabled()
	return segment.active
}

func (segment *Segment) getValue(property Property, defaultValue string) string {
	if value, ok := segment.Properties[property]; ok {
		return parseString(value, defaultValue)
	}
	return defaultValue
}

func (segment *Segment) shouldIncludeFolder() bool {
	cwdIncluded := segment.cwdIncluded()
	cwdExcluded := segment.cwdExcluded()
	return (cwdIncluded && !cwdExcluded)
}

func (segment *Segment) cwdIncluded() bool {
	value, ok := segment.Properties[IncludeFolders]
	if !ok {
		// IncludeFolders isn't specified, everything is included
		return true
	}

	list := parseStringArray(value)

	if len(list) == 0 {
		// IncludeFolders is an empty array, everything is included
		return true
	}

	return dirMatchesOneOf(segment.env, segment.env.getcwd(), list)
}

func (segment *Segment) cwdExcluded() bool {
	value, ok := segment.Properties[ExcludeFolders]
	if !ok {
		value = segment.Properties[IgnoreFolders]
	}
	list := parseStringArray(value)
	return dirMatchesOneOf(segment.env, segment.env.getcwd(), list)
}

func (segment *Segment) getColor(templates []string, defaultColor string) string {
	if len(templates) == 0 {
		return defaultColor
	}
	txtTemplate := &textTemplate{
		Context: segment.writer,
		Env:     segment.env,
	}
	for _, template := range templates {
		txtTemplate.Template = template
		value, err := txtTemplate.render()
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
	color := segment.Properties.getColor(ForegroundOverride, segment.Foreground)
	return segment.getColor(segment.ForegroundTemplates, color)
}

func (segment *Segment) background() string {
	color := segment.Properties.getColor(BackgroundOverride, segment.Background)
	return segment.getColor(segment.BackgroundTemplates, color)
}

func (segment *Segment) mapSegmentWithWriter(env Environment) error {
	segment.env = env
	functions := map[SegmentType]SegmentWriter{
		OWM:           &owm{},
		Session:       &session{},
		Path:          &path{},
		Git:           &git{},
		Plastic:       &plastic{},
		Exit:          &exit{},
		Python:        &python{},
		Root:          &root{},
		Text:          &text{},
		Time:          &tempus{},
		Cmd:           &command{},
		Battery:       &batt{},
		Spotify:       &spotify{},
		ShellInfo:     &shell{},
		Node:          &node{},
		Os:            &osInfo{},
		EnvVar:        &envvar{},
		Az:            &az{},
		Kubectl:       &kubectl{},
		Dotnet:        &dotnet{},
		Terraform:     &terraform{},
		Golang:        &golang{},
		Julia:         &julia{},
		YTM:           &ytm{},
		ExecutionTime: &executiontime{},
		Ruby:          &ruby{},
		Aws:           &aws{},
		Java:          &java{},
		PoshGit:       &poshgit{},
		AZFunc:        &azfunc{},
		Crystal:       &crystal{},
		Dart:          &dart{},
		Nbgv:          &nbgv{},
		Rust:          &rust{},
		SysInfo:       &sysinfo{},
		Angular:       &angular{},
		PHP:           &php{},
		Nightscout:    &nightscout{},
		Strava:        &strava{},
		Wakatime:      &wakatime{},
		WiFi:          &wifi{},
		WinReg:        &winreg{},
		BrewFather:    &brewfather{},
		Ipify:         &ipify{},
	}
	if segment.Properties == nil {
		segment.Properties = make(properties)
	}
	if writer, ok := functions[segment.Type]; ok {
		writer.init(segment.Properties, env)
		segment.writer = writer
		return nil
	}
	return errors.New("unable to map writer")
}

func (segment *Segment) setStringValue(env Environment) {
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
