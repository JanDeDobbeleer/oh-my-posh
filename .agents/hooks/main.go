package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	harnessClaude  = "claude"
	checkCacheFile = "omp-hooks-check"
	checkTimeout   = 9 * time.Minute
	maxOutputChars = 8000
	failureSummary = "Checks failed on the changed files; fix them before finishing:"
	lintInstall    = "golangci-lint not installed, lint skipped: https://golangci-lint.run/docs/welcome/install/"
	noElseHint     = "The codebase does not use else: rewrite with a guard clause, an early return, or a switch."
)

var targetOSes = []string{"linux", "darwin", "windows"}

// gofmt keeps else on the closing brace's line, so this skips the word in strings and comments.
var elseBranch = regexp.MustCompile(`^}\s*else\b`)

type hookPayload struct {
	Cwd            string `json:"cwd"`
	StopHookActive bool   `json:"stop_hook_active"`
}

type changes struct {
	goFiles          []string
	untrackedGoFiles []string
	markdownFiles    []string
	packages         []string
	platformCode     bool
}

type report struct {
	output bytes.Buffer
	failed bool
}

func (r *report) fail(step, output string) {
	r.failed = true
	fmt.Fprintf(&r.output, "%s:\n%s\n", step, strings.TrimSpace(output))
}

func main() {
	harness := parseHarnessFlag(os.Args[1:])
	payload := readPayload(os.Stdin)

	// Copilot CLI also reads .claude/settings.json, so a claude-harness invocation would fire a
	// second time under Copilot; CLAUDE_PROJECT_DIR is only set by Claude Code itself.
	claudeProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if harness == harnessClaude && claudeProjectDir == "" {
		return
	}

	root, ok := resolveRepoRoot(claudeProjectDir, payload.Cwd)
	if !ok {
		return
	}

	check(root, harness, payload)
}

func parseHarnessFlag(args []string) string {
	for i, arg := range args {
		if arg == "--harness" && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(arg, "--harness="); ok {
			return v
		}
	}
	return ""
}

func readPayload(r io.Reader) hookPayload {
	var payload hookPayload
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return payload
	}
	_ = json.Unmarshal(data, &payload)
	return payload
}

func resolveRepoRoot(claudeProjectDir, payloadCwd string) (string, bool) {
	dir := claudeProjectDir
	if dir == "" {
		dir = payloadCwd
	}
	if dir == "" {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	if dir == "" {
		return "", false
	}

	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

func check(root, harness string, payload hookPayload) {
	changed, err := changedFiles(root)
	if err != nil || (len(changed.goFiles) == 0 && len(changed.markdownFiles) == 0) {
		return
	}

	gitDir, err := absoluteGitDir(root)
	if err != nil {
		return
	}
	cachePath := filepath.Join(gitDir, checkCacheFile)
	if cached, err := os.ReadFile(cachePath); err == nil && string(cached) == hashFiles(root, changed) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	var r report
	if len(changed.goFiles) != 0 {
		checkGo(ctx, &r, filepath.Join(root, "src"), changed)
		checkNoElse(ctx, &r, root, changed)
	}
	if len(changed.markdownFiles) != 0 {
		checkMarkdown(ctx, &r, root, changed.markdownFiles)
	}

	if !r.failed {
		// Hash after the checks: the auto-fixes may have rewritten the files.
		_ = os.WriteFile(cachePath, []byte(hashFiles(root, changed)), 0o644)
		return
	}

	reportFailure(harness, payload.StopHookActive, failureSummary+"\n"+truncate(r.output.String(), maxOutputChars))
}

func changedFiles(root string) (changes, error) {
	var c changes

	out, err := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all").Output()
	if err != nil {
		return c, err
	}

	for line := range strings.SplitSeq(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		rel := parseStatusPath(line[3:])
		abs := filepath.Join(root, rel)
		untracked := strings.HasPrefix(line, "??")

		switch {
		case strings.HasPrefix(rel, "src/") && strings.HasSuffix(rel, ".go"):
			addPackage(&c, root, rel)
			if !isFile(abs) {
				continue
			}
			c.goFiles = append(c.goFiles, rel)
			c.platformCode = c.platformCode || isPlatformSpecific(abs)
			if untracked {
				c.untrackedGoFiles = append(c.untrackedGoFiles, rel)
			}
		case strings.HasSuffix(rel, ".md") && isFile(abs):
			c.markdownFiles = append(c.markdownFiles, rel)
		}
	}

	slices.Sort(c.goFiles)
	slices.Sort(c.markdownFiles)
	slices.Sort(c.packages)
	return c, nil
}

// parseStatusPath follows a rename ("old -> new") and strips git's quoting of unusual characters.
func parseStatusPath(field string) string {
	if idx := strings.Index(field, " -> "); idx >= 0 {
		field = field[idx+4:]
	}
	field = strings.TrimSpace(field)
	if unquoted, err := strconv.Unquote(field); err == nil {
		field = unquoted
	}
	return filepath.ToSlash(field)
}

// A deleted file still changes its package, as long as the package has Go files left.
func addPackage(c *changes, root, rel string) {
	dir := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(rel, "src/")))
	goFiles, _ := filepath.Glob(filepath.Join(root, "src", dir, "*.go"))
	if len(goFiles) == 0 {
		return
	}

	pkg := "./" + dir
	if !slices.Contains(c.packages, pkg) {
		c.packages = append(c.packages, pkg)
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isPlatformSpecific(path string) bool {
	name := strings.TrimSuffix(filepath.Base(path), ".go")
	name = strings.TrimSuffix(name, "_test")
	for _, suffix := range []string{"_windows", "_unix", "_darwin", "_linux"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	content, err := os.ReadFile(path)
	return err == nil && bytes.Contains(content, []byte("//go:build"))
}

func absoluteGitDir(root string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func hashFiles(root string, c changes) string {
	h := sha256.New()
	for _, f := range slices.Concat(c.goFiles, c.markdownFiles) {
		h.Write([]byte(f))
		content, _ := os.ReadFile(filepath.Join(root, f))
		h.Write(content)
	}
	for _, pkg := range c.packages {
		h.Write([]byte(pkg))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func checkGo(ctx context.Context, r *report, srcDir string, c changes) {
	lint, lintErr := exec.LookPath("golangci-lint")

	files := make([]string, 0, len(c.goFiles))
	for _, f := range c.goFiles {
		files = append(files, strings.TrimPrefix(f, "src/"))
	}

	if lintErr == nil {
		_, _ = runTool(ctx, srcDir, nil, append([]string{lint, "fmt"}, files...))
	} else {
		_, _ = runTool(ctx, srcDir, nil, append([]string{"gofmt", "-w"}, files...))
	}

	if len(c.packages) == 0 {
		return
	}

	modernize := toolCommand("modernize", "golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest")
	if out, err := runTool(ctx, srcDir, nil, slices.Concat(modernize, []string{"-fix"}, c.packages)); err != nil {
		r.fail("modernize", out)
	}

	// fieldalignment -fix is never used because it drops the comments on reordered struct fields.
	fieldalignment := toolCommand("fieldalignment", "golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest")
	if out, err := runTool(ctx, srcDir, nil, slices.Concat(fieldalignment, c.packages)); err != nil {
		r.fail("fieldalignment", out+"\nReorder the fields by hand, keeping their comments, and use named fields in struct literals.")
	}

	if lintErr == nil {
		if out, err := runTool(ctx, srcDir, nil, slices.Concat([]string{lint, "run"}, c.packages)); err != nil {
			r.fail("golangci-lint", out)
		}
	} else {
		fmt.Fprintln(os.Stderr, lintInstall)
	}

	if out, err := runTool(ctx, srcDir, nil, slices.Concat([]string{"go", "test"}, c.packages)); err != nil {
		r.fail("go test", out)
	}

	if c.platformCode {
		checkOtherPlatforms(ctx, r, srcDir, lint, lintErr == nil, c.packages)
	}
}

// CI builds and lints on every OS, but the toolchain only type-checks files for the host.
func checkOtherPlatforms(ctx context.Context, r *report, srcDir, lint string, hasLint bool, packages []string) {
	for _, goos := range targetOSes {
		if goos == runtime.GOOS {
			continue
		}

		env := []string{"GOOS=" + goos}
		if out, err := runTool(ctx, srcDir, env, slices.Concat([]string{"go", "build"}, packages)); err != nil {
			r.fail("go build (GOOS="+goos+")", out)
			continue
		}

		if !hasLint {
			continue
		}

		if out, err := runTool(ctx, srcDir, env, slices.Concat([]string{lint, "run"}, packages)); err != nil {
			r.fail("golangci-lint (GOOS="+goos+")", out)
		}
	}
}

func checkNoElse(ctx context.Context, r *report, root string, c changes) {
	// Untracked files never appear in git diff, so passing them as pathspecs is harmless.
	args := slices.Concat([]string{"-C", root, "diff", "HEAD", "-U0", "--no-color", "--no-ext-diff", "--dst-prefix=b/", "--"}, c.goFiles)
	// Output, not CombinedOutput: git writes line-ending warnings to stderr.
	diff, _ := exec.CommandContext(ctx, "git", args...).Output()

	hits := elseInDiff(string(diff))
	for _, f := range c.untrackedGoFiles {
		hits = append(hits, elseInFile(root, f)...)
	}

	if len(hits) == 0 {
		return
	}

	r.fail("no-else", strings.Join(hits, "\n")+"\n"+noElseHint)
}

func elseInDiff(diff string) []string {
	var hits []string
	var path string
	var lineNo int
	inHunk := false

	for line := range strings.SplitSeq(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			inHunk = false
		case strings.HasPrefix(line, "@@ "):
			inHunk = true
			lineNo = hunkStart(line)
		case !inHunk:
			if name, ok := strings.CutPrefix(line, "+++ "); ok {
				path = diffPath(name)
			}
		case strings.HasPrefix(line, "+"):
			if hit, ok := elseHit(path, lineNo, line[1:]); ok {
				hits = append(hits, hit)
			}
			lineNo++
		case strings.HasPrefix(line, " "):
			lineNo++
		}
	}

	return hits
}

// The new-file start line sits in the "+c,d" field of "@@ -a,b +c,d @@".
func hunkStart(header string) int {
	fields := strings.Fields(header)
	if len(fields) < 3 {
		return 0
	}

	start, _, _ := strings.Cut(strings.TrimPrefix(fields[2], "+"), ",")
	n, _ := strconv.Atoi(start)
	return n
}

// git appends a tab to names containing a space and quotes names with unusual characters.
func diffPath(name string) string {
	name = strings.TrimSuffix(name, "\t")
	if unquoted, err := strconv.Unquote(name); err == nil {
		name = unquoted
	}
	return strings.TrimPrefix(name, "b/")
}

func elseInFile(root, rel string) []string {
	content, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return nil
	}

	var hits []string
	lineNo := 0
	for line := range strings.Lines(string(content)) {
		lineNo++
		if hit, ok := elseHit(rel, lineNo, line); ok {
			hits = append(hits, hit)
		}
	}

	return hits
}

func elseHit(path string, lineNo int, line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !elseBranch.MatchString(trimmed) {
		return "", false
	}
	return fmt.Sprintf("%s:%d: %s", path, lineNo, trimmed), true
}

func checkMarkdown(ctx context.Context, r *report, root string, files []string) {
	linter := []string{"markdownlint-cli2"}
	if _, err := exec.LookPath("markdownlint-cli2"); err != nil {
		linter = []string{"npx", "--yes", "markdownlint-cli2"}
	}

	_, _ = runTool(ctx, root, nil, slices.Concat(linter, []string{"--fix"}, files))
	if out, err := runTool(ctx, root, nil, slices.Concat(linter, files)); err != nil {
		r.fail("markdownlint", out)
	}
}

// Cloud agent sessions rarely have the analyzers installed, so fall back to go run.
func toolCommand(binary, module string) []string {
	if path, err := exec.LookPath(binary); err == nil {
		return []string{path}
	}
	return []string{"go", "run", module}
}

func runTool(ctx context.Context, dir string, env, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	if len(env) != 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n... (truncated)"
}

// go run collapses the program's exit code to 1, so failures go through the JSON output both
// harnesses understand instead of exit code 2.
func reportFailure(harness string, stopHookActive bool, message string) {
	result := map[string]string{}
	switch {
	case !stopHookActive:
		result["decision"] = "block"
		result["reason"] = message
	case harness == harnessClaude:
		result["systemMessage"] = message
	default:
		fmt.Fprintln(os.Stderr, message)
		return
	}

	payload, _ := json.Marshal(result)
	fmt.Println(string(payload))
}
