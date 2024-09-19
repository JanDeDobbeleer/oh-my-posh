package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZshFeatures(t *testing.T) {
	got := allFeatures.Lines(ZSH).String("// these are the features")

	want := `// these are the features
enable_poshtooltips
_omp_create_widget zle-line-init _omp_zle-line-init
_omp_ftcs_marks=1
"$_omp_executable" upgrade
"$_omp_executable" notice
_omp_cursor_positioning=1`

	assert.Equal(t, want, got)
}
