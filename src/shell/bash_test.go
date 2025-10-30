package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBashFeatures(t *testing.T) {
	got := allFeatures.Lines(BASH).String("// these are the features")

	want := `// these are the features
_omp_ftcs_marks=1
"$_omp_executable" upgrade --auto
"$_omp_executable" notice
_omp_cursor_positioning=1`

	assert.Equal(t, want, got)
}

func TestBashFeaturesWithBLE(t *testing.T) {
	bashBLEsession = true

	got := allFeatures.Lines(BASH).String("// these are the features")

	want := `// these are the features
bleopt prompt_ps1_transient=always
bleopt prompt_ps1_final='$(
    "$_omp_executable" print transient \
        --shell=bash \
        --shell-version="$BASH_VERSION" \
        --escape=false
)'
_omp_ftcs_marks=1
"$_omp_executable" upgrade --auto
"$_omp_executable" notice
bleopt prompt_rps1='$(
	"$_omp_executable" print right \
		--save-cache \
		--shell=bash \
		--shell-version="$BASH_VERSION" \
		--status="$_omp_status" \
		--pipestatus="${_omp_pipestatus[*]}" \
		--no-status="$_omp_no_status" \
		--execution-time="$_omp_execution_time" \
		--stack-count="$_omp_stack_count" \
		--terminal-width="${COLUMNS-0}" \
		--escape=false
)'
_omp_cursor_positioning=1`

	assert.Equal(t, want, got)

	bashBLEsession = false
}

func TestQuotePosixStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `$'/tmp/"omp\'s dir"/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `$'C:/tmp\\omp\'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, QuotePosixStr(tc.str), fmt.Sprintf("QuotePosixStr: %s", tc.str))
	}
}
