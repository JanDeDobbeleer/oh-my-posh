package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNuFeatures(t *testing.T) {
	got := allFeatures.Lines(NU).String("// these are the features")

	want := `// these are the features
$env.TRANSIENT_PROMPT_COMMAND = {|| _omp_get_prompt transient }
^$_omp_executable upgrade
^$_omp_executable notice`

	assert.Equal(t, want, got)
}

func TestQuoteNuStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `"/tmp/\"omp's dir\"/oh-my-posh"`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `"C:/tmp\\omp's dir/oh-my-posh.exe"`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteNuStr(tc.str), fmt.Sprintf("quoteNuStr: %s", tc.str))
	}
}
