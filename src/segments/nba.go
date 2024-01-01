package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

// segment struct, makes templating easier
type Nba struct {
	props properties.Properties
	env   platform.Environment

	NBAData
}

// NBA struct contains parsed API data that care about for the segment
type NBAData struct {
	HomeTeam       string
	AwayTeam       string
	Time           string
	GameDate       string
	StartTimeUTC   string
	GameStatus     GameStatus // 1 = scheduled, 2 = in progress, 3 = finished
	HomeScore      int
	AwayScore      int
	HomeTeamWins   int
	HomeTeamLosses int
	AwayTeamWins   int
	AwayTeamLosses int
}

func (nba *NBAData) HasStats() bool {
	return nba.HomeTeamWins != 0 || nba.HomeTeamLosses != 0 || nba.AwayTeamWins != 0 || nba.AwayTeamLosses != 0
}

func (nba *NBAData) Started() bool {
	return nba.GameStatus == InProgress || nba.GameStatus == Finished
}

const (
	NBASeason  properties.Property = "season"
	TeamName   properties.Property = "team"
	DaysOffset properties.Property = "days_offset"

	ScheduledTemplate  properties.Property = "scheduled_template"
	InProgressTemplate properties.Property = "in_progress_template"
	FinishedTemplate   properties.Property = "finished_template"

	NBAScoreURL    string = "https://cdn.nba.com/static/json/liveData/scoreboard/todaysScoreboard_00.json"
	NBAScheduleURL string = "https://stats.nba.com/stats/internationalbroadcasterschedule?LeagueID=00&Season=%s&Date=%s&RegionID=1&EST=Y"

	Unknown = "Unknown"

	currentNBASeason = "2023"
	NBADateFormat    = "02/01/2006"
)

// Custom type for GameStatus
type GameStatus int

// Constants for GameStatus values
const (
	Scheduled  GameStatus = 1
	InProgress GameStatus = 2
	Finished   GameStatus = 3
	NotFound   GameStatus = 4
)

// Int() method for GameStatus to get its integer representation
// This is a helpful method if people want to come up with their own templates
func (gs GameStatus) Int() int {
	return int(gs)
}

func (gs GameStatus) Valid() bool {
	return gs == Scheduled || gs == InProgress || gs == Finished
}

func (gs GameStatus) String() string {
	switch gs {
	case Scheduled:
		return "Scheduled"
	case InProgress:
		return "In Progress"
	case Finished:
		return "Finished"
	case NotFound:
		return "Not Found"
	default:
		return Unknown
	}
}

// All of the structs needed to retrieve data from the live score endpoint
type ScoreboardResponse struct {
	Scoreboard Scoreboard `json:"scoreboard"`
}

type Scoreboard struct {
	GameDate string `json:"gameDate"`
	Games    []Game `json:"games"`
}

type Game struct {
	GameStatus     int    `json:"gameStatus"`
	GameStatusText string `json:"gameStatusText"`
	GameTimeUTC    string `json:"gameTimeUTC"`
	HomeTeam       Team   `json:"homeTeam"`
	AwayTeam       Team   `json:"awayTeam"`
}

type Team struct {
	TeamTricode string `json:"teamTricode"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
	Score       int    `json:"score"`
}

// All the structs needed to get data from the schedule endpoint
type ScheduleResponse struct {
	ResultSets []ResultSet `json:"resultSets"`
}

type ResultSet struct {
	CompleteGameList []ScheduledGame `json:"CompleteGameList,omitempty"`
}

type ScheduledGame struct {
	VtAbbreviation string `json:"vtAbbreviation"`
	HtAbbreviation string `json:"htAbbreviation"`
	Date           string `json:"date"`
	Time           string `json:"time"`
}

func (nba *Nba) Template() string {
	return " \U000F0806 {{ .HomeTeam}}{{ if .HasStats }} ({{.HomeTeamWins}}-{{.HomeTeamLosses}}){{ end }}{{ if .Started }}:{{.HomeScore}}{{ end }} vs {{ .AwayTeam}}{{ if .HasStats }} ({{.AwayTeamWins}}-{{.AwayTeamLosses}}){{ end }}{{ if .Started }}:{{.AwayScore}}{{ end }} | {{ if not .Started }}{{.GameDate}} | {{ end }}{{.Time}} " //nolint:lll
}

func (nba *Nba) Enabled() bool {
	data, err := nba.getResult()
	if err != nil || !data.GameStatus.Valid() {
		return false
	}

	nba.NBAData = *data

	return true
}

// Returns an empty Game Data struct with the GameStatus set to NotFound
// Helpful for caching the fact that a game was not found for a team
func (nba *Nba) getGameNotFoundData() string {
	return `{
		"HomeTeam":"",
		"AwayTeam":"",
		"Time":"",
		"GameDate":"",
		"StartTimeUTC":"",
		"GameStatus":4,
		"HomeScore":0,
		"AwayScore":0,
		"HomeTeamWins":0,
		"HomeTeamLosses":0,
		"AwayTeamWins":0,
		"AwayTeamLosses":0
	}`
}

// parses through a set of games from the score endpoint and looks for props.team in away or home team
func (nba *Nba) findGameScoreByTeamTricode(games []Game, teamTricode string) (*Game, error) {
	for _, game := range games {
		if game.HomeTeam.TeamTricode == teamTricode || game.AwayTeam.TeamTricode == teamTricode {
			return &game, nil
		}
	}

	return nil, errors.New("no game score found for team")
}

// parses through a set of games from the schedule endpoint and looks for props.team in away or home team
func (nba *Nba) findGameSchedulebyTeamTricode(games []ScheduledGame, teamTricode string) (*ScheduledGame, error) {
	for _, game := range games {
		if game.VtAbbreviation == teamTricode || game.HtAbbreviation == teamTricode {
			return &game, nil
		}
	}

	return nil, errors.New("no scheduled game found for team")
}

// parses the time and date from the schedule endpoint into a UTC time
func (nba *Nba) parseTimetoUTC(timeEST, date string) string {
	combinedTime := date + " " + timeEST
	timeUTC, err := time.Parse("01/02/2006 03:04 PM", combinedTime)
	if err != nil {
		return ""
	}

	return timeUTC.UTC().Format("2006-01-02T15:04:05Z")
}

// retrieves data from the score endpoint
func (nba *Nba) retrieveScoreData(teamName string, httpTimeout int) (*NBAData, error) {
	body, err := nba.env.HTTPRequest(NBAScoreURL, nil, httpTimeout)
	if err != nil {
		return nil, err
	}

	var scoreboardResponse ScoreboardResponse
	err = json.Unmarshal(body, &scoreboardResponse)
	if err != nil {
		return nil, err
	}

	gameInfo, err := nba.findGameScoreByTeamTricode(scoreboardResponse.Scoreboard.Games, teamName)
	if err != nil {
		return nil, err
	}

	return &NBAData{
		AwayTeam:       gameInfo.AwayTeam.TeamTricode,
		HomeTeam:       gameInfo.HomeTeam.TeamTricode,
		Time:           gameInfo.GameStatusText,
		GameDate:       scoreboardResponse.Scoreboard.GameDate,
		StartTimeUTC:   gameInfo.GameTimeUTC,
		GameStatus:     GameStatus(gameInfo.GameStatus),
		HomeScore:      gameInfo.HomeTeam.Score,
		AwayScore:      gameInfo.AwayTeam.Score,
		HomeTeamWins:   gameInfo.HomeTeam.Wins,
		HomeTeamLosses: gameInfo.HomeTeam.Losses,
		AwayTeamWins:   gameInfo.AwayTeam.Wins,
		AwayTeamLosses: gameInfo.AwayTeam.Losses,
	}, nil
}

// Retrieves the data from the schedule endpoint
func (nba *Nba) retrieveScheduleData(teamName string, httpTimeout int) (*NBAData, error) {
	// How many days into the future should we look for a game.
	numDaysToSearch := nba.props.GetInt(DaysOffset, 8)
	nbaSeason := nba.props.GetString(NBASeason, currentNBASeason)
	// Get the current date in America/New_York
	nowNYC := time.Now().In(time.FixedZone("America/New_York", -5*60*60))

	// Check to see if a game is scheduled while the numDaysToSearch is greater than 0
	for numDaysToSearch > 0 {
		dateStr := nowNYC.Format(NBADateFormat)
		urlEndpoint := fmt.Sprintf(NBAScheduleURL, nbaSeason, dateStr)

		body, err := nba.env.HTTPRequest(urlEndpoint, nil, httpTimeout)
		if err != nil {
			return nil, err
		}

		var scheduleResponse *ScheduleResponse
		err = json.Unmarshal(body, &scheduleResponse)
		if err != nil {
			return nil, err
		}

		// Check if we can find a game for the team
		gameInfo, err := nba.findGameSchedulebyTeamTricode(scheduleResponse.ResultSets[1].CompleteGameList, teamName)
		if err != nil {
			// We didn't find a game for the team on this day, so we need to check the next day
			nowNYC = nowNYC.AddDate(0, 0, 1)
			numDaysToSearch--
			continue
		}

		return &NBAData{
			AwayTeam:       gameInfo.VtAbbreviation,
			HomeTeam:       gameInfo.HtAbbreviation,
			Time:           gameInfo.Time + " ET",
			GameDate:       gameInfo.Date,
			StartTimeUTC:   nba.parseTimetoUTC(gameInfo.Time, gameInfo.Date),
			GameStatus:     Scheduled,
			HomeScore:      0,
			AwayScore:      0,
			HomeTeamWins:   0,
			HomeTeamLosses: 0,
			AwayTeamWins:   0,
			AwayTeamLosses: 0,
		}, nil
	}

	return nil, errors.New("no scheduled game found for team within DaysOffset days")
}

// First try to get the data from the score endpoint, if that fails, try the schedule endpoint
// The score endpoint usually goes live within 12 hours of a game starting
func (nba *Nba) getAvailableGameData(teamName string, httpTimeout int) (*NBAData, error) {
	// Get the info from the score endpoint
	data, err := nba.retrieveScoreData(teamName, httpTimeout)
	if err == nil {
		return data, nil
	}

	// If the score endpoint doesn't have anything get data from the schedule endpoint
	data, err = nba.retrieveScheduleData(teamName, httpTimeout)
	if err == nil {
		return data, nil
	}

	return nil, err
}

// Gets the data from the cache if it exists
func (nba *Nba) getCacheValue(key string) (*NBAData, error) {
	if val, found := nba.env.Cache().Get(key); found {
		var nbaData *NBAData
		err := json.Unmarshal([]byte(val), &nbaData)
		if err != nil {
			return nil, err
		}
		return nbaData, nil
	}

	return nil, errors.New("no data in cache")
}

// Gets the data from the cache for a scheduled game if it exists
// Checks whether the game should have started and if so, removes the cache entry
func (nba *Nba) getCachedScheduleValue(key string) (*NBAData, error) {
	data, err := nba.getCacheValue(key)
	if err != nil {
		return nil, errors.New("no data in cache")
	}

	// check if the game was previously not found and we should wait to check again
	if data.GameStatus == NotFound {
		return data, nil
	}

	// check if the current time is after the start time of the game
	// if so, we need to refresh the data
	startTime, err := time.Parse("2006-01-02T15:04:05Z", data.StartTimeUTC)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(startTime) {
		// remove the cache entry
		nba.env.Cache().Delete(key)
		return nil, errors.New("game has already started")
	}

	return data, nil
}

func (nba *Nba) getResult() (*NBAData, error) {
	teamName := nba.props.GetString(TeamName, "")

	cachedScheduleKey := fmt.Sprintf("%s%s", teamName, "schedule")
	cachedScoreKey := fmt.Sprintf("%s%s", teamName, "score")

	httpTimeout := nba.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	// How often you want to query the API to get live score information, defaults to 2 minutes
	cacheScoreTimeout := nba.props.GetInt(properties.CacheTimeout, 2)

	// Cache the schedule information for a day so we don't call the API too often
	cacheScheduleTimeout := nba.props.GetInt(properties.CacheTimeout, 1440)

	// Cache the fact a game was not found for 30 minutes so we don't call the API too often
	cacheNotFoundTimeout := nba.props.GetInt(properties.CacheTimeout, 30)

	nba.env.Debug("Validating cache data for " + teamName)

	if cacheScheduleTimeout > 0 {
		if data, err := nba.getCachedScheduleValue(cachedScheduleKey); err == nil {
			return data, nil
		}
	}

	if cacheScoreTimeout > 0 {
		if data, err := nba.getCacheValue(cachedScoreKey); err == nil {
			return data, nil
		}
	}

	nba.env.Debug("Fetching available data for " + teamName)

	data, err := nba.getAvailableGameData(teamName, httpTimeout)
	if err != nil {
		// cache the fact that we didn't find a game yet for the day for 30m so we don't continuously ping the endpoints
		nba.env.Cache().Set(cachedScheduleKey, nba.getGameNotFoundData(), cacheNotFoundTimeout)
		nba.env.Error(errors.Join(err, fmt.Errorf("unable to get data for team %s", teamName)))
		return nil, err
	}

	if !data.GameStatus.Valid() {
		err := fmt.Errorf("%d is not a valid game status", data.GameStatus)
		nba.env.Error(err)
		return nil, err
	}

	if cacheScheduleTimeout > 0 && data.GameStatus == Scheduled {
		// persist data for team in cache
		cachedData, _ := json.Marshal(data)
		nba.env.Cache().Set(cachedScheduleKey, string(cachedData), cacheScheduleTimeout)
	}

	// if the game is in progress or finished, we can cache the score
	if cacheScoreTimeout > 0 && data.GameStatus == InProgress || data.GameStatus == Finished {
		// persist data for team in cache
		cachedData, _ := json.Marshal(data)
		nba.env.Cache().Set(cachedScoreKey, string(cachedData), cacheScoreTimeout)
	}

	return data, nil
}

func (nba *Nba) Init(props properties.Properties, env platform.Environment) {
	nba.props = props
	nba.env = env
}
