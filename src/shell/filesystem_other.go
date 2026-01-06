//go:build !windows

package shell

// only Windows knows transient file-sharing violations worth retrying,
// on other platforms a write error is always persistent
func canRetryWrite(_ error) bool {
	return false
}
