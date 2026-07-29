package prompt

// Guards the fixture's usefulness, as opposed to its safety (that is
// fixtures_sanitized_test.go). The merged fixture is recorded from a real
// machine, so how much of a template it exercises depends on what that machine
// looked like at the time: record against a clean checkout and the git segment
// comes back all zeroes, which renders but stops covering the ahead/behind and
// working-tree branches entirely. The goldens still pass in that state, which is
// exactly why this has to be asserted rather than noticed.

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixtureExercisesGitStatus(t *testing.T) {
	raw, err := os.ReadFile(fixturePath)
	require.NoErrorf(t, err, "fixture %s not found; regenerate it with: %s", fixturePath, regenerateFixtureCommand)

	var doc struct {
		Segments map[string]struct {
			Data map[string]any `json:"data"`
		} `json:"segments"`
	}

	require.NoError(t, json.Unmarshal(raw, &doc))

	git, ok := doc.Segments["git"]
	require.True(t, ok, "fixture has no git segment; regenerate it with: %s", regenerateFixtureCommand)

	// Ahead covers the branch-status template, Working covers the counts one.
	// Both come from a repository with an upstream and uncommitted changes, so a
	// regeneration against a pristine checkout trips this rather than silently
	// shipping a fixture that renders every git theme as a bare branch name.
	require.NotZero(t, git.Data["Ahead"],
		"git fixture has no commits ahead, so the branch-status template is never exercised; regenerate against a checkout with an upstream and local commits: %s",
		regenerateFixtureCommand)

	working, ok := git.Data["Working"].(map[string]any)
	require.True(t, ok, "git fixture has no Working tree data; regenerate it with: %s", regenerateFixtureCommand)

	var changed float64
	for _, key := range []string{"Added", "Modified", "Deleted", "Untracked"} {
		if count, isNumber := working[key].(float64); isNumber {
			changed += count
		}
	}

	require.NotZero(t, changed,
		"git fixture has a clean working tree, so the file-count template is never exercised; regenerate against a checkout with uncommitted changes: %s",
		regenerateFixtureCommand)
}
