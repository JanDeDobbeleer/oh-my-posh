package segments

import "oh-my-posh/properties"

const (
	// HTTPTimeout timeout used when executing http request
	HTTPTimeout properties.Property = "http_timeout"
	// DefaultHTTPTimeout default timeout used when executing http request
	DefaultHTTPTimeout = 20
	// DefaultCacheTimeout default timeout used when caching data
	DefaultCacheTimeout = 10
)
