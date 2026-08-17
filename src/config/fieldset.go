package config

import (
	"slices"

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

// ResolveFieldSets analyzes every template in the config and stamps each
// renderable segment with the set of top-level context fields those templates
// can read from it - its own templates plus any .Segments.<name> reference
// from another segment, an extra prompt, a tooltip, or the console title.
// MapSegmentWithWriter hands the stamped set to writers implementing
// FieldSetConsumer. Idempotent, so callers can resolve defensively: the
// FieldSetsResolved marker survives the session cache's gob round trip
// (Store stamps before encoding), making this a no-op on restored configs -
// analysis runs once per config content per shell session, not per prompt.
func (cfg *Config) ResolveFieldSets() {
	if cfg.FieldSetsResolved {
		return
	}

	cfg.FieldSetsResolved = true

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
	sources := []string{cfg.ConsoleTitleTemplate}

	for _, segment := range []*Segment{cfg.TransientPrompt, cfg.SecondaryPrompt, cfg.ValidLine, cfg.ErrorLine, cfg.DebugPrompt} {
		if segment == nil {
			continue
		}

		sources = append(sources, segment.templateSources()...)
	}

	return sources
}

// templateSources lists every template string this segment can render,
// mirroring the sources evaluateNeeds walks plus the placeholder and the
// right template (rendered for extra prompts).
func (segment *Segment) templateSources() []string {
	sources := []string{
		segment.Template,
		segment.RightTemplate,
		segment.FallbackTemplate,
		segment.Placeholder,
	}

	sources = append(sources, segment.Templates...)
	sources = append(sources, segment.ForegroundTemplates...)
	sources = append(sources, segment.BackgroundTemplates...)

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
