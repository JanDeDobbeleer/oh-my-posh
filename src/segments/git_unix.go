//go:build !windows

package segments

import "path/filepath"

func resolveGitPath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(base, path)
}
