package engine

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"
)

type codePoints map[int]int

func getGlyphCodePoints() (codePoints, error) {
	var codePoints = make(codePoints)

	client := &http.Client{}
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
	defer cncl()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ohmyposh.dev/codepoints.csv", nil)
	if err != nil {
		return codePoints, &ConnectionError{reason: err.Error()}
	}

	response, err := client.Do(request)
	if err != nil {
		return codePoints, err
	}

	defer response.Body.Close()

	lines, err := csv.NewReader(response.Body).ReadAll()
	if err != nil {
		return codePoints, err
	}

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		oldGlyph, err := strconv.ParseUint(line[0], 16, 32)
		if err != nil {
			continue
		}
		newGlyph, err := strconv.ParseUint(line[1], 16, 32)
		if err != nil {
			continue
		}
		codePoints[int(oldGlyph)] = int(newGlyph)
	}
	return codePoints, nil
}
