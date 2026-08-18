package template

import (
	"slices"
	"strings"
	"text/template/parse"
)

// RefSet is the field-reference information the config-level template
// analysis derives for one segment, as delivered to writers that decide what
// to fetch from it. Fields is always a trustworthy lower bound: every entry
// is a genuine reference (own or cross-segment). Analyzable reports whether
// it is also complete; when false, Sources carries the segment's raw
// template and option texts and Referenced falls back to a conservative
// substring heuristic over them.
type RefSet struct {
	Fields     []string
	Sources    []string
	Analyzable bool
}

// Referenced reports whether a data unit populating the given fields should
// be fetched. Exact references always win. For an unanalyzable set the
// heuristic scans the raw sources for each field name as a standalone
// identifier (case-sensitive, non-identifier runes or text edges on both
// sides), so laundering shapes that still name a field - $g := .Segments.Git
// followed by $g.Working - keep fetching it, while sources that never
// mention a unit's fields don't pay for it. The known limit: a truly opaque
// shape naming no fields ({{ . }} printed whole) fetches nothing extra.
func (r RefSet) Referenced(fields ...string) bool {
	for _, field := range fields {
		if slices.Contains(r.Fields, field) {
			return true
		}
	}

	if r.Analyzable {
		return false
	}

	for _, field := range fields {
		for _, source := range r.Sources {
			if containsIdentifier(source, field) {
				return true
			}
		}
	}

	return false
}

// containsIdentifier reports whether text contains word bounded by
// non-identifier characters (or the text's edges), so "Ahead" never matches
// inside "PushAhead". Field names are ASCII, making byte checks sufficient.
func containsIdentifier(text, word string) bool {
	if word == "" {
		return false
	}

	for start := 0; ; {
		idx := strings.Index(text[start:], word)
		if idx < 0 {
			return false
		}

		idx += start
		end := idx + len(word)

		startsWord := idx == 0 || !isIdentifierByte(text[idx-1])
		endsWord := end == len(text) || !isIdentifierByte(text[end])

		if startsWord && endsWord {
			return true
		}

		start = idx + 1
	}
}

func isIdentifierByte(b byte) bool {
	return b == '_' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// Refs is the outcome of statically analyzing template text for the top-level
// context fields its rendering can touch. Own collects the fields referenced
// on the template's own context, Segments the fields referenced per
// cross-referenced segment (keyed by its .Segments data key). The opaque
// flags mark what escaped the analysis: OwnOpaque the own context,
// SegmentsOpaque a specific segment's context, Opaque the template as a whole
// (parse failure, template include, or the raw segments map itself). A
// consumer must treat an opaque context as potentially touching every field.
type Refs struct {
	Own            map[string]bool
	Segments       map[string]map[string]bool
	SegmentsOpaque map[string]bool
	OwnOpaque      bool
	Opaque         bool
}

func (refs *Refs) recordOwn(field string) {
	refs.Own[field] = true
}

func (refs *Refs) recordSegment(segment, field string) {
	if refs.Segments[segment] == nil {
		refs.Segments[segment] = make(map[string]bool)
	}

	refs.Segments[segment][field] = true
}

// AnalyzeFields statically analyzes template text and reports which top-level
// context fields its rendering can touch. The text goes through the same
// patching the renderer applies (see patchTemplate), so the shape analyzed is
// the shape executed. It never fails: anything the walk cannot follow flips
// the corresponding opaque flag on the result instead.
//
// Results of this analysis are persisted in the session cache (see
// config.ResolveFieldSets): any change to what this function covers or
// reports must be accompanied by a bump of config's fieldSetAnalysisVersion
// so stale persisted stamps refresh.
func AnalyzeFields(text string) (refs *Refs) {
	refs = &Refs{
		Own:            make(map[string]bool),
		Segments:       make(map[string]map[string]bool),
		SegmentsOpaque: make(map[string]bool),
	}

	if !strings.Contains(text, "{{") {
		return refs
	}

	// patchTemplate chokes on shapes the renderer would fail on anyway
	// (a bare {{ .Segments }}, say); analysis must survive them regardless.
	defer func() {
		if r := recover(); r != nil {
			refs.Opaque = true
		}
	}()

	patched := &Text{template: text, trusted: true}
	patched.patchTemplate()

	tree := parse.New("fields")
	tree.Mode = parse.SkipFuncCheck
	treeSet := make(map[string]*parse.Tree)

	if _, err := tree.Parse(patched.template, "", "", treeSet); err != nil {
		refs.Opaque = true
		return refs
	}

	walker := &fieldWalker{refs: refs}
	own := dotRef{kind: dotOwnRoot}

	// treeSet also carries {{define}} blocks; TemplateNode marks the result
	// opaque, so walking their bodies with the own dot only ever adds fields.
	for _, parsed := range treeSet {
		if parsed.Root == nil {
			continue
		}

		walker.walkList(parsed.Root, own)
	}

	return refs
}

type dotKind int

const (
	// dotGrounded is a value whose owning top-level field is already recorded
	// (or that carries no trackable context at all): nothing reached through
	// it can widen the field set further.
	dotGrounded dotKind = iota
	// dotOwnRoot is the template's own, unnarrowed context.
	dotOwnRoot
	// dotSegmentRoot is a cross-referenced segment's unnarrowed context.
	dotSegmentRoot
)

// dotRef classifies what a template value (or the current dot) is bound to
// while walking the tree; segment carries the data key for dotSegmentRoot.
type dotRef struct {
	segment string
	kind    dotKind
}

var grounded = dotRef{kind: dotGrounded}

type fieldWalker struct {
	refs *Refs
}

// consume marks a value as used whole - printed, passed to a function,
// iterated, or assigned to a variable. That defeats field tracking for a
// root context, so the matching opaque flag flips; a grounded value's head
// is already recorded and stays analyzable.
func (w *fieldWalker) consume(value dotRef) {
	switch value.kind {
	case dotOwnRoot:
		w.refs.OwnOpaque = true
	case dotSegmentRoot:
		w.refs.SegmentsOpaque[value.segment] = true
	case dotGrounded:
	}
}

func (w *fieldWalker) walkList(list *parse.ListNode, dot dotRef) {
	if list == nil {
		return
	}

	for _, node := range list.Nodes {
		w.walkNode(node, dot)
	}
}

func (w *fieldWalker) walkNode(node parse.Node, dot dotRef) {
	switch n := node.(type) {
	case *parse.ListNode:
		w.walkList(n, dot)
	case *parse.ActionNode:
		w.consume(w.pipe(n.Pipe, dot))
	case *parse.IfNode:
		w.consume(w.pipe(n.Pipe, dot))
		w.walkList(n.List, dot)
		w.walkList(n.ElseList, dot)
	case *parse.WithNode:
		// with re-roots the dot to its pipeline's value, so a root
		// classification flows into the body instead of being consumed
		w.walkList(n.List, w.pipe(n.Pipe, dot))
		w.walkList(n.ElseList, dot)
	case *parse.RangeNode:
		// iterating a root context walks all of its entries
		w.consume(w.pipe(n.Pipe, dot))
		w.walkList(n.List, grounded)
		w.walkList(n.ElseList, dot)
	case *parse.TemplateNode:
		// an include renders text this walk never sees
		w.refs.Opaque = true

		if n.Pipe != nil {
			w.consume(w.pipe(n.Pipe, dot))
		}
	default:
		// text, comments, break/continue: nothing to track
	}
}

func (w *fieldWalker) pipe(pipe *parse.PipeNode, dot dotRef) dotRef {
	if pipe == nil {
		return grounded
	}

	result := grounded

	for i, cmd := range pipe.Cmds {
		// the previous command's value is piped into this one as an argument
		if i > 0 {
			w.consume(result)
		}

		result = w.command(cmd, dot)
	}

	// a root context assigned to a variable escapes the walk (laundering);
	// a narrowed value's head is already recorded, so its variable is inert
	if len(pipe.Decl) > 0 {
		w.consume(result)
	}

	return result
}

func (w *fieldWalker) command(cmd *parse.CommandNode, dot dotRef) dotRef {
	if len(cmd.Args) == 1 {
		return w.value(cmd.Args[0], dot)
	}

	// the patched form of a cross-segment reference: .Segments.MustGet "Name"
	if name, ok := segmentLookup(cmd, dot); ok {
		return dotRef{kind: dotSegmentRoot, segment: name}
	}

	// a function or method call consumes every argument whole
	for _, arg := range cmd.Args {
		w.consume(w.value(arg, dot))
	}

	return grounded
}

// segmentLookup matches the exact shape patchTemplate rewrites .Segments.Name
// into, so the analysis can keep following the reference instead of going
// opaque on the function call.
func segmentLookup(cmd *parse.CommandNode, dot dotRef) (string, bool) {
	if dot.kind != dotOwnRoot || len(cmd.Args) != 2 {
		return "", false
	}

	field, ok := cmd.Args[0].(*parse.FieldNode)
	if !ok || len(field.Ident) != 2 || field.Ident[0] != "Segments" || field.Ident[1] != "MustGet" {
		return "", false
	}

	name, ok := cmd.Args[1].(*parse.StringNode)
	if !ok {
		return "", false
	}

	return name.Text, true
}

// value classifies a single argument node, recording any field references it
// contains along the way.
func (w *fieldWalker) value(node parse.Node, dot dotRef) dotRef {
	switch n := node.(type) {
	case *parse.DotNode:
		return dot
	case *parse.FieldNode:
		return w.fields(dot, n.Ident)
	case *parse.ChainNode:
		return w.fields(w.value(n.Node, dot), n.Field)
	case *parse.PipeNode:
		return w.pipe(n, dot)
	case *parse.VariableNode:
		// assignments were classified at declaration (see pipe); a variable
		// that survived is narrowed, so nothing below it needs tracking
		return grounded
	default:
		// literals and function identifiers carry no context
		return grounded
	}
}

func (w *fieldWalker) fields(base dotRef, idents []string) dotRef {
	if len(idents) == 0 {
		return base
	}

	switch base.kind {
	case dotGrounded:
		return grounded
	case dotSegmentRoot:
		w.refs.recordSegment(base.segment, idents[0])
		return grounded
	case dotOwnRoot:
		return w.ownFields(idents)
	default:
		return grounded
	}
}

// ownFields resolves a field chain on the own root context, where .Segments
// opens the door to cross-segment references.
func (w *fieldWalker) ownFields(idents []string) dotRef {
	if idents[0] != "Segments" {
		w.refs.recordOwn(idents[0])
		return grounded
	}

	switch {
	case len(idents) == 1:
		// the raw segments map: anything can be read from it
		w.refs.Opaque = true
		return grounded
	case idents[1] == "Contains":
		// presence check only, no field access on any segment
		return grounded
	case idents[1] == "MustGet" || idents[1] == "Get" || idents[1] == "ToSimple":
		// a map method outside the recognized lookup shape (segmentLookup)
		w.refs.Opaque = true
		return grounded
	case len(idents) == 2:
		return dotRef{kind: dotSegmentRoot, segment: idents[1]}
	default:
		w.refs.recordSegment(idents[1], idents[2])
		return grounded
	}
}
