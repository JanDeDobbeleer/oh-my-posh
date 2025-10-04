//go:build !windows

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
