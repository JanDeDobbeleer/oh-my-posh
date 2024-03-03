//go:build !linux && !darwin && !windows

package segments

func (s *Spotify) Enabled() bool {
	return false
}
