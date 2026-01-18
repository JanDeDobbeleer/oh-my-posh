package daemon

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/log"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher watches config files for changes using fsnotify.
// When a file changes, it invalidates the corresponding cache entry.
//
// We watch directories rather than files directly because editors using
// atomic saves (vim, neovim, etc.) delete/rename the original file and
// create a new one. Watching the file directly loses the watch when the
// inode changes, so we'd miss the CREATE event for the new file.
type ConfigWatcher struct {
	watcher     *fsnotify.Watcher
	cache       *ConfigCache
	files       map[string]string
	watchedDirs map[string]bool
	done        chan struct{}
	mu          sync.RWMutex
}

// NewConfigWatcher creates a new config watcher.
func NewConfigWatcher(cache *ConfigCache) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cw := &ConfigWatcher{
		watcher:     watcher,
		cache:       cache,
		files:       make(map[string]string),
		watchedDirs: make(map[string]bool),
		done:        make(chan struct{}),
	}

	go cw.eventLoop()

	return cw, nil
}

// Watch starts watching the given files for a config path.
//
// We watch the parent directory of each file rather than the file itself.
// This is necessary because editors using atomic saves (vim, neovim, VSCode, etc.)
// work by:
//  1. Writing content to a temp file
//  2. Renaming/deleting the original file
//  3. Renaming the temp file to the original name
//
// When watching a file directly, fsnotify tracks the inode. After step 2,
// the watch is removed (the inode is gone). The new file created in step 3
// has a different inode, so we never receive the CREATE event.
//
// By watching the directory, we receive events for all file operations within it,
// including the CREATE event when the new file appears.
func (cw *ConfigWatcher) Watch(configPath string, filePaths []string) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	for _, filePath := range filePaths {
		// Skip remote files
		if strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "http://") {
			continue
		}

		// Helper to add a file to watch
		addWatch := func(path string) {
			// Skip if already tracking this file
			if _, ok := cw.files[path]; ok {
				return
			}

			// Watch the parent directory
			dir := filepath.Dir(path)
			if !cw.watchedDirs[dir] {
				if err := cw.watcher.Add(dir); err != nil {
					log.Debugf("failed to watch directory %s: %v", dir, err)
					return
				}
				cw.watchedDirs[dir] = true
				log.Debugf("watching directory: %s", dir)
			}

			// Track this file so we know which events to handle
			cw.files[path] = configPath
			log.Debugf("tracking config file: %s", path)
		}

		addWatch(filePath)

		// If it's a symlink, also watch the target
		realPath, err := filepath.EvalSymlinks(filePath)
		if err == nil && realPath != filePath {
			addWatch(realPath)
		}
	}

	return nil
}

// Close stops the watcher and cleans up resources.
func (cw *ConfigWatcher) Close() error {
	close(cw.done)
	return cw.watcher.Close()
}

// eventLoop processes file system events.
func (cw *ConfigWatcher) eventLoop() {
	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}
			cw.handleEvent(event)

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			log.Debugf("fsnotify error: %v", err)

		case <-cw.done:
			return
		}
	}
}

// handleEvent processes a file system event.
//
// Since we watch directories (not files), we receive events for all files
// in the directory. We filter to only handle events for files we're tracking.
func (cw *ConfigWatcher) handleEvent(event fsnotify.Event) {
	log.Debugf("fsnotify event: %s %s", event.Op, event.Name)

	cw.mu.RLock()
	configPath, ok := cw.files[event.Name]
	cw.mu.RUnlock()

	if !ok {
		// Event for a file we're not tracking - ignore
		return
	}

	// Any modification event (WRITE, CREATE, REMOVE, RENAME) invalidates the cache.
	// - WRITE: file was modified in place
	// - CREATE: new file appeared (final step of atomic save)
	// - REMOVE/RENAME: file was deleted or renamed away (atomic save in progress,
	//   or user deleted the file)
	//
	// We don't need to re-add watches because we watch the directory, not the file.
	// The directory watch persists across atomic saves.
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
		log.Debugf("config file changed (%s): %s", event.Op, event.Name)
		cw.cache.Invalidate(configPath)
	}
}
