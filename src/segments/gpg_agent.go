package segments

import (
	"errors"
	"iter"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	GPGAgentFetchKeys    options.Option = "fetch_keys"
	GPGAgentKeySeparator options.Option = "key_separator"
	GPGAgentKeyTemplate  options.Option = "key_template"
	GPGAgentMappedKeys   options.Option = "mapped_keys"

	gpgAgentCacheDuration  = "10s"
	gpgAgentKeyInfoKey     = "gpgagent_keyinfo"
	gpgAgentRunningKey     = "gpgagent_running"
	gpgAgentGitKeyPrefix   = "gpgagent_git_"
	gpgAgentSSHSocket      = "S.gpg-agent.ssh"
	gpgAgentKeyInfoCommand = "keyinfo --list"
	gpgAgentInfoCommand    = "GETINFO pid"

	// Fields of a "S KEYINFO" status line, see the gpg-agent documentation. gpg-agent appended the cached,
	// protection, fingerprint, ttl and flags fields over successive releases, so only the first four are certain.
	keyInfoFields       = 6
	keyInfoKeygrip      = 2
	keyInfoCached       = 6
	keyInfoTTL          = 9
	gpgAgentStatusOK    = "OK"
	gpgAgentStatusKey   = "S"
	gpgAgentStatusEvent = "KEYINFO"

	// git config exits 1 when nothing matches the pattern; every other code is a real failure.
	gitConfigNoMatch = 1
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
	return "{{ if .Running }}  {{ if .Count }}{{ .CachedCount }} {{ end }}{{ end }}"
}

func (g *GPGAgent) Enabled() bool {
	g.TTYSet = g.env.Getenv("GPG_TTY") != ""
	g.SSHSupport = strings.HasSuffix(g.env.Getenv("SSH_AUTH_SOCK"), gpgAgentSSHSocket)
	g.readGitSigning()

	if !g.options.Bool(GPGAgentFetchKeys, false) {
		g.Running = agentAnswered(g.query(gpgAgentRunningKey, gpgAgentInfoCommand))
		return true
	}

	output := g.query(gpgAgentKeyInfoKey, gpgAgentKeyInfoCommand)

	g.Running = agentAnswered(output)
	if !g.Running {
		return true
	}

	keys := parseKeyInfo(output)
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

func (g *GPGAgent) query(key, command string) string {
	if output, found := cache.Device.Get[string](key); found {
		return output
	}

	output := g.queryAgent(command)
	cache.Device.Set(key, output, g.cacheDuration())

	return output
}

func (g *GPGAgent) queryAgent(command string) string {
	if !g.env.HasCommand("gpg-connect-agent") {
		return ""
	}

	// --no-autostart keeps the query from launching the agent it is meant to report on.
	output, err := g.env.RunCommand("gpg-connect-agent", "--no-autostart", command, "/bye")
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

	for line := range gpgAgentLines(output) {
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

	output, err := g.env.RunCommand("git", "config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`)
	if err != nil {
		var exitErr interface{ ExitCode() int }
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != gitConfigNoMatch {
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

func gpgAgentLines(output string) iter.Seq[string] {
	return strings.SplitSeq(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
}

// agentAnswered reports whether a gpg-agent replied. gpg-connect-agent exits 0 without an agent and only
// mentions it on stderr, so a closing "OK" status is what proves the agent answered.
func agentAnswered(output string) bool {
	for line := range gpgAgentLines(output) {
		if strings.TrimSpace(line) == gpgAgentStatusOK {
			return true
		}
	}

	return false
}

func parseKeyInfo(output string) []*GPGKey {
	var keys []*GPGKey

	for line := range gpgAgentLines(output) {
		fields := strings.Fields(line)
		if len(fields) < keyInfoFields || fields[0] != gpgAgentStatusKey || fields[1] != gpgAgentStatusEvent {
			continue
		}

		key := &GPGKey{Keygrip: fields[keyInfoKeygrip]}

		if len(fields) > keyInfoCached {
			key.Cached = fields[keyInfoCached] == "1"
		}

		if len(fields) > keyInfoTTL {
			key.TTL, _ = strconv.Atoi(fields[keyInfoTTL])
		}

		keys = append(keys, key)
	}

	return keys
}
