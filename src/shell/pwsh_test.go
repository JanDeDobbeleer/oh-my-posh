package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var allFeatures = Features{Tooltips, LineError, Transient, Jobs, Azure, PoshGit, FTCSMarks, Upgrade, Notice, PromptMark, RPrompt, CursorPositioning}

func TestPwshFeatures(t *testing.T) {
	got := allFeatures.Lines(PWSH).String("")

	want := `
Enable-PoshTooltips
Enable-PoshLineError
Enable-PoshTransientPrompt
$global:_ompJobCount = $true
$global:_ompAzure = $true
$global:_ompPoshGit = $true
$global:_ompFTCSMarks = $true
& $global:_ompExecutable upgrade
& $global:_ompExecutable notice`

	assert.Equal(t, want, got)
}
