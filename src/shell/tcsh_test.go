package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcshFeatures(t *testing.T) {
	got := allFeatures.Lines(TCSH).String("// these are the features")

	want := `// these are the features
"$_omp_executable" upgrade;
"$_omp_executable" notice;`

	assert.Equal(t, want, got)
}

func TestQuoteCshStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"!/oh-my-posh`, expected: `'/tmp/"omp'"'"'s dir"\!/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `'C:/tmp\omp'"'"'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteCshStr(tc.str), fmt.Sprintf("quoteCshStr: %s", tc.str))
	}
}
