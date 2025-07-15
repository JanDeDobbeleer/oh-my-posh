package build

import "time"

var (
	Date    = time.Now().UTC().String()
	Version = "0.0.0-dev"
)
