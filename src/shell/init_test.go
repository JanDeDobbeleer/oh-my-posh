package shell

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func newNuMockEnv(runCommandOutput string, runCommandErr error) *mock.Environment {
	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{
		Shell:      NU,
		ConfigPath: "config.jsonc",
	})
	env.On("GOOS").Return(runtime.WINDOWS)
	env.On("Getenv", testify_.Anything).Return("")
	env.On("RunCommand", "nu", []string{"--commands", nuStreamingProbe}).Return(runCommandOutput, runCommandErr)

	return env
}

func TestGenerateScriptDropsStreamingWhenNuUnsupported(t *testing.T) {
	env := newNuMockEnv("false", nil)

	got := generateScript(env, Streaming)

	assert.NotContains(t, got, "_omp_stream_primary")
}

func TestGenerateScriptKeepsStreamingWhenNuSupported(t *testing.T) {
	env := newNuMockEnv("true", nil)

	got := generateScript(env, Streaming)

	assert.Contains(t, got, "_omp_stream_primary")
}

func TestGenerateScriptDropsStreamingOnProbeError(t *testing.T) {
	env := newNuMockEnv("", assert.AnError)

	got := generateScript(env, Streaming)

	assert.NotContains(t, got, "_omp_stream_primary")
}

func TestGenerateNuScriptDropsStreamingWhenNuUnsupported(t *testing.T) {
	env := newNuMockEnv("false", nil)

	got := generateNuScript(env, Streaming)

	assert.NotContains(t, got, "_omp_stream_primary")
}

func TestGenerateNuScriptKeepsStreamingWhenNuSupported(t *testing.T) {
	env := newNuMockEnv("true", nil)

	got := generateNuScript(env, Streaming)

	assert.Contains(t, got, "_omp_stream_primary")
}

func TestNuSupportsStreamingUsesExecutableOverride(t *testing.T) {
	env := new(mock.Environment)
	env.On("Getenv", nuStreamingExecutableEnv).Return(`D:\nushell-streaming\nu.exe`)
	env.On("RunCommand", `D:\nushell-streaming\nu.exe`, []string{"--commands", nuStreamingProbe}).Return("true", nil)

	assert.True(t, nuSupportsStreaming(env))
}
