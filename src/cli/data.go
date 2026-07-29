package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// Shared between printCmd and imageCmd.
var (
	dataPath   string
	dataDerive bool
	dataOnly   bool
)

func init() {
	const deriveUsage = "derive live state for a recorded --data file too, instead of replaying it hermetically (debugging)"

	printCmd.Flags().BoolVar(&dataDerive, "data-derive", false, deriveUsage)
	imageCmd.Flags().BoolVar(&dataDerive, "data-derive", false, deriveUsage)

	const onlyUsage = "render segments exclusively from --data: a segment the file does not cover renders as absent " +
		"instead of probing the real machine"

	printCmd.Flags().BoolVar(&dataOnly, "data-only", false, onlyUsage)
	imageCmd.Flags().BoolVar(&dataOnly, "data-only", false, onlyUsage)
}

// No-op when dataPath is empty. The env/segment routing onto flags itself - with
// precedence explicit CLI flag > data file > live environment - lives in
// config.Data.ApplyFlags, shared with prompt/golden_test.go's fixture harness,
// which cannot import this package (cli imports prompt).
func applyDataFile(flags *runtime.Flags, changed func(name string) bool) error {
	if dataPath == "" {
		return nil
	}

	data, err := config.LoadData(dataPath)
	if err != nil {
		return err
	}

	if data.Version < config.DataVersion {
		fmt.Fprintf(os.Stderr,
			"warning: %s has no recorder version marker; treating it as hand-written - it is neither hermetic nor deterministic, "+
				"segments left out of it still probe the live environment. Run `oh-my-posh config export data` to record a deterministic file.\n",
			dataPath)
	}

	if dataDerive && data.Version >= config.DataVersion {
		data.Segments = deriveRecordedSegments(data.Segments)
	}

	// --data-only and --data-derive are opposites: one forbids probing
	// entirely, the other forces it even for a recorded file. Asking for both
	// is a contradiction rather than a combination, so say so instead of
	// silently letting one win.
	if dataOnly && dataDerive {
		return errors.New("--data-only and --data-derive contradict each other: one forbids probing the environment, the other forces it")
	}

	flags.DataOnly = dataOnly

	return data.ApplyFlags(flags, changed)
}

// deriveRecordedSegments strips the RecordedSegment envelope back down to each
// segment's raw writer data, forcing Segment.restoreData to treat a recorded
// file the same way it treats a hand-written one: derive live state via
// writer.Enabled(), then overlay the pinned data, instead of the hermetic no-probe
// path. The caller only calls this once data.Version confirms every entry is an
// envelope, so decoding here does not need to guard against a flat entry.
func deriveRecordedSegments(segments map[string]json.RawMessage) map[string]json.RawMessage {
	derived := make(map[string]json.RawMessage, len(segments))

	for key, raw := range segments {
		var recorded config.RecordedSegment
		if err := json.Unmarshal(raw, &recorded); err != nil {
			derived[key] = raw
			continue
		}

		derived[key] = recorded.Data
	}

	return derived
}
