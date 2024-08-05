package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"

	"github.com/alecthomas/assert"
)

func TestUpgrade(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		CurrentVersion  string
		LatestVersion   string
		CachedVersion   string
		ExpectedEnabled bool
		HasCache        bool
	}{
		{
			Case:            "Should upgrade",
			CurrentVersion:  "1.0.0",
			LatestVersion:   "1.0.1",
			ExpectedEnabled: true,
		},
		{
			Case:           "On latest",
			CurrentVersion: "1.0.1",
			LatestVersion:  "1.0.1",
		},
		{
			Case:  "Error on update check",
			Error: errors.New("error"),
		},
		{
			Case:           "On latest, version changed",
			CurrentVersion: "1.0.2",
			LatestVersion:  "1.0.2",
		},
		{
			Case:            "On previous, version changed",
			CurrentVersion:  "1.0.2",
			LatestVersion:   "1.0.3",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		build.Version = tc.CurrentVersion

		json := fmt.Sprintf(`{"tag_name":"v%s"}`, tc.LatestVersion)
		env.On("HTTPRequest", upgrade.RELEASEURL).Return([]byte(json), tc.Error)

		ug := &Upgrade{
			env:   env,
			props: properties.Map{},
		}

		enabled := ug.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
	}
}
