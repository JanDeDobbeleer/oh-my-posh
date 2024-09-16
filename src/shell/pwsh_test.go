package shell

import (
	"fmt"
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

func TestQuotePwshOrElvishStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `'/tmp/"omp''s dir"/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `'C:/tmp\omp''s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quotePwshOrElvishStr(tc.str), fmt.Sprintf("quotePwshOrElvishStr: %s", tc.str))
	}
}
