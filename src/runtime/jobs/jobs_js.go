//go:build js

package jobs

import "os/exec"

// js/wasm has no process model at all: no fork/exec, no process groups, and
// no signals to send a child. Every hook here is a no-op so callers (e.g.
// config.Segment.Execute) can call them unconditionally regardless of
// platform, the same way jobs_other.go already no-ops the primitives that
// don't exist on a given OS - this just extends that to a platform with no
// child processes to manage in the first place.
func CreateJobForGoroutine(_ string) error { return nil }
func AssignPidToGoroutineJob(_ int) error  { return nil }
func CloseGoroutineJob()                   {}
func SetProcessGroup(_ *exec.Cmd)          {}
func RegisterProcess(_ int)                {}
func UnregisterProcess(_ int)              {}
func KillGoroutineChildren(_ uint64) error { return nil }
