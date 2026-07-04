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
	"time"
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

// CreateJobForGoroutine creates a Job object for gid and sets the
// JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE flag so closing/terminating the job
// kills all assigned processes.
func CreateJobForGoroutine(label string) error {
	gid := CurrentGID()
	defer log.Trace(time.Now(), fmt.Sprintf("creating job for goroutine(%s): %d", label, gid))

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

	log.Debugf("successfully added process to job for goroutine: %d, pid: %d", gid, pid)
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

// CloseGoroutineJob releases the Job object (if any) created for the current
// goroutine via CreateJobForGoroutine, along with its recorded pids.
//
// Ownership invariant: a map entry in `jobs` represents an open handle that
// nobody else has closed yet. Removal from the map and closing the handle
// must happen atomically from the caller's perspective, and whichever of
// CloseGoroutineJob/KillGoroutineChildren deletes the entry first is the sole
// owner of the handle from that point on - the other side, finding the entry
// already gone, does nothing further. This guarantees exactly one
// windows.CloseHandle/TerminateJobObject call per handle, so a segment that
// finishes normally right as its timeout fires can never race a double-close
// or have its handle closed out from under a concurrent kill.
//
// This must only be called from the same goroutine that created the job
// (i.e. after the segment's Execute body has returned), since by then any
// child processes started via cmd.Run have already exited - JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
// only affects processes still assigned to the job at close time.
func CloseGoroutineJob() {
	gid := CurrentGID()

	jobsMu.Lock()
	job, hasJob := jobs[gid]
	if hasJob {
		delete(jobs, gid)
	}
	jobsMu.Unlock()

	processesMu.Lock()
	delete(processes, gid)
	processesMu.Unlock()

	if !hasJob {
		// Either no job was ever created for this goroutine, or
		// KillGoroutineChildren already won the race and took ownership of
		// the handle (and already closed it via TerminateJobObject).
		return
	}

	if err := windows.CloseHandle(job); err != nil {
		log.Error(err)
		return
	}

	log.Debugf("closed job object for goroutine: %d", gid)
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
		// Terminate the job which kills all processes in it. This also
		// closes/invalidates the handle's usefulness; TerminateJobObject
		// does not itself close the handle, but since we've already removed
		// it from the map, CloseGoroutineJob running concurrently (e.g. the
		// original goroutine finishing right as the timeout fires) will see
		// no entry and skip closing it here, so we're the sole owner and
		// must close it ourselves to avoid leaking the handle.
		terminated := windows.TerminateJobObject(job, 1) == nil
		if err := windows.CloseHandle(job); err != nil {
			log.Error(err)
		}
		if terminated {
			// cleanup recorded pids as well
			processesMu.Lock()
			delete(processes, gid)
			processesMu.Unlock()
			log.Debugf("successfully terminated job object for goroutine: %d", gid)
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
