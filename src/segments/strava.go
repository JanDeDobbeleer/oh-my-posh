package segments

import (
	"fmt"
	"math"
	"oh-my-posh/environment"
	"oh-my-posh/http"
	"oh-my-posh/properties"
	"time"
)

// StravaAPI is a wrapper around http.Oauth
type StravaAPI interface {
	GetActivities() ([]*StravaData, error)
}

type stravaAPI struct {
	http.OAuth
}

func (s *stravaAPI) GetActivities() ([]*StravaData, error) {
	url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
	return http.OauthResult[[]*StravaData](&s.OAuth, url)
}

// segment struct, makes templating easier
type Strava struct {
	props properties.Properties

	StravaData
	Icon         string
	Ago          string
	Hours        int
	Authenticate bool
	Error        string
	URL          string

	api StravaAPI
}

const (
	RideIcon            properties.Property = "ride_icon"
	RunIcon             properties.Property = "run_icon"
	SkiingIcon          properties.Property = "skiing_icon"
	WorkOutIcon         properties.Property = "workout_icon"
	UnknownActivityIcon properties.Property = "unknown_activity_icon"

	StravaAccessTokenKey  = "strava_access_token"
	StravaRefreshTokenKey = "strava_refresh_token"

	noActivitiesFound = "No activities found"
)

// StravaData struct contains the API data
type StravaData struct {
	ID                   int       `json:"id"`
	Type                 string    `json:"type"`
	StartDate            time.Time `json:"start_date"`
	Name                 string    `json:"name"`
	Distance             float64   `json:"distance"`
	Duration             float64   `json:"moving_time"`
	DeviceWatts          bool      `json:"device_watts"`
	AverageWatts         float64   `json:"average_watts"`
	WeightedAverageWatts float64   `json:"weighted_average_watts"`
	AverageHeartRate     float64   `json:"average_heartrate"`
	MaxHeartRate         float64   `json:"max_heartrate"`
	KudosCount           int       `json:"kudos_count"`
}

func (s *Strava) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ .Ago }}{{ end }} "
}

func (s *Strava) Enabled() bool {
	data, err := s.api.GetActivities()
	if err == nil && len(data) > 0 {
		s.StravaData = *data[0]
		s.Icon = s.getActivityIcon()
		s.Hours = s.getHours()
		s.Ago = s.getAgo()
		s.URL = fmt.Sprintf("https://www.strava.com/activities/%d", s.ID)
		return true
	}
	if err == nil && len(data) == 0 {
		s.Error = noActivitiesFound
		return true
	}
	if _, s.Authenticate = err.(*http.OAuthError); s.Authenticate {
		s.Error = err.(*http.OAuthError).Error()
		return true
	}
	return false
}

func (s *Strava) getHours() int {
	hours := time.Since(s.StartDate).Hours()
	return int(math.Floor(hours))
}

func (s *Strava) getAgo() string {
	if s.Hours > 24 {
		days := int32(math.Floor(float64(s.Hours) / float64(24)))
		return fmt.Sprintf("%d", days) + string('d')
	}
	return fmt.Sprintf("%d", s.Hours) + string("h")
}

func (s *Strava) getActivityIcon() string {
	switch s.Type {
	case "VirtualRide":
		fallthrough
	case "Ride":
		return s.props.GetString(RideIcon, "\uf5a2")
	case "Run":
		return s.props.GetString(RunIcon, "\ufc0c")
	case "NordicSki":
	case "AlpineSki":
	case "BackcountrySki":
		return s.props.GetString(SkiingIcon, "\ue213")
	case "WorkOut":
		return s.props.GetString(WorkOutIcon, "\ue213")
	default:
		return s.props.GetString(UnknownActivityIcon, "\ue213")
	}
	return s.props.GetString(UnknownActivityIcon, "\ue213")
}

func (s *Strava) Init(props properties.Properties, env environment.Environment) {
	s.props = props

	s.api = &stravaAPI{
		OAuth: http.OAuth{
			Props:           props,
			Env:             env,
			AccessTokenKey:  StravaAccessTokenKey,
			RefreshTokenKey: StravaRefreshTokenKey,
			SegmentName:     "strava",
		},
	}
}
