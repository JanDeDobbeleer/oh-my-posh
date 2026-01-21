//go:build !windows

package jobs

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var (
	processesMu sync.Mutex
	processes   = map[uint64]map[int]struct{}{}
)

func CreateJobForGoroutine(_ string) error { return nil }
func AssignPidToGoroutineJob(_ int) error  { return nil }

// setProcessGroup ensures the child process runs in its own process group so
// it can be killed with a group kill (negative pid).
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// registerProcessWithGID keeps track of a started child process for the
// given goroutine id.
func RegisterProcess(pid int) {
	gid := CurrentGID()
	processesMu.Lock()
	m := processes[gid]
	if m == nil {
		m = map[int]struct{}{}
		processes[gid] = m
	}
	m[pid] = struct{}{}
	processesMu.Unlock()
}

func UnregisterProcess(pid int) {
	gid := CurrentGID()
	processesMu.Lock()
	if m, ok := processes[gid]; ok {
		delete(m, pid)
		if len(m) == 0 {
			delete(processes, gid)
		}
	}
	processesMu.Unlock()
}

// KillGoroutineChildren attempts to kill all child processes started by the
// goroutine identified by gid using process groups (PGID). This mirrors the
// previous behavior performed in runtime/cmd.
func KillGoroutineChildren(gid uint64) error {
	processesMu.Lock()
	pidsMap, ok := processes[gid]
	if !ok || len(pidsMap) == 0 {
		processesMu.Unlock()
		return nil
	}
	pids := make([]int, 0, len(pidsMap))
	for pid := range pidsMap {
		pids = append(pids, pid)
	}
	delete(processes, gid)
	processesMu.Unlock()

	var errs []string
	for _, pid := range pids {
		// negative pid kills the process group
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			errs = append(errs, fmt.Sprintf("kill -%d: %v", pid, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to kill child processes: %s", strings.Join(errs, "; "))
	}
	return nil
}
