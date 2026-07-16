package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Overlay mutates a base config map to enable a specific feature. Overlays are combined
// by applying them in order to the same base map.
type Overlay func(cfg map[string]any)

// BaseConfig returns a fresh, deterministic v4 config as a map[string]any: no git/time/
// network segments, a single left-aligned prompt block whose template renders the last
// exit code as "E2E:<code>>". Callers may apply Overlay functions to enable additional
// features before marshaling with WriteConfig.
func BaseConfig() map[string]any {
	return map[string]any{
		"version":     4,
		"final_space": true,
		"upgrade": map[string]any{
			"notice": false,
			"auto":   false,
		},
		"blocks": []any{
			map[string]any{
				"type":      "prompt",
				"alignment": "left",
				"segments": []any{
					map[string]any{
						"type":     "text",
						"style":    "plain",
						"template": "E2E:{{ .Code }}>",
					},
				},
			},
		},
	}
}

// Transient enables the transient prompt feature. It replaces the primary prompt with
// "TR> " once a command has been accepted.
func Transient(cfg map[string]any) {
	cfg["transient_prompt"] = map[string]any{
		"type":     "text",
		"style":    "plain",
		"template": "TR> ",
	}
}

// RPrompt appends a right-aligned prompt block rendering the fixed marker "RMARK".
func RPrompt(cfg map[string]any) {
	blocks, _ := cfg["blocks"].([]any)

	blocks = append(blocks, map[string]any{
		"type": "rprompt",
		"segments": []any{
			map[string]any{
				"type":     "text",
				"style":    "plain",
				"template": "RMARK",
			},
		},
	})

	cfg["blocks"] = blocks
}

// Tooltips enables a single tooltip segment rendering the fixed marker "TIP" for the
// "git" tip word.
func Tooltips(cfg map[string]any) {
	cfg["tooltips"] = []any{
		map[string]any{
			"type":     "text",
			"style":    "plain",
			"template": "TIP",
			"tips":     []any{"git"},
		},
	}
}

// Full combines Transient, RPrompt and Tooltips.
func Full(cfg map[string]any) {
	Transient(cfg)
	RPrompt(cfg)
	Tooltips(cfg)
}

// Colored appends a second text segment to the primary prompt block, rendering the
// fixed marker "CLR" in foreground "#ff0000" on background "#0000ff", for asserting
// rendered screen-cell colors.
func Colored(cfg map[string]any) {
	blocks, _ := cfg["blocks"].([]any)
	block, _ := blocks[0].(map[string]any)
	segments, _ := block["segments"].([]any)

	segments = append(segments, map[string]any{
		"type":       "text",
		"style":      "plain",
		"template":   "CLR",
		"foreground": "#ff0000",
		"background": "#0000ff",
	})

	block["segments"] = segments
}

// ShellIntegration enables shell_integration at the config root, which turns on FTCS
// (Final Term Control Sequence) marks around prompt rendering and command execution.
func ShellIntegration(cfg map[string]any) {
	cfg["shell_integration"] = true
}

// WriteConfig builds a config from BaseConfig with the given overlays applied, in order,
// marshals it to JSON, writes it to a file in t.TempDir(), and returns the absolute path.
func WriteConfig(t *testing.T, overlays ...Overlay) string {
	t.Helper()

	cfg := BaseConfig()
	for _, overlay := range overlays {
		overlay(cfg)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshaling config: %v", err)
	}

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	return path
}
