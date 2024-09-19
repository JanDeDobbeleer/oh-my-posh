package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdFeatures(t *testing.T) {
	got := allFeatures.Lines(CMD).String("// these are the features")

	want := `// these are the features
enable_tooltips()
transient_enabled = true
ftcs_marks_enabled = true
os.execute(string.format('"%s" upgrade', omp_executable))
os.execute(string.format('"%s" notice', omp_executable))
rprompt_enabled = true`

	assert.Equal(t, want, got)
}

func TestEscapeLuaStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: ""},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `/tmp/\"omp\'s dir\"/oh-my-posh`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `C:/tmp\\omp\'s dir/oh-my-posh.exe`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, escapeLuaStr(tc.str), fmt.Sprintf("escapeLuaStr: %s", tc.str))
	}
}
