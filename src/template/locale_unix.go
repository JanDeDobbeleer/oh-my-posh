//go:build !windows

package template

import (
	"strings"
)

func init() {
	localeLayoutsResolver = resolveUnixLocale
}

// resolveUnixLocale runs `locale -k LC_TIME` and extracts the d_fmt and t_fmt
// values, converting them from POSIX strftime format to Go time layout strings.
func resolveUnixLocale() (dateLayout, timeLayout string) {
	output, err := env.RunCommand("locale", "-k", "LC_TIME")
	if err != nil {
		return "", ""
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		value = strings.Trim(value, `"`)

		switch key {
		case "d_fmt":
			if value != "" {
				dateLayout = posixPatternToGoLayout(value)
			}
		case "t_fmt":
			if value != "" {
				timeLayout = posixPatternToGoLayout(value)
			}
		}
	}

	return dateLayout, timeLayout
}
