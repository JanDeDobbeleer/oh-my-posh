//go:build !windows

package segments

import "path/filepath"

// resolveGitPath resolves path relative to base.
func resolveGitPath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
