package auth

// CopilotTokenKey is the cache key under which the GitHub Copilot device-flow
// access token is stored once cli/auth/tui completes the interactive
// authentication flow. It lives here, rather than in that package, because
// segments/copilot.go reads the cached token directly and must not pull in
// bubbletea (via the terminal UI that produced it) just to check a cache
// key — an import graph that also has to compile for wasm, where there is no
// terminal to authenticate against at all.
const CopilotTokenKey = "copilot_token"
