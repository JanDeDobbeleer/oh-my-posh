package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func newCachedTextSegment(env *mock.Environment, alias string, strategy Strategy) *Segment {
	segment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: strategy,
			Duration: cache.Duration("10m"),
		},
		Alias: alias,
		env:   env,
	}
	segment.name = alias

	writer := &segments.Text{}
	writer.Init(nil, nil)
	segment.writer = writer

	return segment
}

func TestSegmentCache(t *testing.T) {
	previousTemplateCache := template.Cache
	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}

	defer func() {
		template.Cache = previousTemplateCache
		cache.Device.DeleteAll()
	}()

	env := new(mock.Environment)
	env.On("Pwd").Return(t.TempDir())

	t.Run("round trip preserves the initialized writer", func(t *testing.T) {
		segment := newCachedTextSegment(env, "my_text_segment", Folder)
		segment.writer.SetText("Hello, Cache!")

		segment.setCache()

		newSegment := newCachedTextSegment(env, "my_text_segment", Folder)
		initializedWriter := newSegment.writer

		restored := newSegment.restoreCache()

		assert.True(t, restored, "cache should be restored")
		// Restoring must overlay the snapshot onto the writer initialized by
		// MapSegmentWithWriter, not replace it: a replacement would drop the
		// writer's runtime state (env, options) and panic on first use.
		assert.Same(t, initializedWriter, newSegment.writer, "restore should reuse the initialized writer")
		assert.Equal(t, "Hello, Cache!", newSegment.writer.Text(), "restored text should match")
	})

	t.Run("cached snapshot is immutable", func(t *testing.T) {
		segment := newCachedTextSegment(env, "immutable_segment", Folder)
		segment.writer.SetText("original")

		segment.setCache()

		// Mutating the writer after caching must not alter the cached snapshot.
		segment.writer.SetText("mutated")

		newSegment := newCachedTextSegment(env, "immutable_segment", Folder)

		assert.True(t, newSegment.restoreCache(), "cache should be restored")
		assert.Equal(t, "original", newSegment.writer.Text(), "cache should hold the state at cache time")
	})

	t.Run("legacy JSON entry is removed", func(t *testing.T) {
		segment := newCachedTextSegment(env, "legacy_segment", Device)

		key, store := segment.cacheKeyAndStore()
		store.Set(key, "legacy_json_string", cache.Duration("10m"))

		assert.False(t, segment.restoreCache(), "legacy cache should not be restored")

		_, found := store.Get[string](key)
		assert.False(t, found, "legacy key should be removed")
	})

	t.Run("unexpected entry type is removed", func(t *testing.T) {
		segment := newCachedTextSegment(env, "unexpected_segment", Device)

		key, store := segment.cacheKeyAndStore()
		store.Set(key, 42, cache.Duration("10m"))

		assert.False(t, segment.restoreCache(), "unexpected cache type should not be restored")

		_, found := store.Get[int](key)
		assert.False(t, found, "unexpected key should be removed")
	})
}

func TestGitMainWorktreeRestoresLiveContextLazilyAfterSegmentCacheHit(t *testing.T) {
	const (
		alias          = "cached-linked-worktree"
		mainWorktree   = "/repo/main"
		linkedWorktree = "/repo/linked"
		commonDir      = mainWorktree + "/.git"
		adminDir       = commonDir + "/worktrees/linked"
	)

	segmentCache := &Cache{Duration: "5h", Strategy: Session}
	source := &Segment{
		Type:  GIT,
		Alias: alias,
		Cache: segmentCache,
		writer: &segments.Git{
			IsWorkTree: true,
		},
	}
	source.setCache()

	key, store := source.cacheKeyAndStore()
	t.Cleanup(func() { store.Delete(key) })

	env := newDataReplayEnv(&runtime.Flags{})
	gitFile := &runtime.FileInfo{
		Path:         linkedWorktree + "/.git",
		ParentFolder: linkedWorktree,
	}
	env.On("HasParentFilePath", ".git", true).Return(gitFile, nil).Once()
	env.On("InWSLSharedDrive").Return(false).Once()
	env.On("GOOS").Return("")
	env.On("HasCommand", "git").Return(true).Once()
	env.On("FileContent", gitFile.Path).Return("gitdir: " + adminDir).Once()
	env.On("FileContent", filepath.Join(adminDir, "gitdir")).Return(linkedWorktree + "/.git").Once()
	env.On("RunCommand", "git", []string{
		"-C", linkedWorktree + "/",
		"--no-optional-locks",
		"-c", "core.quotepath=false",
		"-c", "color.status=false",
		"worktree", "list", "--porcelain", "-z",
	}).Return("worktree "+mainWorktree+"\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00", nil).Once()

	segment := &Segment{
		Type:     GIT,
		Alias:    alias,
		Cache:    segmentCache,
		Template: "{{ .MainWorktree }}|{{ .MainWorktree }}",
	}
	segment.Execute(env)

	env.AssertNotCalled(t, "HasParentFilePath", testifymock.Anything, testifymock.Anything)
	env.AssertNotCalled(t, "HasCommand", testifymock.Anything)
	env.AssertNotCalled(t, "RunCommand", testifymock.Anything, testifymock.Anything)

	writer := segment.Writer().(*segments.Git)
	assert.Equal(t, mainWorktree, writer.MainWorktree())
	assert.Equal(t, mainWorktree, writer.MainWorktree())
	assert.Equal(t, fmt.Sprintf("%s|%s", mainWorktree, mainWorktree), segment.string())
	env.AssertNumberOfCalls(t, "HasParentFilePath", 1)
	env.AssertNumberOfCalls(t, "HasCommand", 1)
	env.AssertNumberOfCalls(t, "RunCommand", 1)
}
