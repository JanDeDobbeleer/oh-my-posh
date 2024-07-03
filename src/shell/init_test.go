package shell

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestConsoleBackgroundColorTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Term     string
	}{
		{Case: "Inside vscode", Expected: "#123456", Term: "vscode"},
		{Case: "Outside vscode", Expected: "", Term: "windowsterminal"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("TemplateCache").Return(&cache.Template{
			Env: map[string]string{
				"TERM_PROGRAM": tc.Term,
			},
		})
		bgColor := ConsoleBackgroundColor(env, "{{ if eq \"vscode\" .Env.TERM_PROGRAM }}#123456{{end}}")
		assert.Equal(t, color.Ansi(tc.Expected), bgColor, tc.Case)
	}
}

func TestQuotePwshStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: ``, expected: `''`},
		{str: `/tmp/oh-my-posh`, expected: `'/tmp/oh-my-posh'`},
		{str: `/tmp/omp's dir/oh-my-posh`, expected: `'/tmp/omp''s dir/oh-my-posh'`},
		{str: `C:\tmp\oh-my-posh.exe`, expected: `'C:\tmp\oh-my-posh.exe'`},
		{str: `C:\tmp\omp's dir\oh-my-posh.exe`, expected: `'C:\tmp\omp''s dir\oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quotePwshStr(tc.str), fmt.Sprintf("quotePwshStr: %s", tc.str))
	}
}

func TestQuotePosixStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: ``, expected: `''`},
		{str: `/tmp/oh-my-posh`, expected: `/tmp/oh-my-posh`},
		{str: `/tmp/omp's dir/oh-my-posh`, expected: `$'/tmp/omp\'s dir/oh-my-posh'`},
		{str: `C:/tmp/oh-my-posh.exe`, expected: `C:/tmp/oh-my-posh.exe`},
		{str: `C:/tmp/omp's dir/oh-my-posh.exe`, expected: `$'C:/tmp/omp\'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quotePosixStr(tc.str), fmt.Sprintf("quotePosixStr: %s", tc.str))
	}
}

func TestQuoteFishStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: ``, expected: `''`},
		{str: `/tmp/oh-my-posh`, expected: `/tmp/oh-my-posh`},
		{str: `/tmp/omp's dir/oh-my-posh`, expected: `'/tmp/omp\'s dir/oh-my-posh'`},
		{str: `C:/tmp/oh-my-posh.exe`, expected: `C:/tmp/oh-my-posh.exe`},
		{str: `C:/tmp/omp's dir/oh-my-posh.exe`, expected: `'C:/tmp/omp\'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteFishStr(tc.str), fmt.Sprintf("quoteFishStr: %s", tc.str))
	}
}

func TestQuoteLuaStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: ``, expected: `''`},
		{str: `/tmp/oh-my-posh`, expected: `'/tmp/oh-my-posh'`},
		{str: `/tmp/omp's dir/oh-my-posh`, expected: `'/tmp/omp\'s dir/oh-my-posh'`},
		{str: `C:/tmp/oh-my-posh.exe`, expected: `'C:/tmp/oh-my-posh.exe'`},
		{str: `C:/tmp/omp's dir/oh-my-posh.exe`, expected: `'C:/tmp/omp\'s dir/oh-my-posh.exe'`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteLuaStr(tc.str), fmt.Sprintf("quoteLuaStr: %s", tc.str))
	}
}

func TestQuoteNuStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: ``, expected: `''`},
		{str: `/tmp/oh-my-posh`, expected: `"/tmp/oh-my-posh"`},
		{str: `/tmp/omp's dir/oh-my-posh`, expected: `"/tmp/omp's dir/oh-my-posh"`},
		{str: `C:/tmp/oh-my-posh.exe`, expected: `"C:/tmp/oh-my-posh.exe"`},
		{str: `C:/tmp/omp's dir/oh-my-posh.exe`, expected: `"C:/tmp/omp's dir/oh-my-posh.exe"`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteNuStr(tc.str), fmt.Sprintf("quoteNuStr: %s", tc.str))
	}
}
