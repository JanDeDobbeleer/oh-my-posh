package text

import (
	"fmt"
	"strings"
)

// Percentage represents a percentage value with gauge visualization.
type Percentage int

// clamp ensures the percentage value is within the valid range of 0-100.
func (p Percentage) clamp() int {
	return min(max(int(p), 0), 100)
}

// Gauge returns a 5-character gauge visualization showing remaining capacity (▰▰▰▰▱ style).
// The gauge displays remaining capacity, so 20% used shows 4 filled blocks (80% remaining).
func (p Percentage) Gauge() string {
	return p.GaugeWith("▰", "▱")
}

// GaugeWith returns a 5-character gauge visualization showing remaining capacity using custom characters.
// marked is the character for filled (remaining) blocks and unmarked is the character for empty (used) blocks.
func (p Percentage) GaugeWith(marked, unmarked string) string {
	percent := p.clamp()

	remainingPercent := 100 - percent
	filledBlocks := (remainingPercent * 5) / 100
	emptyBlocks := 5 - filledBlocks

	return strings.Repeat(marked, filledBlocks) + strings.Repeat(unmarked, emptyBlocks)
}

// GaugeUsed returns a 5-character gauge visualization showing used capacity (▰▱▱▱▱ style).
// The gauge displays used capacity, so 20% used shows 1 filled block (▰▱▱▱▱).
func (p Percentage) GaugeUsed() string {
	return p.GaugeUsedWith("▰", "▱")
}

// GaugeUsedWith returns a 5-character gauge visualization showing used capacity using custom characters.
// marked is the character for filled (used) blocks and unmarked is the character for empty (remaining) blocks.
func (p Percentage) GaugeUsedWith(marked, unmarked string) string {
	percent := p.clamp()

	filledBlocks := (percent * 5) / 100
	emptyBlocks := 5 - filledBlocks

	return strings.Repeat(marked, filledBlocks) + strings.Repeat(unmarked, emptyBlocks)
}

// String returns the percentage as a string without % sign for template compatibility.
func (p Percentage) String() string {
	return fmt.Sprintf("%d", int(p))
}
