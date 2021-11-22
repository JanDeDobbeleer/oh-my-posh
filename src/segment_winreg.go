//go:build windows

// TODO:
//
//  * reg value -> string formatting to own fn?
//  * support for other reg types?
//  * custom formatting for values?  (custom sprintf is too dangerous, what else?)
//

package main

func (r *winreg) enabled() bool {

	var enableSegment bool = false

	registryPath := r.props.getString(RegistryPath, "")
	registryKey := r.props.getString(RegistryKey, "")

	regValue, err := r.env.getWindowsRegistryKeyValue(registryPath, registryKey)

	// Fallback behaviour
	failBehaviour := r.props.getString(QueryFailBehaviour, "hide_segment")
	fallbackString := r.props.getString(QueryFailFallbackString, "")

	if err != nil {
		switch failBehaviour {
		case "hide_segment":
			enableSegment = false
		case "display_fallback_string":
			r.content = fallbackString
			enableSegment = true
		}
	} else {
		r.content = regValue
		enableSegment = true
	}

	return enableSegment
}
