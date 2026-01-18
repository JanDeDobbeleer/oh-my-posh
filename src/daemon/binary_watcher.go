package daemon

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"

	"github.com/fsnotify/fsnotify"
)

// BinaryWatcher watches the oh-my-posh executable for changes using fsnotify.
// When the binary is replaced (e.g. by brew upgrade, go install, or an installer),
// it calls the onChange callback so the daemon can shut down gracefully.
//
// Like ConfigWatcher, we watch the parent directory rather than the file itself,
// because installers replace binaries atomically (delete + rename or rename + rename).
type BinaryWatcher struct {
	watcher  *fsnotify.Watcher
	done     chan struct{}
	binPath  string
	fileName string
	once     sync.Once
}

// NewBinaryWatcher creates a watcher that monitors binPath for changes.
// onChange is called at most once when the binary is replaced.
func NewBinaryWatcher(binPath string, onChange func()) (*BinaryWatcher, error) {
	// Resolve symlinks (e.g. Homebrew: /usr/local/bin/oh-my-posh -> ../Cellar/...)
	resolved, err := filepath.EvalSymlinks(binPath)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(resolved)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, err
	}

	bw := &BinaryWatcher{
		watcher:  watcher,
		binPath:  resolved,
		fileName: filepath.Base(resolved),
		done:     make(chan struct{}),
	}

	go bw.eventLoop(onChange)

	log.Debugf("watching binary: %s (dir: %s)", resolved, dir)

	return bw, nil
}

// Close stops the watcher.
func (bw *BinaryWatcher) Close() error {
	bw.once.Do(func() { close(bw.done) })
	return bw.watcher.Close()
}

// eventLoop processes fsnotify events with debounce.
func (bw *BinaryWatcher) eventLoop(onChange func()) {
	var debounce *time.Timer

	for {
		select {
		case event, ok := <-bw.watcher.Events:
			if !ok {
				return
			}

			if filepath.Base(event.Name) != bw.fileName {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			log.Debugf("binary changed (%s): %s", event.Op, event.Name)

			// Debounce: atomic saves can produce multiple events.
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(1*time.Second, func() {
				log.Debug("binary change confirmed, triggering callback")
				onChange()
			})

		case err, ok := <-bw.watcher.Errors:
			if !ok {
				return
			}
			log.Debugf("binary watcher error: %v", err)

		case <-bw.done:
			if debounce != nil {
				debounce.Stop()
			}
			return
		}
	}
}
