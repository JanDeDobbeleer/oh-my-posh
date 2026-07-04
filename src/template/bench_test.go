package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
)

func setupTemplateBench() {
	env := new(mock.Environment)
	env.On("Shell").Return("pwsh")
	Cache = new(cache.Template)
	Init(env, nil, nil)
}

// BenchmarkRenderPlain benchmarks the fast-path: no {{ }} in the string.
func BenchmarkRenderPlain(b *testing.B) {
	setupTemplateBench()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Render("plain text without any template markers at all", nil)
	}
}

// BenchmarkRenderSimple benchmarks a minimal two-field template.
func BenchmarkRenderSimple(b *testing.B) {
	setupTemplateBench()
	type ctx struct {
		Shell    string
		UserName string
	}
	data := ctx{Shell: "pwsh", UserName: "jandedobbeleer"}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Render("{{ .Shell }} {{ .UserName }}", data)
	}
}

// BenchmarkRenderRepeated benchmarks repeated renders of an identical ~200-char
// template with conditionals and a function call — models per-folder path segments.
func BenchmarkRenderRepeated(b *testing.B) {
	setupTemplateBench()
	type ctx struct {
		Folder  string
		Branch  string
		Version string
		Root    bool
	}
	data := ctx{
		Folder:  "oh-my-posh",
		Branch:  "perf/rendering-engine",
		Root:    false,
		Version: "23.4.1",
	}
	tmpl := `{{ if .Root }}# {{ end }}{{ .Folder }}{{ if .Branch }} on {{ .Branch }}{{ end }} {{ if .Version }}v{{ .Version }}{{ end }}`
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Render(tmpl, data)
	}
}

// BenchmarkPatchTemplate benchmarks patchTemplate alone on a ~300-char template
// that exercises property rewriting (Env vars, struct fields, Segments).
func BenchmarkPatchTemplate(b *testing.B) {
	setupTemplateBench()
	context := map[string]any{
		"Folder":   true,
		"Branch":   true,
		"Version":  true,
		"Status":   true,
		"UserName": true,
		"HostName": true,
		"OS":       true,
		"Working":  true,
		"Staging":  true,
	}
	tmpl := `{{ if .Root }}# {{ end }}{{ .Folder }} {{ .Branch }} {{ .Env.HOME }}` +
		` {{ if or (.Working.Changed) (.Staging.Changed) }}*{{ end }}` +
		` {{ .UserName }}@{{ .HostName }} {{ .OS }} {{ .Version }} {{ .Status }}`
	b.ReportAllocs()
	for b.Loop() {
		t := &Text{
			template: tmpl,
			context:  context,
		}
		t.patchTemplate()
	}
}
