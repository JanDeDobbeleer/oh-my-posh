package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockColors is a simple String implementation for testing TmuxColors.
type mockColors struct {
	result Ansi
}

func (m *mockColors) ToAnsi(_ Ansi, _ bool) Ansi { return m.result }
func (m *mockColors) Resolve(c Ansi) (Ansi, error) { return c, nil }

func TestConvertAnsiToTmux(t *testing.T) {
	tests := []struct {
		name         string
		code         Ansi
		isBackground bool
		expected     Ansi
	}{
		// True color foreground
		{
			name:         "true color fg red",
			code:         "38;2;255;0;0",
			isBackground: false,
			expected:     "fg=#ff0000",
		},
		{
			name:         "true color fg white",
			code:         "38;2;255;255;255",
			isBackground: false,
			expected:     "fg=#ffffff",
		},
		// True color background
		{
			name:         "true color bg blue",
			code:         "48;2;0;0;255",
			isBackground: true,
			expected:     "bg=#0000ff",
		},
		{
			name:         "true color bg mixed",
			code:         "48;2;18;52;86",
			isBackground: true,
			expected:     "bg=#123456",
		},
		// 256-color foreground
		{
			name:         "256-color fg",
			code:         "38;5;42",
			isBackground: false,
			expected:     "fg=colour42",
		},
		// 256-color background
		{
			name:         "256-color bg",
			code:         "48;5;200",
			isBackground: true,
			expected:     "bg=colour200",
		},
		// Named colors (foreground)
		{
			name:         "named black fg",
			code:         "30",
			isBackground: false,
			expected:     "fg=black",
		},
		{
			name:         "named red fg",
			code:         "31",
			isBackground: false,
			expected:     "fg=red",
		},
		{
			name:         "named white fg",
			code:         "37",
			isBackground: false,
			expected:     "fg=white",
		},
		// Named colors (background)
		{
			name:         "named black bg",
			code:         "40",
			isBackground: true,
			expected:     "bg=black",
		},
		{
			name:         "named green bg",
			code:         "42",
			isBackground: true,
			expected:     "bg=green",
		},
		// Bright/high-intensity named colors
		{
			name:         "darkGray fg",
			code:         "90",
			isBackground: false,
			expected:     "fg=darkGray",
		},
		{
			name:         "lightBlue bg",
			code:         "104",
			isBackground: true,
			expected:     "bg=lightBlue",
		},
		// Special values pass through
		{
			name:         "empty color",
			code:         emptyColor,
			isBackground: false,
			expected:     emptyColor,
		},
		{
			name:         "transparent",
			code:         Transparent,
			isBackground: false,
			expected:     Transparent,
		},
		// Unknown code returns empty
		{
			name:         "unknown code",
			code:         "99",
			isBackground: false,
			expected:     emptyColor,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertAnsiToTmux(tc.code, tc.isBackground)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTmuxColorsToAnsi(t *testing.T) {
	tests := []struct {
		name         string
		innerResult  Ansi
		isBackground bool
		expected     Ansi
	}{
		{
			name:         "hex fg via inner",
			innerResult:  "38;2;255;128;0",
			isBackground: false,
			expected:     "fg=#ff8000",
		},
		{
			name:         "hex bg via inner",
			innerResult:  "48;2;0;128;255",
			isBackground: true,
			expected:     "bg=#0080ff",
		},
		{
			name:         "256-color via inner",
			innerResult:  "38;5;100",
			isBackground: false,
			expected:     "fg=colour100",
		},
		{
			name:         "named color via inner",
			innerResult:  "32",
			isBackground: false,
			expected:     "fg=green",
		},
		{
			name:         "transparent passes through",
			innerResult:  Transparent,
			isBackground: false,
			expected:     Transparent,
		},
		{
			name:         "empty passes through",
			innerResult:  emptyColor,
			isBackground: false,
			expected:     emptyColor,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockColors{result: tc.innerResult}
			tc2 := NewTmuxColors(mock)
			result := tc2.ToAnsi("anycolor", tc.isBackground)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTmuxColorsResolve(t *testing.T) {
	mock := &mockColors{}
	tc := NewTmuxColors(mock)
	result, err := tc.Resolve("p:mycolor")
	assert.NoError(t, err)
	assert.Equal(t, Ansi("p:mycolor"), result)
}

func TestTmuxColorsWithDefaults(t *testing.T) {
	// Integration test using a real Defaults instance with TrueColor enabled
	saved := TrueColor
	TrueColor = true
	t.Cleanup(func() { TrueColor = saved })

	defaults := &Defaults{}
	tc := NewTmuxColors(defaults)

	// Hex color → tmux format
	fg := tc.ToAnsi("#ff0000", false)
	assert.Equal(t, Ansi("fg=#ff0000"), fg)

	bg := tc.ToAnsi("#0000ff", true)
	assert.Equal(t, Ansi("bg=#0000ff"), bg)

	// Named color → tmux format
	fgRed := tc.ToAnsi("red", false)
	assert.Equal(t, Ansi("fg=red"), fgRed)

	bgBlue := tc.ToAnsi("blue", true)
	assert.Equal(t, Ansi("bg=blue"), bgBlue)

	// 256-color → tmux format
	fg256 := tc.ToAnsi("42", false)
	assert.Equal(t, Ansi("fg=colour42"), fg256)

	// Transparent passes through
	fgTransparent := tc.ToAnsi(Transparent, false)
	assert.Equal(t, Transparent, fgTransparent)
}
