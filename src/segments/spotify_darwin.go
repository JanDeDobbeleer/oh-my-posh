//go:build darwin

package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const spotifyCacheKey = "spotify_music_player"

func (s *Spotify) Enabled() bool {
	cacheTimeout := s.props.GetInt(properties.CacheTimeout, 0)
	if cacheTimeout > 0 && s.getFromCache() {
		return true
	}

	// Check if running
	running := s.runAppleScriptCommand("application \"Spotify\" is running")
	if running == "false" || running == "" {
		s.Status = stopped
		return false
	}

	s.Status = s.runAppleScriptCommand("tell application \"Spotify\" to player state as string")

	if len(s.Status) == 0 {
		s.Status = stopped
		return false
	}

	if s.Status == stopped {
		return false
	}

	s.Artist = s.runAppleScriptCommand("tell application \"Spotify\" to artist of current track as string")
	s.Track = s.runAppleScriptCommand("tell application \"Spotify\" to name of current track as string")

	s.resolveIcon()

	if cacheTimeout > 0 {
		s.setCache(cacheTimeout)
	}

	return true
}

func (s *Spotify) runAppleScriptCommand(command string) string {
	val, _ := s.env.RunCommand("osascript", "-e", command)
	return val
}

func (s *Spotify) getFromCache() bool {
	str, found := s.env.Cache().Get(spotifyCacheKey)
	if !found {
		return false
	}

	var cachedMusicPlayer MusicPlayer
	err := json.Unmarshal([]byte(str), &cachedMusicPlayer)
	if err != nil {
		return false
	}

	s.MusicPlayer = cachedMusicPlayer
	return true
}

func (s *Spotify) setCache(cacheTimeout int) {
	cache, err := json.Marshal(s.MusicPlayer)
	if err != nil {
		return
	}

	s.env.Cache().Set(spotifyCacheKey, string(cache), cacheTimeout)
}
