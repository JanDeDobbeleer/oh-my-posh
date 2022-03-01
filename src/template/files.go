package template

import "path/filepath"

func glob(pattern string) (bool, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}
