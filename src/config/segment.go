package config

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
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
	Properties             properties.Map `json:"properties,omitempty" toml:"properties,omitempty" yaml:"properties,omitempty"`
	Cache                  *cache.Config  `json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty"`
	Alias                  string         `json:"alias,omitempty" toml:"alias,omitempty" yaml:"alias,omitempty"`
	styleCache             SegmentStyle
	name                   string
	LeadingDiamond         string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty" yaml:"leading_diamond,omitempty"`
	TrailingDiamond        string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty" yaml:"trailing_diamond,omitempty"`
	Template               string         `json:"template,omitempty" toml:"template,omitempty" yaml:"template,omitempty"`
	Foreground             color.Ansi     `json:"foreground,omitempty" toml:"foreground,omitempty" yaml:"foreground,omitempty"`
	TemplatesLogic         template.Logic `json:"templates_logic,omitempty" toml:"templates_logic,omitempty" yaml:"templates_logic,omitempty"`
	PowerlineSymbol        string         `json:"powerline_symbol,omitempty" toml:"powerline_symbol,omitempty" yaml:"powerline_symbol,omitempty"`
	Background             color.Ansi     `json:"background,omitempty" toml:"background,omitempty" yaml:"background,omitempty"`
	Filler                 string         `json:"filler,omitempty" toml:"filler,omitempty" yaml:"filler,omitempty"`
	Type                   SegmentType    `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	Style                  SegmentStyle   `json:"style,omitempty" toml:"style,omitempty" yaml:"style,omitempty"`
	LeadingPowerlineSymbol string         `json:"leading_powerline_symbol,omitempty" toml:"leading_powerline_symbol,omitempty" yaml:"leading_powerline_symbol,omitempty"`
	ForegroundTemplates    template.List  `json:"foreground_templates,omitempty" toml:"foreground_templates,omitempty" yaml:"foreground_templates,omitempty"`
	Tips                   []string       `json:"tips,omitempty" toml:"tips,omitempty" yaml:"tips,omitempty"`
	BackgroundTemplates    template.List  `json:"background_templates,omitempty" toml:"background_templates,omitempty" yaml:"background_templates,omitempty"`
	Templates              template.List  `json:"templates,omitempty" toml:"templates,omitempty" yaml:"templates,omitempty"`
	ExcludeFolders         []string       `json:"exclude_folders,omitempty" toml:"exclude_folders,omitempty" yaml:"exclude_folders,omitempty"`
	IncludeFolders         []string       `json:"include_folders,omitempty" toml:"include_folders,omitempty" yaml:"include_folders,omitempty"`
	Needs                  []string       `json:"-" toml:"-" yaml:"-"`
	MinWidth               int            `json:"min_width,omitempty" toml:"min_width,omitempty" yaml:"min_width,omitempty"`
	MaxWidth               int            `json:"max_width,omitempty" toml:"max_width,omitempty" yaml:"max_width,omitempty"`
	Timeout                time.Duration  `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
	Duration               time.Duration  `json:"-" toml:"-" yaml:"-"`
	NameLength             int            `json:"-" toml:"-" yaml:"-"`
	Interactive            bool           `json:"interactive,omitempty" toml:"interactive,omitempty" yaml:"interactive,omitempty"`
	Enabled                bool           `json:"-" toml:"-" yaml:"-"`
	Newline                bool           `json:"newline,omitempty" toml:"newline,omitempty" yaml:"newline,omitempty"`
	InvertPowerline        bool           `json:"invert_powerline,omitempty" toml:"invert_powerline,omitempty" yaml:"invert_powerline,omitempty"`
	Force                  bool           `json:"force,omitempty" toml:"force,omitempty" yaml:"force,omitempty"`
	restored               bool           `json:"-" toml:"-" yaml:"-"`
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

func (segment *Segment) Execute(env runtime.Environment) {
	// segment timings for debug purposes
	var start time.Time
	if env.Flags().Debug {
		start = time.Now()
		segment.NameLength = len(segment.Name())
		defer func() {
			segment.Duration = time.Since(start)
		}()
	}

	defer segment.evaluateNeeds()

	err := segment.MapSegmentWithWriter(env)
	if err != nil || !segment.shouldIncludeFolder() {
		return
	}

	log.Debugf("segment: %s", segment.Name())

	if segment.isToggled() {
		return
	}

	if segment.restoreCache() {
		return
	}

	if shouldHideForWidth(segment.env, segment.MinWidth, segment.MaxWidth) {
		return
	}

	if segment.Timeout == 0 {
		segment.Enabled = segment.writer.Enabled()
	} else {
		done := make(chan bool)
		go func() {
			segment.Enabled = segment.writer.Enabled()
			done <- true
		}()

		select {
		case <-done:
			// Completed before timeout
		case <-time.After(segment.Timeout * time.Millisecond):
			log.Debugf("timeout after %dms for segment: %s", segment.Timeout, segment.Name())
			return
		}
	}

	if segment.Enabled {
		template.Cache.AddSegmentData(segment.Name(), segment.writer)
	}
}

func (segment *Segment) Render(index int) bool {
	if !segment.Enabled {
		return false
	}

	segment.writer.SetIndex(index)

	text := segment.string()
	segment.Enabled = segment.Force || len(strings.ReplaceAll(text, " ", "")) > 0

	if !segment.Enabled {
		template.Cache.RemoveSegmentData(segment.Name())
		return false
	}

	segment.SetText(text)
	segment.setCache()

	// We do this to make `.Text` available for a cross-segment reference in an extra prompt.
	template.Cache.AddSegmentData(segment.Name(), segment.writer)

	return true
}

func (segment *Segment) Text() string {
	return segment.writer.Text()
}

func (segment *Segment) SetText(text string) {
	segment.writer.SetText(text)
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

func (segment *Segment) hasCache() bool {
	return segment.Cache != nil && !segment.Cache.Duration.IsEmpty()
}

func (segment *Segment) isToggled() bool {
	toggles, OK := segment.env.Session().Get(cache.TOGGLECACHE)
	if !OK || len(toggles) == 0 {
		log.Debug("no toggles found")
		return false
	}

	list := strings.SplitSeq(toggles, ",")
	for toggle := range list {
		if SegmentType(toggle) == segment.Type || toggle == segment.Alias {
			log.Debugf("segment toggled off: %s", segment.Name())
			return true
		}
	}

	return false
}

func (segment *Segment) restoreCache() bool {
	if !segment.hasCache() {
		return false
	}

	cacheKey := segment.cacheKey()
	data, OK := segment.env.Session().Get(cacheKey)
	if !OK {
		log.Debugf("no cache found for segment: %s, key: %s", segment.Name(), cacheKey)
		return false
	}

	err := json.Unmarshal([]byte(data), &segment.writer)
	if err != nil {
		log.Error(err)
	}

	segment.Enabled = true
	template.Cache.AddSegmentData(segment.Name(), segment.writer)

	log.Debug("restored segment from cache: ", segment.Name())

	segment.restored = true

	return true
}

func (segment *Segment) setCache() {
	if segment.restored || !segment.hasCache() {
		return
	}

	data, err := json.Marshal(segment.writer)
	if err != nil {
		log.Error(err)
		return
	}

	segment.env.Session().Set(segment.cacheKey(), string(data), segment.Cache.Duration)
}

func (segment *Segment) cacheKey() string {
	format := "segment_cache_%s"
	switch segment.Cache.Strategy {
	case cache.Session:
		return fmt.Sprintf(format, segment.Name())
	case cache.Folder:
		fallthrough
	default:
		return fmt.Sprintf(format, strings.Join([]string{segment.Name(), segment.folderKey()}, "_"))
	}
}

func (segment *Segment) folderKey() string {
	ctx, ok := segment.writer.(cache.Context)
	if !ok {
		return segment.env.Pwd()
	}

	if key, OK := ctx.CacheKey(); OK {
		return key
	}

	return segment.env.Pwd()
}

func (segment *Segment) string() string {
	result := segment.Templates.Resolve(segment.writer, "", segment.TemplatesLogic)
	if len(result) != 0 {
		return result
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
	if len(segment.IncludeFolders) == 0 {
		return true
	}

	return segment.env.DirMatchesOneOf(segment.env.Pwd(), segment.IncludeFolders)
}

func (segment *Segment) cwdExcluded() bool {
	return segment.env.DirMatchesOneOf(segment.env.Pwd(), segment.ExcludeFolders)
}

func (segment *Segment) evaluateNeeds() {
	value := segment.Template

	if len(segment.ForegroundTemplates) != 0 {
		value += strings.Join(segment.ForegroundTemplates, "")
	}

	if len(segment.BackgroundTemplates) != 0 {
		value += strings.Join(segment.BackgroundTemplates, "")
	}

	if len(segment.Templates) != 0 {
		value += strings.Join(segment.Templates, "")
	}

	if !strings.Contains(value, ".Segments.") {
		return
	}

	matches := regex.FindAllNamedRegexMatch(`\.Segments\.(?P<NAME>[a-zA-Z0-9]+)`, value)
	for _, name := range matches {
		segmentName := name["NAME"]

		if len(name) == 0 || slices.Contains(segment.Needs, segmentName) {
			continue
		}

		segment.Needs = append(segment.Needs, segmentName)
	}
}
