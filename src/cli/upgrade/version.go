package upgrade

import "strings"

// IsMajorUpgrade reports whether latest bumps the major version relative to
// current. It is a pure version comparison with no terminal UI involved, so
// unlike Install and the Stage reporting around it, it stayed in this
// package rather than moving to cli/upgrade/tui: notice.go (this package)
// calls it directly to decide whether to suggest --force, and callers like
// cli/upgrade.go need it without pulling in a UI at all.
func IsMajorUpgrade(current, latest string) bool {
	if current == "" {
		return false
	}

	getMajorNumber := func(version string) string {
		major, _, _ := strings.Cut(version, ".")
		return major
	}

	return getMajorNumber(current) != getMajorNumber(latest)
}
