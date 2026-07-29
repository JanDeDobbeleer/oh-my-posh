package text

import (
	"fmt"
	"strings"
)

type Percentage int

func (p Percentage) clamp() int {
	return min(max(int(p), 0), 100)
}

// Shows remaining capacity, so 20% used displays 4 filled blocks (80% remaining).
func (p Percentage) Gauge() string {
	return p.GaugeWith("▰", "▱")
}

// Shows remaining capacity: marked fills the remaining blocks, unmarked the used ones.
func (p Percentage) GaugeWith(marked, unmarked string) string {
	percent := p.clamp()

	remainingPercent := 100 - percent
	filledBlocks := (remainingPercent * 5) / 100
	emptyBlocks := 5 - filledBlocks

	return strings.Repeat(marked, filledBlocks) + strings.Repeat(unmarked, emptyBlocks)
}

// Shows used capacity, so 20% used displays 1 filled block.
func (p Percentage) GaugeUsed() string {
	return p.GaugeUsedWith("▰", "▱")
}

// Shows used capacity: marked fills the used blocks, unmarked the remaining ones.
func (p Percentage) GaugeUsedWith(marked, unmarked string) string {
	percent := p.clamp()

	filledBlocks := (percent * 5) / 100
	emptyBlocks := 5 - filledBlocks

	return strings.Repeat(marked, filledBlocks) + strings.Repeat(unmarked, emptyBlocks)
}

// Without a % sign, for template compatibility.
func (p Percentage) String() string {
	return fmt.Sprintf("%d", int(p))
}
