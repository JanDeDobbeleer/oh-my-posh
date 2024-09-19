package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFishFeatures(t *testing.T) {
	got := allFeatures.Lines(FISH).String("// these are the features")

	want := `// these are the features
enable_poshtooltips
set --global _omp_transient_prompt 1
set --global _omp_ftcs_marks 1
"$_omp_executable" upgrade
"$_omp_executable" notice
set --global _omp_prompt_mark 1`

	assert.Equal(t, want, got)
}

func TestQuoteFishStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `'/tmp/"omp\'s dir"/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `'C:/tmp\\omp\'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteFishStr(tc.str), fmt.Sprintf("quoteFishStr: %s", tc.str))
	}
}
