package segments

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	thousand = 1000
	million  = 1000000

	gaugeMarkedChar   options.Option = "gauge_marked_char"
	gaugeUnmarkedChar options.Option = "gauge_unmarked_char"
)

// formatTokenCount formats a token count as a human-readable string ("1.2K", "3.4M", or raw).
func formatTokenCount(n int) string {
	if n < thousand {
		return fmt.Sprintf("%d", n)
	}

	if n < million {
		return fmt.Sprintf("%.1fK", float64(n)/float64(thousand))
	}

	return fmt.Sprintf("%.1fM", float64(n)/float64(million))
}
