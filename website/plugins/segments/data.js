/**
 * Curated tables for the segment catalog. Everything here is hand-derived and reviewed —
 * nothing in this file is computed from the registry or the docs. `website/plugins/segments/index.js`
 * asserts every key below matches a real doc id and throws naming the offender otherwise, so a
 * stale entry here fails loudly instead of silently disappearing from the catalog.
 */

// Doc id -> auth tier. Keyed by doc id (not config `type`) so the two alias cases
// (golang/go, copilot-cli/copilot_cli) can't create ambiguity: neither alias needs auth,
// and the two segments that do (`copilot`, `ytm`) have doc id === type anyway.
const AUTH_TIERS = {
  // health: OAuth device/browser flow against the vendor
  strava: 'oauth',
  withings: 'oauth',
  // cli: credentials obtained via `oh-my-posh auth <arg>` (src/cli/auth.go ValidArgs)
  copilot: 'cli',
  ytm: 'cli',
  // music/web: a plain API key pasted into the segment config
  lastfm: 'apikey',
  owm: 'apikey',
  todoist: 'apikey',
  brewfather: 'apikey',
  // web/health: a token embedded in a URL the segment calls
  nightscout: 'url',
  wakatime: 'url',
  http: 'url',
};

// Tier id -> user-facing label. Order here is the fixed display order for `authTiers`.
const AUTH_TIER_LABELS = {
  oauth: 'Sign in via ohmyposh.dev',
  cli: 'Run oh-my-posh auth',
  apikey: 'API key in your config',
  url: 'Token in your URL',
};

// Group directory name -> display label. Groups without an entry here fall back to the
// capitalised directory name (see index.js); only the two that capitalisation gets wrong
// need an override.
const GROUP_LABELS = {
  scm: 'Source control',
  cli: 'CLI',
};

// Doc id -> replacement description, used only where the doc's own "## What" paragraph does not
// survive being lifted out of its page. Three reasons appear below, and nothing else qualifies:
// the paragraph describes the upstream product rather than the segment, it opens with marketing
// copy, or it runs long enough to break the card grid. Everything not listed here uses the doc
// verbatim, which is what keeps this table small and the docs the source of truth.
const DESCRIPTION_OVERRIDES = {
  // "Display text." and "Display SysInfo." are circular — they restate the segment name.
  text: 'Display any static text or template output.',
  sysinfo: 'Display memory use and system load for the current machine.',
  // Missing its full stop; every other description ends in one.
  helm: 'Display the active Helm version.',

  // These describe the upstream service, not what the segment puts in your prompt.
  strava: 'Display your latest Strava activity, with an optional colour cue when it has been a while.',
  nightscout: 'Display live blood sugar readings from a Nightscout server.',
  withings: 'Display your last measured weight, sleep or step count from Withings.',
  brewfather: 'Display the status of your active Brewfather batch.',
  nba: 'Display the schedule and score for your favourite NBA team.',
  orthodoxcal: 'Display the Orthodox fasting level and feast for the current day.',
  ramadan: 'Display Sehar and Iftar times during Ramadan, with a countdown to the next.',

  // Long enough to blow out a card. Trimmed to the first clause, which is the whole point.
  claude: 'Display Claude Code session state: model, context window usage, tokens and cost.',
  'copilot-cli': 'Display GitHub Copilot CLI session state: model, context window usage and tokens.',
  copilot: 'Display your GitHub Copilot quota: premium interactions, completions and chat usage.',
  vimode: 'Display the current Vi mode in shells using Vi key bindings.',
  aspire: 'Display the Aspire AppHost in the current repo and whether it is running.',
  taskwarrior: 'Display Taskwarrior output for any set of configured commands.',
  language: 'Display the version of any executable, without writing a dedicated segment.',

  // All four carry the same "make sure your X executable is up-to-date" tail, which is
  // troubleshooting advice for the doc page, not a description of the segment.
  git: 'Display Git information when in a repository, including subfolders.',
  svn: 'Display Subversion information when in a repository, including subfolders.',
  plastic: 'Display Plastic SCM information when in a repository, including subfolders.',
  mercurial: 'Display Mercurial information when in a repository.',

  // Trails off into "based on the following logic", which refers to a list that does not travel.
  umbraco: 'Display the Umbraco version in use in the current working directory.',
};

module.exports = {
  AUTH_TIERS,
  AUTH_TIER_LABELS,
  GROUP_LABELS,
  DESCRIPTION_OVERRIDES,
};
