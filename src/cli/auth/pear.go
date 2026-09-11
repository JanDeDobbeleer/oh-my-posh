package auth

// PEARBASEURL and PEARTOKEN are the Pear Desktop API Server's base URL and
// the cache key its access token is stored under, once cli/auth/tui completes
// the interactive authentication flow. They live here, rather than in that
// package, because segments/pear.go builds request URLs and reads the cached
// token directly and must not pull in bubbletea (via the terminal UI that
// produced the token) just for that — an import graph that also has to
// compile for wasm, where there is no terminal to authenticate against at all.
const (
	PEARBASEURL = "http://127.0.0.1:26538"
	PEARTOKEN   = "pear_token"
)
