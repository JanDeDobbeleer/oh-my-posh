package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Session represents an active shell session connected to the daemon.
type Session struct {
	cancel context.CancelFunc
	UUID   string
	Shell  string
	PID    int
}

// SessionManager tracks active shell sessions and watches for their termination.
// When all sessions end, it triggers daemon shutdown.
type SessionManager struct {
	sessions     map[int]*Session
	onUnregister func(int)
	onEmpty      func()
	mu           sync.RWMutex
}

// NewSessionManager creates a new session manager.
// The onUnregister callback is called when a session is removed.
// The onEmpty callback is called when the last session ends.
func NewSessionManager(onUnregister func(int), onEmpty func()) *SessionManager {
	return &SessionManager{
		sessions:     make(map[int]*Session),
		onUnregister: onUnregister,
		onEmpty:      onEmpty,
	}
}

// Register adds a new session for the given PID.
// If the PID is already registered, this is a no-op.
// Starts watching for process termination.
func (sm *SessionManager) Register(pid int, uuid, shell string) {
	log.Debugf("SessionManager.Register: PID %d", pid)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[pid]; exists {
		log.Debugf("Session for PID %d already registered", pid)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := &Session{
		PID:    pid,
		UUID:   uuid,
		Shell:  shell,
		cancel: cancel,
	}
	sm.sessions[pid] = session

	log.Debugf("Registered session for PID %d (total: %d)", pid, len(sm.sessions))

	// Start watching for process exit in background
	go sm.watchProcess(ctx, pid)
}

// Unregister removes a session for the given PID.
// Called when shell explicitly unregisters (e.g., on exit trap).
func (sm *SessionManager) Unregister(pid int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.unregisterLocked(pid)
}

// unregisterLocked removes a session (must hold lock).
func (sm *SessionManager) unregisterLocked(pid int) {
	session, exists := sm.sessions[pid]
	if !exists {
		return
	}

	// Cancel the process watcher
	session.cancel()
	delete(sm.sessions, pid)

	log.Debugf("Unregistered session for PID %d (remaining: %d)", pid, len(sm.sessions))

	if sm.onUnregister != nil {
		sm.onUnregister(pid)
	}

	if len(sm.sessions) == 0 && sm.onEmpty != nil {
		log.Debug("All sessions ended, triggering shutdown")
		sm.onEmpty()
	}
}

// Count returns the number of active sessions.
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// watchProcess monitors a process and calls unregister when it exits.
// Implementation is platform-specific (see session_*.go files).
func (sm *SessionManager) watchProcess(ctx context.Context, pid int) {
	log.Debugf("SessionManager.watchProcess: starting for PID %d", pid)
	// Wait for process to exit (platform-specific implementation)
	waitForProcessExit(ctx, pid)
	log.Debugf("SessionManager.watchProcess: waitForProcessExit returned for PID %d", pid)

	// Process exited - unregister the session
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if still registered (might have been explicitly unregistered)
	if _, exists := sm.sessions[pid]; exists {
		log.Debugf("Process %d exited, removing session", pid)
		sm.unregisterLocked(pid)
	} else {
		log.Debugf("Process %d exited, but session already removed", pid)
	}
}

// pollForProcessExit is a fallback mechanism that periodically checks
// if a process is running. It blocks until the process exits or context is cancelled.
func pollForProcessExit(ctx context.Context, pid int) {
	log.Debugf("Using polling fallback for PID %d", pid)

	// Check immediately once before waiting
	if !IsProcessRunning(pid) {
		log.Debugf("Process %d exit detected via polling (immediate)", pid)
		return
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !IsProcessRunning(pid) {
				log.Debugf("Process %d exit detected via polling", pid)
				return
			}
		}
	}
}
