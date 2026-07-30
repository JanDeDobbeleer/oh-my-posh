//go:build !js

package runtime

import (
	load "github.com/shirou/gopsutil/v4/load"

	disk "github.com/shirou/gopsutil/v4/disk"
)

func (term *Terminal) SystemInfo() (*SystemInfo, error) {
	s := &SystemInfo{}

	mem, err := term.Memory()
	if err != nil {
		return nil, err
	}
	s.Memory = *mem

	loadStat, err := load.Avg()
	if err == nil {
		s.Load1 = loadStat.Load1
		s.Load5 = loadStat.Load5
		s.Load15 = loadStat.Load15
	}

	diskIO, err := disk.IOCounters()
	if err == nil {
		s.Disks = diskIO
	}
	return s, nil
}
