//go:build illumos || solaris

package battery

// Get returns information about all batteries in the system.
//
// Battery reporting is not implemented on illumos/solaris; this always
// returns a NoBatteryError so the battery segment disables itself cleanly.
func Get() (*Info, error) {
	return nil, &NoBatteryError{}
}
