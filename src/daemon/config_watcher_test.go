package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigWatcher(t *testing.T) {
	cache := NewConfigCache()
	watcher, err := NewConfigWatcher(cache)
	require.NoError(t, err)
	defer watcher.Close()

	// Create a test config file
	configPath := createTestConfig(t)
	configPath, err = filepath.EvalSymlinks(configPath)
	require.NoError(t, err)

	// Add to cache first
	cfg, err := config.Parse(configPath)
	require.NoError(t, err)
	cache.Set(configPath, cfg, []string{configPath})

	err = watcher.Watch(configPath, []string{configPath})
	require.NoError(t, err)

	// Verify watching - files maps file path to config path
	watcher.mu.RLock()
	assert.Contains(t, watcher.files, configPath)
	assert.Equal(t, configPath, watcher.files[configPath])
	watcher.mu.RUnlock()

	// Modify the file
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	// Change the template value to trigger a hash change
	newContent := []byte(strings.Replace(string(content), "hello", "world", 1))
	err = os.WriteFile(configPath, newContent, 0644)
	require.NoError(t, err)

	// Wait for watcher to pick up change and invalidate cache
	// We use a longer timeout because fsnotify can be slow
	success := false
	for range 20 {
		_, ok := cache.Get(configPath)
		if !ok {
			success = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, success, "Cache should have been invalidated after file change")
}

func TestConfigWatcherInheritance(t *testing.T) {
	cache := NewConfigCache()
	watcher, err := NewConfigWatcher(cache)
	require.NoError(t, err)
	defer watcher.Close()

	// Create parent config
	parentContent := `{
		"version": 4,
		"palette": { "red": "#ff0000" }
	}`
	parentFile, err := os.CreateTemp("", "omp-parent-*.json")
	require.NoError(t, err)
	_, _ = parentFile.WriteString(parentContent)
	parentPath, _ := filepath.EvalSymlinks(parentFile.Name())
	parentFile.Close()
	defer os.Remove(parentPath)

	// Create child config that extends parent
	escapedParentPath := strings.ReplaceAll(parentPath, "\\", "\\\\")
	childContent := fmt.Sprintf(`{
		"version": 4,
		"extends": "%s",
		"blocks": [{ "type": "prompt", "segments": [{ "type": "text", "template": "hello" }] }]
	}`, escapedParentPath)
	childFile, err := os.CreateTemp("", "omp-child-*.json")
	require.NoError(t, err)
	_, _ = childFile.WriteString(childContent)
	childPath, _ := filepath.EvalSymlinks(childFile.Name())
	childFile.Close()
	defer os.Remove(childPath)

	// Parse child config
	cfg, err := config.Parse(childPath)
	require.NoError(t, err)
	assert.Contains(t, cfg.FilePaths, parentPath)

	// Add to cache and start watching
	cache.Set(childPath, cfg, cfg.FilePaths)
	err = watcher.Watch(childPath, cfg.FilePaths)
	require.NoError(t, err)

	// Verify both files are watched and map to child config
	watcher.mu.RLock()
	assert.Equal(t, childPath, watcher.files[childPath])
	assert.Equal(t, childPath, watcher.files[parentPath])
	watcher.mu.RUnlock()

	// Modify parent file
	newParentContent := `{
		"version": 4,
		"palette": { "red": "#bb0000" }
	}`
	err = os.WriteFile(parentFile.Name(), []byte(newParentContent), 0644)
	require.NoError(t, err)

	// Wait for watcher to invalidate child config
	success := false
	for range 20 {
		_, ok := cache.Get(childPath)
		if !ok {
			success = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, success, "Child cache should have been invalidated after parent file change")
}

func TestConfigWatcherSymlink(t *testing.T) {
	cache := NewConfigCache()
	watcher, err := NewConfigWatcher(cache)
	require.NoError(t, err)
	defer watcher.Close()

	// Create real config file
	realDir := t.TempDir()
	realConfig := filepath.Join(realDir, "real_config.json")
	content := `{ "version": 4 }`
	err = os.WriteFile(realConfig, []byte(content), 0644)
	require.NoError(t, err)

	// Create symlink in another directory
	linkDir := t.TempDir()
	linkConfig := filepath.Join(linkDir, "link_config.json")
	err = os.Symlink(realConfig, linkConfig)
	require.NoError(t, err)

	// Parse via symlink
	cfg, err := config.Parse(linkConfig)
	require.NoError(t, err)

	cache.Set(linkConfig, cfg, []string{linkConfig})

	err = watcher.Watch(linkConfig, []string{linkConfig})
	require.NoError(t, err)

	// Modify REAL file
	newContent := `{ "version": 4, "modified": true }`
	err = os.WriteFile(realConfig, []byte(newContent), 0644)
	require.NoError(t, err)

	// Wait for invalidation
	success := false
	for range 20 {
		_, ok := cache.Get(linkConfig)
		if !ok {
			success = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, success, "Cache should be invalidated when symlink target is modified")
}

// TestConfigWatcherAtomicSave tests that the watcher correctly detects changes
// when editors use atomic saves (write to temp file, then rename).
// This is how vim, neovim, VSCode, and many other editors save files.
func TestConfigWatcherAtomicSave(t *testing.T) {
	cache := NewConfigCache()
	watcher, err := NewConfigWatcher(cache)
	require.NoError(t, err)
	defer watcher.Close()

	// Create a test config file
	configPath := createTestConfig(t)
	configPath, err = filepath.EvalSymlinks(configPath)
	require.NoError(t, err)

	// Add to cache first
	cfg, err := config.Parse(configPath)
	require.NoError(t, err)
	cache.Set(configPath, cfg, []string{configPath})

	err = watcher.Watch(configPath, []string{configPath})
	require.NoError(t, err)

	// Simulate atomic save (like vim/neovim):
	// 1. Write new content to a temp file
	// 2. Rename/remove the original file
	// 3. Rename the temp file to the original name
	dir := filepath.Dir(configPath)
	tempPath := filepath.Join(dir, "config.tmp")

	newContent := `{ "version": 4, "blocks": [{ "type": "prompt", "segments": [{ "type": "text", "template": "atomic save test" }] }] }`
	err = os.WriteFile(tempPath, []byte(newContent), 0644)
	require.NoError(t, err)

	// Remove original (some editors do this instead of rename)
	err = os.Remove(configPath)
	require.NoError(t, err)

	// Rename temp to original
	err = os.Rename(tempPath, configPath)
	require.NoError(t, err)

	// Wait for watcher to pick up change and invalidate cache
	success := false
	for range 20 {
		_, ok := cache.Get(configPath)
		if !ok {
			success = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, success, "Cache should have been invalidated after atomic save")

	// Verify subsequent atomic saves also work (the bug was that only the first worked)
	cfg, err = config.Parse(configPath)
	require.NoError(t, err)
	cache.Set(configPath, cfg, []string{configPath})

	// Second atomic save
	tempPath2 := filepath.Join(dir, "config2.tmp")
	newContent2 := `{ "version": 4, "blocks": [{ "type": "prompt", "segments": [{ "type": "text", "template": "second atomic save" }] }] }`
	err = os.WriteFile(tempPath2, []byte(newContent2), 0644)
	require.NoError(t, err)

	err = os.Remove(configPath)
	require.NoError(t, err)

	err = os.Rename(tempPath2, configPath)
	require.NoError(t, err)

	// Wait for second invalidation
	success = false
	for range 20 {
		_, ok := cache.Get(configPath)
		if !ok {
			success = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, success, "Cache should have been invalidated after second atomic save")
}

func TestConfigWatcherSkipsRemoteFiles(t *testing.T) {
	cache := NewConfigCache()
	watcher, err := NewConfigWatcher(cache)
	require.NoError(t, err)
	defer watcher.Close()

	configPath := "/local/config.json"
	remotePaths := []string{
		"https://example.com/config.json",
		"http://example.com/config.json",
		configPath,
	}

	// This should not error even though the local file doesn't exist
	// because we skip remote files
	err = watcher.Watch(configPath, remotePaths)
	require.NoError(t, err)

	// Only local file should be in the files map (even if watch failed)
	watcher.mu.RLock()
	_, hasHTTPS := watcher.files["https://example.com/config.json"]
	_, hasHTTP := watcher.files["http://example.com/config.json"]
	watcher.mu.RUnlock()

	assert.False(t, hasHTTPS, "HTTPS URLs should not be watched")
	assert.False(t, hasHTTP, "HTTP URLs should not be watched")
}
