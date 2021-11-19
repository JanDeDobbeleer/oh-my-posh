package main

type regquery struct {
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

func (r *regquery) string() string {
	newText := r.content
	return newText
}

func (r *regquery) init(props *properties, env environmentInfo) {
	r.props = props
	r.env = env
}
