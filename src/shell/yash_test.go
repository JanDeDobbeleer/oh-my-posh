package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYashFeatures(t *testing.T) {
	got := allFeatures.Lines(YASH).String("// these are the features")

	want := `// these are the features
_omp_ftcs_marks=1
"$_omp_executable" upgrade --auto
"$_omp_executable" notice
YASH_PS1R='$(
    "$_omp_executable" print right \
        --save-cache \
        --shell=yash \
        --shell-version="$YASH_VERSION" \
        --status="$_omp_status" \
        --no-status="$_omp_no_status" \
        --execution-time="$_omp_execution_time" \
        --job-count="$_omp_job_count" \
        --terminal-width="${COLUMNS:-0}"
)'`

	assert.Equal(t, want, got)
}

func TestYashFeaturesUnsupported(t *testing.T) {
	unsupported := []Features{
		Transient,
		Tooltips,
		KeyHandlers,
		CursorPositioning,
		Async,
		Streaming,
		VIMode,
		LineError,
		Jobs,
		Azure,
		PoshGit,
		PromptMark,
	}

	for _, feature := range unsupported {
		assert.Empty(t, string(feature.Yash()), "expected no code for feature %v", feature)
	}
}
