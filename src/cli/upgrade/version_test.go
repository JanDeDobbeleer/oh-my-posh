package upgrade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMajorUpgrade(t *testing.T) {
	cases := []struct {
		Case           string
		CurrentVersion string
		LatestVersion  string
		Expected       bool
	}{
		{Case: "Same version", Expected: false, CurrentVersion: "v3.0.0", LatestVersion: "v3.0.0"},
		{Case: "Breaking change", Expected: true, CurrentVersion: "v3.0.0", LatestVersion: "v4.0.0"},
		{Case: "Empty version, mostly development build", Expected: false, LatestVersion: "v4.0.0"},
	}

	for _, tc := range cases {
		canUpgrade := IsMajorUpgrade(tc.CurrentVersion, tc.LatestVersion)
		assert.Equal(t, tc.Expected, canUpgrade, tc.Case)
	}
}
