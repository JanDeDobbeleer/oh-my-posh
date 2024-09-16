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
"$_omp_executable" upgrade
"$_omp_executable" notice
_omp_cursor_positioning=1`

	assert.Equal(t, want, got)
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
