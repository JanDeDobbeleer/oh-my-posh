package config

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/jsonutil"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	json "github.com/goccy/go-json"
	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v3"
)

const (
	defaultHash = "default"
)

// Load returns the default configuration including possible user overrides
func Load(configFile, sh string, migrate bool) (*Config, string) {
	defer log.Trace(time.Now())

	configFile, err := filePath(configFile)
	if err != nil {
		log.Error(err)
		return Default(true), defaultHash
	}

	cfg, hash := readConfig(configFile)

	// only migrate automatically when the switch isn't set
	if !migrate && cfg.Version < Version {
		cfg.BackupAndMigrate()
	}

	if cfg.Upgrade == nil {
		cfg.Upgrade = &upgrade.Config{
			Source:        upgrade.CDN,
			DisplayNotice: cfg.UpgradeNotice,
			Auto:          cfg.AutoUpgrade,
			Interval:      cache.ONEWEEK,
		}
	}

	if cfg.Upgrade.Interval.IsEmpty() {
		cfg.Upgrade.Interval = cache.ONEWEEK
	}

	cfg.Source = configFile

	if cfg.extended {
		fileName := fmt.Sprintf("%s.omp.json", hash)
		cfg.Source = filepath.Join(cache.Path(), fileName)
		cfg.Write(JSON)
	}

	if !cfg.ShellIntegration {
		return cfg, hash
	}

	// bash  - ok
	// fish  - ok
	// pwsh  - ok
	// zsh   - ok
	// cmd   - ok, as of v1.4.25 (chrisant996/clink#457, fixed in chrisant996/clink@8a5d7ea)
	// nu    - built-in (and bugged) feature - nushell/nushell#5585, https://www.nushell.sh/blog/2022-08-16-nushell-0_67.html#shell-integration-fdncred-and-tyriar
	// elv   - broken OSC sequences
	// xonsh - broken OSC sequences
	switch sh {
	case shell.ELVISH, shell.XONSH, shell.NU:
		cfg.ShellIntegration = false
	}

	return cfg, hash
}

func filePath(config string) (string, error) {
	defer log.Trace(time.Now())

	// if the config flag is set, we'll use that over POSH_THEME
	// in our internal shell logic, we'll always use the POSH_THEME
	// due to not using --config to set the configuration
	hasConfig := len(config) > 0

	if poshTheme := os.Getenv("POSH_THEME"); len(poshTheme) > 0 && !hasConfig {
		log.Debug("config set using POSH_THEME:", poshTheme)
		return poshTheme, nil
	}

	if !hasConfig {
		return "", errors.New("no config file specified")
	}

	if strings.HasPrefix(config, "https://") {
		filePath, err := Download(cache.Path(), config)
		if err != nil {
			log.Error(err)
			return "", err
		}

		return filePath, nil
	}

	isCygwin := func() bool {
		return runtime.GOOS == "windows" && len(os.Getenv("OSTYPE")) > 0
	}

	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if isCygwin() {
		log.Debug("cygwin detected, using full path for config")
		return config, nil
	}

	configFile := path.ReplaceTildePrefixWithHomeDir(config)

	abs, err := filepath.Abs(configFile)
	if err != nil {
		log.Error(err)
		return filepath.Clean(configFile), nil
	}

	return abs, nil
}

func readConfig(configFile string) (*Config, string) {
	defer log.Trace(time.Now())

	if len(configFile) == 0 {
		log.Debug("no config file specified, using default")
		return Default(false), defaultHash
	}

	var cfg Config
	cfg.Source = configFile
	cfg.Format = strings.TrimPrefix(filepath.Ext(configFile), ".")

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Error(err)
		return Default(true), defaultHash
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
		log.Error(err)
		return Default(true), defaultHash
	}

	// Calculate FNV-1a hash of the raw config data
	data = append(data, []byte(configFile)...) // Include the file path in the hash to enable file modification detection
	hasher := fnv.New64a()
	hasher.Write(data)
	hash := strconv.FormatUint(hasher.Sum64(), 16)

	if len(cfg.Extends) == 0 {
		return &cfg, hash
	}

	basePath, err := filePath(cfg.Extends)
	if err != nil {
		log.Error(err)
		return &cfg, hash
	}

	base, baseHash := readConfig(basePath)
	err = base.merge(&cfg)
	if err != nil {
		log.Error(err)
	}

	return base, fmt.Sprintf("%s.%s", hash, baseHash)
}
