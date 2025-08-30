package config

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	runtimelib "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/jsonutil"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"

	json "github.com/goccy/go-json"
	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v3"
)

const (
	defaultHash = "default"
)

// custom no config error
var ErrNoConfig = errors.New("no config file specified")

func Load(configFile string, migrate bool) (*Config, string) {
	defer log.Trace(time.Now())

	configFile, err := filePath(configFile)
	if err != nil {
		log.Error(err)
		warning := !errors.Is(err, ErrNoConfig)
		return Default(warning), defaultHash
	}

	cfg, hash := readConfig(configFile)

	// only migrate automatically when the switch isn't set
	if !migrate && cfg.Version < Version {
		cfg.BackupAndMigrate()
	}

	cfg.Source = configFile

	processExtendedConfig(cfg, hash)

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

	return cfg, hash
}

func filePath(config string) (string, error) {
	defer log.Trace(time.Now())

	// if the config flag is set, we'll use that over POSH_THEME
	// in our internal shell logic, we'll always use the POSH_THEME
	// due to not using --config to set the configuration
	hasConfig := config != ""

	if poshTheme := os.Getenv("POSH_THEME"); poshTheme != "" && !hasConfig {
		log.Debug("config set using POSH_THEME:", poshTheme)
		return poshTheme, nil
	}

	if !hasConfig {
		return "", ErrNoConfig
	}

	if url, OK := isTheme(config); OK {
		log.Debug("theme detected, using theme file")
		config = url
	}

	if strings.HasPrefix(config, "https://") {
		filePath, err := Download(cache.Path(), config)
		if err != nil {
			log.Error(err)
			return "", err
		}

		return filePath, nil
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

	if configFile == "" {
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
	case YAML, YML:
		cfg.Format = YAML
		err = yaml.Unmarshal(data, &cfg)
	case JSONC, JSON:
		cfg.Format = JSON

		str := jsonutil.StripComments(string(data))
		data = []byte(str)

		decoder := json.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&cfg)
	case TOML, TML:
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
	hasher := fnv.New64a()
	hasher.Write(data)
	hasher.Write([]byte(configFile)) // Include the file path in the hash to enable file modification detection
	hash := strconv.FormatUint(hasher.Sum64(), 16)

	if cfg.Extends == "" {
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

// isCygwin checks if we're running in Cygwin environment
func isCygwin() bool {
	return runtimelib.GOOS == "windows" && len(os.Getenv("OSTYPE")) > 0
}

// processExtendedConfig writes extended config to cache if needed
func processExtendedConfig(cfg *Config, hash string) {
	if !cfg.extended {
		return
	}

	fileName := fmt.Sprintf("%s.omp.json", hash)
	cfg.Source = filepath.Join(cache.Path(), fileName)
	cfg.Write(JSON)
}

func isTheme(config string) (string, bool) {
	themes := map[string]string{
		"1_shell":                  "1_shell.omp.json",
		"m365princess":             "M365Princess.omp.json",
		"agnoster":                 "agnoster.omp.json",
		"agnoster.minimal":         "agnoster.minimal.omp.json",
		"agnosterplus":             "agnosterplus.omp.json",
		"aliens":                   "aliens.omp.json",
		"amro":                     "amro.omp.json",
		"atomic":                   "atomic.omp.json",
		"atomicbit":                "atomicBit.omp.json",
		"avit":                     "avit.omp.json",
		"blue-owl":                 "blue-owl.omp.json",
		"blueish":                  "blueish.omp.json",
		"bubbles":                  "bubbles.omp.json",
		"bubblesextra":             "bubblesextra.omp.json",
		"bubblesline":              "bubblesline.omp.json",
		"capr4n":                   "capr4n.omp.json",
		"catppuccin":               "catppuccin.omp.json",
		"catppuccin_frappe":        "catppuccin_frappe.omp.json",
		"catppuccin_latte":         "catppuccin_latte.omp.json",
		"catppuccin_macchiato":     "catppuccin_macchiato.omp.json",
		"catppuccin_mocha":         "catppuccin_mocha.omp.json",
		"cert":                     "cert.omp.json",
		"chips":                    "chips.omp.json",
		"cinnamon":                 "cinnamon.omp.json",
		"clean-detailed":           "clean-detailed.omp.json",
		"cloud-context":            "cloud-context.omp.json",
		"cloud-native-azure":       "cloud-native-azure.omp.json",
		"cobalt2":                  "cobalt2.omp.json",
		"craver":                   "craver.omp.json",
		"darkblood":                "darkblood.omp.json",
		"devious-diamonds":         "devious-diamonds.omp.yaml",
		"di4am0nd":                 "di4am0nd.omp.json",
		"dracula":                  "dracula.omp.json",
		"easy-term":                "easy-term.omp.json",
		"emodipt":                  "emodipt.omp.json",
		"emodipt-extend":           "emodipt-extend.omp.json",
		"fish":                     "fish.omp.json",
		"free-ukraine":             "free-ukraine.omp.json",
		"froczh":                   "froczh.omp.json",
		"glowsticks":               "glowsticks.omp.yaml",
		"gmay":                     "gmay.omp.json",
		"grandpa-style":            "grandpa-style.omp.json",
		"gruvbox":                  "gruvbox.omp.json",
		"half-life":                "half-life.omp.json",
		"honukai":                  "honukai.omp.json",
		"hotstick.minimal":         "hotstick.minimal.omp.json",
		"hul10":                    "hul10.omp.json",
		"hunk":                     "hunk.omp.json",
		"huvix":                    "huvix.omp.json",
		"if_tea":                   "if_tea.omp.json",
		"illusi0n":                 "illusi0n.omp.json",
		"iterm2":                   "iterm2.omp.json",
		"jandedobbeleer":           "jandedobbeleer.omp.json",
		"jblab_2021":               "jblab_2021.omp.json",
		"jonnychipz":               "jonnychipz.omp.json",
		"json":                     "json.omp.json",
		"jtracey93":                "jtracey93.omp.json",
		"jv_sitecorian":            "jv_sitecorian.omp.json",
		"kali":                     "kali.omp.json",
		"kushal":                   "kushal.omp.json",
		"lambda":                   "lambda.omp.json",
		"lambdageneration":         "lambdageneration.omp.json",
		"larserikfinholt":          "larserikfinholt.omp.json",
		"lightgreen":               "lightgreen.omp.json",
		"marcduiker":               "marcduiker.omp.json",
		"markbull":                 "markbull.omp.json",
		"material":                 "material.omp.json",
		"microverse-power":         "microverse-power.omp.json",
		"mojada":                   "mojada.omp.json",
		"montys":                   "montys.omp.json",
		"mt":                       "mt.omp.json",
		"multiverse-neon":          "multiverse-neon.omp.json",
		"negligible":               "negligible.omp.json",
		"neko":                     "neko.omp.json",
		"night-owl":                "night-owl.omp.json",
		"nordtron":                 "nordtron.omp.json",
		"nu4a":                     "nu4a.omp.json",
		"onehalf.minimal":          "onehalf.minimal.omp.json",
		"paradox":                  "paradox.omp.json",
		"pararussel":               "pararussel.omp.json",
		"patriksvensson":           "patriksvensson.omp.json",
		"peru":                     "peru.omp.json",
		"pixelrobots":              "pixelrobots.omp.json",
		"plague":                   "plague.omp.json",
		"poshmon":                  "poshmon.omp.json",
		"powerlevel10k_classic":    "powerlevel10k_classic.omp.json",
		"powerlevel10k_lean":       "powerlevel10k_lean.omp.json",
		"powerlevel10k_modern":     "powerlevel10k_modern.omp.json",
		"powerlevel10k_rainbow":    "powerlevel10k_rainbow.omp.json",
		"powerline":                "powerline.omp.json",
		"probua.minimal":           "probua.minimal.omp.json",
		"pure":                     "pure.omp.json",
		"quick-term":               "quick-term.omp.json",
		"remk":                     "remk.omp.json",
		"robbyrussell":             "robbyrussell.omp.json",
		"rudolfs-dark":             "rudolfs-dark.omp.json",
		"rudolfs-light":            "rudolfs-light.omp.json",
		"sim-web":                  "sim-web.omp.json",
		"slim":                     "slim.omp.json",
		"slimfat":                  "slimfat.omp.json",
		"smoothie":                 "smoothie.omp.json",
		"sonicboom_dark":           "sonicboom_dark.omp.json",
		"sonicboom_light":          "sonicboom_light.omp.json",
		"sorin":                    "sorin.omp.json",
		"space":                    "space.omp.json",
		"spaceship":                "spaceship.omp.json",
		"star":                     "star.omp.json",
		"stelbent-compact.minimal": "stelbent-compact.minimal.omp.json",
		"stelbent.minimal":         "stelbent.minimal.omp.json",
		"takuya":                   "takuya.omp.json",
		"the-unnamed":              "the-unnamed.omp.json",
		"thecyberden":              "thecyberden.omp.json",
		"tiwahu":                   "tiwahu.omp.json",
		"tokyo":                    "tokyo.omp.json",
		"tokyonight_storm":         "tokyonight_storm.omp.json",
		"tonybaloney":              "tonybaloney.omp.json",
		"uew":                      "uew.omp.json",
		"unicorn":                  "unicorn.omp.json",
		"velvet":                   "velvet.omp.json",
		"wholespace":               "wholespace.omp.json",
		"wopian":                   "wopian.omp.json",
		"xtoys":                    "xtoys.omp.json",
		"ys":                       "ys.omp.json",
		"zash":                     "zash.omp.json",
	}

	config = strings.ToLower(config)
	if theme, ok := themes[config]; ok {
		log.Debug("theme found:", config)
		url := fmt.Sprintf("https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/refs/tags/v%s/themes/%s", build.Version, theme)
		return url, true
	}

	log.Debug("theme not found for:", config)

	return "", false
}
