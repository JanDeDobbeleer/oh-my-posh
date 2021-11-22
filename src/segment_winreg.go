//go:build windows

// TODO:
//
//  * reg value -> string formatting to own fn?
//  * support for other reg types?
//  * custom formatting for values?  (custom sprintf is too dangerous, what else?)
//

package main

type winreg struct {
	props     *properties
	env       environmentInfo
	content   string
	errorInfo string
}

const (

	// Properties defining how to get to the key.
	RegistryPath Property = "registry_path" // path from the supplied root under which the key exists
	RegistryKey  Property = "registry_key"  // key within full reg path formed from two above

	// Properties defining fallback behaviour in case the query fails.
	QueryFailBehaviour      Property = "query_fail_behaviour"       // what to do, can be one of "hide_segment", "display_fallback_string", "show_debug_info"
	QueryFailFallbackString Property = "query_fail_fallback_string" // what to display if above is "display_fallback_string"
)

func (r *winreg) string() string {
	newText := r.content
	return newText
}

func (r *winreg) init(props *properties, env environmentInfo) {
	r.props = props
	r.env = env
}

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
