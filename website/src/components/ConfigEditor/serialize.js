// Parses/stringifies a plain JS value against oh-my-posh's three accepted config formats. Shared
// by the studio (to fold an appended segment into whatever the reader already has open - see
// Studio/index.js) and each segment doc's own editor (to wrap its single edited segment into a
// renderable document - see Config.js). The format strings are the same ones config.ParseBytes
// takes and, not coincidentally, Prism's language ids too.
import YAML from 'yaml';
import TOML from 'smol-toml';

export function parseConfig(format, text) {
  switch (format) {
    case 'yaml':
      return YAML.parse(text);
    case 'toml':
      return TOML.parse(text);
    case 'json':
    default:
      return JSON.parse(text);
  }
}

// The three serializers disagree about trailing newlines - YAML.stringify always ends with one,
// JSON.stringify and TOML.stringify never do - which showed up as the editor being a blank line
// taller in YAML than in JSON for the same document. Normalizing to exactly one gives every
// format the same single empty row under the last character, which is also the height the editor
// now sizes itself to (see .editor in styles.module.css).
export function stringifyConfig(format, value) {
  switch (format) {
    case 'yaml':
      return withTrailingNewline(YAML.stringify(value));
    case 'toml':
      return withTrailingNewline(TOML.stringify(value));
    case 'json':
    default:
      return withTrailingNewline(JSON.stringify(value, null, 2));
  }
}

export function withTrailingNewline(text) {
  return `${text.replace(/\s+$/, '')}\n`;
}

// Rewrites the same document from one format into another, so switching the format tab converts
// what is in the editor instead of only relabelling it.
//
// Both editors used to swap in the next format's pre-written starter, but only while the text was
// still untouched - the moment a reader changed anything, switching left their JSON in the editor
// and told the renderer to read it as TOML, which failed with a parse error and no way back
// except retyping. Round-tripping through the parsers keeps the edits and changes the syntax.
//
// Returns null when the text cannot be converted: it does not parse (a config caught mid-edit),
// or it holds something the target format cannot express - TOML has no null, for instance. The
// caller is expected to leave both the text and the selected format alone in that case, since
// switching to a format the document cannot be written in is what caused the original bug.
export function convertConfig(from, to, text) {
  if (from === to) {
    return text;
  }

  try {
    return stringifyConfig(to, parseConfig(from, text));
  } catch {
    return null;
  }
}
