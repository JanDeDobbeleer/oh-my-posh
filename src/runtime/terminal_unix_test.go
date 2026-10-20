//go:build !windows && !js

package runtime

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
)

func TestSetSwapMemory(t *testing.T) {
	cases := []struct {
		SwapStat            *mem.SwapMemoryStat
		MemoryStat          *mem.VirtualMemoryStat
		Name                string
		ExpectedSwapTotal   uint64
		ExpectedSwapFree    uint64
		ExpectedSwapPercent float64
	}{
		{
			Name: "swap stats are available",
			SwapStat: &mem.SwapMemoryStat{
				Total:       8 * 1024,
				Free:        4 * 1024,
				UsedPercent: 50,
			},
			MemoryStat: &mem.VirtualMemoryStat{
				SwapTotal: 16 * 1024,
				SwapFree:  8 * 1024,
			},
			ExpectedSwapTotal:   8 * 1024,
			ExpectedSwapFree:    4 * 1024,
			ExpectedSwapPercent: 50,
		},
		{
			Name: "virtual memory swap stats are used when swap stats are unavailable",
			MemoryStat: &mem.VirtualMemoryStat{
				SwapTotal: 8 * 1024,
				SwapFree:  2 * 1024,
			},
			ExpectedSwapTotal:   8 * 1024,
			ExpectedSwapFree:    2 * 1024,
			ExpectedSwapPercent: 75,
		},
		{
			Name:       "no swap remains zero when swap stats are unavailable",
			MemoryStat: &mem.VirtualMemoryStat{},
		},
		{
			Name: "free swap greater than total does not underflow",
			MemoryStat: &mem.VirtualMemoryStat{
				SwapTotal: 2 * 1024,
				SwapFree:  4 * 1024,
			},
			ExpectedSwapTotal:   2 * 1024,
			ExpectedSwapFree:    4 * 1024,
			ExpectedSwapPercent: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			memory := &Memory{}

			setSwapMemory(memory, tc.SwapStat, tc.MemoryStat)

			assert.Equal(t, tc.ExpectedSwapTotal, memory.SwapTotalMemory)
			assert.Equal(t, tc.ExpectedSwapFree, memory.SwapFreeMemory)
			assert.InDelta(t, tc.ExpectedSwapPercent, memory.SwapPercentUsed, 0.01)
		})
	}
}

func TestMemoryPercentageCalculation(t *testing.T) {
	cases := []struct {
		Name            string
		Total           uint64
		Available       uint64
		ExpectedPercent float64
	}{
		{
			Name:            "50% usage",
			Total:           8 * 1024 * 1024 * 1024,
			Available:       4 * 1024 * 1024 * 1024,
			ExpectedPercent: 50.0,
		},
		{
			Name:            "37% usage (from issue)",
			Total:           8079691776,
			Available:       5093384192,
			ExpectedPercent: 36.96,
		},
		{
			Name:            "25% usage",
			Total:           16 * 1024 * 1024 * 1024,
			Available:       12 * 1024 * 1024 * 1024,
			ExpectedPercent: 25.0,
		},
		{
			Name:            "75% usage",
			Total:           8 * 1024 * 1024 * 1024,
			Available:       2 * 1024 * 1024 * 1024,
			ExpectedPercent: 75.0,
		},
		{
			Name:            "0% usage",
			Total:           8 * 1024 * 1024 * 1024,
			Available:       8 * 1024 * 1024 * 1024,
			ExpectedPercent: 0.0,
		},
		{
			Name:            "100% usage",
			Total:           8 * 1024 * 1024 * 1024,
			Available:       0,
			ExpectedPercent: 100.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var percentUsed float64
			if tc.Total > 0 {
				percentUsed = float64(tc.Total-tc.Available) / float64(tc.Total) * 100
			}

			assert.InDelta(t, tc.ExpectedPercent, percentUsed, 0.01, tc.Name)
		})
	}
}

func TestResolveTerminalWidth(t *testing.T) {
	terminalErr := errors.New("not a tty")
	cases := []struct {
		TerminalError   error
		Name            string
		Columns         string
		TerminalWidth   uint
		ExpectedWidth   uint
		ExpectedToError bool
	}{
		{
			Name:          "terminal width is available",
			Columns:       "120",
			TerminalWidth: 80,
			ExpectedWidth: 80,
		},
		{
			Name:          "COLUMNS replaces terminal error",
			Columns:       "120",
			TerminalError: terminalErr,
			ExpectedWidth: 120,
		},
		{
			Name:            "terminal and COLUMNS both fail",
			Columns:         "invalid",
			TerminalError:   terminalErr,
			ExpectedToError: true,
		},
		{
			Name:            "zero COLUMNS is invalid",
			Columns:         "0",
			TerminalError:   terminalErr,
			ExpectedToError: true,
		},
		{
			Name:            "negative COLUMNS is invalid",
			Columns:         "-1",
			TerminalError:   terminalErr,
			ExpectedToError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			width, err := resolveTerminalWidth(tc.TerminalWidth, tc.TerminalError, tc.Columns)

			assert.Equal(t, tc.ExpectedWidth, width)
			if tc.ExpectedToError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, terminalErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
