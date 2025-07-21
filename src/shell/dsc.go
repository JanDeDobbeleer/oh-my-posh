package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

const (
	initCommandRegex = `oh-my-posh(?:\.exe)?\s+init`
)

func DSC() *dsc.Resource[*Shell] {
	return &dsc.Resource[*Shell]{
		JSONSchemaURL: "https://ohmyposh.dev/dsc.shell.schema.json",
	}
}

type Shell struct {
	Command string `json:"command,omitempty" jsonschema:"title=Command,description=The oh-my-posh init command to run"`
	Name    string `json:"name,omitempty" jsonschema:"title=Shell name,description=The name of the shell"`
}

func (s *Shell) Equal(shell *Shell) bool {
	if shell == nil {
		return false
	}

	return s.Name == shell.Name
}

func (s *Shell) Resolve() (*Shell, bool) {
	return nil, false
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
	case BASH:
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}

		return filepath.Join(home, ".bash_profile"), nil
	case ZSH:
		return filepath.Join(home, ".zshrc"), nil
	case FISH:
		configDir := filepath.Join(home, ".config", "fish")
		return filepath.Join(configDir, "config.fish"), nil
	case PWSH, PWSH5:
		return cmd.Run(s.Name, "-NoProfile", "-Command", "$PROFILE")
	case NU:
		return cmd.Run("nu", "-c", "$nu.config-path")
	case ELVISH:
		return filepath.Join(home, ".elvish", "rc.elv"), nil
	case XONSH:
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

	if !os.IsNotExist(err) {
		return nil
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
	log.Debug("current shell config content:\n", content)

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
	content = strings.Join(lines, "\n")
	log.Debug("updated shell config content:\n", content)

	return content, true
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
	case BASH, ZSH:
		return fmt.Sprintf(`eval "$(%s)"`, s.Command)
	case FISH:
		return s.Command + " | source"
	case PWSH, PWSH5:
		return s.Command + " | Invoke-Expression"
	case ELVISH:
		return fmt.Sprintf(`eval (%s)`, s.Command)
	case XONSH:
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
