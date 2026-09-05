package template

import (
	"bytes"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"text/template/parse"
	"unicode"
	"unicode/utf8"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/generics"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Data any

type context struct {
	Data
	Getenv func(string) string
	cache.Template
}

func (c *context) init(t *Text) {
	c.Data = t.context
	c.Template = *Cache

	if t.trusted {
		c.Getenv = env.Getenv
		return
	}

	// Untrusted templates never get real env values: .Env.NAME is rewritten by
	// patchTemplate into (call .Getenv "NAME"), but a raw {{ call .Getenv "NAME" }}
	// bypasses that rewrite entirely, so the binding itself must be the boundary.
	c.Getenv = func(string) string { return "" }
}

var (
	renderPool *generics.Pool[*renderer]
)

// renderer holds only per-render mutable state: an output buffer and a context.
// The parsed template is now looked up from the shared parsedTemplates cache.
type renderer struct {
	context *context
	buffer  bytes.Buffer
}

func (t *renderer) release() {
	t.buffer.Reset()
	t.context.Data = nil
	renderPool.Put(t)
}

// Encodes the raw (unpatched) template text, the trust level (parsed
// templates are bound to a func map at Parse time, so a trusted and a
// restricted render of identical text must never share a cache entry), and,
// when context is non-nil, the reflect.Type of the context so that two
// different struct types whose patchTemplate output would differ (via
// hasField) get separate cache entries. For map[string]any contexts the key
// also includes the sorted exported key names, since patchTemplate output
// depends on which keys are present.
func templateCacheKey(rawText string, trusted bool, ctx any) string {
	key := rawText + "\x00" + strconv.FormatBool(trusted)

	if ctx == nil {
		return key
	}

	t := reflect.TypeOf(ctx)
	if t.Kind() == reflect.Map {
		if m, ok := ctx.(map[string]any); ok {
			keys := make([]string, 0, len(m))
			for k := range m {
				if r, _ := utf8.DecodeRuneInString(k); unicode.IsUpper(r) {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			return key + "\x00" + t.String() + "\x00" + strings.Join(keys, "\x01")
		}
	}

	return key + "\x00" + t.String()
}

// The first call for a given key parses and stores it; subsequent calls
// return the cached value. Concurrent first-renders of the same template
// may both parse, but LoadOrStore ensures only one result is shared.
func parsedTemplate(text *Text) (*template.Template, error) {
	// Key on the raw, unpatched template text plus the trust level and context
	// type so that cache hits can skip patchTemplate entirely: patching is only
	// needed the first time a given (raw template, trust level, context type)
	// combination is seen.
	key := templateCacheKey(text.template, text.trusted, text.context)

	if cached, ok := parsedTemplates.Load(key); ok {
		return cached.(*template.Template), nil
	}

	// Cache miss: patch the raw template into its executable form.
	text.patchTemplate()

	// Parse into a fresh template with the func map matching this render's trust
	// level. missingkey=zero so a missing map key yields a typed nil that
	// escapeActionValue can render as <no value> instead of erroring on an
	// invalid reflect.Value — same visible output as the default option.
	tmpl, err := template.New("cache").Funcs(funcMap(text.trusted)).Option("missingkey=zero").Parse(text.template)
	if err != nil {
		return nil, err
	}

	escapePrintActions(tmpl)

	// Store; if another goroutine already stored an equivalent template, use theirs.
	actual, _ := parsedTemplates.LoadOrStore(key, tmpl)
	return actual.(*template.Template), nil
}

// escapePrintActions appends escapeActionValue to every print action's
// pipeline, so interpolated values get their chevrons neutralized (unless the
// value is Markup) while literal template text keeps its anchors. This is the
// boundary between user-authored markup and attacker-influenceable data
// (branch names, manifest fields, folder names): without it both land in the
// writer as indistinguishable text and data can forge anchors.
func escapePrintActions(tmpl *template.Template) {
	for _, t := range tmpl.Templates() {
		if t.Tree == nil {
			continue
		}

		escapeListActions(t.Tree, t.Root)
	}
}

func escapeListActions(tree *parse.Tree, list *parse.ListNode) {
	if list == nil {
		return
	}

	for _, node := range list.Nodes {
		switch n := node.(type) {
		case *parse.ActionNode:
			escapePipe(tree, n.Pipe)
		case *parse.IfNode:
			escapeBranchActions(tree, &n.BranchNode)
		case *parse.RangeNode:
			escapeBranchActions(tree, &n.BranchNode)
		case *parse.WithNode:
			escapeBranchActions(tree, &n.BranchNode)
		}
	}
}

func escapeBranchActions(tree *parse.Tree, branch *parse.BranchNode) {
	escapeListActions(tree, branch.List)
	escapeListActions(tree, branch.ElseList)
}

func escapePipe(tree *parse.Tree, pipe *parse.PipeNode) {
	// assignments and declarations produce no output
	if pipe == nil || len(pipe.Decl) != 0 || len(pipe.Cmds) == 0 {
		return
	}

	pos := pipe.Cmds[len(pipe.Cmds)-1].Position()
	ident := parse.NewIdentifier(escapeFuncName).SetTree(tree).SetPos(pos)
	cmd := &parse.CommandNode{
		NodeType: parse.NodeCommand,
		Pos:      pos,
		Args:     []parse.Node{ident},
	}
	pipe.Cmds = append(pipe.Cmds, cmd)
}

func (t *renderer) execute(text *Text) (string, error) {
	tmpl, err := parsedTemplate(text)
	if err != nil {
		log.Error(err)
		return "", errors.New(InvalidTemplate)
	}

	t.context.init(text)

	err = tmpl.Execute(&t.buffer, t.context)
	if err != nil {
		log.Error(err)
		return "", errors.New(IncorrectTemplate)
	}

	output := t.buffer.String()

	// issue with missingkey=zero ignored for map[string]any
	// https://github.com/golang/go/issues/24963
	output = strings.ReplaceAll(output, noValue, "")

	return output, nil
}
