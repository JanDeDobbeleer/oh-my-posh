package ipc

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/stretchr/testify/assert"
)

func TestFlagsToProto(t *testing.T) {
	flags := &runtime.Flags{
		Type:          "primary",
		PipeStatus:    "0 0",
		ConfigPath:    "/home/user/.poshthemes/config.json",
		PSWD:          "~/projects",
		Shell:         "zsh",
		ShellVersion:  "5.9",
		PWD:           "/home/user/projects",
		AbsolutePWD:   "/home/user/projects",
		ErrorCode:     1,
		PromptCount:   42,
		Column:        80,
		TerminalWidth: 120,
		ExecutionTime: 1500.5,
		StackCount:    2,
		ConfigHash:    12345678,
		JobCount:      3,
		HasExtra:      true,
		Strict:        true,
		Debug:         true,
		Cleared:       true,
		NoExitCode:    true,
		Init:          true,
		Migrate:       true,
		Eval:          true,
		Escape:        true,
		IsPrimary:     true,
		Plain:         true,
		Force:         true,
	}

	proto := FlagsToProto(flags)

	assert.Equal(t, "primary", proto.Type)
	assert.Equal(t, "0 0", proto.PipeStatus)
	assert.Equal(t, "/home/user/.poshthemes/config.json", proto.ConfigPath)
	assert.Equal(t, "~/projects", proto.Pswd)
	assert.Equal(t, "zsh", proto.Shell)
	assert.Equal(t, "5.9", proto.ShellVersion)
	assert.Equal(t, "/home/user/projects", proto.Pwd)
	assert.Equal(t, "/home/user/projects", proto.AbsolutePwd)
	assert.Equal(t, int32(1), proto.ErrorCode)
	assert.Equal(t, int32(42), proto.PromptCount)
	assert.Equal(t, int32(80), proto.Column)
	assert.Equal(t, int32(120), proto.TerminalWidth)
	assert.Equal(t, 1500.5, proto.ExecutionTime)
	assert.Equal(t, int32(2), proto.StackCount)
	assert.Equal(t, uint64(12345678), proto.ConfigHash)
	assert.Equal(t, int32(3), proto.JobCount)
	assert.True(t, proto.HasExtra)
	assert.True(t, proto.Strict)
	assert.True(t, proto.Debug)
	assert.True(t, proto.Cleared)
	assert.True(t, proto.NoExitCode)
	assert.True(t, proto.Init)
	assert.True(t, proto.Migrate)
	assert.True(t, proto.Eval)
	assert.True(t, proto.Escape)
	assert.True(t, proto.IsPrimary)
	assert.True(t, proto.Plain)
	assert.True(t, proto.Force)
}

func TestProtoToFlags(t *testing.T) {
	proto := &Flags{
		Type:          "right",
		PipeStatus:    "0 1 0",
		ConfigPath:    "/config/theme.yaml",
		Pswd:          "~/dev",
		Shell:         "bash",
		ShellVersion:  "5.1",
		Pwd:           "/home/user/dev",
		AbsolutePwd:   "/home/user/dev",
		ErrorCode:     127,
		PromptCount:   100,
		Column:        40,
		TerminalWidth: 200,
		ExecutionTime: 500.25,
		StackCount:    1,
		ConfigHash:    87654321,
		JobCount:      5,
		HasExtra:      false,
		Strict:        false,
		Debug:         false,
		Cleared:       false,
		NoExitCode:    false,
		Init:          false,
		Migrate:       false,
		Eval:          false,
		Escape:        false,
		IsPrimary:     false,
		Plain:         false,
		Force:         false,
	}

	flags := ProtoToFlags(proto)

	assert.Equal(t, "right", flags.Type)
	assert.Equal(t, "0 1 0", flags.PipeStatus)
	assert.Equal(t, "/config/theme.yaml", flags.ConfigPath)
	assert.Equal(t, "~/dev", flags.PSWD)
	assert.Equal(t, "bash", flags.Shell)
	assert.Equal(t, "5.1", flags.ShellVersion)
	assert.Equal(t, "/home/user/dev", flags.PWD)
	assert.Equal(t, "/home/user/dev", flags.AbsolutePWD)
	assert.Equal(t, 127, flags.ErrorCode)
	assert.Equal(t, 100, flags.PromptCount)
	assert.Equal(t, 40, flags.Column)
	assert.Equal(t, 200, flags.TerminalWidth)
	assert.Equal(t, 500.25, flags.ExecutionTime)
	assert.Equal(t, 1, flags.StackCount)
	assert.Equal(t, uint64(87654321), flags.ConfigHash)
	assert.Equal(t, 5, flags.JobCount)
	assert.False(t, flags.HasExtra)
	assert.False(t, flags.Strict)
	assert.False(t, flags.Debug)
	assert.False(t, flags.Cleared)
	assert.False(t, flags.NoExitCode)
	assert.False(t, flags.Init)
	assert.False(t, flags.Migrate)
	assert.False(t, flags.Eval)
	assert.False(t, flags.Escape)
	assert.False(t, flags.IsPrimary)
	assert.False(t, flags.Plain)
	assert.False(t, flags.Force)
}

func TestFlagsRoundTrip(t *testing.T) {
	// Test that converting to proto and back preserves all values
	original := &runtime.Flags{
		Type:          "transient",
		PipeStatus:    "0",
		ConfigPath:    "/path/to/config",
		PSWD:          "~",
		Shell:         "fish",
		ShellVersion:  "3.6.0",
		PWD:           "/home/user",
		AbsolutePWD:   "/home/user",
		ErrorCode:     0,
		PromptCount:   1,
		Column:        0,
		TerminalWidth: 80,
		ExecutionTime: 0,
		StackCount:    0,
		ConfigHash:    0,
		JobCount:      0,
		HasExtra:      true,
		Strict:        false,
		Debug:         true,
		Cleared:       false,
		NoExitCode:    true,
		Init:          false,
		Migrate:       true,
		Eval:          false,
		Escape:        true,
		IsPrimary:     true,
		Plain:         false,
		Force:         true,
	}

	proto := FlagsToProto(original)
	roundTrip := ProtoToFlags(proto)

	assert.Equal(t, original.Type, roundTrip.Type)
	assert.Equal(t, original.PipeStatus, roundTrip.PipeStatus)
	assert.Equal(t, original.ConfigPath, roundTrip.ConfigPath)
	assert.Equal(t, original.PSWD, roundTrip.PSWD)
	assert.Equal(t, original.Shell, roundTrip.Shell)
	assert.Equal(t, original.ShellVersion, roundTrip.ShellVersion)
	assert.Equal(t, original.PWD, roundTrip.PWD)
	assert.Equal(t, original.AbsolutePWD, roundTrip.AbsolutePWD)
	assert.Equal(t, original.ErrorCode, roundTrip.ErrorCode)
	assert.Equal(t, original.PromptCount, roundTrip.PromptCount)
	assert.Equal(t, original.Column, roundTrip.Column)
	assert.Equal(t, original.TerminalWidth, roundTrip.TerminalWidth)
	assert.Equal(t, original.ExecutionTime, roundTrip.ExecutionTime)
	assert.Equal(t, original.StackCount, roundTrip.StackCount)
	assert.Equal(t, original.ConfigHash, roundTrip.ConfigHash)
	assert.Equal(t, original.JobCount, roundTrip.JobCount)
	assert.Equal(t, original.HasExtra, roundTrip.HasExtra)
	assert.Equal(t, original.Strict, roundTrip.Strict)
	assert.Equal(t, original.Debug, roundTrip.Debug)
	assert.Equal(t, original.Cleared, roundTrip.Cleared)
	assert.Equal(t, original.NoExitCode, roundTrip.NoExitCode)
	assert.Equal(t, original.Init, roundTrip.Init)
	assert.Equal(t, original.Migrate, roundTrip.Migrate)
	assert.Equal(t, original.Eval, roundTrip.Eval)
	assert.Equal(t, original.Escape, roundTrip.Escape)
	assert.Equal(t, original.IsPrimary, roundTrip.IsPrimary)
	assert.Equal(t, original.Plain, roundTrip.Plain)
	assert.Equal(t, original.Force, roundTrip.Force)
}

func TestFlagsToProtoNilInput(t *testing.T) {
	// Ensure nil input doesn't panic - it will cause a panic on dereference
	// This documents expected behavior
	assert.Panics(t, func() {
		FlagsToProto(nil)
	})
}

func TestProtoToFlagsNilInput(t *testing.T) {
	// Ensure nil input doesn't panic - it will cause a panic on dereference
	// This documents expected behavior
	assert.Panics(t, func() {
		ProtoToFlags(nil)
	})
}

func TestProtocolVersion(t *testing.T) {
	// Ensure protocol version is defined and positive
	assert.Greater(t, ProtocolVersion, uint32(0))
}
