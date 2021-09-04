package main

import (
	"strconv"

	mem "github.com/pbnjay/memory"
)

type memory struct {
	props       *properties
	env         environmentInfo
	TotalMemory uint64
	FreeMemory  uint64
}

const (
	Precision Property = "precision"
)

func (n *memory) enabled() bool {
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
	n.TotalMemory = mem.TotalMemory()
	n.FreeMemory = mem.FreeMemory()
}
