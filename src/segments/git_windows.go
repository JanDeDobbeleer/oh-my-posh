package segments

import "path/filepath"

// resolveGitPath resolves path relative to base.
func resolveGitPath(base, path string) string {
	if len(path) == 0 {
		return base
	}
	if filepath.IsAbs(path) {
		return path
	}
	// Note that git on Windows uses slashes exclusively. And it's okay
	// because Windows actually accepts both directory separators. More
	// importantly, however, parts of the git segment depend on those
	// slashes.
	if path[0] == '/' {
		// path is a disk-relative path.
		return filepath.VolumeName(base) + path
	}
	return filepath.ToSlash(filepath.Join(base, path))
}
