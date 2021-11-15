//go:build !windows

package main

func (r *regquery) enabled() bool {
	return false
}
