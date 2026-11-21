package prompt

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// viModeEnv keeps the vi mode between Setenv and Getenv like the real environment does.
type viModeEnv struct {
	*mock.Environment
	flags *runtime.Flags
	mode  string
}

func (e *viModeEnv) Flags() *runtime.Flags { return e.flags }

func (e *viModeEnv) Getenv(key string) string {
	if key == segments.VIModeEnv {
		return e.mode
	}

	return e.Environment.Getenv(key)
}

func (e *viModeEnv) Setenv(key, value string) {
	if key == segments.VIModeEnv {
		e.mode = value
	}
}

func TestPrimaryVariants(t *testing.T) {
	cases := []struct {
		Case      string
		Common    string
		Previous  string
		Expected  map[string]string
		Absent    map[string]string
		Segments  []*config.Segment
		Modes     []string
		FromCache bool
	}{
		{
			Case:     "one prompt per mode",
			Modes:    []string{"viins", "vicmd"},
			Segments: []*config.Segment{{Type: config.VIMODE, Template: "{{ .Mode }}"}},
			Expected: map[string]string{"viins": "insert", "vicmd": "normal"},
		},
		{
			Case:      "rendered from the executed segments",
			Modes:     []string{"viins", "vicmd"},
			FromCache: true,
			Segments:  []*config.Segment{{Type: config.VIMODE, Template: "{{ .Mode }}"}},
			Expected:  map[string]string{"viins": "insert", "vicmd": "normal"},
		},
		{
			Case:     "other segments are the same in every mode",
			Modes:    []string{"viins", "vicmd"},
			Segments: []*config.Segment{{Type: config.TEXT, Template: "static"}, {Type: config.VIMODE, Template: " {{ .Mode }}"}},
			Common:   "static",
			Expected: map[string]string{"viins": "insert", "vicmd": "normal"},
		},
		{
			Case:  "style that depends on the mode",
			Modes: []string{"viins", "vicmd"},
			Segments: []*config.Segment{{
				Type:           config.VIMODE,
				Template:       "{{ .Mode }}",
				Style:          `{{ if eq .Mode "normal" }}diamond{{ else }}plain{{ end }}`,
				LeadingDiamond: "<",
			}},
			Expected: map[string]string{"viins": "insert", "vicmd": "<"},
			Absent:   map[string]string{"viins": "<"},
		},
		{
			Case:     "no vimode segment",
			Modes:    []string{"viins", "vicmd"},
			Segments: []*config.Segment{{Type: config.TEXT, Template: "static"}},
			Common:   "static",
			Expected: map[string]string{"viins": "static", "vicmd": "static"},
		},
		{
			Case:     "mode set before is restored",
			Modes:    []string{"viins", "vicmd"},
			Previous: "emacs",
			Segments: []*config.Segment{{Type: config.VIMODE, Template: "{{ .Mode }}"}},
			Expected: map[string]string{"viins": "insert", "vicmd": "normal"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			base := setupStreamingTestEnv()
			env := &viModeEnv{Environment: base, flags: &runtime.Flags{VIModes: tc.Modes}, mode: tc.Previous}

			for _, segment := range tc.Segments {
				if segment.Style == "" {
					segment.Style = config.Plain
				}
			}

			engine := &Engine{
				Config: &config.Config{
					Blocks: []*config.Block{{Type: config.Prompt, Alignment: config.Left, Segments: tc.Segments}},
				},
				Env: env,
			}

			if tc.FromCache {
				// the segments have been executed by an earlier pass, like in a streaming cycle
				engine.Primary()
			}

			render := engine.Primary
			if tc.FromCache {
				render = engine.renderFromBlocks
			}

			var prompts map[string]string
			require.NoError(t, json.Unmarshal([]byte(render()), &prompts))

			require.Len(t, prompts, len(tc.Expected))
			for mode, want := range tc.Expected {
				assert.Contains(t, prompts[mode], want, mode)
				assert.Contains(t, prompts[mode], tc.Common, mode)
			}

			for mode, absent := range tc.Absent {
				assert.NotContains(t, prompts[mode], absent, mode)
			}

			assert.Equal(t, tc.Previous, env.mode)
		})
	}
}
