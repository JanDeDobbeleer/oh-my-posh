package segments

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	GPGAgentKeySeparator options.Option = "key_separator"
	GPGAgentKeyTemplate  options.Option = "key_template"
	GPGAgentMappedKeys   options.Option = "mapped_keys"

	gpgAgentCacheDuration = "10s"
	gpgAgentKeyInfoKey    = "gpgagent_keyinfo"
	gpgAgentGitKeyPrefix  = "gpgagent_git_"
	gpgAgentSSHSocket     = "S.gpg-agent.ssh"

	// Fields of a "S KEYINFO" status line, see the gpg-agent documentation.
	keyInfoFields       = 11
	keyInfoKeygrip      = 2
	keyInfoCached       = 6
	keyInfoTTL          = 9
	gpgAgentStatusOK    = "OK"
	gpgAgentStatusKey   = "S"
	gpgAgentStatusEvent = "KEYINFO"
)

type GPGAgent struct {
	Base

	Keys        template.Markup
	SigningKey  string
	Count       int
	CachedCount int
	Running     bool
	SignCommits bool
	TTYSet      bool
	SSHSupport  bool
}

type GPGKey struct {
	Alias   template.Markup
	Keygrip string
	TTL     int
	Cached  bool
}

func (g *GPGAgent) Template() string {
	return "{{ if .Running }}  {{ .CachedCount }} {{ end }}"
}

func (g *GPGAgent) Enabled() bool {
	g.TTYSet = g.env.Getenv("GPG_TTY") != ""
	g.SSHSupport = strings.HasSuffix(g.env.Getenv("SSH_AUTH_SOCK"), gpgAgentSSHSocket)
	g.readGitSigning()

	keys, running := parseKeyInfo(g.keyInfo())
	g.Running = running
	if !running {
		return true
	}

	g.Count = len(keys)
	for _, key := range keys {
		if key.Cached {
			g.CachedCount++
		}
	}

	g.Keys = g.renderKeys(keys)

	return true
}

func (g *GPGAgent) cacheDuration() cache.Duration {
	return cache.Duration(g.options.String(options.CacheDuration, gpgAgentCacheDuration))
}

func (g *GPGAgent) keyInfo() string {
	if output, found := cache.Device.Get[string](gpgAgentKeyInfoKey); found {
		return output
	}

	output := g.queryAgent()
	cache.Device.Set(gpgAgentKeyInfoKey, output, g.cacheDuration())

	return output
}

func (g *GPGAgent) queryAgent() string {
	if !g.env.HasCommand("gpg-connect-agent") {
		return ""
	}

	// --no-autostart keeps the query from launching the agent it is meant to report on.
	output, err := g.env.RunCommand("gpg-connect-agent", "--no-autostart", "keyinfo --list", "/bye")
	if err != nil {
		log.Error(err)
		return ""
	}

	return output
}

func (g *GPGAgent) readGitSigning() {
	key := gpgAgentGitKeyPrefix + g.env.Pwd()

	output, found := cache.Session.Get[string](key)
	if !found {
		output = g.queryGitConfig()
		cache.Session.Set(key, output, g.cacheDuration())
	}

	for line := range strings.SplitSeq(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		name, value, _ := strings.Cut(strings.TrimSpace(line), " ")

		switch name {
		case "user.signingkey":
			g.SigningKey = value
		case "commit.gpgsign":
			g.SignCommits = isGitTrue(value)
		}
	}
}

func (g *GPGAgent) queryGitConfig() string {
	if !g.env.HasCommand("git") {
		return ""
	}

	// git exits 1 when neither key is set, which is a valid answer rather than a failure.
	output, err := g.env.RunCommand("git", "config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`)
	if err != nil {
		var exitErr interface{ ExitCode() int }
		if !errors.As(err, &exitErr) {
			log.Error(err)
		}

		return ""
	}

	return output
}

func isGitTrue(value string) bool {
	switch strings.ToLower(value) {
	case "true", "yes", "on", "1":
		return true
	default:
		return false
	}
}

func (g *GPGAgent) renderKeys(keys []*GPGKey) template.Markup {
	aliases := g.options.KeyValueMap(GPGAgentMappedKeys, nil)
	separator := template.RawMarkup(g.options.String(GPGAgentKeySeparator, " "))
	keyTemplate := "{{ if .Cached }}{{ or .Alias .Keygrip }}{{ end }}"

	parts := make([]template.Markup, 0, 2*len(keys))
	for _, key := range keys {
		key.Alias = template.RawMarkup(aliases[key.Keygrip])

		rendered := template.RawMarkup(g.options.Template(GPGAgentKeyTemplate, keyTemplate, key))
		if rendered == "" {
			continue
		}

		if len(parts) > 0 {
			parts = append(parts, separator)
		}

		parts = append(parts, rendered)
	}

	return template.JoinMarkup(parts...)
}

// parseKeyInfo reads the output of `gpg-connect-agent "keyinfo --list" /bye`. gpg-connect-agent exits 0
// without an agent and only reports it on stderr, so a closing "OK" status is what proves the agent answered.
func parseKeyInfo(output string) ([]*GPGKey, bool) {
	var keys []*GPGKey
	running := false

	for line := range strings.SplitSeq(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 1 && fields[0] == gpgAgentStatusOK {
			running = true
			continue
		}

		if len(fields) < keyInfoFields || fields[0] != gpgAgentStatusKey || fields[1] != gpgAgentStatusEvent {
			continue
		}

		ttl, _ := strconv.Atoi(fields[keyInfoTTL])
		keys = append(keys, &GPGKey{
			Keygrip: fields[keyInfoKeygrip],
			Cached:  fields[keyInfoCached] == "1",
			TTL:     ttl,
		})
	}

	return keys, running
}
