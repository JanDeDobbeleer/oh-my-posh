package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, tc.expected, QuotePosixStr(tc.str), fmt.Sprintf("quotePosixStr: %s", tc.str))
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
