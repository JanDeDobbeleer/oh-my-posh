package segments

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	codexMaxSessionFiles          = 25
	codexMaxSessionDirs           = codexMaxSessionFiles
	codexMaxRequestedSessionFiles = 250
)

var errCodexNoTokenCount = errors.New("no token_count event found")

// CodexLocalStatusOptions configures local Codex session transcript discovery.
type CodexLocalStatusOptions struct {
	CodexHome   string
	SessionRoot string
	SessionID   string
}

type codexSessionFile struct {
	path    string
	sortKey string
}

type codexJSONLEvent struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type codexSessionMeta struct {
	ID              string `json:"id"`
	CWD             string `json:"cwd"`
	CLIVersion      string `json:"cli_version"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
	ApprovalMode    string `json:"approval_policy"`
	SandboxPolicy   string `json:"sandbox_mode"`
}

type codexConfigStatus struct {
	Model                string `toml:"model"`
	ModelReasoningEffort string `toml:"model_reasoning_effort"`
	ApprovalPolicy       string `toml:"approval_policy"`
	SandboxMode          string `toml:"sandbox_mode"`
}

// ResolveCodexLocalStatusOptions fills default local Codex discovery paths.
func ResolveCodexLocalStatusOptions(options CodexLocalStatusOptions, codexHomeEnv, userHome string) CodexLocalStatusOptions {
	if options.CodexHome == "" {
		options.CodexHome = defaultCodexHome(codexHomeEnv, userHome)
	}

	if options.SessionRoot == "" && options.CodexHome != "" {
		options.SessionRoot = filepath.Join(options.CodexHome, "sessions")
	}

	return options
}

// CodexStatusFromLocalSessions returns the latest token_count status from local Codex session transcripts.
func CodexStatusFromLocalSessions(options CodexLocalStatusOptions) (CodexData, error) {
	if options.SessionRoot == "" {
		return CodexData{}, errors.New("codex sessions directory could not be determined; set CODEX_HOME, pass --codex-home or --session-root, or configure codex_home/session_root")
	}

	files, err := codexSessionFiles(options.SessionRoot, options.SessionID)
	if err != nil {
		return CodexData{}, err
	}

	if len(files) == 0 {
		return CodexData{}, fmt.Errorf("no codex session files found in %s; start Codex once, set CODEX_HOME, pass --codex-home or --session-root, or configure codex_home/session_root", options.SessionRoot)
	}

	cfg := loadCodexConfigStatus(options.CodexHome)
	var lastErr error

	for _, file := range files {
		data, err := codexStatusFromSessionFile(file.path, cfg)
		if err == nil && options.SessionID != "" && data.ThreadID != options.SessionID {
			continue
		}

		if err == nil && data.hasStatus() {
			return data, nil
		}

		if err != nil && !errors.Is(err, errCodexNoTokenCount) {
			lastErr = fmt.Errorf("%s: %w", file.path, err)
		}
	}

	if lastErr != nil {
		return CodexData{}, lastErr
	}

	if options.SessionID != "" {
		return CodexData{}, fmt.Errorf("%w for codex session %s", errCodexNoTokenCount, options.SessionID)
	}

	return CodexData{}, fmt.Errorf("%w in the newest codex session files", errCodexNoTokenCount)
}

func codexSessionFiles(root, sessionID string) ([]codexSessionFile, error) {
	files := []codexSessionFile{}
	limit := codexSessionFileLimit(sessionID)

	dirs, err := codexSessionSearchDirs(root)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, file := range codexSessionFilesInDir(dir, entries) {
			files = append(files, file)
			if len(files) > limit {
				sortCodexSessionFiles(files)
				files = files[:limit]
			}
		}

		if sessionID == "" && len(files) >= limit {
			break
		}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range codexSessionFilesInDir(root, entries) {
		files = append(files, file)
		if len(files) > limit {
			sortCodexSessionFiles(files)
			files = files[:limit]
		}
	}

	sortCodexSessionFiles(files)
	return files, nil
}

func codexSessionFileLimit(sessionID string) int {
	if sessionID != "" {
		return codexMaxRequestedSessionFiles
	}

	return codexMaxSessionFiles
}

func codexSessionSearchDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	years := filterCodexSessionDirs(entries, 4)
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	dirs := []string{}
	for _, year := range years {
		yearPath := filepath.Join(root, year)
		monthEntries, err := os.ReadDir(yearPath)
		if err != nil {
			return nil, err
		}

		months := filterCodexSessionDirs(monthEntries, 2)
		sort.Sort(sort.Reverse(sort.StringSlice(months)))

		for _, month := range months {
			monthPath := filepath.Join(yearPath, month)
			dayEntries, err := os.ReadDir(monthPath)
			if err != nil {
				return nil, err
			}

			days := filterCodexSessionDirs(dayEntries, 2)
			sort.Sort(sort.Reverse(sort.StringSlice(days)))

			for _, day := range days {
				dirs = append(dirs, filepath.Join(monthPath, day))
				if len(dirs) >= codexMaxSessionDirs {
					return dirs, nil
				}
			}
		}
	}

	return dirs, nil
}

func filterCodexSessionDirs(entries []os.DirEntry, length int) []string {
	dirs := []string{}
	for _, entry := range entries {
		if entry.IsDir() && isDigits(entry.Name(), length) {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs
}

func isDigits(value string, length int) bool {
	if len(value) != length {
		return false
	}

	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func codexSessionFilesInDir(dir string, entries []os.DirEntry) []codexSessionFile {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	files := []codexSessionFile{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		files = append(files, codexSessionFile{
			path:    path,
			sortKey: entry.Name(),
		})
	}

	return files
}

func sortCodexSessionFiles(files []codexSessionFile) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].sortKey == files[j].sortKey {
			return files[i].path > files[j].path
		}

		return files[i].sortKey > files[j].sortKey
	})
}

func codexStatusFromSessionFile(path string, cfg codexConfigStatus) (CodexData, error) {
	file, err := os.Open(path)
	if err != nil {
		return CodexData{}, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var meta codexSessionMeta
	var latest json.RawMessage

	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			event, parseErr := parseCodexJSONLEvent(line)
			if parseErr != nil {
				return CodexData{}, parseErr
			}

			switch event.Type {
			case "session_meta":
				if parseErr := json.Unmarshal(event.Payload, &meta); parseErr != nil {
					return CodexData{}, parseErr
				}
			case "event_msg":
				matches, parseErr := isCodexTokenCountPayload(event.Payload)
				if parseErr != nil {
					return CodexData{}, parseErr
				}

				if matches {
					latest = event.Payload
				}
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return CodexData{}, err
		}
	}

	if latest == nil {
		return CodexData{}, errCodexNoTokenCount
	}

	var data CodexData
	if err := json.Unmarshal(latest, &data); err != nil {
		return CodexData{}, err
	}

	enrichCodexStatusPayload(&data, &meta, cfg)

	return data, nil
}

func parseCodexJSONLEvent(line []byte) (codexJSONLEvent, error) {
	var event codexJSONLEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return codexJSONLEvent{}, err
	}

	return event, nil
}

func isCodexTokenCountPayload(payload json.RawMessage) (bool, error) {
	var typed struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &typed); err != nil {
		return false, err
	}

	return typed.Type == "token_count", nil
}

func enrichCodexStatusPayload(payload *CodexData, meta *codexSessionMeta, cfg codexConfigStatus) {
	setStringIfMissing(&payload.ThreadID, meta.ID)
	setStringIfMissing(&payload.CWD, meta.CWD)
	setStringIfMissing(&payload.Version, meta.CLIVersion)
	setStringIfMissing(&payload.ApprovalMode, firstNonEmpty(meta.ApprovalMode, cfg.ApprovalPolicy))
	setStringIfMissing(&payload.SandboxPolicy, firstNonEmpty(meta.SandboxPolicy, cfg.SandboxMode))
	setStringIfMissing(&payload.ReasoningEffort, firstNonEmpty(meta.ReasoningEffort, cfg.ModelReasoningEffort))

	model := firstNonEmpty(meta.Model, cfg.Model)
	if model == "" {
		return
	}

	setStringIfMissing(&payload.Model.ID, model)
	setStringIfMissing(&payload.Model.DisplayName, model)
}

func setStringIfMissing(target *string, value string) {
	if *target != "" || value == "" {
		return
	}

	*target = value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func defaultCodexHome(codexHomeEnv, userHome string) string {
	if codexHomeEnv != "" {
		return codexHomeEnv
	}

	if userHome == "" {
		return ""
	}

	return filepath.Join(userHome, ".codex")
}

func loadCodexConfigStatus(codexHome string) codexConfigStatus {
	if codexHome == "" {
		return codexConfigStatus{}
	}

	data, err := os.ReadFile(filepath.Join(codexHome, "config.toml"))
	if err != nil {
		return codexConfigStatus{}
	}

	var cfg codexConfigStatus
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return codexConfigStatus{}
	}

	return cfg
}
