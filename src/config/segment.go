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
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"go.yaml.in/yaml/v3"
	c "golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SegmentStyle the style of segment, for more information, see the constants
type SegmentStyle string

func (s *SegmentStyle) resolve(context any) SegmentStyle {
	value, err := template.Render(string(*s), context)

	// default to Plain
	if err != nil || value == "" {
		return Plain
	}

	return SegmentStyle(value)
}

type Segment struct {
	writer  SegmentWriter
	env     runtime.Environment
	Options options.Map `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty"`
	// Properties is deprecated, use Options instead. This field exists for TOML backward compatibility
	// since go-toml/v2 doesn't support custom unmarshalers. It will be migrated to Options after loading.
	Properties             options.Map `json:"-" toml:"properties,omitempty" yaml:"-"`
	Cache                  *Cache      `json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty"`
	Alias                  string      `json:"alias,omitempty" toml:"alias,omitempty" yaml:"alias,omitempty"`
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
	Timeout                int            `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
	MaxWidth               int            `json:"max_width,omitempty" toml:"max_width,omitempty" yaml:"max_width,omitempty"`
	MinWidth               int            `json:"min_width,omitempty" toml:"min_width,omitempty" yaml:"min_width,omitempty"`
	Duration               time.Duration  `json:"-" toml:"-" yaml:"-"`
	NameLength             int            `json:"-" toml:"-" yaml:"-"`
	Index                  int            `json:"index,omitempty" toml:"index,omitempty" yaml:"index,omitempty"`
	Interactive            bool           `json:"interactive,omitempty" toml:"interactive,omitempty" yaml:"interactive,omitempty"`
	Enabled                bool           `json:"-" toml:"-" yaml:"-"`
	Newline                bool           `json:"newline,omitempty" toml:"newline,omitempty" yaml:"newline,omitempty"`
	InvertPowerline        bool           `json:"invert_powerline,omitempty" toml:"invert_powerline,omitempty" yaml:"invert_powerline,omitempty"`
	Force                  bool           `json:"force,omitempty" toml:"force,omitempty" yaml:"force,omitempty"`
	Restored               bool           `json:"-" toml:"-" yaml:"-"`
	Toggled                bool           `json:"toggled,omitempty" toml:"toggled,omitempty" yaml:"toggled,omitempty"`
	Pending                bool           `json:"-" toml:"-" yaml:"-"`
}

// segmentAlias is used to avoid recursion during unmarshaling
type segmentAlias Segment

// segmentAux is a helper struct that captures the legacy 'properties' field
type segmentAux struct {
	Properties options.Map `json:"properties,omitempty" yaml:"properties,omitempty" toml:"properties,omitempty"`
	*segmentAlias
}

func (segment *Segment) UnmarshalJSON(data []byte) error {
	aux := &segmentAux{
		segmentAlias: (*segmentAlias)(segment),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Migrate 'properties' to 'options' if present
	if len(aux.Properties) > 0 && len(segment.Options) == 0 {
		segment.Options = aux.Properties
	}

	return nil
}

func (segment *Segment) UnmarshalYAML(node *yaml.Node) error {
	// Decode into a map to handle field renaming
	var raw map[string]any
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// If 'properties' exists and 'options' doesn't, rename it
	if props, hasProps := raw["properties"]; hasProps {
		if _, hasOptions := raw["options"]; !hasOptions {
			raw["options"] = props
			delete(raw, "properties")
		}
	}

	// Re-encode and decode into the struct
	modifiedNode := &yaml.Node{}
	if err := modifiedNode.Encode(raw); err != nil {
		return err
	}

	return modifiedNode.Decode((*segmentAlias)(segment))
}

// MigratePropertiesToOptions migrates the deprecated Properties field to Options.
// This is needed for TOML configs since go-toml/v2 doesn't support custom unmarshalers.
func (segment *Segment) MigratePropertiesToOptions() {
	if len(segment.Properties) > 0 && len(segment.Options) == 0 {
		segment.Options = segment.Properties
		segment.Properties = nil
	}
}

func (segment *Segment) Name() string {
	if len(segment.name) != 0 {
		return segment.name
	}

	name := segment.Alias
	if name == "" {
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

	// In streaming mode, use cache for initial display but continue executing for fresh data
	cacheRestored := segment.restoreCache()
	if cacheRestored && !env.Flags().Streaming {
		return
	}

	if shouldHideForWidth(segment.env, segment.MinWidth, segment.MaxWidth) {
		return
	}

	defer func() {
		if segment.Enabled {
			template.Cache.AddSegmentData(segment.Name(), segment.writer)
		}
	}()

	// Create Job for this goroutine so child processes can be tracked and killed on timeout
	if err := runjobs.CreateJobForGoroutine(segment.Name()); err != nil {
		log.Errorf("failed to create job for goroutine (segment: %s): %v", segment.Name(), err)
	}

	// In streaming mode, don't write to Enabled if segment is pending to avoid data race
	// Render() will be the sole controller of Enabled state for pending segments
	if env.Flags().Streaming && segment.Pending {
		return
	}

	segment.Enabled = segment.writer.Enabled()
}

func (segment *Segment) Render(index int, force bool) bool {
	// Allow pending segments to render (they'll show "..." text)
	if !segment.Pending && !segment.Enabled && !force {
		return false
	}

	if force {
		segment.Force = true
	}

	segment.writer.SetIndex(index)

	text := segment.string()

	// Only update Enabled if segment is NOT pending (avoid race with Execute goroutine)
	if !segment.Pending {
		segment.Enabled = segment.Force || len(strings.ReplaceAll(text, " ", "")) > 0

		if !segment.Enabled {
			template.Cache.RemoveSegmentData(segment.Name())
			return false
		}
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

	return segment.TrailingDiamond == ""
}

func (segment *Segment) hasCache() bool {
	return segment.Cache != nil && !segment.Cache.Duration.IsEmpty()
}

func (segment *Segment) isToggled() bool {
	togglesMap, OK := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
	if !OK || len(togglesMap) == 0 {
		log.Debug("no toggles found")
		return false
	}

	segmentName := segment.Alias
	if segmentName == "" {
		segmentName = string(segment.Type)
	}

	if togglesMap[segmentName] {
		log.Debugf("segment toggled off: %s", segment.Name())
		return true
	}

	return false
}

func (segment *Segment) restoreCache() bool {
	if !segment.hasCache() {
		return false
	}

	key, store := segment.cacheKeyAndStore()
	data, OK := cache.Get[string](store, key)
	if !OK {
		log.Debugf("no cache found for segment: %s, key: %s", segment.Name(), key)
		return false
	}

	err := json.Unmarshal([]byte(data), &segment.writer)
	if err != nil {
		log.Error(err)
	}

	segment.Enabled = true
	template.Cache.AddSegmentData(segment.Name(), segment.writer)

	log.Debug("restored segment from cache: ", segment.Name())

	segment.Restored = true

	return true
}

func (segment *Segment) setCache() {
	if segment.Restored || !segment.hasCache() {
		return
	}

	// Never cache pending state to avoid polluting cache with incomplete data
	if segment.Pending {
		return
	}

	data, err := json.Marshal(segment.writer)
	if err != nil {
		log.Error(err)
		return
	}

	// TODO: check if we can make segmentwriter a generic Type indicator
	// that way we can actually get the value straight from cache.Get
	// and marchalling is obsolete
	key, store := segment.cacheKeyAndStore()
	cache.Set(store, key, string(data), segment.Cache.Duration)
}

func (segment *Segment) cacheKeyAndStore() (string, cache.Store) {
	format := "segment_cache_%s"
	switch segment.Cache.Strategy {
	case Session:
		return fmt.Sprintf(format, segment.Name()), cache.Session
	case Device:
		return fmt.Sprintf(format, segment.Name()), cache.Device
	case Folder:
		fallthrough
	default:
		return fmt.Sprintf(format, strings.Join([]string{segment.Name(), segment.folderKey()}, "_")), cache.Device
	}
}

func (segment *Segment) folderKey() string {
	key, ok := segment.writer.CacheKey()
	if !ok {
		return segment.env.Pwd()
	}

	return key
}

func (segment *Segment) string() string {
	// Use simple pending text if segment is still pending
	if segment.Pending {
		return "..."
	}

	result := segment.Templates.Resolve(segment.writer, "", segment.TemplatesLogic)
	if len(result) != 0 {
		return result
	}

	if segment.Template == "" {
		segment.Template = segment.writer.Template()
	}

	text, err := template.Render(segment.Template, segment.writer)
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

func (segment *Segment) key() any {
	if segment.Index > 0 {
		return segment.Index - 1
	}

	return segment.Name()
}
