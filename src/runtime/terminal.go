package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	httplib "net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

type Terminal struct {
	CmdFlags      *Flags
	cmdCache      *cache.Command
	lsDirMap      *maps.Concurrent[[]fs.DirEntry]
	dirIndexMap   *maps.Concurrent[*dirIndex]
	parentFileMap *maps.Concurrent[parentFilePathResult]
	cwd           string
	host          string
	networks      []*Connection
}

// parentFilePathResult memoizes a HasParentFilePath outcome (hit or miss) for
// the duration of one prompt invocation, so the activation gate and a
// segment's own Enabled() doing the same upward search only walk the
// directory tree once.
type parentFilePathResult struct {
	info *FileInfo
	err  error
}

func (term *Terminal) Init(flags *Flags) {
	defer log.Trace(time.Now())

	term.CmdFlags = flags

	if term.CmdFlags == nil {
		term.CmdFlags = &Flags{}
	}

	term.lsDirMap = maps.NewConcurrent[[]fs.DirEntry]()
	term.dirIndexMap = maps.NewConcurrent[*dirIndex]()
	term.parentFileMap = maps.NewConcurrent[parentFilePathResult]()

	term.setPromptCount()

	term.setPwd()

	term.cmdCache = &cache.Command{
		Commands: maps.NewConcurrent[string](),
	}
}

func (term *Terminal) Getenv(key string) string {
	defer log.Trace(time.Now(), key)

	// The data file's env section carries template values (UserName, PWD, ...),
	// not OS variables, so there is nothing to substitute here - and a browser
	// has no environment either. Answering empty is what makes the CLI under
	// DataOnly and the wasm build agree; reading the real environment would
	// leave a segment keyed on, say, TERM_PROGRAM rendering one thing here and
	// another there, from the same config and the same data.
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return ""
	}

	val := os.Getenv(key)
	log.Debug(val)
	return val
}

func (term *Terminal) Pwd() string {
	return term.cwd
}

func (term *Terminal) setPwd() {
	defer log.Trace(time.Now())

	correctPath := func(pwd string) string {
		if term.GOOS() != WINDOWS {
			return pwd
		}

		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		driveLetter, err := regex.GetCompiledRegex(`^[a-z]:`)
		if err == nil {
			return driveLetter.ReplaceAllStringFunc(pwd, strings.ToUpper)
		}

		return pwd
	}

	if term.CmdFlags != nil && term.CmdFlags.PWD != "" {
		term.cwd = path.Clean(term.CmdFlags.PWD)
		log.Debug(term.cwd)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}

	term.cwd = correctPath(dir)
	log.Debug(term.cwd)
}

// errDataOnly is what every environment probe answers with when
// Flags.DataOnly is set. DataOnly started life in config.Segment.restoreData,
// suppressing a segment the recorded data does not cover - but that only
// governs the segment's own Enabled(). A writer field computed lazily by a
// method the template calls still reached the machine long afterwards:
// segments/git.go's StashCount() reads logs/refs/stash off disk, and
// stashCount is unexported so no recorded data ever restores it. Rendering
// jandedobbeleer under wasm, where there is no filesystem, failed that
// template outright while the CLI happily read the real repository.
//
// Gating the environment itself rather than each such method is what makes
// the guarantee hold for segments nobody has audited: there is no way to
// write one that probes, because the probe primitives themselves refuse.
var errDataOnly = errors.New("environment access is disabled: rendering from recorded data only")

func (term *Terminal) HasFiles(pattern string) bool {
	return term.HasFilesInDir(term.Pwd(), pattern)
}

func (term *Terminal) HasFilesInDir(dir, pattern string) bool {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return false
	}
	defer log.Trace(time.Now(), pattern)

	dirEntries, err := term.readDir(dir)
	if err != nil {
		log.Error(err)
		log.Debug("false")
		return false
	}

	pattern = strings.ToLower(pattern)

	idx := term.dirIndex(dir, dirEntries)
	if matched, ok := idx.match(pattern); ok {
		if matched {
			log.Debug("true")
			return true
		}

		log.Debug("false")
		return false
	}

	return linearMatch(dirEntries, pattern)
}

// readDir returns dir's listing, populating the per-Terminal cache from disk
// on first use. A directory read once during a prompt invocation never
// changes underneath it, so later calls for the same dir reuse the cached
// entries instead of hitting the filesystem again.
func (term *Terminal) readDir(dir string) ([]fs.DirEntry, error) {
	if files, OK := term.lsDirMap.Get(dir); OK && len(files) > 0 {
		return files, nil
	}

	fileSystem := os.DirFS(dir)

	dirEntries, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, err
	}

	term.lsDirMap.Set(dir, dirEntries)

	return dirEntries, nil
}

// dirIndex returns the cached dirIndex for dir, building it from dirEntries
// on first use. Two callers racing on the same never-before-seen dir may
// both build it; the maps.Concurrent Set that loses the race is discarded,
// which is harmless since both builds produce an equal index - the same
// tradeoff lsDirMap already makes for the raw listing.
func (term *Terminal) dirIndex(dir string, dirEntries []fs.DirEntry) *dirIndex {
	if idx, OK := term.dirIndexMap.Get(dir); OK {
		return idx
	}

	idx := newDirIndex(dirEntries)

	// readDir deliberately never caches an empty listing - an empty directory
	// is re-read on every probe - so an index built from one must not be
	// cached either: a file appearing mid-render would show up in the fresh
	// listing (and in linearMatch fallbacks) while a cached empty index kept
	// answering false for the fast-path shapes.
	if len(dirEntries) > 0 {
		term.dirIndexMap.Set(dir, idx)
	}

	return idx
}

// dirIndex is the inverted view of a directory listing that HasFilesInDir
// consults before falling back to a linear filepath.Match scan. Building it
// once per directory turns the common query shapes - a literal file name, or
// a "*.ext" glob - into a single map lookup instead of a scan repeated for
// every segment and every glob it declares.
//
// It excludes directories and lower-cases every name, mirroring linearMatch's
// own comparison exactly.
type dirIndex struct {
	// names holds every file's lower-cased name, answering literal (no
	// metacharacter) patterns.
	names map[string]struct{}
	// suffixes holds every dotted suffix of every file's lower-cased name -
	// for "a.b.c" that is ".c" and ".b.c" - so both "*.c" and multi-dot
	// patterns like "*.gradle.kts" resolve with a single lookup.
	suffixes map[string]struct{}
}

// newDirIndex builds a dirIndex from a directory listing.
func newDirIndex(dirEntries []fs.DirEntry) *dirIndex {
	idx := &dirIndex{
		names:    make(map[string]struct{}, len(dirEntries)),
		suffixes: make(map[string]struct{}),
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())
		idx.names[name] = struct{}{}

		for i, r := range name {
			if r != '.' {
				continue
			}

			idx.suffixes[name[i:]] = struct{}{}
		}
	}

	return idx
}

// match answers pattern - already lower-cased by the caller, the same
// contract linearMatch relies on - from the index when its shape allows a
// direct lookup. ok is false for anything else, telling the caller to fall
// back to linearMatch; that fallback is always correct, so a false negative
// here only costs performance, never correctness.
func (idx *dirIndex) match(pattern string) (matched, ok bool) {
	if suffix, isSuffix := suffixPattern(pattern); isSuffix {
		_, hit := idx.suffixes[suffix]
		return hit, true
	}

	if !hasGlobMeta(pattern) {
		_, hit := idx.names[pattern]
		return hit, true
	}

	return false, false
}

// hasGlobMeta reports whether pattern contains a filepath.Match
// metacharacter - wildcard, single-character wildcard, character class, or
// escape. A pattern with none of these is a plain literal: filepath.Match
// degrades to a straight string comparison against the (single-element,
// separator-free) directory entry name.
func hasGlobMeta(pattern string) bool {
	return strings.ContainsAny(pattern, `*?[\`)
}

// suffixPattern reports whether pattern has the shape "*" followed by a
// dotted suffix with no further metacharacters - "*.go", "*.gradle.kts" -
// and if so returns that suffix, dot included, ready for a
// dirIndex.suffixes lookup.
//
// Patterns the index cannot decide this way - bare "*", or a leading "*"
// whose suffix does not start with "." such as "*txt" - fall through to
// linearMatch instead of being special-cased here: they are rare in
// practice, and the index only ever tracks dot-anchored suffixes.
func suffixPattern(pattern string) (suffix string, ok bool) {
	if len(pattern) < 2 || pattern[0] != '*' || pattern[1] != '.' {
		return "", false
	}

	if hasGlobMeta(pattern[1:]) {
		return "", false
	}

	return pattern[1:], true
}

// linearMatch is the pre-index HasFilesInDir scan: filepath.Match against
// every non-directory entry, case-insensitively. It is the fallback for
// pattern shapes dirIndex.match cannot decide, and the reference
// implementation the differential tests check the index against.
func linearMatch(dirEntries []fs.DirEntry, pattern string) bool {
	for _, match := range dirEntries {
		if match.IsDir() {
			continue
		}

		matchFileName, err := filepath.Match(pattern, strings.ToLower(match.Name()))
		if err != nil {
			log.Error(err)
			log.Debug("false")
			return false
		}

		if matchFileName {
			log.Debug("true")
			return true
		}
	}

	log.Debug("false")
	return false
}

func (term *Terminal) HasFileInParentDirs(pattern string, depth uint) bool {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return false
	}
	defer log.Trace(time.Now(), pattern, fmt.Sprint(depth))
	currentFolder := term.Pwd()

	for c := 0; c < int(depth); c++ {
		if term.HasFilesInDir(currentFolder, pattern) {
			log.Debug("true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			log.Debug("false")
			return false
		}
	}
	log.Debug("false")
	return false
}

func (term *Terminal) HasFolder(folder string) bool {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return false
	}
	defer log.Trace(time.Now(), folder)
	f, err := os.Stat(folder)
	if err != nil {
		log.Debug("false")
		return false
	}
	isDir := f.IsDir()
	log.Debugf("%t", isDir)
	return isDir
}

// StatFile reports the modification time and size of the file at filePath.
// It works the same way on every platform (a plain os.Stat), so a cache key
// built from it invalidates correctly on Windows and darwin too, not just
// unix.
func (term *Terminal) StatFile(filePath string) (FileStat, error) {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return FileStat{}, errDataOnly
	}
	defer log.Trace(time.Now(), filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		log.Error(err)
		return FileStat{}, err
	}

	stat := FileStat{ModTime: info.ModTime().Unix(), Size: info.Size()}
	log.Debugf("%+v", stat)
	return stat, nil
}

func (term *Terminal) ResolveSymlink(input string) (string, error) {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return "", errDataOnly
	}
	defer log.Trace(time.Now(), input)
	link, err := filepath.EvalSymlinks(input)
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Debug(link)
	return link, nil
}

func (term *Terminal) FileContent(file string) string {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return ""
	}
	defer log.Trace(time.Now(), file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(term.Pwd(), file)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error(err)
		return ""
	}

	fileContent := string(content)
	log.Debug(fileContent)

	return fileContent
}

func (term *Terminal) LsDir(input string) []fs.DirEntry {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return nil
	}
	defer log.Trace(time.Now(), input)

	entries, err := os.ReadDir(input)
	if err != nil {
		log.Error(err)
		return nil
	}

	log.Debugf("%v", entries)
	return entries
}

func (term *Terminal) User() string {
	defer log.Trace(time.Now())
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	log.Debug(user)
	return user
}

func (term *Terminal) Host() (string, error) {
	defer log.Trace(time.Now())
	if len(term.host) != 0 {
		return term.host, nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		log.Error(err)
		return "", err
	}

	hostName = cleanHostName(hostName)
	log.Debug(hostName)
	term.host = hostName

	return hostName, nil
}

func (term *Terminal) GOOS() string {
	defer log.Trace(time.Now())
	return runtime.GOOS
}

func (term *Terminal) Home() string {
	return path.Home()
}

func (term *Terminal) RunCommand(command string, args ...string) (string, error) {
	return term.RunCommandWithEnv(command, nil, args...)
}

func (term *Terminal) RunCommandWithEnv(command string, envs []string, args ...string) (string, error) {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return "", errDataOnly
	}
	defer log.Trace(time.Now(), append([]string{command}, args...)...)

	if cacheCommand, ok := term.cmdCache.Get(command); ok {
		command = cacheCommand
	}

	output, err := cmd.RunWithEnv(command, envs, args...)
	if err != nil {
		log.Error(err)
	}

	log.Debug(output)
	return output, err
}

func (term *Terminal) RunShellCommand(shell, command string) string {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return ""
	}
	defer log.Trace(time.Now())

	if out, err := term.RunCommand(shell, "-c", command); err == nil {
		return out
	}

	return ""
}

func (term *Terminal) CommandPath(command string) string {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return ""
	}
	defer log.Trace(time.Now(), command)

	// L1: in-memory, unbounded for the lifetime of this process.
	if cmdPath, ok := term.cmdCache.Get(command); ok {
		log.Debug(cmdPath)
		return cmdPath
	}

	// L2: session-persisted lookups, shared across prompt invocations within
	// the same shell session. Avoids re-running exec.LookPath (a PATH x
	// PATHEXT stat storm on Windows, worst for missing commands) on every
	// prompt render.
	if cachedPath, found, ok := cache.GetPersistedCommandPath(command); ok {
		if !found {
			log.Debug("command not found (cached)")
			return ""
		}

		// Revalidate cheaply; a stale/moved binary should fall through to a
		// fresh LookPath rather than returning a dead path.
		if _, err := os.Stat(cachedPath); err == nil {
			term.cmdCache.Set(command, cachedPath)
			log.Debug(cachedPath)
			return cachedPath
		}

		log.Debugf("cached command path no longer valid: %s", cachedPath)
	}

	cmdPath, err := exec.LookPath(command)
	if err == nil {
		term.cmdCache.Set(command, cmdPath)
		cache.PersistCommandPath(command, cmdPath, true)
		log.Debug(cmdPath)
		return cmdPath
	}

	cache.PersistCommandPath(command, "", false)

	log.Error(err)
	return ""
}

func (term *Terminal) HasCommand(command string) bool {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return false
	}
	defer log.Trace(time.Now(), command)

	if cmdPath := term.CommandPath(command); cmdPath != "" {
		return true
	}

	return false
}

func (term *Terminal) StatusCodes() (int, string) {
	defer log.Trace(time.Now())

	if term.CmdFlags.Shell != CMD || !term.CmdFlags.NoExitCode {
		return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
	}

	errorCode := term.Getenv("=ExitCode")
	log.Debug(errorCode)
	term.CmdFlags.ErrorCode, _ = strconv.Atoi(errorCode)

	return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
}

func (term *Terminal) ExecutionTime() float64 {
	defer log.Trace(time.Now())
	if term.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return term.CmdFlags.ExecutionTime
}

func (term *Terminal) Flags() *Flags {
	defer log.Trace(time.Now())
	return term.CmdFlags
}

func (term *Terminal) Shell() string {
	defer log.Trace(time.Now())
	if len(term.CmdFlags.Shell) != 0 {
		return term.CmdFlags.Shell
	}

	log.Debug("no shell name provided in flags, trying to detect it")

	name := term.shellProcessName()
	if len(name) == 0 {
		return UNKNOWN
	}

	log.Debug("process name: " + name)

	// Cache the shell value to speed things up.
	term.CmdFlags.Shell = strings.Trim(strings.TrimSuffix(name, ".exe"), " ")
	return term.CmdFlags.Shell
}

func (term *Terminal) unWrapError(err error) error {
	cause := err
	for {
		type nested interface{ Unwrap() error }
		unwrap, ok := cause.(nested)
		if !ok {
			break
		}
		cause = unwrap.Unwrap()
	}
	return cause
}

func (term *Terminal) HTTPRequest(targetURL string, body io.Reader, timeout int, requestModifiers ...http.RequestModifier) ([]byte, error) {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return nil, errDataOnly
	}
	defer log.Trace(time.Now(), targetURL)

	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodGet, targetURL, body)
	if err != nil {
		return nil, err
	}

	for _, modifier := range requestModifiers {
		modifier(request)
	}

	if term.CmdFlags.Debug {
		log.Debug(dumpRequest(request))
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, term.unWrapError(err)
	}

	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		err := &http.Error{
			StatusCode: response.StatusCode,
		}
		log.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Debug(string(responseBody))

	return responseBody, nil
}

func (term *Terminal) HasParentFilePath(parent string, followSymlinks bool) (*FileInfo, error) {
	if term.CmdFlags != nil && term.CmdFlags.DataOnly {
		return nil, errDataOnly
	}
	defer log.Trace(time.Now(), parent)

	key := parent + "|" + strconv.FormatBool(followSymlinks)
	if term.parentFileMap != nil {
		if result, OK := term.parentFileMap.Get(key); OK {
			return result.info, result.err
		}
	}

	info, err := term.findParentFilePath(parent, followSymlinks)

	if term.parentFileMap != nil {
		term.parentFileMap.Set(key, parentFilePathResult{info: info, err: err})
	}

	return info, err
}

func (term *Terminal) findParentFilePath(parent string, followSymlinks bool) (*FileInfo, error) {
	pwd := term.Pwd()
	if followSymlinks {
		if actual, err := term.ResolveSymlink(pwd); err == nil {
			pwd = actual
		}
	}

	for {
		fileSystem := os.DirFS(pwd)
		info, err := fs.Stat(fileSystem, parent)
		if err == nil {
			return &FileInfo{
				ParentFolder: pwd,
				Path:         filepath.Join(pwd, parent),
				IsDir:        info.IsDir(),
			}, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		if dir := filepath.Dir(pwd); dir != pwd {
			pwd = dir
			continue
		}

		log.Error(err)
		return nil, errors.New("no match at root level")
	}
}

func (term *Terminal) StackCount() int {
	defer log.Trace(time.Now())

	if term.CmdFlags.StackCount < 0 {
		return 0
	}

	return term.CmdFlags.StackCount
}

func (term *Terminal) Logs() string {
	return log.String()
}

func (term *Terminal) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			log.Error(errors.New("panic"))
			match = false
		}
	}()
	match = dirMatchesOneOf(dir, term.Home(), term.GOOS(), regexes)
	return
}

func dirMatchesOneOf(dir, home, goos string, regexes []string) bool {
	if len(regexes) == 0 {
		return false
	}

	if goos == WINDOWS {
		dir = strings.ReplaceAll(dir, "\\", "/")
		home = strings.ReplaceAll(home, "\\", "/")
	}

	for _, element := range regexes {
		normalized := strings.ReplaceAll(element, "\\\\", "/")
		if strings.HasPrefix(normalized, "~") {
			rem := normalized[1:]
			if rem == "" || rem[0] == '/' {
				normalized = home + rem
			}
		}
		pattern := fmt.Sprintf("^%s$", normalized)
		if goos == WINDOWS || goos == DARWIN {
			pattern = "(?i)" + pattern
		}
		matched := regex.MatchString(pattern, dir)
		if matched {
			return true
		}
	}
	return false
}

func (term *Terminal) setPromptCount() {
	defer log.Trace(time.Now())

	var count int
	if val, found := cache.Session.Get[int](cache.PROMPTCOUNTCACHE); found {
		count = val
	}

	// Only update the count if we're generating a primary prompt.
	if term.CmdFlags.Type == PRIMARY {
		count++
		cache.Session.Set(cache.PROMPTCOUNTCACHE, count, cache.ONEDAY)
	}

	term.CmdFlags.PromptCount = count
}

func (term *Terminal) CursorPosition() (row, col int) {
	if number, err := strconv.Atoi(term.Getenv("POSH_CURSOR_LINE")); err == nil {
		row = number
	}

	if number, err := strconv.Atoi(term.Getenv("POSH_CURSOR_COLUMN")); err != nil {
		col = number
	}

	return
}

func cleanHostName(hostName string) string {
	garbage := []string{
		".lan",
		".local",
		".localdomain",
	}
	for _, g := range garbage {
		if strings.HasSuffix(hostName, g) {
			hostName = strings.Replace(hostName, g, "", 1)
		}
	}
	return hostName
}
