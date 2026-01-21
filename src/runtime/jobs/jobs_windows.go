//go:build windows

package jobs

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"golang.org/x/sys/windows"
)

var (
	jobsMu      sync.Mutex
	jobs        = map[uint64]windows.Handle{}
	processesMu sync.Mutex
	processes   = map[uint64]map[int]struct{}{}
)

func init() {
	// nothing to do here; runtime/cmd will call RegisterProcess when a
	// child is started. We previously used AssignPidHandler to bridge this
	// but now the jobs package exposes RegisterProcess directly.
}

// CreateJobForGoroutine creates a Job object for gid and sets the
// JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE flag so closing/terminating the job
// kills all assigned processes.
func CreateJobForGoroutine() error {
	gid := CurrentGID()
	jobsMu.Lock()
	if _, ok := jobs[gid]; ok {
		jobsMu.Unlock()
		return nil
	}
	jobsMu.Unlock()

	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}

	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	size := uint32(unsafe.Sizeof(info))
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), size); err != nil {
		_ = windows.CloseHandle(job)
	}

	jobsMu.Lock()
	jobs[gid] = job
	jobsMu.Unlock()

	return nil
}

// registerProcessWithGID keeps track of a started child process for the
// given goroutine id and attempts to assign it to the Job object if present.
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

	// Try to assign to job if exists (best-effort)
	jobsMu.Lock()
	job, ok := jobs[gid]
	jobsMu.Unlock()
	if !ok {
		log.Debugf("no job found for goroutine %d when assigning pid %d", gid, pid)
		return
	}

	proc, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		err = windows.CloseHandle(proc)
		if err != nil {
			log.Error(err)
		}
	}()

	if err = windows.AssignProcessToJobObject(job, proc); err != nil {
		log.Error(err)
	}
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

// KillGoroutineChildren will first try to terminate a Job if present, and
// otherwise will fall back to taskkill for each recorded pid.
func KillGoroutineChildren(gid uint64) error {
	// if Job exists, prefer terminating the Job
	jobsMu.Lock()
	job, hasJob := jobs[gid]
	if hasJob {
		delete(jobs, gid)
	}
	jobsMu.Unlock()
	if hasJob {
		// Terminate the job which kills all processes in it
		if err := windows.TerminateJobObject(job, 1); err == nil {
			// cleanup recorded pids as well
			processesMu.Lock()
			delete(processes, gid)
			processesMu.Unlock()
			return nil
		}
	}

	// No job or terminate failed; fall back to per-pid taskkill
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
		if err := exec.CommandContext(context.Background(), "taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run(); err != nil {
			errs = append(errs, fmt.Sprintf("taskkill %d: %v", pid, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to kill child processes: %s", strings.Join(errs, "; "))
	}
	return nil
}

// setProcessGroup ensures the child process runs in its own process group
// (CREATE_NEW_PROCESS_GROUP) so it can be terminated as a group.
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
