package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXonshFeatures(t *testing.T) {
	got := allFeatures.Lines(XONSH).String("// these are the features")

	want := `// these are the features
@(_omp_executable) upgrade
@(_omp_executable) notice`

	assert.Equal(t, want, got)
}

func TestQuotePythonStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `'/tmp/"omp'"'"'s dir"/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `'C:/tmp\\omp'"'"'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quotePythonStr(tc.str), fmt.Sprintf("quotePythonStr: %s", tc.str))
	}
}
