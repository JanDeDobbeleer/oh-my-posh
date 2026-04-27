package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
)

func TestTooltipFallback_NoCacheReturnsEmpty(t *testing.T) {
	cache.Delete(cache.Session, RPromptKey)
	cache.Delete(cache.Session, RPromptLengthKey)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.ZSH)

	terminal.Init(shell.ZSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Empty(t, got)
}

func TestTooltipFallback_NoMatchReturnsRPromptText(t *testing.T) {
	cache.Delete(cache.Session, RPromptKey)
	cache.Delete(cache.Session, RPromptLengthKey)
	cache.Set(cache.Session, RPromptKey, "my-rprompt", cache.INFINITE)
	cache.Set(cache.Session, RPromptLengthKey, 10, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.ZSH)

	terminal.Init(shell.ZSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Equal(t, "my-rprompt", got)
}

func TestTooltipFallback_PwshNoMatchReturnsCursorPositionedRPrompt(t *testing.T) {
	cache.Delete(cache.Session, RPromptKey)
	cache.Delete(cache.Session, RPromptLengthKey)
	cache.Set(cache.Session, RPromptKey, "my-rprompt", cache.INFINITE)
	cache.Set(cache.Session, RPromptLengthKey, 10, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Column: 5})
	env.On("TerminalWidth").Return(200, nil)

	terminal.Init(shell.PWSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Contains(t, got, "my-rprompt")
	assert.Contains(t, got, terminal.SaveCursorPosition())
	assert.Contains(t, got, terminal.RestoreCursorPosition())
}

func TestTooltipFallback_PwshNoRoomReturnsEmpty(t *testing.T) {
	cache.Delete(cache.Session, RPromptKey)
	cache.Delete(cache.Session, RPromptLengthKey)
	cache.Set(cache.Session, RPromptKey, "my-rprompt", cache.INFINITE)
	// rprompt length exceeds available terminal space
	cache.Set(cache.Session, RPromptLengthKey, 200, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Column: 100})
	env.On("TerminalWidth").Return(200, nil)

	terminal.Init(shell.PWSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Empty(t, got)
}
