package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"
)

// segment struct, makes templating easier
type strava struct {
	props Properties
	env   Environment

	StravaData
	Icon         string
	Ago          string
	Hours        int
	Authenticate bool
	Error        string
	URL          string
}

const (
	RideIcon            Property = "ride_icon"
	RunIcon             Property = "run_icon"
	SkiingIcon          Property = "skiing_icon"
	WorkOutIcon         Property = "workout_icon"
	UnknownActivityIcon Property = "unknown_activity_icon"

	StravaAccessToken  = "strava_access_token"
	StravaRefreshToken = "strava_refresh_token"

	Timeout             = "timeout"
	InvalidRefreshToken = "invalid refresh token"
	TokenRefreshFailed  = "token refresh error"
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

type TokenExchange struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthError struct {
	message string
}

func (a *AuthError) Error() string {
	return a.message
}

func (s *strava) enabled() bool {
	data, err := s.getResult()
	if err == nil {
		s.StravaData = *data
		s.Icon = s.getActivityIcon()
		s.Hours = s.getHours()
		s.Ago = s.getAgo()
		return true
	}
	if _, s.Authenticate = err.(*AuthError); s.Authenticate {
		s.Error = err.(*AuthError).Error()
		return true
	}
	return false
}

func (s *strava) getHours() int {
	hours := time.Since(s.StartDate).Hours()
	return int(math.Floor(hours))
}

func (s *strava) getAgo() string {
	if s.Hours > 24 {
		days := int32(math.Floor(float64(s.Hours) / float64(24)))
		return fmt.Sprintf("%d", days) + string('d')
	}
	return fmt.Sprintf("%d", s.Hours) + string("h")
}

func (s *strava) getActivityIcon() string {
	switch s.Type {
	case "VirtualRide":
		fallthrough
	case "Ride":
		return s.props.getString(RideIcon, "\uf5a2")
	case "Run":
		return s.props.getString(RunIcon, "\ufc0c")
	case "NordicSki":
	case "AlpineSki":
	case "BackcountrySki":
		return s.props.getString(SkiingIcon, "\ue213")
	case "WorkOut":
		return s.props.getString(WorkOutIcon, "\ue213")
	default:
		return s.props.getString(UnknownActivityIcon, "\ue213")
	}
	return s.props.getString(UnknownActivityIcon, "\ue213")
}

func (s *strava) string() string {
	if s.Error != "" {
		return s.Error
	}
	segmentTemplate := s.props.getString(SegmentTemplate, "{{ .Ago }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  s,
		Env:      s.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	s.URL = fmt.Sprintf("https://www.strava.com/activities/%d", s.ID)
	return text
}

func (s *strava) getAccessToken() (string, error) {
	// get directly from cache
	if acccessToken, OK := s.env.cache().get(StravaAccessToken); OK {
		return acccessToken, nil
	}
	// use cached refersh token to get new access token
	if refreshToken, OK := s.env.cache().get(StravaRefreshToken); OK {
		if acccessToken, err := s.refreshToken(refreshToken); err == nil {
			return acccessToken, nil
		}
	}
	// use initial refresh token from property
	refreshToken := s.props.getString(RefreshToken, "")
	if len(refreshToken) == 0 {
		return "", &AuthError{
			message: InvalidRefreshToken,
		}
	}
	// no need to let the user provide access token, we'll always verify the refresh token
	acccessToken, err := s.refreshToken(refreshToken)
	return acccessToken, err
}

func (s *strava) refreshToken(refreshToken string) (string, error) {
	httpTimeout := s.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	url := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=strava&token=%s", refreshToken)
	body, err := s.env.HTTPRequest(url, httpTimeout)
	if err != nil {
		return "", &AuthError{
			// This might happen if /api was asleep. Assume the user will just retry
			message: Timeout,
		}
	}
	tokens := &TokenExchange{}
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return "", &AuthError{
			message: TokenRefreshFailed,
		}
	}
	// add tokens to cache
	s.env.cache().set(StravaAccessToken, tokens.AccessToken, tokens.ExpiresIn/60)
	s.env.cache().set(StravaRefreshToken, tokens.RefreshToken, 2*525960) // it should never expire unless revoked, default to 2 year
	return tokens.AccessToken, nil
}

func (s *strava) getResult() (*StravaData, error) {
	parseSingleElement := func(data []byte) (*StravaData, error) {
		var result []*StravaData
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			return nil, errors.New("no elements in the array")
		}
		return result[0], nil
	}
	getCacheValue := func(key string) (*StravaData, error) {
		val, found := s.env.cache().get(key)
		// we got something from the cache
		if found {
			if data, err := parseSingleElement([]byte(val)); err == nil {
				return data, nil
			}
		}
		return nil, errors.New("no data in cache")
	}

	// We only want the last activity
	url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
	httpTimeout := s.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	// No need to check more the every 30 min
	cacheTimeout := s.props.getInt(CacheTimeout, 30)
	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}
	accessToken, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}
	addAuthHeader := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer "+accessToken)
	}
	body, err := s.env.HTTPRequest(url, httpTimeout, addAuthHeader)
	if err != nil {
		return nil, err
	}
	var arr []*StravaData
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return nil, err
	}
	data, err := parseSingleElement(body)
	if err != nil {
		return nil, err
	}
	if cacheTimeout > 0 {
		// persist new sugars in cache
		s.env.cache().set(url, string(body), cacheTimeout)
	}
	return data, nil
}

func (s *strava) init(props Properties, env Environment) {
	s.props = props
	s.env = env
}
