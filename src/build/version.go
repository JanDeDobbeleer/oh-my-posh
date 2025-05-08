package build

import "time"

var (
	Date    = time.Now().UTC().String()
	Version = "dev"
)
