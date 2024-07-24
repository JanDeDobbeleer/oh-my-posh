package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElvishFeatures(t *testing.T) {
	got := allFeatures.Lines(ELVISH).String("// these are the features")

	want := `// these are the features
$_omp_executable upgrade
$_omp_executable notice`

	assert.Equal(t, want, got)
}
