package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

// FieldRefs is embedded by writers whose data fetching is derived from the
// config's template analysis: it receives the segment's reference set
// (satisfying the engine's FieldSetConsumer contract) and answers, per data
// unit, whether any of the fields that unit populates is referenced -
// exactly for an analyzable config, by template.RefSet's substring
// heuristic otherwise. There are no fetch options anymore: what the
// templates render is what gets fetched.
type FieldRefs struct {
	// Refs is exported only so gob can snapshot writers embedding this
	// struct (gob refuses structs without exported fields, and no-op gob
	// hooks here would promote onto the embedding writer and swallow its
	// entire snapshot). A restored snapshot's value always equals the
	// freshly delivered one: the segment cache key fingerprints the same
	// analysis this carries. Treat as internal.
	Refs template.RefSet
}

// SetReferencedFields stores the reference set the engine derived for this
// segment's templates.
func (f *FieldRefs) SetReferencedFields(refs template.RefSet) {
	f.Refs = refs
}

// fetchUnit reports whether the data unit populating fields should run.
func (f *FieldRefs) fetchUnit(fields ...string) bool {
	return f.Refs.Referenced(fields...)
}

// fetchUnitFailOpen is fetchUnit for a data unit whose historic default was
// ON (version fetching). The polarity deliberately differs from fetchUnit:
// the substring heuristic exists to switch a default-off unit on when the
// analysis loses track, but for a default-on unit the safe unanalyzable
// answer is simply "fetch" - wrong display costs more than the subprocess.
// A never-delivered set (library or test usage) fetches for the same reason.
func (f *FieldRefs) fetchUnitFailOpen(fields ...string) bool {
	if !f.Refs.Analyzable {
		return true
	}

	return f.Refs.Referenced(fields...)
}
