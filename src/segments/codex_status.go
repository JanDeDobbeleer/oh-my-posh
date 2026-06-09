package segments

import (
	"bufio"
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

const codexMaxSessionFiles = 25

var errCodexNoTokenCount = errors.New("no token_count event found")

// CodexLocalStatusOptions configures local Codex session transcript discovery.
type CodexLocalStatusOptions struct {
	CodexHome   string
	SessionRoot string
	SessionID   string
}

type codexSessionFile struct {
	path    string
	modTime int64
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
		return CodexData{}, errors.New("codex sessions directory could not be determined")
	}

	files, err := codexSessionFiles(options.SessionRoot, options.SessionID)
	if err != nil {
		return CodexData{}, err
	}

	if len(files) == 0 {
		return CodexData{}, fmt.Errorf("no codex session files found in %s", options.SessionRoot)
	}

	cfg := loadCodexConfigStatus(options.CodexHome)
	var lastErr error

	for _, file := range files {
		data, err := codexStatusFromSessionFile(file.path, cfg)
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
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			return nil
		}

		if sessionID != "" && !strings.Contains(entry.Name(), sessionID) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		files = appendNewestCodexSessionFile(files, codexSessionFile{
			path:    path,
			modTime: info.ModTime().UnixNano(),
		}, sessionID == "")

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	return files, nil
}

func appendNewestCodexSessionFile(files []codexSessionFile, file codexSessionFile, bound bool) []codexSessionFile {
	files = append(files, file)
	if !bound || len(files) <= codexMaxSessionFiles {
		return files
	}

	oldestIndex := 0
	for i := 1; i < len(files); i++ {
		if files[i].modTime < files[oldestIndex].modTime {
			oldestIndex = i
		}
	}

	return append(files[:oldestIndex], files[oldestIndex+1:]...)
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
		if len(line) > 0 {
			event, ok := parseCodexJSONLEvent(line)
			if ok {
				switch event.Type {
				case "session_meta":
					_ = json.Unmarshal(event.Payload, &meta)
				case "event_msg":
					if isCodexTokenCountPayload(event.Payload) {
						latest = event.Payload
					}
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

func parseCodexJSONLEvent(line []byte) (codexJSONLEvent, bool) {
	var event codexJSONLEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return codexJSONLEvent{}, false
	}

	return event, true
}

func isCodexTokenCountPayload(payload json.RawMessage) bool {
	var typed struct {
		Type string `json:"type"`
	}

	return json.Unmarshal(payload, &typed) == nil && typed.Type == "token_count"
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
