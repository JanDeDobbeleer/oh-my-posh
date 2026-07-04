//go:build windows

package jobs

import "testing"

// TestCloseGoroutineJob verifies that CloseGoroutineJob removes both the job
// and its recorded pids for the current goroutine, closes the underlying
// handle exactly once, and is a safe no-op if called again (e.g. because
// KillGoroutineChildren already won the race and closed it first).
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

// TestCloseGoroutineJobNoJob verifies calling CloseGoroutineJob without a
// prior CreateJobForGoroutine call (e.g. a segment with no timeout) is a
// harmless no-op.
func TestCloseGoroutineJobNoJob(t *testing.T) {
	CloseGoroutineJob()
}
