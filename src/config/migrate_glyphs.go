package config

import (
	"context"
	"fmt"
	"io"
	httplib "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

type ConnectionError struct {
	reason string
}

func (f *ConnectionError) Error() string {
	return f.reason
}

type codePoints map[uint64]uint64

func getGlyphCodePoints() (codePoints, error) {
	var codePoints = make(codePoints)

	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodGet, "https://ohmyposh.dev/codepoints.csv", nil)
	if err != nil {
		return codePoints, &ConnectionError{reason: err.Error()}
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		return codePoints, err
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return codePoints, err
	}

	lines := strings.SplitSeq(string(bytes), "\n")

	for line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		oldGlyph, err := strconv.ParseUint(fields[0], 16, 32)
		if err != nil {
			continue
		}

		newGlyph, err := strconv.ParseUint(fields[1], 16, 32)
		if err != nil {
			continue
		}

		codePoints[oldGlyph] = newGlyph
	}

	return codePoints, nil
}

func escapeGlyphs(s string, migrate bool) string {
	shouldExclude := func(r rune) bool {
		if r < 0x1000 { // Basic Multilingual Plane
			return true
		}
		if r > 0x1F600 && r < 0x1F64F { // Emoticons
			return true
		}
		if r > 0x1F300 && r < 0x1F5FF { // Misc Symbols and Pictographs
			return true
		}
		if r > 0x1F680 && r < 0x1F6FF { // Transport and Map
			return true
		}
		if r > 0x2600 && r < 0x26FF { // Misc symbols
			return true
		}
		if r > 0x2700 && r < 0x27BF { // Dingbats
			return true
		}
		if r > 0xFE00 && r < 0xFE0F { // Variation Selectors
			return true
		}
		if r > 0x1F900 && r < 0x1F9FF { // Supplemental Symbols and Pictographs
			return true
		}
		if r > 0x1F1E6 && r < 0x1F1FF { // Flags
			return true
		}
		return false
	}

	var cp codePoints
	var err error
	if migrate {
		cp, err = getGlyphCodePoints()
		if err != nil {
			migrate = false
		}
	}

	var builder strings.Builder
	for _, r := range s {
		// exclude regular characters and emojis
		if shouldExclude(r) {
			builder.WriteRune(r)
			continue
		}

		if migrate {
			if val, OK := cp[uint64(r)]; OK {
				r = rune(val)
			}
		}

		if r > 0x10000 {
			// calculate surrogate pairs
			one := 0xd800 + (((r - 0x10000) >> 10) & 0x3ff)
			two := 0xdc00 + ((r - 0x10000) & 0x3ff)
			quoted := fmt.Sprintf("\\u%04x\\u%04x", one, two)
			builder.WriteString(quoted)
			continue
		}

		quoted := fmt.Sprintf("\\u%04x", r)
		builder.WriteString(quoted)
	}
	return builder.String()
}
