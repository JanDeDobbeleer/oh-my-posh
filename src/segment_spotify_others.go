//go:build !darwin && !windows

package main

func (s *spotify) enabled() bool {
	return false
}
