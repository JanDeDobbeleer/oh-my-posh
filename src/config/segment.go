package config

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	c "golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SegmentStyle the style of segment, for more information, see the constants
type SegmentStyle string

func (s *SegmentStyle) resolve(context any) SegmentStyle {
	txtTemplate := &template.Text{
		Context: context,
	}

	txtTemplate.Template = string(*s)
	value, err := txtTemplate.Render()

	// default to Plain
	if err != nil || len(value) == 0 {
		return Plain
	}

	return SegmentStyle(value)
}

type Segment struct {
	writer                 SegmentWriter
	env                    runtime.Environment
	Properties             properties.Map `json:"properties,omitempty" toml:"properties,omitempty"`
	LeadingPowerlineSymbol string         `json:"leading_powerline_symbol,omitempty" toml:"leading_powerline_symbol,omitempty"`
	Background             color.Ansi     `json:"background" toml:"background"`
	Style                  SegmentStyle   `json:"style,omitempty" toml:"style,omitempty"`
	Text                   string         `json:"-" toml:"-"`
	name                   string
	LeadingDiamond         string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty"`
	TrailingDiamond        string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty"`
	Template               string         `json:"template,omitempty" toml:"template,omitempty"`
	Foreground             color.Ansi     `json:"foreground" toml:"foreground"`
	TemplatesLogic         template.Logic `json:"templates_logic,omitempty" toml:"templates_logic,omitempty"`
	PowerlineSymbol        string         `json:"powerline_symbol,omitempty" toml:"powerline_symbol,omitempty"`
	Filler                 string         `json:"filler,omitempty" toml:"filler,omitempty"`
	Alias                  string         `json:"alias,omitempty" toml:"alias,omitempty"`
	Type                   SegmentType    `json:"type,omitempty" toml:"type,omitempty"`
	styleCache             SegmentStyle
	CacheDuration          cache.Duration `json:"cache_duration,omitempty" toml:"cache_duration,omitempty"`
	BackgroundTemplates    template.List  `json:"background_templates,omitempty" toml:"background_templates,omitempty"`
	ForegroundTemplates    template.List  `json:"foreground_templates,omitempty" toml:"foreground_templates,omitempty"`
	Tips                   []string       `json:"tips,omitempty" toml:"tips,omitempty"`
	Templates              template.List  `json:"templates,omitempty" toml:"templates,omitempty"`
	MinWidth               int            `json:"min_width,omitempty" toml:"min_width,omitempty"`
	MaxWidth               int            `json:"max_width,omitempty" toml:"max_width,omitempty"`
	Duration               time.Duration  `json:"-" toml:"-"`
	NameLength             int            `json:"-" toml:"-"`
	Interactive            bool           `json:"interactive,omitempty" toml:"interactive,omitempty"`
	Enabled                bool           `json:"-" toml:"-"`
	Newline                bool           `json:"newline,omitempty" toml:"newline,omitempty"`
	InvertPowerline        bool           `json:"invert_powerline,omitempty" toml:"invert_powerline,omitempty"`
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

func (segment *Segment) SetEnabled(env runtime.Environment) {
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

	// segment timings for debug purposes
	var start time.Time
	if env.Flags().Debug {
		start = time.Now()
		segment.NameLength = len(segment.Name())
		defer func() {
			segment.Duration = time.Since(start)
		}()
	}

	err := segment.MapSegmentWithWriter(env)
	if err != nil || !segment.shouldIncludeFolder() {
		return
	}

	segment.env.DebugF("segment: %s", segment.Name())

	if segment.isToggled() {
		return
	}

	if segment.restoreCache() {
		return
	}

	if shouldHideForWidth(segment.env, segment.MinWidth, segment.MaxWidth) {
		return
	}

	if segment.writer.Enabled() {
		segment.Enabled = true
		env.TemplateCache().AddSegmentData(segment.Name(), segment.writer)
	}
}

func (segment *Segment) isToggled() bool {
	toggles, OK := segment.env.Session().Get(cache.TOGGLECACHE)
	if !OK || len(toggles) == 0 {
		return false
	}

	list := strings.Split(toggles, ",")
	for _, toggle := range list {
		if SegmentType(toggle) == segment.Type || toggle == segment.Alias {
			segment.env.DebugF("segment toggled off: %s", segment.Name())
			return true
		}
	}

	return false
}

func (segment *Segment) restoreCache() bool {
	if segment.CacheDuration.IsEmpty() {
		return false
	}

	text, OK := segment.env.Session().Get(segment.cacheKey())
	if !OK {
		return false
	}

	segment.env.DebugF("restored %s segment from cache", segment.Name())
	segment.Text = text
	segment.Enabled = true

	data, OK := segment.env.Session().Get(segment.writerCacheKey())
	if !OK {
		return true
	}

	err := json.Unmarshal([]byte(data), &segment.writer)
	if err != nil {
		segment.env.Error(err)
	}

	segment.env.TemplateCache().AddSegmentData(segment.Name(), segment.writer)

	return true
}

func (segment *Segment) setCache() {
	if segment.CacheDuration.IsEmpty() {
		return
	}

	segment.env.Session().Set(segment.cacheKey(), segment.Text, segment.CacheDuration)

	data, err := json.Marshal(segment.writer)
	if err != nil {
		segment.env.Error(err)
		return
	}

	segment.env.Session().Set(segment.writerCacheKey(), string(data), segment.CacheDuration)
}

func (segment *Segment) cacheKey() string {
	return fmt.Sprintf("segment_cache_%s_%s", segment.Name(), segment.env.Pwd())
}

func (segment *Segment) writerCacheKey() string {
	return fmt.Sprintf("segment_cache_writer_%s_%s", segment.Name(), segment.env.Pwd())
}

func (segment *Segment) SetText() {
	if !segment.Enabled {
		return
	}

	segment.Text = segment.string()
	segment.Enabled = len(strings.ReplaceAll(segment.Text, " ", "")) > 0

	if !segment.Enabled {
		segment.env.TemplateCache().RemoveSegmentData(segment.Name())
		return
	}

	segment.setCache()
}

func (segment *Segment) string() string {
	if !segment.Templates.Empty() {
		templatesResult := segment.Templates.Resolve(segment.writer, "", segment.TemplatesLogic)
		if len(segment.Template) == 0 {
			return templatesResult
		}
	}

	if len(segment.Template) == 0 {
		segment.Template = segment.writer.Template()
	}

	tmpl := &template.Text{
		Template: segment.Template,
		Context:  segment.writer,
	}

	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (segment *Segment) shouldIncludeFolder() bool {
	if segment.env == nil {
		return true
	}

	cwdIncluded := segment.cwdIncluded()
	cwdExcluded := segment.cwdExcluded()

	return cwdIncluded && !cwdExcluded
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

func (segment *Segment) ResolveForeground() color.Ansi {
	if len(segment.ForegroundTemplates) != 0 {
		match := segment.ForegroundTemplates.FirstMatch(segment.writer, segment.Foreground.String())
		segment.Foreground = color.Ansi(match)
	}

	return segment.Foreground
}

func (segment *Segment) ResolveBackground() color.Ansi {
	if len(segment.BackgroundTemplates) != 0 {
		match := segment.BackgroundTemplates.FirstMatch(segment.writer, segment.Background.String())
		segment.Background = color.Ansi(match)
	}

	return segment.Background
}

func (segment *Segment) ResolveStyle() SegmentStyle {
	if len(segment.styleCache) != 0 {
		return segment.styleCache
	}

	segment.styleCache = segment.Style.resolve(segment.writer)

	return segment.styleCache
}

func (segment *Segment) IsPowerline() bool {
	style := segment.ResolveStyle()
	return style == Powerline || style == Accordion
}

func (segment *Segment) HasEmptyDiamondAtEnd() bool {
	if segment.ResolveStyle() != Diamond {
		return false
	}

	return len(segment.TrailingDiamond) == 0
}
