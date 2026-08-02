package gitstatus

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/format/index"
)

type indexEntryRef struct {
	entry *index.Entry
	seen  bool
}

// scanner holds the shared, read-mostly state of one parallel worktree walk.
// Counters are atomics; every index entry has a unique path so each
// indexEntryRef is only ever touched by one goroutine.
type scanner struct {
	entries          map[string]*indexEntryRef
	lowerEntries     map[string]*indexEntryRef
	unmerged         map[string]bool
	sem              chan struct{}
	untrackedMode    string
	sortedPaths      []string
	lowerSortedPaths []string
	indexModTime     time.Time
	wg               sync.WaitGroup
	untracked        atomic.Int64
	modified         atomic.Int64
	added            atomic.Int64
	rehashes         atomic.Int64
	caseInsensitive  bool
}

// scanWorktree walks the worktree in parallel (one goroutine per tracked
// directory, bounded by NumCPU), comparing every tracked entry against the
// on-disk state and collecting untracked files/directories, then fills in
// result.Working accordingly.
func scanWorktree(opts Options, idx *index.Index, indexModTime time.Time, untrackedMode string, basePatterns []gitignore.Pattern, result *Result) {
	s := &scanner{
		entries:       make(map[string]*indexEntryRef, len(idx.Entries)),
		unmerged:      map[string]bool{},
		sem:           make(chan struct{}, runtime.NumCPU()),
		untrackedMode: untrackedMode,
		indexModTime:  indexModTime,
		// Case-insensitive filesystems can report a tracked path back with
		// different casing than the index stores, which would otherwise look
		// like an untracked file paired with a deleted one.
		caseInsensitive: runtime.GOOS == "windows" || runtime.GOOS == "darwin",
	}

	unmergedStages := map[string]int{}

	for _, e := range idx.Entries {
		if e.Stage != mergedStage {
			unmergedStages[e.Name] |= 1 << uint(e.Stage)
			continue
		}
		s.entries[e.Name] = &indexEntryRef{entry: e}
		s.sortedPaths = append(s.sortedPaths, e.Name)
	}
	sort.Strings(s.sortedPaths)

	if s.caseInsensitive {
		s.lowerEntries = make(map[string]*indexEntryRef, len(s.entries))
		s.lowerSortedPaths = make([]string, 0, len(s.sortedPaths))
		for path, ref := range s.entries {
			s.lowerEntries[strings.ToLower(path)] = ref
		}
		for _, p := range s.sortedPaths {
			s.lowerSortedPaths = append(s.lowerSortedPaths, strings.ToLower(p))
		}
		sort.Strings(s.lowerSortedPaths)
	}

	for name, stages := range unmergedStages {
		s.unmerged[name] = true
		staging, working := conflictXY(stages)
		addCode(&result.Staging, staging)
		addCode(&result.Working, working)
	}

	s.walk(opts.RepoRoot, "", basePatterns)
	s.wg.Wait()

	result.Working.Untracked += int(s.untracked.Load())
	result.Working.Modified += int(s.modified.Load())
	result.Working.Added += int(s.added.Load())

	for _, p := range s.sortedPaths {
		ref := s.entries[p]
		if !ref.seen && !ref.entry.SkipWorktree {
			result.Working.Deleted++
		}
	}
}

func (s *scanner) lookup(rel string) (*indexEntryRef, bool) {
	if ref, ok := s.entries[rel]; ok {
		return ref, true
	}
	if !s.caseInsensitive {
		return nil, false
	}
	ref, ok := s.lowerEntries[strings.ToLower(rel)]
	return ref, ok
}

func (s *scanner) hasTrackedPrefix(dirSlash string) bool {
	if i := sort.SearchStrings(s.sortedPaths, dirSlash); i < len(s.sortedPaths) && strings.HasPrefix(s.sortedPaths[i], dirSlash) {
		return true
	}
	if !s.caseInsensitive {
		return false
	}
	lowerSlash := strings.ToLower(dirSlash)
	i := sort.SearchStrings(s.lowerSortedPaths, lowerSlash)
	return i < len(s.lowerSortedPaths) && strings.HasPrefix(s.lowerSortedPaths[i], lowerSlash)
}

// descend walks a subdirectory, on its own goroutine when the concurrency
// budget allows, inline otherwise.
func (s *scanner) descend(dir, name, childRel string, patterns []gitignore.Pattern) {
	sub := filepath.Join(dir, name)
	select {
	case s.sem <- struct{}{}:
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer func() { <-s.sem }()
			s.walk(sub, childRel, patterns)
		}()
	default:
		s.walk(sub, childRel, patterns)
	}
}

func (s *scanner) walk(dir, rel string, patterns []gitignore.Pattern) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	if ps := readIgnoreFile(filepath.Join(dir, ".gitignore"), splitDomain(rel)); ps != nil {
		// Full-capacity reslice: appending copies into a new backing
		// array instead of sharing one with sibling goroutines.
		patterns = append(patterns[:len(patterns):len(patterns)], ps...)
	}

	matcher := gitignore.NewMatcher(patterns)
	match := func(childRel string, isDir bool) bool {
		return matcher.Match(strings.Split(childRel, "/"), isDir)
	}

	for _, de := range des {
		name := de.Name()
		if rel == "" && name == ".git" {
			continue
		}

		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}

		// Checked before de.IsDir(): a tracked symlink-to-directory can
		// be reported as a directory by ReadDir (Windows junctions), in
		// which case it must still be treated as the tracked file it
		// is, not descended into.
		if ref, tracked := s.lookup(childRel); tracked {
			if de.IsDir() && ref.entry.Mode != filemode.Symlink {
				// tracked file replaced by a directory: porcelain reports
				// the file as deleted (via the unseen pass) and hides the
				// directory's contents
				continue
			}
			compareTracked(dir, name, ref, de, s.indexModTime, &s.modified, &s.added, &s.rehashes)
			continue
		}

		if s.unmerged[childRel] {
			continue
		}

		if de.IsDir() {
			s.walkDir(dir, name, childRel, patterns, match)
			continue
		}

		if s.untrackedMode == "no" {
			continue
		}
		if match(childRel, false) {
			continue
		}
		s.untracked.Add(1)
	}
}

// walkDir handles a directory entry that holds no tracked file itself:
// descend when tracked files live beneath it, otherwise apply the untracked
// mode.
func (s *scanner) walkDir(dir, name, childRel string, patterns []gitignore.Pattern, match func(string, bool) bool) {
	if s.hasTrackedPrefix(childRel + "/") {
		s.descend(dir, name, childRel, patterns)
		return
	}

	if s.untrackedMode == "no" {
		return
	}

	if match(childRel, true) {
		return
	}

	if s.untrackedMode == "all" {
		s.descend(dir, name, childRel, patterns)
		return
	}

	// "normal": collapse to one untracked entry iff the directory contains
	// at least one non-ignored file.
	if dirHasVisibleFile(filepath.Join(dir, name), childRel, match) {
		s.untracked.Add(1)
	}
}

// compareTracked resolves the working-tree state of a single tracked entry.
// Every index entry has a unique path, so exactly one goroutine ever visits
// a given ref; writing ref.seen here needs no synchronization, only the
// wg.Wait() happens-before edge before it's read back on the caller side.
func compareTracked(dir, name string, ref *indexEntryRef, de os.DirEntry, indexModTime time.Time, modified, added, rehashes *atomic.Int64) {
	ref.seen = true
	e := ref.entry

	if e.SkipWorktree {
		return
	}

	if e.IntentToAdd {
		added.Add(1)
		return
	}

	fullPath := filepath.Join(dir, name)

	if e.Mode == filemode.Symlink {
		if !symlinkMatches(fullPath, e.Hash) {
			modified.Add(1)
		}
		return
	}

	info, err := de.Info()
	if err != nil {
		return
	}

	if uint32(info.Size()) != e.Size {
		modified.Add(1)
		return
	}

	mt := info.ModTime()
	clean := mt.Unix() == e.ModifiedAt.Unix() && (mt.Nanosecond() == e.ModifiedAt.Nanosecond() || e.ModifiedAt.Nanosecond() == 0)
	// Racy check: an index entry recorded at or after the index file's own
	// mtime can't be trusted from stat comparison alone (the file could
	// have changed again within the same timestamp tick), so force a
	// rehash even though the stat matched.
	racy := !e.ModifiedAt.Before(indexModTime)
	if clean && !racy {
		return
	}

	rehashes.Add(1)
	if !blobMatches(fullPath, e.Hash) {
		modified.Add(1)
	}
}

func symlinkMatches(path string, want plumbing.Hash) bool {
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	return blobSHA([]byte(target)) == want
}

// blobMatches hashes the file as a git blob and compares. Falls back to a
// CRLF-normalized hash to absorb autocrlf checkouts.
func blobMatches(path string, want plumbing.Hash) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	if blobSHA(data) == want {
		return true
	}

	if bytes.Contains(data, []byte("\r\n")) {
		normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		return blobSHA(normalized) == want
	}

	return false
}

func blobSHA(data []byte) plumbing.Hash {
	h := plumbing.NewHasher(plumbing.BlobObject, int64(len(data)))
	h.Write(data)
	return h.Sum()
}
