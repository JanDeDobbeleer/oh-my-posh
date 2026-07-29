// The studio's starter config: small enough to read in one glance, but shaped like a prompt
// somebody would actually write - a path and a session alongside a couple of language segments -
// so the preview shows something recognisable on first paint without fetching a theme.
//
// Two kinds of segment render here, and it is worth knowing which is which when editing this.
// git, node and python come from website/segment_data.json's "segments" map, so they show its
// recorded sample values. path and session own no entry there at all; they render from the
// "env" section's pinned PWD and UserName, because the wasm module renders with
// runtime.Flags.DataOnly set and that only stops a segment reaching the *machine* - a segment
// that needs nothing but the recorded data still renders normally. A segment with neither an
// entry nor a machine-free derivation (a cloud context, say) reports itself disabled and simply
// does not appear.
export const STARTER_CONFIG = `# yaml-language-server: $schema=https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json
version: 4
final_space: true
blocks:
  - type: prompt
    alignment: left
    segments:
      - type: session
        style: plain
        foreground: p:green
        template: "{{ .UserName }} "

      - type: path
        style: plain
        foreground: p:blue
        options:
          style: folder
        template: "{{ .Path }} "

      - type: git
        style: plain
        foreground: p:yellow
        template: "{{ .HEAD }}{{ if .BranchStatus }} {{ .BranchStatus }}{{ end }} "

      - type: node
        style: plain
        foreground: p:green
        template: " node {{ .Full }} "

      - type: python
        style: plain
        foreground: p:red
        template: " py {{ .Full }} "
palette:
  blue: "#61AFEF"
  green: "#98C379"
  yellow: "#E5C07B"
  red: "#E06C75"
`;

// The same prompt in the other two formats oh-my-posh accepts. A format switch that only
// changed the parser would drop whoever used it straight into a parse error, since the text
// on screen is still the previous format's - so each format brings its own starter, and the
// switch swaps text only while it is still untouched (see index.js's handleFormatChange).
export const STARTER_CONFIG_JSON = `{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "version": 4,
  "final_space": true,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "type": "session",
          "style": "plain",
          "foreground": "p:green",
          "template": "{{ .UserName }} "
        },
        {
          "type": "path",
          "style": "plain",
          "foreground": "p:blue",
          "options": { "style": "folder" },
          "template": "{{ .Path }} "
        },
        {
          "type": "git",
          "style": "plain",
          "foreground": "p:yellow",
          "template": "{{ .HEAD }}{{ if .BranchStatus }} {{ .BranchStatus }}{{ end }} "
        },
        {
          "type": "node",
          "style": "plain",
          "foreground": "p:green",
          "template": " node {{ .Full }} "
        },
        {
          "type": "python",
          "style": "plain",
          "foreground": "p:red",
          "template": " py {{ .Full }} "
        }
      ]
    }
  ],
  "palette": {
    "blue": "#61AFEF",
    "green": "#98C379",
    "yellow": "#E5C07B",
    "red": "#E06C75"
  }
}
`;

export const STARTER_CONFIG_TOML = `#:schema https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json
version = 4
final_space = true

[[blocks]]
type = 'prompt'
alignment = 'left'

  [[blocks.segments]]
  type = 'session'
  style = 'plain'
  foreground = 'p:green'
  template = '{{ .UserName }} '

  [[blocks.segments]]
  type = 'path'
  style = 'plain'
  foreground = 'p:blue'
  template = '{{ .Path }} '
  [blocks.segments.options]
  style = 'folder'

  [[blocks.segments]]
  type = 'git'
  style = 'plain'
  foreground = 'p:yellow'
  template = '{{ .HEAD }}{{ if .BranchStatus }} {{ .BranchStatus }}{{ end }} '

  [[blocks.segments]]
  type = 'node'
  style = 'plain'
  foreground = 'p:green'
  template = ' node {{ .Full }} '

  [[blocks.segments]]
  type = 'python'
  style = 'plain'
  foreground = 'p:red'
  template = ' py {{ .Full }} '

[palette]
blue = '#61AFEF'
green = '#98C379'
yellow = '#E5C07B'
red = '#E06C75'
`;

// Keyed by the format string config.ParseBytes takes as its first argument, which for all
// three happens to also be Prism's language id, so no format-to-language mapping is needed.
export const STARTERS = {
  yaml: STARTER_CONFIG,
  json: STARTER_CONFIG_JSON,
  toml: STARTER_CONFIG_TOML,
};

// json first, matching the order and the default a segment doc's editor uses (SEGMENT_FORMATS in
// Config.js), so the two editors a reader moves between don't reorder their tabs underneath them.
export const CONFIG_FORMATS = ['json', 'yaml', 'toml'];

export const CONFIG_FORMAT = 'json';
