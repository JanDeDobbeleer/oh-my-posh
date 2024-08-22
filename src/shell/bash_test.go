package shell

import (
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
