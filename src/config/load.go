package config

import (
	"bytes"
	"fmt"
	stdOS "os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/goutil/jsonutil"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	json "github.com/goccy/go-json"
	yaml "github.com/goccy/go-yaml"
	toml "github.com/pelletier/go-toml/v2"
)

// LoadConfig returns the default configuration including possible user overrides
func Load(env runtime.Environment) *Config {
	cfg := loadConfig(env)

	// only migrate automatically when the switch isn't set
	if !env.Flags().Migrate && cfg.Version < Version {
		cfg.BackupAndMigrate()
	}

	if !cfg.ShellIntegration {
		return cfg
	}

	// bash  - ok
	// fish  - ok
	// pwsh  - ok
	// zsh   - ok
	// cmd   - ok, as of v1.4.25 (chrisant996/clink#457, fixed in chrisant996/clink@8a5d7ea)
	// nu    - built-in (and bugged) feature - nushell/nushell#5585, https://www.nushell.sh/blog/2022-08-16-nushell-0_67.html#shell-integration-fdncred-and-tyriar
	// elv   - broken OSC sequences
	// xonsh - broken OSC sequences
	// tcsh  - overall broken, FTCS_COMMAND_EXECUTED could be added to POSH_POSTCMD in the future
	switch env.Shell() {
	case shell.ELVISH, shell.XONSH, shell.TCSH, shell.NU:
		cfg.ShellIntegration = false
	}

	return cfg
}

func loadConfig(env runtime.Environment) *Config {
	defer env.Trace(time.Now())
	configFile := env.Flags().Config

	if len(configFile) == 0 {
		env.Debug("no config file specified, using default")
		return Default(env, false)
	}

	var cfg Config
	cfg.origin = configFile
	cfg.Format = strings.TrimPrefix(filepath.Ext(configFile), ".")
	cfg.env = env

	data, err := stdOS.ReadFile(configFile)
	if err != nil {
		env.Error(err)
		return Default(env, true)
	}

	switch cfg.Format {
	case "yml", "yaml":
		cfg.Format = YAML
		err = yaml.Unmarshal(data, &cfg)
	case "jsonc", "json":
		cfg.Format = JSON

		str := jsonutil.StripComments(string(data))
		data = []byte(str)

		decoder := json.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&cfg)
	case "toml", "tml":
		cfg.Format = TOML
		err = toml.Unmarshal(data, &cfg)
	default:
		err = fmt.Errorf("unsupported config file format: %s", cfg.Format)
	}

	if err != nil {
		env.Error(err)
		return Default(env, true)
	}

	return &cfg
}
