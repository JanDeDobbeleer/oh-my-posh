package main

import (
	"strconv"

	mem "github.com/jan0660/memory"
)

type memory struct {
	props       *properties
	env         environmentInfo
	TotalMemory uint64
	FreeMemory  uint64
}

const (
	// Precision number of decimal places to show
	Precision Property = "precision"
	// UseAvailable if available memory should be used instead of free on Linux
	UseAvailable Property = "use_available"
	// MemoryType either physical or swap
	MemoryType Property = "memory_type"
)

func (n *memory) enabled() bool {
	if n.TotalMemory == 0 || n.FreeMemory == 0 {
		// failed to get memory information
		return false
	}
	return true
}

func (n *memory) string() string {
	// 100.0 / total * used
	percentage := 100.0 / float64(n.TotalMemory) * float64(n.TotalMemory-n.FreeMemory)
	text := strconv.FormatFloat(percentage, 'f', n.props.getInt(Precision, 0), 64)
	return text
}

func (n *memory) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
	if props.getString(MemoryType, "physical") == "physical" {
		n.TotalMemory = mem.TotalMemory()
		if props.getBool(UseAvailable, true) {
			n.FreeMemory = mem.AvailableMemory()
			return
		}
		n.FreeMemory = mem.FreeMemory()
		return
	}
	n.TotalMemory = mem.TotalSwap()
	n.FreeMemory = mem.FreeSwap()
}
