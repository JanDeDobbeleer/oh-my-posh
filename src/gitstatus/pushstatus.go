package gitstatus

import "fmt"

// AheadBehind reports how many commits oursHex is ahead of and behind
// theirsHex, both full commit hashes, in the repository whose common git
// dir is commonGitDir. It's the same computation Load uses for the
// upstream ahead/behind counts, exposed standalone so callers with an
// already-resolved pair of commits (such as the push-remote comparison)
// don't need to go through the full status Load.
func AheadBehind(commonGitDir, oursHex, theirsHex string) (ahead, behind int, err error) {
	ours, ok := parseHash(oursHex)
	if !ok {
		return 0, 0, fmt.Errorf("gitstatus: invalid hash %q", oursHex)
	}

	theirs, ok := parseHash(theirsHex)
	if !ok {
		return 0, 0, fmt.Errorf("gitstatus: invalid hash %q", theirsHex)
	}

	store := newObjectStore(commonGitDir)
	defer store.close()

	return aheadBehind(store, ours, theirs)
}
