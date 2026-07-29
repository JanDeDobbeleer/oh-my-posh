package prompt

// Backstop for the sanitizer in cli/sanitize.go: this is not the sanitization
// mechanism itself, it is what catches the next person who regenerates a
// fixture on their own machine (with `config export data` but without
// --sanitize, or with a sanitizer that regressed) before it reaches a public
// repo.
//
// TestFixturesContainNoIdentityByShape is the load-bearing test here. An
// earlier version of this backstop checked a denylist of field names
// (AccessKeyID, TenantID, HomeTenantID, Email) - and missed a real email
// address recorded at segments.az.data.user.name, because the sanitizer's own
// az scrub was matching the wrong (Go-field-name) casing against the actual
// (camelCase-tagged) serialized key, so "user"/"name" never matched "User"/
// "Name" and neither did this test. A field-name denylist and a field-name
// sanitizer share the same blind spot: the next mis-cased or newly-added field
// slips past both identically. Checking every string value in the tree by
// *shape* - looks like an email, looks like an absolute path, matches this
// machine's username/hostname - regardless of which key holds it, does not
// depend on having enumerated the right field name in advance.
import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sanitizedEmail = "alice@contoso.com"

var (
	// Anchored to the whole value (not searched as a substring) so a
	// structured value that merely embeds an email-shaped fragment - a git
	// SSH remote like git@github.com:org/repo.git, which is not an email - is
	// not misreported as a leak. The fields that actually carry an email in
	// these writers (git.User.Email, az user.name) always store it as the
	// entire string value, never concatenated with anything else.
	emailShape = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

	// Prefix-anchored: these only need to recognize the *start* of an
	// absolute path, since real paths carry arbitrary content after it
	// (drive letter, or a home/user directory root).
	windowsPathShape = regexp.MustCompile(`(?i)^[a-z]:[\\/]`)
	uncPathShape     = regexp.MustCompile(`^\\\\`)
	unixPathShape    = regexp.MustCompile(`^/(home|Users|root)(/|$)`)
)

// identityPattern word-bounds needle so a legitimate, non-identity string that
// merely contains needle as a substring - like the git remote
// https://github.com/JanDeDobbeleer/oh-my-posh, which contains "jande"
// case-insensitively - is not misreported as a leak. This is exactly the trap
// a naive case-insensitive string replace/search falls into; using \b instead
// is what avoids it.
func identityPattern(needle string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(needle) + `\b`)
}

func looksLikeAbsolutePath(s string) bool {
	return windowsPathShape.MatchString(s) || uncPathShape.MatchString(s) || unixPathShape.MatchString(s)
}

// walkJSONStrings calls visit for every string value found anywhere in the
// decoded document, at any nesting depth, through both objects and arrays,
// with a JSON-pointer-ish path for a readable failure. Map/array keys and
// indices are not inspected themselves - only string leaf values - since the
// leak this exists to catch (segments.az.data.user.name) sits in a value, and
// the whole point is not to depend on already knowing which key to check.
func walkJSONStrings(node any, path string, visit func(path, value string)) {
	switch v := node.(type) {
	case string:
		visit(path, v)
	case map[string]any:
		for key, value := range v {
			walkJSONStrings(value, path+"."+key, visit)
		}
	case []any:
		for i, value := range v {
			walkJSONStrings(value, path+"["+strconv.Itoa(i)+"]", visit)
		}
	}
}

// TestFixturesContainNoIdentityByShape walks every committed fixture's full
// decoded JSON tree and checks each string value by shape: does it look like
// an email address, an absolute filesystem path, or contain this machine's
// username/hostname - regardless of which key it is stored under. See the
// package comment above for why shape (not a field-name denylist) is the
// right check here.
func TestFixturesContainNoIdentityByShape(t *testing.T) {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}

	hostname, err := os.Hostname()
	require.NoError(t, err)

	hostShort, _, _ := strings.Cut(hostname, ".")

	file := fixturePath

	raw, err := os.ReadFile(file)
	require.NoErrorf(t, err, "fixture %s not found; regenerate it with: %s", file, regenerateFixtureCommand)

	var doc any
	require.NoErrorf(t, json.Unmarshal(raw, &doc), "%s is not valid JSON", file)

	walkJSONStrings(doc, "$", func(path, value string) {
		if value == "" {
			return
		}

		if emailShape.MatchString(value) {
			assert.Equalf(t, sanitizedEmail, value,
				"%s: %s is an email address (%q) that is not the synthetic placeholder - regenerate with: %s",
				file, path, value, regenerateFixtureCommand)
		}

		if looksLikeAbsolutePath(value) {
			t.Errorf("%s: %s looks like an absolute filesystem path (%q) - regenerate with: %s",
				file, path, value, regenerateFixtureCommand)
		}

		if username != "" && identityPattern(username).MatchString(value) {
			t.Errorf("%s: %s contains the recording machine's username (%q)", file, path, value)
		}

		if hostname != "" && identityPattern(hostname).MatchString(value) {
			t.Errorf("%s: %s contains the recording machine's hostname (%q)", file, path, value)
		}

		if hostShort != "" && hostShort != hostname && identityPattern(hostShort).MatchString(value) {
			t.Errorf("%s: %s contains the recording machine's hostname (%q)", file, path, value)
		}
	})
}

// TestFixturesDoNotLeakRecordingMachineIdentity is a second, independent check
// over the raw file bytes rather than the decoded tree - it also catches the
// home directory, which cannot be recognized by shape the way an absolute
// path prefix can (a bare "/data/build" is a perfectly ordinary absolute path
// on Linux; only comparing against *this* machine's actual home directory
// tells them apart).
func TestFixturesDoNotLeakRecordingMachineIdentity(t *testing.T) {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}

	hostname, err := os.Hostname()
	require.NoError(t, err)
	// A hostname can carry a domain suffix (surface-pro.local); check the
	// short form too so a leak isn't missed just because the fixture omits it.
	hostShort, _, _ := strings.Cut(hostname, ".")

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	homeForward := filepath.ToSlash(home)

	needles := map[string]string{
		"username": username,
		"hostname": hostname,
	}

	if hostShort != "" && hostShort != hostname {
		needles["hostname (short)"] = hostShort
	}

	file := fixturePath

	raw, err := os.ReadFile(file)
	require.NoErrorf(t, err, "fixture %s not found; regenerate it with: %s", file, regenerateFixtureCommand)

	for label, needle := range needles {
		if needle == "" {
			continue
		}

		assert.Falsef(t, identityPattern(needle).Match(raw),
			"%s contains the recording machine's %s (%q) - regenerate with: %s",
			file, label, needle, regenerateFixtureCommand)
	}

	// The home directory appears embedded in a longer path
	// (.../dev/oh-my-posh), so it cannot be word-bounded the same way;
	// a plain substring check is correct here since a home directory
	// string is not the kind of thing that legitimately appears inside
	// unrelated recorded data the way a username can.
	if home != "" {
		assert.NotContainsf(t, string(raw), home, "%s contains the recording machine's home directory", file)
		assert.NotContainsf(t, string(raw), homeForward, "%s contains the recording machine's home directory", file)
	}
}
