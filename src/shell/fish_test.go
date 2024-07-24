package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFishFeatures(t *testing.T) {
	got := allFeatures.Lines(FISH).String("// these are the features")

	want := `// these are the features
enable_poshtooltips
set --global _omp_transient_prompt 1
set --global _omp_ftcs_marks 1
$_omp_executable upgrade
$_omp_executable notice
set --global _omp_prompt_mark 1`

	assert.Equal(t, want, got)
}
