package config

import (
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

// FieldSetConsumer is implemented by segment writers that adapt what they
// fetch to the fields the config's templates actually reference.
// SetReferencedFields runs right after Init when the config was analyzed
// (see Config.ResolveFieldSets): fields is the sorted union of top-level
// context fields any template in the config can read from this segment, and
// analyzable reports whether that set can be trusted - false means at least
// one template escaped the analysis, so the writer must fall back to its
// option defaults.
type FieldSetConsumer interface {
	SetReferencedFields(fields []string, analyzable bool)
}

// analyzeFields is the template analyzer behind ResolveFieldSets, swappable
// in tests to observe whether an analysis actually ran.
var analyzeFields = template.AnalyzeFields

// fieldSetAnalysisVersion is the analyzer generation stamped into configs
// alongside their field sets. BUMP THIS whenever the analysis changes -
// template.AnalyzeFields' semantics, the source lists walked here, or the
// stamp shape - so a config stamped by another binary generation is treated
// as unstamped (see Get) and a long-lived session (tmux) re-analyzes once
// after a binary upgrade instead of trusting stale stamps.
const fieldSetAnalysisVersion = 2

// ResolveFieldSets analyzes every template in the config and stamps each
// renderable segment with the set of top-level context fields those templates
// can read from it - its own templates plus any .Segments.<name> reference
// from another segment, an extra prompt, a tooltip, or a global template
// string (console title, fillers, palette values, ...).
// MapSegmentWithWriter hands the stamped set to writers implementing
// FieldSetConsumer. Idempotent, so callers can resolve defensively: the
// FieldSetsVersion marker survives the session cache's gob round trip (Store
// stamps before encoding), making this a no-op on restored configs -
// analysis runs once per config content per shell session, not per prompt.
func (cfg *Config) ResolveFieldSets() {
	if cfg.FieldSetsVersion == fieldSetAnalysisVersion {
		return
	}

	cfg.FieldSetsVersion = fieldSetAnalysisVersion

	analysis := newFieldAnalysis()

	// templates rendered against the global context: their own-dot fields
	// belong to no segment, only their cross-segment references count
	for _, text := range cfg.globalTemplateSources() {
		analysis.analyzeGlobal(text)
	}

	segments := cfg.renderableSegments()

	for _, segment := range segments {
		analysis.analyzeSegment(segment)
	}

	for _, segment := range segments {
		analysis.stamp(segment)
	}
}

// renderableSegments returns the segments that evaluate a writer of their
// own: block segments and tooltips. Extra prompt segments (transient,
// secondary, ...) are containers for global templates, not writers, and are
// handled by globalTemplateSources instead.
func (cfg *Config) renderableSegments() []*Segment {
	var segments []*Segment

	for _, block := range cfg.Blocks {
		segments = append(segments, block.Segments...)
	}

	return append(segments, cfg.Tooltips...)
}

func (cfg *Config) globalTemplateSources() []string {
	// nil-context template strings can still reach any segment's data
	// through the global context's .Segments, so they all go through the
	// cross-segment extraction pass
	sources := []string{
		cfg.ConsoleTitleTemplate,
		cfg.PWD,
		cfg.CursorStyle,
		string(cfg.TerminalBackground),
	}

	for _, block := range cfg.Blocks {
		sources = append(sources, block.Filler)
	}

	for _, value := range cfg.Palette {
		sources = append(sources, string(value))
	}

	if cfg.Palettes != nil {
		sources = append(sources, cfg.Palettes.Template)

		for _, palette := range cfg.Palettes.List {
			for _, value := range palette {
				sources = append(sources, string(value))
			}
		}
	}

	for _, segment := range []*Segment{cfg.TransientPrompt, cfg.SecondaryPrompt, cfg.ValidLine, cfg.ErrorLine, cfg.DebugPrompt} {
		if segment == nil {
			continue
		}

		sources = append(sources, segment.templateSources()...)
	}

	return sources
}

// templateSources lists every template string this segment can render,
// mirroring the sources evaluateNeeds walks plus the placeholder, the right
// template and filler (rendered for extra prompts), and the style (resolved
// as a template against the segment's own writer, see SegmentStyle.resolve).
func (segment *Segment) templateSources() []string {
	sources := []string{
		segment.Template,
		segment.RightTemplate,
		segment.FallbackTemplate,
		segment.Placeholder,
		segment.Filler,
		string(segment.Style),
	}

	sources = append(sources, segment.Templates...)
	sources = append(sources, segment.ForegroundTemplates...)
	sources = append(sources, segment.BackgroundTemplates...)

	return sources
}

// templatedOptionValues collects the segment's option values (nested ones
// included) that carry template syntax. These render through
// options.Map.Template against contexts this analysis cannot model per
// option, so analyzeSegment treats any hit conservatively.
func (segment *Segment) templatedOptionValues() []string {
	var sources []string

	var collect func(value any)
	collect = func(value any) {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "{{") {
				sources = append(sources, v)
			}
		case map[string]any:
			for _, nested := range v {
				collect(nested)
			}
		case map[any]any:
			for _, nested := range v {
				collect(nested)
			}
		case options.Map:
			for _, nested := range v {
				collect(nested)
			}
		case []any:
			for _, nested := range v {
				collect(nested)
			}
		case []string:
			for _, nested := range v {
				collect(nested)
			}
		}
	}

	for _, value := range segment.Options {
		collect(value)
	}

	return sources
}

// analysisTemplate resolves the template a segment renders when none is
// configured: the writer's default. The throwaway writer is a plain struct
// construction; on builds without writers there is none, and neither is
// there anything to gate on the recorded-data render path.
func (segment *Segment) analysisTemplate() string {
	if segment.Template != "" {
		return segment.Template
	}

	writer, err := newSegmentWriter(segment.Type)
	if err != nil || writer == nil {
		return ""
	}

	return writer.Template()
}

type fieldAnalysis struct {
	own         map[*Segment]map[string]bool
	ownOpaque   map[*Segment]bool
	cross       map[string]map[string]bool
	crossOpaque map[string]bool
	opaque      bool
}

func newFieldAnalysis() *fieldAnalysis {
	return &fieldAnalysis{
		own:         make(map[*Segment]map[string]bool),
		ownOpaque:   make(map[*Segment]bool),
		cross:       make(map[string]map[string]bool),
		crossOpaque: make(map[string]bool),
	}
}

func (a *fieldAnalysis) analyzeSegment(segment *Segment) {
	ownSet := make(map[string]bool)
	a.own[segment] = ownSet

	// the writer's default template renders when none is configured, so it
	// replaces the empty Template entry at index 0
	sources := segment.templateSources()
	sources[0] = segment.analysisTemplate()

	for _, text := range sources {
		refs := analyzeFields(text)

		for field := range refs.Own {
			ownSet[field] = true
		}

		if refs.OwnOpaque {
			a.ownOpaque[segment] = true
		}

		a.mergeCross(refs)
	}

	// Templated option values (branch_template and friends) render against
	// per-option contexts this walk cannot model, so any hit is handled
	// conservatively: the segment's own set becomes untrustworthy - falling
	// back to the explicit fetch options, which is display-correct by
	// construction - while cross-segment references inside those options
	// still count toward the segments they name. refs.Own/OwnOpaque are
	// deliberately dropped here: the option's own dot is not this segment's
	// context.
	for _, text := range segment.templatedOptionValues() {
		a.ownOpaque[segment] = true

		a.mergeCross(analyzeFields(text))
	}
}

func (a *fieldAnalysis) analyzeGlobal(text string) {
	refs := analyzeFields(text)

	// the global context embeds .Segments, so losing track of the own dot
	// here can hide a reference to any segment
	if refs.OwnOpaque {
		a.opaque = true
	}

	a.mergeCross(refs)
}

func (a *fieldAnalysis) mergeCross(refs *template.Refs) {
	if refs.Opaque {
		a.opaque = true
	}

	for name, fields := range refs.Segments {
		if a.cross[name] == nil {
			a.cross[name] = make(map[string]bool)
		}

		for field := range fields {
			a.cross[name][field] = true
		}
	}

	for name := range refs.SegmentsOpaque {
		a.crossOpaque[name] = true
	}
}

// stamp fixes the analysis outcome onto the segment: the sorted union of its
// own references and the cross-references to its data key, plus whether that
// union is trustworthy. Cross-references resolve by Name(), the key
// AddSegmentData stores segment data under.
func (a *fieldAnalysis) stamp(segment *Segment) {
	name := segment.Name()

	set := a.own[segment]
	for field := range a.cross[name] {
		set[field] = true
	}

	fields := make([]string, 0, len(set))
	for field := range set {
		fields = append(fields, field)
	}

	slices.Sort(fields)

	segment.ReferencedFields = fields
	segment.FieldsAnalyzable = !a.opaque && !a.ownOpaque[segment] && !a.crossOpaque[name]
}
