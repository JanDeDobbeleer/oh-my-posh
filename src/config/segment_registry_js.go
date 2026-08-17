//go:build js && wasm

package config

// newSegmentWriter answers that this build has no writers.
//
// The browser build renders a config from recorded data and never runs a segment's detection - it
// has no filesystem, no processes and no network to detect anything with. Every writer it could
// construct would immediately find nothing.
//
// Constructing them anyway is expensive in the one dimension a browser cares about. The registry
// on the other side of this build tag is a package-level map holding a constructor for each of
// the 118 segment types, beside an init that gob-registers all of them; both are reachable from
// any config parse, so between them they pin every segment package and everything those pull in -
// cloud SDKs, HTTP clients, the lot. Measured against a build with them: 4.2MB raw, 0.5MB brotli,
// for code no visitor can reach.
//
// A nil writer is not a failure here. Segment.templateContext falls back to the recorded data,
// which carries what the template reads - fields and method results alike (see
// cli/segment_data.go) - so a segment renders from that instead.
func newSegmentWriter(_ SegmentType) (SegmentWriter, error) {
	return nil, nil
}
