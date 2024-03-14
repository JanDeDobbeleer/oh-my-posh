package engine

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

type codePoints map[uint64]uint64

func getGlyphCodePoints() (codePoints, error) {
	var codePoints = make(codePoints)

	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
	defer cncl()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ohmyposh.dev/codepoints.csv", nil)
	if err != nil {
		return codePoints, &ConnectionError{reason: err.Error()}
	}

	response, err := platform.Client.Do(request)
	if err != nil {
		return codePoints, err
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return codePoints, err
	}

	lines := strings.Split(string(bytes), "\n")

	for _, line := range lines {
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
