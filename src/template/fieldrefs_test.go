package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeFields(t *testing.T) {
	cases := []struct {
		ExpectedSegments       map[string][]string
		ExpectedSegmentsOpaque map[string]bool
		Case                   string
		Template               string
		ExpectedOwn            []string
		ExpectedOwnOpaque      bool
		ExpectedOpaque         bool
	}{
		{
			Case:     "no template syntax",
			Template: "plain text",
		},
		{
			Case:        "simple field",
			Template:    "{{ .HEAD }}",
			ExpectedOwn: []string{"HEAD"},
		},
		{
			Case:        "nested field records the head",
			Template:    "{{ .Working.Changed }}",
			ExpectedOwn: []string{"Working"},
		},
		{
			Case:        "multiple fields with conditions",
			Template:    "{{ .HEAD }}{{ if .BranchStatus }} {{ .BranchStatus }}{{ end }}{{ if gt .Ahead 0 }}up{{ end }}",
			ExpectedOwn: []string{"HEAD", "BranchStatus", "Ahead"},
		},
		{
			Case:        "with re-roots the dot",
			Template:    "{{ with .Working }}{{ .Changed }}{{ end }}",
			ExpectedOwn: []string{"Working"},
		},
		{
			Case:        "with body whole-dot stays grounded",
			Template:    "{{ with .Working }}{{ . }}{{ end }}",
			ExpectedOwn: []string{"Working"},
		},
		{
			Case:        "range over a field",
			Template:    "{{ range .Tags }}{{ . }}{{ end }}",
			ExpectedOwn: []string{"Tags"},
		},
		{
			Case:        "variable from a field is inert",
			Template:    "{{ $rel := .RelativeDir }}{{ if $rel }}{{ $rel }}{{ end }}",
			ExpectedOwn: []string{"RelativeDir"},
		},
		{
			Case:        "function arguments record their heads",
			Template:    "{{ url .UpstreamIcon .UpstreamURL }}",
			ExpectedOwn: []string{"UpstreamIcon", "UpstreamURL"},
		},
		{
			Case:        "env access via patching stays analyzable",
			Template:    "{{ .Env.MY-VAR }}{{ .HEAD }}",
			ExpectedOwn: []string{"HEAD", "Getenv"},
		},
		{
			Case:             "cross-segment field",
			Template:         "{{ .Segments.Git.Working.Changed }}",
			ExpectedSegments: map[string][]string{"Git": {"Working"}},
		},
		{
			Case:             "cross-segment with alias name",
			Template:         "{{ .Segments.scm.Ahead }}",
			ExpectedSegments: map[string][]string{"scm": {"Ahead"}},
		},
		{
			Case:             "cross-segment inside with",
			Template:         "{{ with .Segments.Git }}{{ .Ahead }}{{ .Behind }}{{ end }}",
			ExpectedSegments: map[string][]string{"Git": {"Ahead", "Behind"}},
		},
		{
			// the patcher rewrites this into the (.Segments.MustGet "Git")
			// chain the walker's segmentLookup recognizes
			Case:             "deep cross-segment method chain",
			Template:         "{{ .Segments.Git.Staging.String }}",
			ExpectedSegments: map[string][]string{"Git": {"Staging"}},
		},
		{
			Case:     "segments contains is a presence check",
			Template: `{{ if .Segments.Contains "Git" }}yes{{ end }}`,
		},
		{
			Case:              "whole-dot print",
			Template:          "{{ . }}",
			ExpectedOwnOpaque: true,
		},
		{
			Case:              "whole dot passed to a function",
			Template:          `{{ printf "%v" . }}`,
			ExpectedOwnOpaque: true,
		},
		{
			Case:              "index over the dot",
			Template:          `{{ index . "Working" }}`,
			ExpectedOwnOpaque: true,
		},
		{
			Case:              "variable laundering the dot",
			Template:          "{{ $x := . }}{{ $x.Working }}",
			ExpectedOwnOpaque: true,
		},
		{
			Case:              "range over the dot",
			Template:          "{{ range . }}{{ . }}{{ end }}",
			ExpectedOwnOpaque: true,
		},
		{
			Case:                   "whole segment printed",
			Template:               "{{ .Segments.Git }}",
			ExpectedSegmentsOpaque: map[string]bool{"Git": true},
		},
		{
			Case:                   "segment truthiness",
			Template:               "{{ if .Segments.Git }}repo{{ end }}",
			ExpectedSegmentsOpaque: map[string]bool{"Git": true},
		},
		{
			Case:           "template include",
			Template:       `{{ template "x" }}`,
			ExpectedOpaque: true,
		},
		{
			Case:           "parse failure",
			Template:       "{{ .Working",
			ExpectedOpaque: true,
		},
		{
			Case:           "bare segments map",
			Template:       "{{ .Segments }}",
			ExpectedOpaque: true,
		},
	}

	for _, tc := range cases {
		refs := AnalyzeFields(tc.Template)

		own := make(map[string]bool, len(tc.ExpectedOwn))
		for _, field := range tc.ExpectedOwn {
			own[field] = true
		}

		assert.Equal(t, own, refs.Own, tc.Case)

		segments := make(map[string]map[string]bool, len(tc.ExpectedSegments))
		for name, fields := range tc.ExpectedSegments {
			segments[name] = make(map[string]bool, len(fields))
			for _, field := range fields {
				segments[name][field] = true
			}
		}

		assert.Equal(t, segments, refs.Segments, tc.Case)

		segmentsOpaque := tc.ExpectedSegmentsOpaque
		if segmentsOpaque == nil {
			segmentsOpaque = map[string]bool{}
		}

		assert.Equal(t, segmentsOpaque, refs.SegmentsOpaque, tc.Case)
		assert.Equal(t, tc.ExpectedOwnOpaque, refs.OwnOpaque, tc.Case)
		assert.Equal(t, tc.ExpectedOpaque, refs.Opaque, tc.Case)
	}
}
