//go:build js

package battery

// Get always reports no battery on js/wasm. There is no operating-system
// battery API to query when this package is compiled to run inside a
// browser (or any other wasm host): battery_windows_nix.go's systemGetAll
// has no implementation for this platform, so this file exists purely to
// satisfy the package's exported surface for the config and segments
// packages that pull it in, not to provide real readings.
func Get() (*Info, error) {
	return nil, ErrNotFound
}
