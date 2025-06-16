package build

import "time"

var (
	Date    = time.Now().UTC().String()
	Version = "v0.0.0-dev"
)
