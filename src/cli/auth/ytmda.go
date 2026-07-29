package auth

// YTMDABASEURL and YTMDATOKEN are the YouTube Music Desktop App Companion
// API's base URL and the cache key its access token is stored under, once
// cli/auth/tui completes the interactive authentication flow. They live
// here, rather than in that package, because segments/ytm.go builds request
// URLs and reads the cached token directly and must not pull in bubbletea
// (via the terminal UI that produced the token) just for that — an import
// graph that also has to compile for wasm, where there is no terminal to
// authenticate against at all.
const (
	YTMDABASEURL = "http://localhost:9863/api/v1"
	YTMDATOKEN   = "ytmda_token"
)
