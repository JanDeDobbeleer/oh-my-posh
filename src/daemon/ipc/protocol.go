package ipc

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative daemon.proto

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// ProtocolVersion is the current protocol version.
// Increment when making breaking changes to the protocol.
const ProtocolVersion uint32 = 1

// FlagsToProto converts runtime.Flags to proto Flags.
func FlagsToProto(f *runtime.Flags) *Flags {
	return &Flags{
		Type:          f.Type,
		PipeStatus:    f.PipeStatus,
		ConfigPath:    f.ConfigPath,
		Pswd:          f.PSWD,
		Shell:         f.Shell,
		ShellVersion:  f.ShellVersion,
		Pwd:           f.PWD,
		AbsolutePwd:   f.AbsolutePWD,
		ErrorCode:     int32(f.ErrorCode),
		PromptCount:   int32(f.PromptCount),
		Column:        int32(f.Column),
		TerminalWidth: int32(f.TerminalWidth),
		ExecutionTime: f.ExecutionTime,
		StackCount:    int32(f.StackCount),
		ConfigHash:    f.ConfigHash,
		JobCount:      int32(f.JobCount),
		HasExtra:      f.HasExtra,
		Strict:        f.Strict,
		Debug:         f.Debug,
		Cleared:       f.Cleared,
		NoExitCode:    f.NoExitCode,
		Init:          f.Init,
		Migrate:       f.Migrate,
		Eval:          f.Eval,
		Escape:        f.Escape,
		IsPrimary:     f.IsPrimary,
		Plain:         f.Plain,
		Force:         f.Force,
	}
}

// ProtoToFlags converts proto Flags to runtime.Flags.
func ProtoToFlags(f *Flags) *runtime.Flags {
	return &runtime.Flags{
		Type:          f.Type,
		PipeStatus:    f.PipeStatus,
		ConfigPath:    f.ConfigPath,
		PSWD:          f.Pswd,
		Shell:         f.Shell,
		ShellVersion:  f.ShellVersion,
		PWD:           f.Pwd,
		AbsolutePWD:   f.AbsolutePwd,
		ErrorCode:     int(f.ErrorCode),
		PromptCount:   int(f.PromptCount),
		Column:        int(f.Column),
		TerminalWidth: int(f.TerminalWidth),
		ExecutionTime: f.ExecutionTime,
		StackCount:    int(f.StackCount),
		ConfigHash:    f.ConfigHash,
		JobCount:      int(f.JobCount),
		HasExtra:      f.HasExtra,
		Strict:        f.Strict,
		Debug:         f.Debug,
		Cleared:       f.Cleared,
		NoExitCode:    f.NoExitCode,
		Init:          f.Init,
		Migrate:       f.Migrate,
		Eval:          f.Eval,
		Escape:        f.Escape,
		IsPrimary:     f.IsPrimary,
		Plain:         f.Plain,
		Force:         f.Force,
	}
}
