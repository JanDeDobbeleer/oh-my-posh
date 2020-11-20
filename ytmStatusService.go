package main

type ytmPlayingState int

const (
	playing ytmPlayingState = iota
	paused
	stopped
)

type ytmStatus struct {
	state  ytmPlayingState
	author string
	title  string
}

type ytmStatusService interface {
	Get() (*ytmStatus, error)
}
