//go:build windows

package template

import (
	"strings"
)

func init() {
	localeLayoutsResolver = resolveWindowsLocale
}

// resolveWindowsLocale reads the user's short-date and short-time patterns
// from the Windows registry and converts them to Go time layout strings.
//
// Registry keys read:
//
//	HKCU\Control Panel\International\sShortDate
//	HKCU\Control Panel\International\sShortTime
func resolveWindowsLocale() (dateLayout, timeLayout string) {
	const keyBase = `HKCU\Control Panel\International\`

	if v, err := env.WindowsRegistryKeyValue(keyBase + "sShortDate"); err == nil && v != nil {
		raw := strings.TrimSpace(v.String)
		if raw != "" {
			dateLayout = windowsPatternToGoLayout(raw)
		}
	}

	if v, err := env.WindowsRegistryKeyValue(keyBase + "sShortTime"); err == nil && v != nil {
		raw := strings.TrimSpace(v.String)
		if raw != "" {
			timeLayout = windowsPatternToGoLayout(raw)
		}
	}

	return dateLayout, timeLayout
}
