package prompt

import (
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
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
