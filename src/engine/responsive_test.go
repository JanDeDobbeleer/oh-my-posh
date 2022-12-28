package engine

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"

	"github.com/stretchr/testify/assert"
)

func TestShouldHideForWidth(t *testing.T) {
	cases := []struct {
		Case     string
		MinWidth int
		MaxWidth int
		Width    int
		Error    error
		Expected bool
	}{
		{Case: "No settings"},
		{Case: "Min cols - hide", MinWidth: 10, Width: 9, Expected: true},
		{Case: "Min cols - show", MinWidth: 10, Width: 20, Expected: false},
		{Case: "Max cols - hide", MaxWidth: 10, Width: 11, Expected: true},
		{Case: "Max cols - show", MaxWidth: 10, Width: 8, Expected: false},
		{Case: "Min & Max cols - hide", MinWidth: 10, MaxWidth: 20, Width: 21, Expected: true},
		{Case: "Min & Max cols - hide 2", MinWidth: 10, MaxWidth: 20, Width: 8, Expected: true},
		{Case: "Min & Max cols - show", MinWidth: 10, MaxWidth: 20, Width: 11, Expected: false},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TerminalWidth").Return(tc.Width, tc.Error)
		got := shouldHideForWidth(env, tc.MinWidth, tc.MaxWidth)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
