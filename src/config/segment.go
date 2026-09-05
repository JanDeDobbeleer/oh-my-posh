package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"strconv"
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
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"go.yaml.in/yaml/v3"
)

type SegmentStyle string

func (s *SegmentStyle) resolve(context any) SegmentStyle {
	value, err := template.RenderTrusted(string(*s), context)

	// default to Plain
	if err != nil || value == "" {
		return Plain
	}

	return SegmentStyle(value)
}

type Segment struct {
	writer SegmentWriter
	// data is what templates evaluate against when there is no writer to evaluate against: a
	// build with no segment packages linked restores recorded data into a plain map instead of
	// into a writer struct. Nil everywhere else, and templateContext picks whichever of the two is
	// present. Go templates resolve a name against a map key exactly as they resolve it against a
	// field or a method, so a recorded value reads the same either way.
	data map[string]any
	// text is where the rendered text lives when there is no writer to hold it: the writer
	// normally stores it (SegmentWriter.SetText/Text), which a build with no writers cannot do.
	text                   string
	env                    runtime.Environment
	Options                options.Map `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty"`
	Properties             options.Map `json:"-" toml:"properties,omitempty" yaml:"-"`
	Cache                  *Cache      `json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty"`
	presentFields          map[string]bool
	Alias                  string `json:"alias,omitempty" toml:"alias,omitempty" yaml:"alias,omitempty"`
	styleCache             SegmentStyle
	foregroundCache        color.Ansi
	backgroundCache        color.Ansi
	name                   string
	LeadingDiamond         string         `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty" yaml:"leading_diamond,omitempty"`
	TrailingDiamond        string         `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty" yaml:"trailing_diamond,omitempty"`
	Template               string         `json:"template,omitempty" toml:"template,omitempty" yaml:"template,omitempty"`
	RightTemplate          string         `json:"right_template,omitempty" toml:"right_template,omitempty" yaml:"right_template,omitempty"`
	Foreground             color.Ansi     `json:"foreground,omitempty" toml:"foreground,omitempty" yaml:"foreground,omitempty"`
	TemplatesLogic         template.Logic `json:"templates_logic,omitempty" toml:"templates_logic,omitempty" yaml:"templates_logic,omitempty"`
	PowerlineSymbol        string         `json:"powerline_symbol,omitempty" toml:"powerline_symbol,omitempty" yaml:"powerline_symbol,omitempty"`
	Background             color.Ansi     `json:"background,omitempty" toml:"background,omitempty" yaml:"background,omitempty"`
	Filler                 string         `json:"filler,omitempty" toml:"filler,omitempty" yaml:"filler,omitempty"`
	Type                   SegmentType    `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	Style                  SegmentStyle   `json:"style,omitempty" toml:"style,omitempty" yaml:"style,omitempty"`
	LeadingPowerlineSymbol string         `json:"leading_powerline_symbol,omitempty" toml:"leading_powerline_symbol,omitempty" yaml:"leading_powerline_symbol,omitempty"`
	Placeholder            string         `json:"placeholder,omitempty" toml:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	FallbackTemplate       string         `json:"fallback_template,omitempty" toml:"fallback_template,omitempty" yaml:"fallback_template,omitempty"`
	Tips                   []string       `json:"tips,omitempty" toml:"tips,omitempty" yaml:"tips,omitempty"`
	BackgroundTemplates    template.List  `json:"background_templates,omitempty" toml:"background_templates,omitempty" yaml:"background_templates,omitempty"`
	Templates              template.List  `json:"templates,omitempty" toml:"templates,omitempty" yaml:"templates,omitempty"`
	ExcludeFolders         []string       `json:"exclude_folders,omitempty" toml:"exclude_folders,omitempty" yaml:"exclude_folders,omitempty"`
	IncludeFolders         []string       `json:"include_folders,omitempty" toml:"include_folders,omitempty" yaml:"include_folders,omitempty"`
	Needs                  []string       `json:"-" toml:"-" yaml:"-"`
	ForegroundTemplates    template.List  `json:"foreground_templates,omitempty" toml:"foreground_templates,omitempty" yaml:"foreground_templates,omitempty"`
	pendingData            json.RawMessage
	// ReferencedFields is the sorted, analysis-derived set of top-level fields
	// the config's templates can read from this segment, stamped by
	// Config.ResolveFieldSets and only trustworthy when FieldsAnalyzable is
	// true; see FieldSetConsumer. Exported (but kept out of every config
	// format, like Needs) so the session cache's gob round trip preserves
	// the analysis instead of forcing a re-run on every render.
	ReferencedFields []string `json:"-" toml:"-" yaml:"-"`
	// HeuristicSources is the whole-config text corpus the fallback
	// heuristic scans when FieldsAnalyzable is false - stamped (and
	// gob-persisted) because the reference that defeated the analysis can
	// live outside this segment, in texts the segment cannot reconstruct
	// from its own fields. Nil for analyzable segments and for configs that
	// never went through ResolveFieldSets.
	HeuristicSources    []string      `json:"-" toml:"-" yaml:"-"`
	Index               int           `json:"index,omitempty" toml:"index,omitempty" yaml:"index,omitempty"`
	MinWidth            int           `json:"min_width,omitempty" toml:"min_width,omitempty" yaml:"min_width,omitempty"`
	Duration            time.Duration `json:"-" toml:"-" yaml:"-"`
	NameLength          int           `json:"-" toml:"-" yaml:"-"`
	MaxWidth            int           `json:"max_width,omitempty" toml:"max_width,omitempty" yaml:"max_width,omitempty"`
	Timeout             int           `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
	Newline             bool          `json:"newline,omitempty" toml:"newline,omitempty" yaml:"newline,omitempty"`
	Enabled             bool          `json:"-" toml:"-" yaml:"-"`
	InvertPowerline     bool          `json:"invert_powerline,omitempty" toml:"invert_powerline,omitempty" yaml:"invert_powerline,omitempty"`
	Force               bool          `json:"force,omitempty" toml:"force,omitempty" yaml:"force,omitempty"`
	restored            bool          `json:"-" toml:"-" yaml:"-"`
	Toggled             bool          `json:"toggled,omitempty" toml:"toggled,omitempty" yaml:"toggled,omitempty"`
	Pending             bool          `json:"-" toml:"-" yaml:"-"`
	Killed              bool          `json:"-" toml:"-" yaml:"-"`
	Interactive         bool          `json:"interactive,omitempty" toml:"interactive,omitempty" yaml:"interactive,omitempty"`
	MultilineKeepPrompt bool          `json:"multiline_keepprompt,omitempty" toml:"multiline_keepprompt,omitempty" yaml:"multiline_keepprompt,omitempty"`
	foregroundResolved  bool
	backgroundResolved  bool
	needsEvaluated      bool
	evaluated           bool
	FieldsAnalyzable    bool `json:"-" toml:"-" yaml:"-"`
}

// A nil presentFields map means presence was never recorded, in which case every
// field is treated as present, preserving merge's legacy unconditional-overwrite
// behavior for such segments. name is the json tag key.
func (segment *Segment) fieldPresent(name string) bool {
	if segment.presentFields == nil {
		return true
	}

	return segment.presentFields[name]
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

// Needed for TOML configs since go-toml/v2 doesn't support custom unmarshalers.
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
		name = text.Title(string(segment.Type))
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

	if segment.restoreData() {
		return
	}

	cacheRestored := segment.restoreCache()
	if cacheRestored && !env.Flags().Streaming {
		// A hand-written entry stashed itself in pendingData above instead of
		// short-circuiting, expecting the overlay to run once live/derived
		// state is available (see overlayData). A cache hit is exactly such
		// state and returns early right here, before the writer.Enabled() call
		// further down ever runs - so without this, pinned data would silently
		// lose to a live cache hit instead of winning as documented. Safe to
		// call unconditionally: json.Unmarshal into the cache-restored writer
		// only touches fields pendingData set, and it is a no-op when
		// pendingData is empty (the common case: no hand-written override for
		// this segment).
		segment.overlayData()
		return
	}

	if shouldHideForWidth(segment.env, segment.MinWidth, segment.MaxWidth) {
		return
	}

	if !segment.gateActive() {
		if segment.FallbackTemplate != "" {
			// Contract change (deliberate): a fallback template used to force
			// the full Enabled() evaluation so it could render against the
			// evaluated writer. Now the gate wins: the segment counts as
			// evaluated without Enabled() ever running, and the fallback
			// renders against the zero-state writer.
			segment.evaluated = true
			log.Debugf("segment gated (inactive), fallback renders against zero state: %s", segment.Name())
			return
		}

		log.Debugf("segment gated (inactive): %s", segment.Name())
		return
	}

	defer func() {
		if segment.Enabled {
			template.Cache.AddSegmentData(segment.Name(), segment.templateContext())
		}
	}()

	// Only segments with a timeout can ever be killed via
	// KillGoroutineChildren (see prompt/segments.go executeSegmentWithTimeout),
	// so only those need a Job object to track/terminate their child
	// processes. Skipping this for the common case (no timeout configured)
	// avoids two syscalls + a map insert per segment per prompt.
	if segment.Timeout > 0 {
		if err := runjobs.CreateJobForGoroutine(segment.Name()); err != nil {
			log.Errorf("failed to create job for goroutine (segment: %s): %v", segment.Name(), err)
		}

		// Release the Job object (Windows) once this goroutine is done
		// spawning/waiting on children. cmd.Run blocks until its child exits,
		// so by the time Execute returns - on any path below, including a
		// panic unwind - no process we intended to keep alive is still
		// assigned to the job, making it safe to close despite
		// JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE. If the timeout path already
		// terminated/closed the job concurrently, this is a no-op; see the
		// ownership invariant documented on jobs.CloseGoroutineJob.
		defer runjobs.CloseGoroutineJob()
	}

	switch {
	case segment.writer != nil:
		segment.Enabled = segment.writer.Enabled()
	case !segment.restored:
		// No writer to ask and nothing recorded about this one, so the template is the only thing
		// that can decide. Render already settles it that way - a segment is on when its text is
		// not blank - and this is what lets a segment needing no data at all, a text segment
		// being the obvious case, still draw itself.
		segment.Enabled = true
	}

	segment.evaluated = true

	segment.overlayData()
}

// gateActive evaluates the writer's activation gate: a cheap, declarative
// pre-check that proves a segment cannot possibly be enabled in the current
// working directory, letting Execute skip the (potentially expensive)
// writer.Enabled() probe entirely. Activation() is part of the SegmentWriter
// contract; writers inherit the ungated (Always) default from segments.Base.
//
// Two segment shapes bypass the gate and run the full path regardless of its
// outcome: forced segments (Force), and segments pinned via a hand-written
// data file (pendingData), whose overlay expects writer.Enabled() to derive
// live state first. A nil writer (the js/wasm build) has no gate to consult.
// Deliberately not bypassed: a segment with a cache config whose restore
// missed gates like any other - the next cache fill simply waits until the
// segment can activate again.
func (segment *Segment) gateActive() bool {
	if segment.Force || len(segment.pendingData) > 0 {
		return true
	}

	if segment.writer == nil {
		return true
	}

	activation := segment.writer.Activation()
	return activation.Active(segment.env)
}

// overlayData applies data pinned in a hand-written (unmarked) file on top of the
// writer's live-computed state. It must run after writer.Enabled() above, never
// before: roughly half of all segment writers compute template-visible state
// inside Enabled() (Time.Format and friends), and overlaying first would only
// have that live computation clobber the pinned values right back. json.Unmarshal
// into the already-initialized writer only touches fields present in
// pendingData, so pinned values win and everything else keeps its live-derived
// value.
//
// Matches restoreData's existing contract for a recorded entry: a pinned
// segment renders even where its own live check would suppress it (battery on a
// machine without one, say). That is intentional and stays; it just now also
// applies to hand-written files.
func (segment *Segment) overlayData() {
	if len(segment.pendingData) == 0 {
		return
	}

	if err := json.Unmarshal(segment.pendingData, &segment.writer); err != nil {
		log.Error(err)
		return
	}

	segment.Enabled = true
	segment.restored = true

	log.Debug("derived and overlaid segment from data: ", segment.Name())
}

func (segment *Segment) Render(index int, force bool) bool {
	// Foreground/background may be overridden directly (e.g. color cycling) between
	// render passes, so the memoized values must not survive across calls to Render.
	// This reset must precede the early return below: a disabled segment without a
	// fallback template would otherwise serve a stale collapsed color from a prior
	// pass to parent-color and separator consumers.
	segment.foregroundResolved = false
	segment.backgroundResolved = false

	// Allow pending segments to render (they'll show "..." text)
	if !segment.Pending && !segment.Enabled && !force {
		return segment.renderFallback(index)
	}

	if force {
		segment.Force = true
	}

	segment.setIndex(index)

	rendered := segment.string()

	// Only update Enabled if segment is NOT pending (avoid race with Execute goroutine)
	if !segment.Pending {
		segment.Enabled = segment.Force || strings.ContainsFunc(rendered, func(r rune) bool { return r != ' ' })

		if !segment.Enabled {
			template.Cache.RemoveSegmentData(segment.Name())
			return false
		}
	}

	segment.SetText(rendered)
	segment.setCache()

	// We do this to make `.Text` available for a cross-segment reference in an extra prompt.
	template.Cache.AddSegmentData(segment.Name(), segment.templateContext())

	return true
}

// renderFallback attempts to render FallbackTemplate when a segment would
// otherwise be hidden because its writer's Enabled() returned false. It
// requires the writer to have completed its evaluation; segments hidden for other
// reasons (toggled off, folder include/exclude, width constraints, or a
// writer mapping error) never set evaluated, so they keep the silent-omission
// behavior. Segments killed by their timeout are excluded via Killed rather
// than evaluated, because their Execute goroutine keeps running after the
// kill and may still complete the evaluation before rendering starts.
func (segment *Segment) renderFallback(index int) bool {
	if segment.FallbackTemplate == "" || !segment.evaluated || segment.Killed {
		return false
	}

	rendered, err := template.RenderTrusted(segment.FallbackTemplate, segment.writer)
	if err != nil {
		rendered = err.Error()
	}

	if !strings.ContainsFunc(rendered, func(r rune) bool { return r != ' ' }) {
		return false
	}

	// Foreground/background may be overridden directly (e.g. color cycling) between
	// render passes, so the memoized values must not survive across calls to Render.
	segment.foregroundResolved = false
	segment.backgroundResolved = false

	segment.setIndex(index)
	segment.Enabled = true
	segment.SetText(rendered)

	// Intentionally skip setCache(): the writer is zero/partially hydrated
	// here, and caching it would make restoreCache() later resurrect a
	// disabled writer as enabled while rendering the main template against
	// empty data.
	template.Cache.AddSegmentData(segment.Name(), segment.templateContext())

	return true
}

func (segment *Segment) Text() string {
	if segment.writer == nil {
		return segment.text
	}

	return segment.writer.Text()
}

func (segment *Segment) SetText(value string) {
	if segment.writer == nil {
		segment.text = value
		return
	}

	segment.writer.SetText(value)
}

func (segment *Segment) setIndex(index int) {
	if segment.writer == nil {
		return
	}

	segment.writer.SetIndex(index)
}

func (segment *Segment) ResolveForeground() color.Ansi {
	if segment.foregroundResolved {
		return segment.foregroundCache
	}

	if len(segment.ForegroundTemplates) != 0 {
		match := segment.ForegroundTemplates.FirstMatch(segment.templateContext(), segment.Foreground.String())
		segment.Foreground = color.Ansi(match)
	}

	segment.foregroundCache = segment.Foreground
	segment.foregroundResolved = true

	return segment.foregroundCache
}

func (segment *Segment) ResolveBackground() color.Ansi {
	if segment.backgroundResolved {
		return segment.backgroundCache
	}

	if len(segment.BackgroundTemplates) != 0 {
		match := segment.BackgroundTemplates.FirstMatch(segment.templateContext(), segment.Background.String())
		segment.Background = color.Ansi(match)
	}

	segment.backgroundCache = segment.Background
	segment.backgroundResolved = true

	return segment.backgroundCache
}

// CollapseBackground overrides the segment's resolved background color cache, bypassing
// template resolution, without touching the raw Background field (which ResolveBackground
// falls back to as a template-miss default on a later render pass). It is used by the prompt
// engine to collapse a gradient background to a solid color once per segment, before anything
// renders, so every later call to ResolveBackground this render (separators, diamonds, parent
// color references) sees the same solid color instead of the raw gradient string.
func (segment *Segment) CollapseBackground(background color.Ansi) {
	segment.backgroundCache = background
	segment.backgroundResolved = true
}

// CollapseForeground is CollapseBackground's foreground counterpart.
func (segment *Segment) CollapseForeground(foreground color.Ansi) {
	segment.foregroundCache = foreground
	segment.foregroundResolved = true
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

func (segment *Segment) DataKey() string {
	if segment.Alias != "" {
		return segment.Alias
	}

	return string(segment.Type)
}

func (segment *Segment) Writer() SegmentWriter {
	return segment.writer
}

func (segment *Segment) isToggled() bool {
	togglesMap, OK := cache.Session.Get[map[string]bool](cache.TOGGLECACHE)
	if !OK || len(togglesMap) == 0 {
		log.Debug("no toggles found")
		return false
	}

	if togglesMap[segment.DataKey()] {
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

	data, OK := store.Get[any](key)
	if !OK {
		log.Debugf("no cache found for segment: %s, key: %s", segment.Name(), key)
		return false
	}

	switch v := data.(type) {
	case []byte:
		// Decode into the writer initialized by MapSegmentWithWriter instead of
		// replacing it: the snapshot only carries exported fields, while the
		// writer's unexported runtime state (env, options) must stay intact or
		// any method relying on it panics after a restore.
		if err := gob.NewDecoder(bytes.NewReader(v)).Decode(segment.writer); err != nil {
			log.Error(err)
			store.Delete(key)
			return false
		}
	case string:
		// legacy JSON cache entry, remove it so it gets re-cached in the new format
		log.Debugf("removing legacy cache key: %s", key)
		store.Delete(key)
		return false
	default:
		log.Debugf("unexpected cache type for segment: %s, key: %s", segment.Name(), key)
		store.Delete(key)
		return false
	}

	segment.Enabled = true
	template.Cache.AddSegmentData(segment.Name(), segment.templateContext())

	log.Debug("restored segment from cache: ", segment.Name())

	segment.restored = true

	return true
}

// restoreData replays a segment's writer state from the data file supplied via
// runtime.Flags.SegmentData, bypassing the real runtime entirely. This lets
// segments render from a recorded fixture instead of probing the environment.
//
// Two shapes reach here, distinguished by structure rather than a flag threaded
// through runtime.Flags (which this package does not own, and cannot extend): a
// RecordedSegment envelope - {"enabled":...,"data":...}, exactly those two keys -
// written by `config export data` for a versioned file, or the flat writer JSON
// a hand-written file has always stored directly.
//
// A recorded, enabled entry is unmarshaled and returns true here, same as
// before: no probe. A recorded-but-disabled entry is suppressed here with no
// probe either - that is what makes replay hermetic. A flat entry cannot be
// trusted this early: it stashes itself in pendingData and returns false so
// Execute keeps running - shouldHideForWidth, the Job setup, and
// writer.Enabled() all still fire - and overlayData applies it afterward.
// There are two call sites for that overlay, both in Execute: a device/session
// cache hit in restoreCache also produces state pendingData needs to win
// over, and returns before writer.Enabled() runs, so it overlays right there;
// everything else overlays after writer.Enabled() further down.
//
// Nothing here is conditional on runtime.Flags.DataOnly, deliberately. That
// flag makes the *environment* refuse every probe (see runtime.Terminal), so a
// segment with no recorded entry still runs and still falls through to
// writer.Enabled() - it simply finds nothing when it reaches for the machine,
// and reports itself disabled. Suppressing it here as well was tried and
// backed out: it also killed the segments that need no machine at all. path
// and session derive from the pinned PWD and user in the data file's env
// section, so they render correctly under DataOnly, and a rule keyed on
// "is there an entry under this segment's DataKey" made both vanish from the
// studio, which is exactly where they matter most.
func (segment *Segment) restoreData() bool {
	raw, OK := segment.env.Flags().SegmentData[segment.DataKey()]
	if !OK {
		return false
	}

	recorded, isRecorded := decodeRecordedSegment(raw)
	if !isRecorded {
		segment.pendingData = raw
		return false
	}

	if !recorded.Enabled {
		segment.Enabled = false
		segment.restored = true

		log.Debug("suppressing recorded-but-disabled segment: ", segment.Name())

		return true
	}

	if err := segment.restoreInto(recorded.Data, recorded.Methods); err != nil {
		log.Error(err)
		return false
	}

	segment.Enabled = true
	segment.restored = true

	template.Cache.AddSegmentData(segment.Name(), segment.templateContext())

	log.Debug("restored segment from data: ", segment.Name())

	return true
}

// restoreInto unmarshals a recorded segment's data into whatever this build renders from: the
// writer struct where one was constructed, and a plain map where none was. The map is not a
// degraded form - a template resolves a name against a map key exactly as it resolves it against
// a struct field, so both carry the same recorded values to the same templates. What a map cannot
// carry is a method result, which is why the recorder writes those out as data too.
//
// methods is the overlay for the values whose method results could not be written as data without
// changing what the writer sees (RecordedSegment.Methods explains which). It is read here and
// nowhere else: a writer brings its own methods, so only the map ever needs it.
func (segment *Segment) restoreInto(raw, methods json.RawMessage) error {
	if segment.writer != nil {
		return json.Unmarshal(raw, &segment.writer)
	}

	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	if len(methods) != 0 {
		overlay := make(map[string]any)
		if err := json.Unmarshal(methods, &overlay); err != nil {
			return err
		}

		MergeRecordedMethods(data, overlay)
	}

	// Markup fields were recorded tagged (see template.Markup's JSON form);
	// revive them or every recorded anchor renders escaped on this path.
	segment.data = template.ReviveMarkup(normalizeNumbers(data)).(map[string]any)

	return nil
}

// normalizeNumbers turns whole float64s back into ints.
//
// JSON has one number type, so unmarshalling into a map makes every number a float64 - where
// unmarshalling into a writer struct would have produced whatever the field declared. Templates
// notice: sysinfo renders `{{ round .PhysicalPercentUsed .Precision }}`, and round wants its
// precision as an int, so a Precision that arrived as 2.0 fails the whole template where the
// struct's int 2 succeeded.
//
// Whole numbers become int and the rest stay float64. int rather than int64 because a template
// function's parameter type has to match exactly: sprig's round takes an int for its precision,
// and Go templates will not narrow an int64 to reach it, so `{{ round .PhysicalPercentUsed
// .Precision }}` failed on an int64 exactly as it failed on a float64.
//
// Lossy in one direction - a float field holding exactly 2.0 marshals as "2" and comes back an
// int - but the functions taking a float accept an int and widen it, while the ones taking an int
// reject a float outright. int is the guess that still renders.
func normalizeNumbers(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			typed[key] = normalizeNumbers(nested)
		}

		return typed
	case []any:
		for i, nested := range typed {
			typed[i] = normalizeNumbers(nested)
		}

		return typed
	case float64:
		if typed == math.Trunc(typed) && !math.IsInf(typed, 0) {
			return int(typed)
		}

		return typed
	default:
		return value
	}
}

// decodeRecordedSegment reports whether raw is exactly a RecordedSegment
// envelope: a JSON object holding "enabled" and "data", and nothing beyond an
// optional "methods". Anything short of that - a flat hand-written entry, or
// malformed JSON - is left for the caller to treat as unmarked data.
func decodeRecordedSegment(raw json.RawMessage) (RecordedSegment, bool) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return RecordedSegment{}, false
	}

	enabledRaw, hasEnabled := fields["enabled"]
	dataRaw, hasData := fields["data"]
	if !hasEnabled || !hasData {
		return RecordedSegment{}, false
	}

	methodsRaw, hasMethods := fields["methods"]

	if len(fields) != 2 && (len(fields) != 3 || !hasMethods) {
		return RecordedSegment{}, false
	}

	var enabled bool
	if err := json.Unmarshal(enabledRaw, &enabled); err != nil {
		return RecordedSegment{}, false
	}

	return RecordedSegment{Data: dataRaw, Methods: methodsRaw, Enabled: enabled}, true
}

func (segment *Segment) setCache() {
	if segment.restored || !segment.hasCache() {
		return
	}

	// Never cache pending state to avoid polluting cache with incomplete data
	if segment.Pending {
		return
	}

	// Store a gob snapshot rather than the writer itself. The writer is a live
	// object that render passes keep mutating, so caching it directly would
	// share mutable state through the cache; and once persisted to disk it
	// would come back without its unexported runtime state (env, options).
	// The snapshot is immutable and restoreCache overlays it onto a freshly
	// initialized writer. An encode failure only skips caching this segment.
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(segment.writer); err != nil {
		log.Error(err)
		return
	}

	key, store := segment.cacheKeyAndStore()
	store.Set(key, data.Bytes(), segment.Cache.Duration)
}

func (segment *Segment) cacheKeyAndStore() (string, cache.Store) {
	name := segment.Name()

	// A field-set consuming writer fetches different data per referenced-field
	// set, so a gob snapshot taken under one set must never be restored under
	// another: fold the set's fingerprint into the key. Every other writer
	// keeps its key untouched.
	if _, ok := segment.writer.(FieldSetConsumer); ok {
		name = strings.Join([]string{name, segment.fieldSetFingerprint()}, "_")
	}

	format := "segment_cache_%s"
	switch segment.Cache.Strategy {
	case Session:
		return fmt.Sprintf(format, name), cache.Session
	case Device:
		return fmt.Sprintf(format, name), cache.Device
	case Folder:
		fallthrough
	default:
		return fmt.Sprintf(format, strings.Join([]string{name, segment.folderKey()}, "_")), cache.Device
	}
}

// fieldSetFingerprint condenses the stamped field set (and whether it is
// trustworthy) into a short stable token for the segment cache key.
// ReferencedFields is sorted by ResolveFieldSets, so equal sets always
// fingerprint identically. An unanalyzable set fetches by the heuristic
// over the raw sources instead, so those sources join the fingerprint -
// two configs with identical (empty) field sets but different templated
// content must never share a snapshot.
func (segment *Segment) fieldSetFingerprint() string {
	h := fnv.New64a()

	if segment.FieldsAnalyzable {
		_, _ = h.Write([]byte{1})
	}

	for _, field := range segment.ReferencedFields {
		_, _ = h.Write([]byte(field))
		_, _ = h.Write([]byte{0})
	}

	if !segment.FieldsAnalyzable {
		for _, source := range segment.fallbackSources() {
			_, _ = h.Write([]byte(source))
			_, _ = h.Write([]byte{0})
		}
	}

	return strconv.FormatUint(h.Sum64(), 36)
}

func (segment *Segment) folderKey() string {
	if segment.writer == nil {
		return segment.env.Pwd()
	}

	key, ok := segment.writer.CacheKey()
	if !ok {
		return segment.env.Pwd()
	}

	return key
}

// templateContext is what a template evaluates against: the writer when one was constructed, the
// recorded data map when none was (see Segment.data).
func (segment *Segment) templateContext() any {
	if segment.data != nil {
		return segment.data
	}

	return segment.writer
}

func (segment *Segment) string() string {
	// Use simple pending text if segment is still pending
	if segment.Pending {
		if segment.Placeholder != "" {
			return segment.Placeholder
		}

		return "..."
	}

	context := segment.templateContext()

	result := segment.Templates.Resolve(context, "", segment.TemplatesLogic)
	if len(result) != 0 {
		return result
	}

	if segment.Template == "" && segment.writer != nil {
		segment.Template = segment.writer.Template()
	}

	rendered, err := template.RenderTrusted(segment.Template, context)
	if err != nil {
		return err.Error()
	}

	return rendered
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
	if segment.needsEvaluated {
		return
	}

	segment.needsEvaluated = true

	value := segment.Template
	if value == "" && segment.writer != nil {
		value = segment.writer.Template()
	}

	if len(segment.ForegroundTemplates) != 0 {
		value += strings.Join(segment.ForegroundTemplates, "")
	}

	if len(segment.BackgroundTemplates) != 0 {
		value += strings.Join(segment.BackgroundTemplates, "")
	}

	if len(segment.Templates) != 0 {
		value += strings.Join(segment.Templates, "")
	}

	if segment.FallbackTemplate != "" {
		value += segment.FallbackTemplate
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
