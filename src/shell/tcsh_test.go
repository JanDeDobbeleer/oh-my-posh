package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcshFeatures(t *testing.T) {
	got := allFeatures.Lines(TCSH).String("// these are the features")

	want := `// these are the features
$POSH_COMMAND upgrade;
$POSH_COMMAND notice;`

	assert.Equal(t, want, got)
}
