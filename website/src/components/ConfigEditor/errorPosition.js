// Re-parses a config that failed to parse, purely to recover a real message and, where possible,
// a position to draw a squiggly underline under (see ConfigEditor/index.js's errorLocation prop).
// This matters most for the studio (Studio/index.js), whose only other source of an error is the
// wasm module's generic "CONFIG PARSE ERROR" sentinel - config.ParseBytes discards the real
// underlying parser error, see src/config/load.go - so this is the one place a specific message
// is available at all. The shared source of truth for whether a config is actually valid stays
// wherever it already was (parseConfig for Config.js's segment editor, the wasm renderer for the
// studio); this can degrade to a location-less result, or be ignored entirely, without ever
// contradicting them.
import YAML from 'yaml';
import TOML from 'smol-toml';

// offset is 0-based; the returned line/column are both 1-based, matching every other parser's
// own position fields below.
function offsetToLineColumn(text, offset) {
  let line = 1;
  let column = 1;

  for (let i = 0; i < offset && i < text.length; i += 1) {
    if (text[i] === '\n') {
      line += 1;
      column = 1;
    } else {
      column += 1;
    }
  }

  return { line, column };
}

// Widens a single reported column into a span worth underlining. Parsers that only report where
// a token should have been (a missing comma, say) point at whitespace or past the end of the
// line - there is nothing there to underline, so this walks back to the last real character
// instead. Otherwise it walks forward to the end of the offending token.
function widenSpan(lineText, column) {
  const atGap = column > lineText.length || /\s/.test(lineText[column - 1] || '');

  if (atGap) {
    let start = column;
    while (start > 1 && /\s/.test(lineText[start - 2] || '')) {
      start -= 1;
    }
    start = Math.max(1, start - 1);
    return { column: start, endColumn: start + 1 };
  }

  let end = column;
  while (end <= lineText.length && !/\s/.test(lineText[end - 1] || '')) {
    end += 1;
  }

  return { column, endColumn: Math.max(end, column + 1) };
}

// Returns { message, line, column, endColumn } for the first syntax error in `text`, or null if
// `text` parses cleanly. `message` is always a usable string whenever an object is returned;
// line/column/endColumn (1-based, endColumn exclusive) are null instead when the parser's error
// doesn't expose a usable position, so a caller can still show the message with no underline.
// Only ever describes a single line - a squiggle spanning multiple lines (an unterminated YAML
// string, say) would need a much more elaborate overlay for very little benefit over pointing at
// where the problem starts.
export function getSyntaxError(format, text) {
  if (!text) {
    return null;
  }

  if (format === 'yaml') {
    try {
      YAML.parse(text);
      return null;
    } catch (err) {
      const [start] = err.linePos || [];
      if (!start) {
        return { message: err.message, line: null, column: null, endColumn: null };
      }

      const lineText = text.split('\n')[start.line - 1] || '';
      const { column, endColumn } = widenSpan(lineText, start.col);
      return { message: err.message, line: start.line, column, endColumn };
    }
  }

  if (format === 'toml') {
    try {
      TOML.parse(text);
      return null;
    } catch (err) {
      if (typeof err.line !== 'number' || typeof err.column !== 'number') {
        return { message: err.message, line: null, column: null, endColumn: null };
      }

      const lineText = text.split('\n')[err.line - 1] || '';
      const { column, endColumn } = widenSpan(lineText, err.column);
      return { message: err.message, line: err.line, column, endColumn };
    }
  }

  // json (default) - JSON.parse carries no structured position, but V8 (every browser this site
  // targets) embeds one in the message text itself, e.g. "... in JSON at position 5 (line 2
  // column 4)". Falls back to the byte offset alone (still enough to derive line/column) on
  // engines that only report that much, and to a location-less result - message only, no
  // underline - on one that reports neither.
  try {
    JSON.parse(text);
    return null;
  } catch (err) {
    const match = /position (\d+)(?: \(line (\d+) column (\d+)\))?/.exec(err.message || '');
    if (!match) {
      return { message: err.message, line: null, column: null, endColumn: null };
    }

    const { line, column } = match[2]
      ? { line: Number(match[2]), column: Number(match[3]) }
      : offsetToLineColumn(text, Number(match[1]));

    const lineText = text.split('\n')[line - 1] || '';
    const { column: startColumn, endColumn } = widenSpan(lineText, column);
    return { message: err.message, line, column: startColumn, endColumn };
  }
}
