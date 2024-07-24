package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXonshFeatures(t *testing.T) {
	got := allFeatures.Lines(XONSH).String("// these are the features")

	want := `// these are the features
@($POSH_EXECUTABLE) upgrade
@($POSH_EXECUTABLE) notice`

	assert.Equal(t, want, got)
}
