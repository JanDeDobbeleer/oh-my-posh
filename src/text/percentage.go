package text

import (
	"fmt"
	"strings"
)

// Percentage represents a percentage value with gauge visualization.
type Percentage int

// Gauge returns a 5-character gauge visualization showing remaining capacity (▰▰▰▰▱ style).
// The gauge displays remaining capacity, so 20% used shows 4 filled blocks (80% remaining).
func (p Percentage) Gauge() string {
	percent := min(max(int(p), 0), 100)

	// Calculate remaining percentage for gauge display
	remainingPercent := 100 - percent

	// 5 blocks total, calculate how many should be filled (representing remaining capacity)
	filledBlocks := (remainingPercent * 5) / 100
	emptyBlocks := 5 - filledBlocks

	// Use ▰ for filled blocks (remaining) and ▱ for empty blocks (used)
	return strings.Repeat("▰", filledBlocks) + strings.Repeat("▱", emptyBlocks)
}

// String returns the percentage as a string without % sign for template compatibility.
func (p Percentage) String() string {
	return fmt.Sprintf("%d", int(p))
}
