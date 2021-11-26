package main

import (
	cpu "github.com/shirou/gopsutil/v3/cpu"
	load "github.com/shirou/gopsutil/v3/load"
	mem "github.com/shirou/gopsutil/v3/mem"
)

type sysinfo struct {
	props     properties
	env       environmentInfo
	Precision int
	// mem
	PhysicalTotalMemory uint64
	PhysicalFreeMemory  uint64
	PhysicalPercentUsed float64
	SwapTotalMemory     uint64
	SwapFreeMemory      uint64
	SwapPercentUsed     float64
	// cpu
	Times float64
	CPU   []cpu.InfoStat
	// load
	Load1  float64
	Load5  float64
	Load15 float64
}

const (
	// Precision number of decimal places to show
	Precision Property = "precision"
)

func (s *sysinfo) enabled() bool {
	if s.PhysicalPercentUsed == 0 && s.SwapPercentUsed == 0 {
		return false
	}
	return true
}

func (s *sysinfo) string() string {
	// keep old memory segment template
	segmentTemplate := s.props.getString(SegmentTemplate, "{{ round .PhysicalPercentUsed .Precision }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  s,
		Env:      s.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (s *sysinfo) init(props properties, env environmentInfo) {
	s.props = props
	s.env = env
	s.Precision = s.props.getInt(Precision, 2)
	// mem
	memStat, err := mem.VirtualMemory()
	if err == nil {
		s.PhysicalTotalMemory = memStat.Total
		s.PhysicalFreeMemory = memStat.Free
		s.PhysicalPercentUsed = memStat.UsedPercent
	}
	swapStat, err := mem.SwapMemory()
	if err == nil {
		s.SwapTotalMemory = swapStat.Total
		s.SwapFreeMemory = swapStat.Free
		s.SwapPercentUsed = swapStat.UsedPercent
	}
	// load
	loadStat, err := load.Avg()
	if err == nil {
		s.Load1 = loadStat.Load1
		s.Load5 = loadStat.Load5
		s.Load15 = loadStat.Load15
	}
	// times
	processorTimes, err := cpu.Percent(0, false)
	if err == nil {
		s.Times = processorTimes[0]
	}
	// cpu
	processors, err := cpu.Info()
	if err == nil {
		s.CPU = processors
	}
}
