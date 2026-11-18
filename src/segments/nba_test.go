package segments

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestData(file string) string {
	content, _ := os.ReadFile(fmt.Sprintf("../test/%s", file))
	return string(content)
}

func TestNBASegment(t *testing.T) {
	jsonScheduleData := getTestData("nba/schedule.json")
	jsonScoreData := getTestData("nba/score.json")
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	cases := []struct {
		Error           error
		Case            string
		JSONResponse    string
		ExpectedString  string
		TeamName        string
		CacheTimeout    int
		DaysOffset      int
		ExpectedEnabled bool
		CacheFoundFail  bool
	}{
		{
			Case:            "Team (Home Team) Scheduled Game",
			JSONResponse:    jsonScheduleData,
			TeamName:        "LAL",
			ExpectedString:  "󰠆 LAL vs PHX | 10/26/2023 | 10:00 PM ET",
			ExpectedEnabled: true,
			DaysOffset:      8,
		},
		{
			Case:            "Team (Away Team) Scheduled Game",
			JSONResponse:    jsonScheduleData,
			TeamName:        "PHX",
			ExpectedString:  "󰠆 LAL vs PHX | 10/26/2023 | 10:00 PM ET",
			DaysOffset:      4,
			ExpectedEnabled: true,
		},
		{
			Case:            "Team (Home Team) Live Game",
			JSONResponse:    jsonScoreData,
			TeamName:        "CHA",
			ExpectedString:  "󰠆 CHA (1-0):13 vs BOS (0-1):8 | Q1 8:23",
			ExpectedEnabled: true,
		},
		{
			Case:            "Team (Away Team) Live Game",
			JSONResponse:    jsonScoreData,
			TeamName:        "BOS",
			ExpectedString:  "󰠆 CHA (1-0):13 vs BOS (0-1):8 | Q1 8:23",
			ExpectedEnabled: true,
		},
		{
			Case:            "Team not Found",
			JSONResponse:    jsonScheduleData,
			DaysOffset:      8,
			TeamName:        "INVALID",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := options.Map{
			TeamName:   tc.TeamName,
			DaysOffset: tc.DaysOffset,
		}

		env.On("HTTPRequest", NBAScoreURL).Return([]byte(tc.JSONResponse), tc.Error)

		nowNYC := time.Now().In(loc)
		// The schedule search keeps its starting season even when dates cross into July.
		season := currentNBASeason(nowNYC)
		for i := 0; i < tc.DaysOffset; i++ {
			dateStr := nowNYC.AddDate(0, 0, i).Format(NBADateFormat)
			scheduleURLEndpoint := fmt.Sprintf(NBAScheduleURL, season, dateStr)
			env.On("HTTPRequest", scheduleURLEndpoint).Return([]byte(tc.JSONResponse), tc.Error)
		}

		nba := &Nba{}
		nba.Init(props, env)

		enabled := nba.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(env, nba.Template(), nba), tc.Case)
	}
}
