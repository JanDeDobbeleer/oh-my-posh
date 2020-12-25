// +build !darwin
// +build !windows

package main

func (s *spotify) enabled() bool {
	return false
}
