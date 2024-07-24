package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdFeatures(t *testing.T) {
	got := allFeatures.Lines(CMD).String("// these are the features")

	want := `// these are the features
enable_tooltips()
transient_enabled = true
os.execute(string.format('%s upgrade', omp_exe()))
os.execute(string.format('%s notice', omp_exe()))
rprompt_enabled = true`

	assert.Equal(t, want, got)
}
