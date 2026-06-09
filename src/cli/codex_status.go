package cli

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

	"github.com/spf13/cobra"
)

const codexMaxSessionFiles = 25

type codexLocalStatusOptions struct {
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

func codexStatusDataSource(cmd *cobra.Command) ([]byte, error) {
	options, err := codexLocalStatusOptionsFromCommand(cmd)
	if err != nil {
		return nil, err
	}

	return codexStatusFromLocalSessions(options)
}

func codexLocalStatusOptionsFromCommand(cmd *cobra.Command) (codexLocalStatusOptions, error) {
	sessionID, err := cmd.Flags().GetString("session")
	if err != nil {
		return codexLocalStatusOptions{}, err
	}

	codexHome, err := cmd.Flags().GetString("codex-home")
	if err != nil {
		return codexLocalStatusOptions{}, err
	}

	sessionRoot, err := cmd.Flags().GetString("session-root")
	if err != nil {
		return codexLocalStatusOptions{}, err
	}

	if codexHome == "" {
		codexHome = defaultCodexHome()
	}

	if sessionRoot == "" && codexHome != "" {
		sessionRoot = filepath.Join(codexHome, "sessions")
	}

	return codexLocalStatusOptions{
		CodexHome:   codexHome,
		SessionRoot: sessionRoot,
		SessionID:   sessionID,
	}, nil
}

func defaultCodexHome() string {
	if home := os.Getenv("CODEX_HOME"); home != "" {
		return home
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".codex")
}

func codexStatusFromLocalSessions(options codexLocalStatusOptions) ([]byte, error) {
	if options.SessionRoot == "" {
		return nil, errors.New("codex sessions directory could not be determined")
	}

	files, err := codexSessionFiles(options.SessionRoot, options.SessionID)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Codex session files found in %s", options.SessionRoot)
	}

	cfg := loadCodexConfigStatus(options.CodexHome)
	for _, file := range files {
		data, err := codexStatusFromSessionFile(file.path, cfg)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}

	if options.SessionID != "" {
		return nil, fmt.Errorf("no token_count event found for Codex session %s", options.SessionID)
	}

	return nil, fmt.Errorf("no token_count event found in the newest Codex session files")
}

func codexSessionFiles(root, sessionID string) ([]codexSessionFile, error) {
	files := []codexSessionFile{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
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
			return nil
		}

		files = append(files, codexSessionFile{
			path:    path,
			modTime: info.ModTime().UnixNano(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	if sessionID == "" && len(files) > codexMaxSessionFiles {
		files = files[:codexMaxSessionFiles]
	}

	return files, nil
}

func codexStatusFromSessionFile(path string, cfg codexConfigStatus) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var meta codexSessionMeta
	var latest map[string]any

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			event, ok := parseCodexJSONLEvent(line)
			if ok {
				switch event.Type {
				case "session_meta":
					_ = json.Unmarshal(event.Payload, &meta)
				case "event_msg":
					payload, ok := parseCodexTokenCountPayload(event.Payload)
					if ok {
						latest = payload
					}
				}
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	if latest == nil {
		return nil, errors.New("no token_count event found")
	}

	enrichCodexStatusPayload(latest, &meta, cfg)

	return json.Marshal(latest)
}

func parseCodexJSONLEvent(line []byte) (codexJSONLEvent, bool) {
	var event codexJSONLEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return codexJSONLEvent{}, false
	}

	return event, true
}

func parseCodexTokenCountPayload(payload json.RawMessage) (map[string]any, bool) {
	var typed struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &typed); err != nil || typed.Type != "token_count" {
		return nil, false
	}

	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, false
	}

	return result, true
}

func enrichCodexStatusPayload(payload map[string]any, meta *codexSessionMeta, cfg codexConfigStatus) {
	setIfMissing(payload, "thread_id", meta.ID)
	setIfMissing(payload, "cwd", meta.CWD)
	setIfMissing(payload, "version", meta.CLIVersion)
	setIfMissing(payload, "approval_mode", firstNonEmpty(meta.ApprovalMode, cfg.ApprovalPolicy))
	setIfMissing(payload, "sandbox_policy", firstNonEmpty(meta.SandboxPolicy, cfg.SandboxMode))
	setIfMissing(payload, "reasoning_effort", firstNonEmpty(meta.ReasoningEffort, cfg.ModelReasoningEffort))

	if _, ok := payload["model"]; ok {
		return
	}

	model := firstNonEmpty(meta.Model, cfg.Model)
	if model == "" {
		return
	}

	payload["model"] = map[string]string{
		"id":           model,
		"display_name": model,
	}
}

func setIfMissing(payload map[string]any, key, value string) {
	if value == "" {
		return
	}

	if existing, ok := payload[key]; ok && existing != nil {
		if text, ok := existing.(string); !ok || text != "" {
			return
		}
	}

	payload[key] = value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
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
