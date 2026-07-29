//go:build windows

package jobs

import "testing"

// Removes both the job and its recorded pids for the current goroutine, closes the underlying
// handle exactly once, and is safe to call again (e.g. if KillGoroutineChildren already won the
// race and closed it first).
func TestCloseGoroutineJob(t *testing.T) {
	if err := CreateJobForGoroutine("test"); err != nil {
		t.Fatalf("CreateJobForGoroutine returned error: %v", err)
	}

	gid := CurrentGID()

	jobsMu.Lock()
	_, hasJob := jobs[gid]
	jobsMu.Unlock()

	if !hasJob {
		t.Fatalf("expected job to be registered for gid %d", gid)
	}

	// register a fake pid entry to exercise the processes map cleanup
	processesMu.Lock()
	processes[gid] = map[int]struct{}{1234: {}}
	processesMu.Unlock()

	CloseGoroutineJob()

	jobsMu.Lock()
	_, hasJob = jobs[gid]
	jobsMu.Unlock()

	if hasJob {
		t.Fatalf("expected job to be removed for gid %d after close", gid)
	}

	processesMu.Lock()
	_, hasProcesses := processes[gid]
	processesMu.Unlock()

	if hasProcesses {
		t.Fatalf("expected processes entry to be removed for gid %d after close", gid)
	}

	// calling again must be a safe no-op (no double-close panic/error)
	CloseGoroutineJob()
}

// Calling CloseGoroutineJob without a prior CreateJobForGoroutine call (e.g. a segment with no
// timeout) must be a harmless no-op.
func TestCloseGoroutineJobNoJob(t *testing.T) {
	CloseGoroutineJob()
}
