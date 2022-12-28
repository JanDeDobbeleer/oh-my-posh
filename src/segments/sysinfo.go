package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"

	cpu "github.com/shirou/gopsutil/v3/cpu"
	disk "github.com/shirou/gopsutil/v3/disk"
	load "github.com/shirou/gopsutil/v3/load"
	mem "github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	props properties.Properties
	env   platform.Environment

	Precision int
	// mem
	PhysicalTotalMemory     uint64
	PhysicalAvailableMemory uint64
	PhysicalFreeMemory      uint64
	PhysicalPercentUsed     float64
	SwapTotalMemory         uint64
	SwapFreeMemory          uint64
	SwapPercentUsed         float64
	// cpu
	Times float64
	CPU   []cpu.InfoStat
	// load
	Load1  float64
	Load5  float64
	Load15 float64
	// disk
	Disks map[string]disk.IOCountersStat
}

const (
	// Precision number of decimal places to show
	Precision properties.Property = "precision"
)

func (s *SystemInfo) Template() string {
	return " {{ round .PhysicalPercentUsed .Precision }} "
}

func (s *SystemInfo) Enabled() bool {
	if s.PhysicalPercentUsed == 0 && s.SwapPercentUsed == 0 {
		return false
	}
	return true
}

func (s *SystemInfo) Init(props properties.Properties, env platform.Environment) {
	s.props = props
	s.env = env
	s.Precision = s.props.GetInt(Precision, 2)
	// mem
	memStat, err := mem.VirtualMemory()
	if err == nil {
		s.PhysicalTotalMemory = memStat.Total
		s.PhysicalAvailableMemory = memStat.Available
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
	if err == nil && len(processorTimes) > 0 {
		s.Times = processorTimes[0]
	}
	// cpu
	processors, err := cpu.Info()
	if err == nil {
		s.CPU = processors
	}
	// disk
	diskIO, err := disk.IOCounters()
	if err == nil {
		s.Disks = diskIO
	}
}
