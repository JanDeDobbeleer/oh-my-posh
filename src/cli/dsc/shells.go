package dsc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

const (
	initCommandRegex = `oh-my-posh(?:\.exe)?\s+init`
)

type Shells []*Shell

type Shell struct {
	Command string `json:"command,omitempty"`
	Name    string `json:"name,omitempty"`
}

func (s *Shells) Exists(name string) bool {
	for _, shell := range *s {
		if shell.Name == name {
			return true
		}
	}

	return false
}

func (s *Shells) Add(name, command string) {
	if s.Exists(name) {
		log.Debug("Shell already exists:", name)
		return
	}

	log.Debugf("adding shell %s with command %s", name, command)

	*s = append(*s, &Shell{
		Command: command,
		Name:    name,
	})
}

func (s *Shells) Apply() error {
	log.Debug("applying shells")

	var err error

	for _, shell := range *s {
		if applyErr := shell.Apply(); applyErr != nil {
			log.Error(applyErr)
			err = errors.Join(err, applyErr)
		}
	}

	log.Debug("shells applied")

	return err
}

func (s *Shell) Apply() error {
	if s.Command == "" {
		return nil
	}

	log.Debug("applying shell configuration with command: %s", s.Command)

	// Get the shell configuration file path
	configPath, err := s.getShellConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get shell config path: %w", err)
	}

	if err := s.validateShellConfigPath(configPath); err != nil {
		return err
	}

	// Read current configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Debug("failed to read shell config file")
		return err
	}

	contentStr, updated := s.updateShellConfig(string(content))
	if !updated {
		log.Debug("shell config already up to date, skipping write")
		return nil
	}

	return os.WriteFile(configPath, []byte(contentStr), 0644)
}

func (s *Shell) getShellConfigPath() (string, error) {
	home := path.Home()
	if home == "" {
		return "", fmt.Errorf("failed to get home directory")
	}

	switch s.Name {
	case shell.BASH:
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}

		return filepath.Join(home, ".bash_profile"), nil
	case shell.ZSH:
		return filepath.Join(home, ".zshrc"), nil
	case shell.FISH:
		configDir := filepath.Join(home, ".config", "fish")
		return filepath.Join(configDir, "config.fish"), nil
	case shell.PWSH, shell.PWSH5:
		return cmd.Run(s.Name, "-NoProfile", "-Command", "$PROFILE")
	case shell.NU:
		return cmd.Run("nu", "-c", "$nu.config-path")
	case shell.ELVISH:
		return filepath.Join(home, ".elvish", "rc.elv"), nil
	case shell.XONSH:
		return filepath.Join(home, ".xonshrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell type: %s", s.Name)
	}
}

func (s *Shell) validateShellConfigPath(configPath string) error {
	log.Debug("validating shell config path:", configPath)

	_, err := os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	log.Debug("shell config file does not exist")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		log.Debug("failed to create shell config directory")
		return err
	}

	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		log.Debug("failed to create shell config file")
		return err
	}

	return nil
}

func (s *Shell) updateShellConfig(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	initLinePos := s.getLastInitLinePosition(lines)

	if initLinePos < 0 {
		return s.addInitLine(content), true
	}

	initLineStr := lines[initLinePos]
	shellCommand := s.shellCommand()

	// validate if we have the same command
	if strings.Contains(initLineStr, shellCommand) {
		log.Debug("oh-my-posh already correctly configured")
		return content, false
	}

	lines[initLinePos] = whitespacePrefix(initLineStr) + shellCommand

	return strings.Join(lines, "\n"), true
}

func (s *Shell) addInitLine(content string) string {
	log.Debug("oh-my-posh not initialized, adding initialization")

	// Add the initialization command to the end of the file
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += s.shellCommand() + "\n"

	return content
}

func (s *Shell) getLastInitLinePosition(lines []string) int {
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if regex.MatchString(initCommandRegex, line) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			return i
		}
	}

	return -1
}

func (s *Shell) shellCommand() string {
	switch s.Name {
	case shell.BASH, shell.ZSH:
		return fmt.Sprintf(`eval "$(%s)"`, s.Command)
	case shell.FISH:
		return s.Command + " | source"
	case shell.PWSH, shell.PWSH5:
		return s.Command + " | Invoke-Expression"
	case shell.ELVISH:
		return fmt.Sprintf(`eval (%s)`, s.Command)
	case shell.XONSH:
		return fmt.Sprintf(`execx($(%s))`, s.Command)
	default:
		return s.Command
	}
}

func whitespacePrefix(s string) string {
	var builder strings.Builder

	for _, char := range s {
		if char == ' ' || char == '\t' {
			builder.WriteRune(char)
			continue
		}

		break
	}

	return builder.String()
}
