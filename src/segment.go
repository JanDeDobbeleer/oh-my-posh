package main

import (
	"errors"
	"fmt"
)

// Segment represent a single segment and it's configuration
type Segment struct {
	Type            SegmentType              `json:"type"`
	Style           SegmentStyle             `json:"style"`
	PowerlineSymbol string                   `json:"powerline_symbol"`
	InvertPowerline bool                     `json:"invert_powerline"`
	Foreground      string                   `json:"foreground"`
	Background      string                   `json:"background"`
	LeadingDiamond  string                   `json:"leading_diamond"`
	TrailingDiamond string                   `json:"trailing_diamond"`
	Properties      map[Property]interface{} `json:"properties"`
	props           *properties
	writer          SegmentWriter
	stringValue     string
	active          bool
}

// SegmentWriter is the interface used to define what and if to write to the prompt
type SegmentWriter interface {
	enabled() bool
	string() string
	init(props *properties, env environmentInfo)
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

func (segment *Segment) shouldIgnoreFolder(cwd string) bool {
	if value, ok := segment.Properties[IgnoreFolders]; ok {
		list := parseStringArray(value)
		for _, element := range list {
			pattern := fmt.Sprintf("^%s$", element)
			matched := matchString(pattern, cwd)
			return matched
		}
		return false
	}
	return false
}

func (segment *Segment) mapSegmentWithWriter(env environmentInfo) error {
	functions := map[SegmentType]SegmentWriter{
		Session:       &session{},
		Path:          &path{},
		Git:           &git{},
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
	}
	if writer, ok := functions[segment.Type]; ok {
		props := &properties{
			values:     segment.Properties,
			foreground: segment.Foreground,
			background: segment.Background,
		}
		writer.init(props, env)
		segment.writer = writer
		segment.props = props
		return nil
	}
	return errors.New("unable to map writer")
}

func (segment *Segment) setStringValue(env environmentInfo, cwd string) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		// display a message explaining omp failed(with the err)
		message := fmt.Sprintf("oh-my-posh fatal error rendering %s segment:%s", segment.Type, err)
		fmt.Println(message)
		segment.stringValue = "error"
		segment.active = true
	}()
	err := segment.mapSegmentWithWriter(env)
	if err != nil || segment.shouldIgnoreFolder(cwd) {
		return
	}
	if segment.enabled() {
		segment.stringValue = segment.string()
	}
}
