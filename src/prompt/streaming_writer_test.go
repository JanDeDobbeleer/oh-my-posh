package prompt

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// streamTestType is a test-only segment type that drives the streaming machinery through a
// controllable fake writer instead of a real segment's Enabled() implementation.
const streamTestType config.SegmentType = "streamtest"

// streamTestControl steers one fake writer instance, looked up by the segment's "id" option.
type streamTestControl struct {
	release chan struct{} // Enabled blocks on this until closed; nil means no wait.
	delay   time.Duration // Sleep after release, before writing fields.
	panics  bool          // Panic instead of returning, after release and delay.
}

// streamTestControls maps a segment's "id" option to the control steering its writer instance.
// It has to be a package-level map: writer instances are constructed by the config package's
// registry (config.Segments), which has no way to thread per-test state through a constructor.
var streamTestControls sync.Map

// streamTestWriter is a fake SegmentWriter whose Enabled() can be told to block, sleep, or panic,
// so a test can drive the exact timing the streaming producer must handle correctly.
type streamTestWriter struct {
	segments.Base
	control *streamTestControl
	Ref     string
	Items   []string
	Count   int
}

func init() {
	// Test-only writer type; registered for the whole test binary, never removed.
	config.Segments[streamTestType] = func() config.SegmentWriter { return &streamTestWriter{} }
}

func (w *streamTestWriter) Init(props options.Provider, env runtime.Environment) {
	w.Base.Init(props, env)

	control, ok := streamTestControls.Load(props.String("id", ""))
	if !ok {
		return
	}

	w.control = control.(*streamTestControl)
}

func (w *streamTestWriter) Template() string {
	return "{{ .Ref }}/{{ .Count }}"
}

func (w *streamTestWriter) Enabled() bool {
	if w.control != nil {
		if w.control.release != nil {
			<-w.control.release
		}

		time.Sleep(w.control.delay)

		if w.control.panics {
			panic("streamtest: enabled panicked")
		}
	}

	// Multi-word fields on purpose: a string and a slice are what a concurrent reader would
	// tear, so these writes are what the race detector must see against a pending render.
	w.Ref = "resolved"
	w.Items = []string{"a", "b", "c"}
	w.Count = 42

	return true
}

// streamTestSegment builds a config.Segment wired to the streamTestType writer identified by id.
// Its two foreground templates read writer fields the fake writer writes late (Ref, then Items),
// so a render that touches the writer while the segment is pending is exactly what the race
// detector needs to see.
func streamTestSegment(id string) *config.Segment {
	return &config.Segment{
		Type:        streamTestType,
		Options:     options.Map{"id": id},
		Style:       config.Powerline,
		Placeholder: "PENDING-" + id,
		Foreground:  "#ffffff",
		Background:  "#000000",
		ForegroundTemplates: []string{
			"{{ if .Ref }}#00ff00{{ end }}",
			"{{ if gt (len .Items) 0 }}#ff0000{{ end }}",
		},
	}
}

// TestStreamPrimary_StreamingEnvTimeout drives the timeout the engine applies to pending
// segments through newEngine: POSH_STREAMING has to enable streaming for themes without a
// streaming property, while an explicit config value wins and a non-streaming render
// ignores the environment entirely.
func TestStreamPrimary_StreamingEnvTimeout(t *testing.T) {
	cases := []struct {
		Case            string
		EnvStreaming    string
		Delay           time.Duration
		ConfigStreaming int
		Streaming       bool
		Placeholder     bool
	}{
		{
			Case:         "environment enables streaming",
			Streaming:    true,
			EnvStreaming: "20",
			Delay:        200 * time.Millisecond,
			Placeholder:  true,
		},
		{
			Case:            "config value wins over the environment",
			Streaming:       true,
			ConfigStreaming: 20,
			EnvStreaming:    "500",
			Delay:           200 * time.Millisecond,
			Placeholder:     true,
		},
		{
			Case:        "unset environment leaves the segment without a timeout",
			Streaming:   true,
			Delay:       50 * time.Millisecond,
			Placeholder: false,
		},
		{
			Case:         "non-streaming render ignores the environment",
			Streaming:    false,
			EnvStreaming: "20",
			Delay:        50 * time.Millisecond,
			Placeholder:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			t.Setenv("POSH_STREAMING", tc.EnvStreaming)

			const id = "envtimeout"
			streamTestControls.Store(id, &streamTestControl{delay: tc.Delay})
			t.Cleanup(func() { streamTestControls.Delete(id) })

			flags := &runtime.Flags{
				Shell:     shell.GENERIC,
				IsPrimary: true,
				Streaming: tc.Streaming,
			}

			env := &runtime.Terminal{}
			env.Init(flags)

			cfg := &config.Config{
				Streaming: tc.ConfigStreaming,
				Blocks: []*config.Block{
					{
						Type:      config.Prompt,
						Alignment: config.Left,
						Segments:  []*config.Segment{streamTestSegment(id)},
					},
				},
			}

			records := collectChannelOutput(newEngine(cfg, env).StreamPrimary(), 2*time.Second)

			var primary string
			for _, record := range records {
				if strings.HasPrefix(record, TransientMarker) {
					continue
				}
				primary = record
				break
			}

			require.NotEmpty(t, primary, "streamed no primary record")

			placeholder := "PENDING-" + id
			if tc.Placeholder {
				assert.Contains(t, primary, placeholder)
				return
			}

			assert.NotContains(t, primary, placeholder)
			assert.Contains(t, primary, "resolved")
		})
	}
}
