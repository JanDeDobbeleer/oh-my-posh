package shell

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

var allFeatures = Tooltips | LineError | Transient | Jobs | Azure | PoshGit | FTCSMarks |
	Upgrade | Notice | PromptMark | RPrompt | CursorPositioning | KeyHandlers | Streaming | VIMode | TransientRPrompt

func TestPwshFeatures(t *testing.T) {
	got := allFeatures.Lines(PWSH).String("")

	want := `
$global:_ompJobCount = $true
$global:_ompAzure = $true
$global:_ompPoshGit = $true
Enable-PoshLineError
Enable-PoshTooltips
$global:_ompTransientPrompt = $true
$global:_ompFTCSMarks = $true
& $global:_ompExecutable upgrade --auto
& $global:_ompExecutable notice
Enable-PoshStreaming
Enable-KeyHandlers
Enable-PoshVIMode`

	assert.Equal(t, want, got)
}

func TestPwshStreamingGuard(t *testing.T) {
	assert.Contains(t, pwshInit, "if ($env:POSH_DISABLE_STREAMING -eq '1') {")
	assert.Contains(t, pwshInit, "return")
}

func TestSourceCommandAsyncPwsh(t *testing.T) {
	got := sourceCommandAsync(PWSH, "C:/cache/init.pwsh.ps1")

	want := "if (-not (Get-Variable -Name _ompOriginalPromptFunction -Scope Global -ErrorAction Ignore -ValueOnly)) { " +
		"$global:_ompOriginalPromptFunction = $Function:prompt; " +
		"$global:_ompPromptFunction = $null; " +
		"$global:_ompInitialized = $false; " +
		"function prompt() { if (-not $global:_ompInitialized) { $global:_ompAsyncInit = $true; & 'C:/cache/init.pwsh.ps1'; return }; " +
		"if ($global:_ompPromptFunction) { & $global:_ompPromptFunction } } " +
		"}"

	assert.Equal(t, want, got)
}

func TestQuotePwshStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `'/tmp/"omp''s dir"/oh-my-posh'`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `'C:/tmp\omp''s dir/oh-my-posh.exe'`},
		// non-ASCII runes are emitted as [char] expressions so the output
		// survives PowerShell's OEM code page decoding of native stdout
		{str: `C:/Users/MyNameE²/init.ps1`, expected: `"$('C:/Users/MyNameE' + [char]0xB2 + '/init.ps1')"`},
		{str: `C:/Users/Ø²/omp's.ps1`, expected: `"$('C:/Users/' + [char]0xD8 + [char]0xB2 + '/omp''s.ps1')"`},
		{str: `²init.ps1`, expected: `"$('' + [char]0xB2 + 'init.ps1')"`},
		{str: `C:/Users/📁/init.ps1`, expected: `"$('C:/Users/' + [char]::ConvertFromUtf32(0x1F4C1) + '/init.ps1')"`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quotePwshStr(tc.str), fmt.Sprintf("quotePwshStr: %s", tc.str))
	}
}

// PowerShell decodes native command output using [Console]::OutputEncoding,
// which defaults to the legacy OEM code page on Windows. Everything written
// to stdout for pwsh must therefore be pure ASCII, no matter which characters
// the injected paths contain.
func TestSessionScriptPwshIsPureASCII(t *testing.T) {
	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{Shell: PWSH, ConfigPath: "C:/Users/MyNameE²/omp.json"})

	got := sessionScript(env)

	assert.Contains(t, got, `$env:POSH_CONFIG = "$('C:/Users/MyNameE' + [char]0xB2 + '/omp.json')";`)

	for i, b := range []byte(got) {
		assert.Less(t, b, uint8(0x80), fmt.Sprintf("non-ASCII byte at index %d in: %s", i, got))
	}
}
