// CodeMirror 6 completion + hover, built directly on the lezer syntax tree instead of the old
// completion.js's char-by-char text scanners (see schemaResolution.js's own header comment) or
// the third-party json-schema-library resolver a prior version of this editor delegated to,
// which was broken on this schema (see schemaResolution.js's own git-history reference).
// @codemirror/lang-json and @codemirror/lang-yaml both keep producing a real (if partial/error-
// flagged) tree through mid-edit states, which is exactly the case this whole rewrite exists for:
// a blank line inside an object, a dangling unterminated key, a value slot that's just a bare
// `key:`. Walking the tree handles all of those as a side effect of how lezer's incremental
// parser already recovers from them, rather than needing bespoke recovery logic of our own for
// each shape.
import { syntaxTree } from '@codemirror/language';
import { startCompletion } from '@codemirror/autocomplete';
import { EditorView, hoverTooltip } from '@codemirror/view';
import {
  resolveSchema,
  deriveValueSchema,
  mergeTypeBranch,
  getItemSchema,
  inferSchemaType,
  buildPropertyHint,
  getScopedRootSchema,
} from './schemaResolution';

function children(node) {
  const out = [];
  for (let child = node.firstChild; child; child = child.nextSibling) {
    out.push(child);
  }
  return out;
}

function textOf(state, node) {
  return state.doc.sliceString(node.from, node.to);
}

// Column of `offset` on its own line - used only to compare indentation between a blank line and
// a dangling `key:` above it (see the BlockMapping branch in walkYaml), never to scan content.
function lineIndent(doc, offset) {
  return offset - doc.lineAt(offset).from;
}

// @codemirror/lang-json's PropertyName/String nodes always include their surrounding quotes in
// their own text (an unterminated one becomes an error node instead - see below) - strip them
// for schema/property-name lookups without caring whether the closing quote is even there yet.
function stripJsonQuotes(text) {
  if (text.startsWith('"')) {
    return text.endsWith('"') && text.length > 1 ? text.slice(1, -1) : text.slice(1);
  }
  return text;
}

function yamlLiteralText(state, node) {
  const text = textOf(state, node);
  if (node.name === 'QuotedLiteral') {
    if (text.length > 1 && (text.startsWith('"') || text.startsWith("'")) && text.endsWith(text[0])) {
      return text.slice(1, -1);
    }
    return text.replace(/^['"]/, '');
  }
  return text;
}

// Collects the property/pair names directly inside a container node - the same "sibling keys
// already present" exclusion the old engine tracked by hand (see completion.js's
// collectForwardKeys), but trivial here: the tree already has every sibling as a direct child
// regardless of which side of the cursor it's on, unlike a forward-only text scan.
function siblingKeys(state, containerNode, isJson) {
  const entryName = isJson ? 'Property' : 'Pair';
  const keys = [];
  children(containerNode).forEach((kid) => {
    if (kid.name !== entryName) {
      return;
    }
    const kids = children(kid);
    if (isJson) {
      const nameNode = kids.find((k) => k.name === 'PropertyName');
      if (nameNode) {
        keys.push(stripJsonQuotes(textOf(state, nameNode)));
      }
    } else {
      const keyNode = kids.find((k) => k.name === 'Key');
      const literal = keyNode && children(keyNode)[0];
      if (keyNode) {
        keys.push(literal ? yamlLiteralText(state, literal) : textOf(state, keyNode));
      }
    }
  });
  return keys;
}

// mergeTypeBranch only does something for a schema that actually declares per-`type`
// conditionals (definitions.segment, today) - reading every object/mapping's OWN "type"
// property here and applying it unconditionally is what makes a segment's `options:` mapping
// offer its type-specific properties (e.g. path's `folder_icon`) without this walker needing to
// know which schemas are "the kind with branches" and which aren't.
function applyOwnTypeBranch(state, containerNode, schemaNode, isJson) {
  const entryName = isJson ? 'Property' : 'Pair';
  for (const kid of children(containerNode)) {
    if (kid.name !== entryName) {
      continue;
    }
    const kids = children(kid);

    if (isJson) {
      const nameNode = kids.find((k) => k.name === 'PropertyName');
      if (!nameNode || stripJsonQuotes(textOf(state, nameNode)) !== 'type') {
        continue;
      }
      const valueNode = kids.find((k) => k !== nameNode && k.name !== ':');
      if (!valueNode) {
        continue;
      }
      const value = valueNode.name === 'String' ? stripJsonQuotes(textOf(state, valueNode)) : textOf(state, valueNode);
      return mergeTypeBranch(schemaNode, value);
    }

    const keyNode = kids.find((k) => k.name === 'Key');
    if (!keyNode) {
      continue;
    }
    const keyLiteral = children(keyNode)[0];
    const keyName = keyLiteral ? yamlLiteralText(state, keyLiteral) : textOf(state, keyNode);
    if (keyName !== 'type') {
      continue;
    }
    const valueNode = kids.find((k) => k.name === 'Literal' || k.name === 'QuotedLiteral');
    if (!valueNode) {
      continue;
    }
    return mergeTypeBranch(schemaNode, yamlLiteralText(state, valueNode));
  }

  return schemaNode;
}

// A blank/dangling line at the very end of the document sits past the range of every real node
// (lezer never extends a block container's range to cover trailing whitespace it hasn't seen
// content resume after) - `resolveInner` lands on Stream/Document instead of the mapping the
// reader is actually inside. Falls back to the deepest still-open mapping/sequence by following
// the tree's own rightmost spine, so typing at the true end of the file still resolves against
// *some* real container instead of nothing. (A blank line that also dedents back out to a
// shallower level at the very end of the file is the one shape this doesn't recover - the tree
// simply carries no signal for it beyond raw indentation, which is exactly the kind of char-by-
// char scanning this rewrite exists to avoid re-introducing.)
function findDeepestOpenYamlContainer(tree) {
  let candidate = null;
  let cur = tree.topNode.lastChild;
  while (cur) {
    if (cur.name === 'BlockMapping' || cur.name === 'BlockSequence') {
      candidate = cur;
      cur = cur.lastChild;
    } else if (cur.name === 'Document' || cur.name === 'Pair' || cur.name === 'Item') {
      cur = cur.lastChild;
    } else {
      break;
    }
  }
  return candidate;
}

// Walks from the document root down to `pos`, resolving the schema at each level (object
// property steps via deriveValueSchema, array element steps via getItemSchema, each object's
// own `type` sibling via applyOwnTypeBranch) and classifying what's at `pos` itself: a property
// KEY being typed, or a property VALUE. Returns null when `pos` doesn't land anywhere completion
// can make sense of (e.g. inside a comment, or - see findDeepestOpenYamlContainer - a dedent back
// out to a shallower level at the very end of the file).
function walkYaml(state, pos, rootSchema) {
  const tree = syntaxTree(state);
  let node = tree.resolveInner(pos, -1);
  if (node.name === 'Stream' || node.name === 'Document') {
    node = findDeepestOpenYamlContainer(tree) || node;
  }
  if (node.name === 'Stream' || node.name === 'Document') {
    return null;
  }

  const path = [];
  for (let n = node; n; n = n.parent) {
    path.unshift(n);
  }

  let containerSchema = rootSchema;
  let containerNode = null;

  path.forEach((cur) => {
    if (cur.name === 'BlockMapping') {
      containerSchema = applyOwnTypeBranch(state, cur, containerSchema, false);
      containerNode = cur;
    } else if (cur.name === 'Pair') {
      const keyNode = children(cur).find((k) => k.name === 'Key');
      // Strictly BEFORE pos: with the cursor sitting exactly at the end of a key still being
      // typed (lezer wraps even a dangling colon-less line into a Pair/Key when it recovers),
      // `<=` would descend into that half-typed name as if it were a real property and clobber
      // the container schema to {} - the key branches below then have nothing to suggest from.
      // A genuine value position always has at least the `:` between key end and cursor.
      if (keyNode && keyNode.to < pos) {
        const literal = children(keyNode)[0];
        const propertyName = literal ? yamlLiteralText(state, literal) : textOf(state, keyNode);
        containerSchema = deriveValueSchema(containerSchema, propertyName);
      }
    } else if (cur.name === 'BlockSequence') {
      containerSchema = getItemSchema(containerSchema);
    }
  });

  const final = path[path.length - 1];
  const parent = final.parent;

  // An existing, fully-typed key - re-triggering completion (or hovering) over it. containerSchema
  // is still the ENCLOSING mapping's own schema (the Pair step above only advances past a key
  // whose range ends at/before `pos`, which isn't true while `pos` is still inside it). A Key
  // node's own range is IDENTICAL to its Literal/QuotedLiteral child's (there's no extra
  // punctuation between them, unlike a Pair which also owns the trailing `:`), so `resolveInner`
  // always descends past Key into that child when `pos` sits inside it - `final` here is
  // therefore the literal, not the Key wrapper, whenever `pos` isn't sitting exactly on Key's own
  // (identical) boundary.
  const keyLiteralChild = (final.name === 'Literal' || final.name === 'QuotedLiteral') && parent?.name === 'Key';
  if (final.name === 'Key' || keyLiteralChild) {
    const keyNode = keyLiteralChild ? parent : final;
    const literal = keyLiteralChild ? final : children(final)[0];
    const raw = literal ? textOf(state, literal) : textOf(state, keyNode);
    const quoted = literal?.name === 'QuotedLiteral';
    const tokenFrom = quoted ? literal.from + 1 : keyNode.from;
    return {
      kind: 'key',
      schema: containerSchema,
      partialText: state.doc.sliceString(tokenFrom, pos),
      from: tokenFrom,
      quoteOpen: quoted,
      usedKeys: containerNode ? siblingKeys(state, containerNode, false) : [],
      keyToken: { from: tokenFrom, to: quoted ? literal.to - 1 : keyNode.to, name: raw.replace(/^['"]|['"]$/g, '') },
    };
  }

  // A dangling bare key with no `:` yet (`        f`) - lezer-yaml can't attach it to a Pair, so
  // it surfaces as an error node directly under the enclosing mapping, a sibling of the real Pairs.
  if (final.type.isError && parent?.name === 'BlockMapping') {
    return {
      kind: 'key',
      schema: containerSchema,
      partialText: state.doc.sliceString(final.from, pos),
      from: final.from,
      usedKeys: siblingKeys(state, parent, false),
    };
  }

  // A blank indented line inside a mapping - resolveInner lands on the mapping itself, since the
  // gap between two Pair children (or before the first/after the last) belongs to no child node.
  // BUT a `key:` with nothing typed after it yet collapses into this exact same shape, whether
  // `pos` sits right after the colon on the SAME line (`style: `) or on a deeper-indented blank
  // line below it that was clearly meant to become that key's own nested mapping (`options:`
  // followed by a blank indented line, before anything under it exists yet) - the Pair's own
  // range ends at the `:` either way, leaving everything after it as unclaimed mapping content.
  // Both are told apart from a genuinely blank line (nothing dangling immediately above `pos`)
  // by hand, and from each other by comparing indentation and the dangling key's own value type.
  if (final.name === 'BlockMapping') {
    const danglingPair = children(final).slice().reverse().find((k) => k.name === 'Pair' && k.to <= pos);
    const dkids = danglingPair ? children(danglingPair) : [];
    const keyNode = dkids.find((k) => k.name === 'Key');
    const hasValue = dkids.some((k) => k.name !== 'Key' && k.name !== ':');

    if (danglingPair && keyNode && !hasValue) {
      const literal = children(keyNode)[0];
      const propertyName = literal ? yamlLiteralText(state, literal) : null;
      const sameLine = !state.doc.sliceString(danglingPair.to, pos).includes('\n');
      const nested = lineIndent(state.doc, pos) > lineIndent(state.doc, keyNode.from);

      if (propertyName && (sameLine || nested)) {
        const valueSchema = deriveValueSchema(containerSchema, propertyName);
        const isObjectValue = inferSchemaType(resolveSchema(valueSchema)) === 'object';

        if (nested && isObjectValue) {
          return { kind: 'key', schema: valueSchema, partialText: '', from: pos, usedKeys: [] };
        }
        return { kind: 'value', schema: valueSchema, propertyName, partialText: '', from: pos };
      }
    }

    return {
      kind: 'key',
      schema: containerSchema,
      partialText: '',
      from: pos,
      usedKeys: siblingKeys(state, final, false),
    };
  }

  // `key:` with nothing typed after it yet (not even a space) - the Pair has no value child at
  // all, so `pos` resolves to the Pair itself. containerSchema already advanced past this pair's
  // own key (the Pair step above ran, since the key necessarily ends before `pos` here).
  if (final.name === 'Pair') {
    const keyNode = children(final).find((k) => k.name === 'Key');
    const literal = keyNode && children(keyNode)[0];
    return {
      kind: 'value',
      schema: containerSchema,
      propertyName: literal ? yamlLiteralText(state, literal) : null,
      partialText: '',
      from: pos,
    };
  }

  // A scalar value already/being typed (`style: pl`, `foreground: p:blue`) - containerSchema was
  // already advanced to this property's own value schema by the Pair step above.
  if ((final.name === 'Literal' || final.name === 'QuotedLiteral') && parent?.name === 'Pair') {
    const quoted = final.name === 'QuotedLiteral';
    const tokenFrom = quoted ? final.from + 1 : final.from;
    const keyNode = children(parent).find((k) => k.name === 'Key');
    const literal = keyNode && children(keyNode)[0];
    // Same bare-literal rule as walkJson: filtering true/false by the literal already sitting
    // there would filter the other one out - YAML has no distinct node names for booleans, so
    // the token's own text decides.
    const bareLiteral = !quoted && /^(true|false|null|\d+(\.\d+)?)$/.test(textOf(state, final));
    return {
      kind: 'value',
      schema: containerSchema,
      propertyName: literal ? yamlLiteralText(state, literal) : null,
      partialText: bareLiteral ? '' : state.doc.sliceString(tokenFrom, pos),
      from: tokenFrom,
      quoteOpen: quoted,
    };
  }

  // Right after `- ` in a block sequence, before any content of the new item exists yet -
  // containerSchema is already this sequence's item schema (the BlockSequence step above).
  // Segment-shaped items (an object) offer property names; anything else offers a bare value.
  if (final.name === 'BlockSequence' || final.name === 'Item') {
    const isObjectItem = inferSchemaType(resolveSchema(containerSchema)) === 'object';
    return {
      kind: isObjectItem ? 'key' : 'value',
      schema: containerSchema,
      propertyName: null,
      partialText: '',
      from: pos,
      usedKeys: [],
    };
  }

  return null;
}

function walkJson(state, pos, rootSchema) {
  const tree = syntaxTree(state);
  const node = tree.resolveInner(pos, -1);

  const path = [];
  for (let n = node; n; n = n.parent) {
    path.unshift(n);
  }

  let containerSchema = rootSchema;
  let containerNode = null;

  path.forEach((cur) => {
    if (cur.name === 'Object') {
      containerSchema = applyOwnTypeBranch(state, cur, containerSchema, true);
      containerNode = cur;
    } else if (cur.name === 'Property') {
      const nameNode = children(cur).find((k) => k.name === 'PropertyName');
      // Strictly BEFORE pos - same off-by-one as walkYaml's Pair step: at pos == nameNode.to
      // (cursor right on the name's closing quote) this is still a KEY position, and stepping
      // into the possibly-partial name would resolve the container schema to {}.
      if (nameNode && nameNode.to < pos) {
        containerSchema = deriveValueSchema(containerSchema, stripJsonQuotes(textOf(state, nameNode)));
      }
    } else if (cur.name === 'Array') {
      containerSchema = getItemSchema(containerSchema);
    }
  });

  const final = path[path.length - 1];
  const parent = final.parent;

  // An existing, fully-closed property name being re-visited.
  if (final.name === 'PropertyName') {
    return {
      kind: 'key',
      schema: containerSchema,
      partialText: state.doc.sliceString(final.from + 1, pos),
      from: final.from + 1,
      quoteOpen: true,
      usedKeys: containerNode ? siblingKeys(state, containerNode, true) : [],
      keyToken: { from: final.from + 1, to: final.to - 1, name: stripJsonQuotes(textOf(state, final)) },
    };
  }

  // An error node under Object/Property: either an unterminated `"partial` key/value string, or
  // (a trailing comma then nothing before the closing brace) a genuinely empty, zero-width node.
  // Only the former has a leading quote to skip - the zero-width case has no text to skip at all.
  if (final.type.isError) {
    const text = textOf(state, final);
    const hasQuote = text.startsWith('"');
    const tokenFrom = hasQuote ? final.from + 1 : final.from;

    if (parent?.name === 'Object') {
      return {
        kind: 'key',
        schema: containerSchema,
        partialText: state.doc.sliceString(tokenFrom, pos),
        from: tokenFrom,
        quoteOpen: hasQuote,
        usedKeys: siblingKeys(state, parent, true),
      };
    }
    if (parent?.name === 'Property') {
      const nameNode = children(parent).find((k) => k.name === 'PropertyName');
      return {
        kind: 'value',
        schema: containerSchema,
        propertyName: nameNode ? stripJsonQuotes(textOf(state, nameNode)) : null,
        partialText: state.doc.sliceString(tokenFrom, pos),
        from: tokenFrom,
        quoteOpen: hasQuote,
      };
    }
  }

  // A blank/comma-adjacent position inside an object with no partial token at all yet.
  if (final.name === 'Object') {
    return {
      kind: 'key',
      schema: containerSchema,
      partialText: '',
      from: pos,
      quoteOpen: false,
      usedKeys: siblingKeys(state, final, true),
    };
  }

  // `"prop":` with no value yet - Property has only PropertyName + ':' as children.
  if (final.name === 'Property') {
    const nameNode = children(final).find((k) => k.name === 'PropertyName');
    return {
      kind: 'value',
      schema: containerSchema,
      propertyName: nameNode ? stripJsonQuotes(textOf(state, nameNode)) : null,
      partialText: '',
      from: pos,
      quoteOpen: false,
    };
  }

  // Inside an (empty) array - e.g. right after a seeded `[]`'s opening bracket. The path loop
  // above already stepped containerSchema down to the array's ITEM schema, so string-enum items
  // get their value popup here; anything else yields no options and the popup simply stays shut.
  if (final.name === 'Array') {
    return {
      kind: 'value',
      schema: containerSchema,
      propertyName: null,
      partialText: '',
      from: pos,
      quoteOpen: false,
    };
  }

  // An already-closed string/number/bool value being re-visited. A bare literal (true/false/
  // null/number - typically one this editor just seeded as a default) deliberately reports an
  // EMPTY partialText while still replacing the whole token: filtering true/false by the literal
  // that's already there would just filter the other one out, defeating the point of offering
  // both. Strings keep prefix filtering - there, narrowing as the reader types is the point.
  if (parent?.name === 'Property' && final.name !== 'PropertyName') {
    const nameNode = children(parent).find((k) => k.name === 'PropertyName');
    const quoted = final.name === 'String';
    const bareLiteral = ['True', 'False', 'Null', 'Number'].includes(final.name);
    const tokenFrom = quoted ? final.from + 1 : final.from;
    return {
      kind: 'value',
      schema: containerSchema,
      propertyName: nameNode ? stripJsonQuotes(textOf(state, nameNode)) : null,
      partialText: bareLiteral ? '' : state.doc.sliceString(tokenFrom, pos),
      from: bareLiteral ? final.from : tokenFrom,
      quoteOpen: quoted,
    };
  }

  return null;
}

function detectIsJson(state) {
  return syntaxTree(state).topNode.name === 'JsonText';
}

export const schemaHintTheme = EditorView.baseTheme({
  '.cm-omp-schema-hint': {
    padding: '0.5rem 0.65rem',
    maxWidth: '18rem',
    fontSize: '0.85rem',
    lineHeight: '1.4',
  },
  '.cm-omp-schema-hint-title': {
    fontWeight: '600',
    marginBottom: '0.25rem',
  },
  '.cm-omp-schema-hint-examples': {
    marginTop: '0.25rem',
    opacity: '0.75',
  },
});

function buildHintDom(title, resolvedProp) {
  const { description, examples } = buildPropertyHint(resolvedProp);
  const dom = document.createElement('div');
  dom.className = 'cm-omp-schema-hint';

  const titleEl = document.createElement('div');
  titleEl.className = 'cm-omp-schema-hint-title';
  titleEl.textContent = title;
  dom.append(titleEl);

  if (description) {
    const text = document.createElement('div');
    text.textContent = description;
    dom.append(text);
  }

  if (examples) {
    const ex = document.createElement('div');
    ex.className = 'cm-omp-schema-hint-examples';
    ex.textContent = `Examples: ${examples.join(', ')}`;
    dom.append(ex);
  }

  return dom;
}

// closeBrackets auto-pairs a typed quote, so when the walker says the cursor sits inside an open
// string (quoteOpen), the quote right after the replaced range is OUR token's auto-inserted
// closer, not the next token's opener - it must be consumed by the replacement or it survives as
// a stray `"` after the inserted text (`"maps": "`). Only ever done under quoteOpen: without it,
// a quote at `end` really is an unrelated neighbour's opening quote and must stay.
function consumeAutoClosedQuote(doc, start, end) {
  const quote = doc.sliceString(start - 1, start);
  if ((quote === '"' || quote === "'") && doc.sliceString(end, end + 1) === quote) {
    return end + 1;
  }
  return end;
}

// Whether a `,` must follow whatever gets inserted at `end` for the JSON to stay valid: yes
// whenever the next non-whitespace thing is another member/element, no when it's a closer, an
// existing comma, or the end of the document (a trailing comma there would itself be invalid).
function missingJsonComma(doc, end) {
  const rest = doc.sliceString(end, Math.min(doc.length, end + 500));
  const next = rest.match(/^\s*(\S)/);
  return next && !/[,}\]]/.test(next[1]) ? ',' : '';
}

// What to put in the value slot of a just-completed JSON property so the document is valid the
// moment the completion lands, even if the reader stops right there - the old editor's chain
// seeding, rebuilt. cursorStart/cursorEnd are offsets into `text`: equal for a plain caret (e.g.
// inside the seeded `""`), a real span for a seeded default literal so overtyping it (`true` over
// a defaulted `false`) is a single keystroke. `chain` opens the value popup right away - only
// when it would actually have something to show.
function buildJsonValueSeed(resolvedProp) {
  // Chained seeds (a value popup opens right away) must leave a plain CARET at the end of the
  // literal, never a range selection over it: @codemirror/autocomplete anchors both its query
  // position and its replace range at selection.main.FROM (see its `cur()` helper), so a
  // selection spanning the literal collapses the accept-range to zero width and the picked
  // value gets INSERTED next to the seeded one instead of replacing it. The walker's
  // whole-token rule for bare literals is what makes accepting over the caret replace the
  // seed. Only the unchained number seed below keeps the overtype selection - no popup is
  // involved there, so the quirk never triggers.
  if (Array.isArray(resolvedProp.enum)) {
    if (typeof resolvedProp.enum[0] === 'string') {
      return { text: '""', cursorStart: 1, cursorEnd: 1, chain: true };
    }
    const literal = JSON.stringify(resolvedProp.default !== undefined ? resolvedProp.default : resolvedProp.enum[0]);
    return { text: literal, cursorStart: literal.length, cursorEnd: literal.length, chain: true };
  }

  const type = inferSchemaType(resolvedProp);

  if (type === 'boolean') {
    const literal = JSON.stringify(resolvedProp.default !== undefined ? resolvedProp.default : false);
    return { text: literal, cursorStart: literal.length, cursorEnd: literal.length, chain: true };
  }
  if (type === 'integer' || type === 'number') {
    const literal = JSON.stringify(resolvedProp.default !== undefined ? resolvedProp.default : 0);
    return { text: literal, cursorStart: 0, cursorEnd: literal.length, chain: false };
  }
  if (type === 'object') {
    // Chaining straight into the nested object's own property popup only makes sense when the
    // schema actually names properties to offer.
    return { text: '{}', cursorStart: 1, cursorEnd: 1, chain: Object.keys(resolvedProp.properties || {}).length > 0 };
  }
  if (type === 'array') {
    // Item completion inside the seeded `[]` comes from walkJson's Array branch; chain only for
    // string-enum items, where the popup has values to offer immediately.
    return { text: '[]', cursorStart: 1, cursorEnd: 1, chain: Array.isArray(getItemSchema(resolvedProp).enum) };
  }
  // string, and the untyped/anyOf leftovers where a string is the least-wrong valid placeholder.
  return { text: '""', cursorStart: 1, cursorEnd: 1, chain: false };
}

function buildPropertyOption(name, propSchema, isJson, quoteOpen) {
  const resolvedProp = resolveSchema(propSchema);
  const title = resolvedProp.title || name;
  const chainType = inferSchemaType(resolvedProp);

  return {
    label: name,
    type: 'property',
    detail: resolvedProp.title || '',
    info: () => buildHintDom(title, resolvedProp),
    apply(view, _completion, from, to) {
      const doc = view.state.doc;
      let start = from;
      let end = to;
      if (quoteOpen) {
        end = consumeAutoClosedQuote(doc, start, end);
        if (!isJson) {
          // A YAML key the reader started quoting: schema property names never need quotes, so
          // swallow the opening quote too instead of leaving `"maps: `.
          start -= 1;
        }
      }
      // Re-completing over a key whose `:` already exists (or whose value slot the caller
      // seeded) must not double the separator - and must not seed a second value either.
      const hasColon = /^[ \t]*:/.test(doc.sliceString(end, end + 8));
      const base = isJson ? (quoteOpen ? `${name}"` : `"${name}"`) : name;

      let insertText;
      let anchor;
      let head;
      let shouldChain = false;

      if (isJson && !hasColon) {
        const seed = buildJsonValueSeed(resolvedProp);
        const comma = missingJsonComma(doc, end);
        const valueStart = start + base.length + 2;
        insertText = `${base}: ${seed.text}${comma}`;
        anchor = valueStart + seed.cursorStart;
        head = valueStart + seed.cursorEnd;
        shouldChain = seed.chain;
      } else {
        insertText = `${base}${hasColon ? '' : ': '}`;
        anchor = start + insertText.length;
        head = anchor;
        // YAML stays valid with an empty value slot (`key: ` is a null), so no seeding - just
        // open the value popup when it has something to show.
        shouldChain = !hasColon && (chainType === 'boolean' || Array.isArray(resolvedProp.enum));
      }

      view.dispatch({
        changes: { from: start, to: end, insert: insertText },
        selection: { anchor, head },
      });
      if (shouldChain) {
        startCompletion(view);
      }
    },
  };
}

function buildEnumOption(value, isJson, quoteOpen) {
  const label = String(value);
  const isString = typeof value === 'string';
  return {
    label,
    type: 'keyword',
    apply(view, _completion, from, to) {
      const doc = view.state.doc;
      let start = from;
      let end = to;
      let insertText;
      if (quoteOpen) {
        const quote = doc.sliceString(start - 1, start);
        end = consumeAutoClosedQuote(doc, start, end);
        if (isString) {
          // The opening quote the reader typed stays; re-close it ourselves since the
          // auto-paired closer (if any) was consumed above.
          insertText = `${label}${quote === "'" ? "'" : '"'}`;
        } else {
          // A bare literal (true/false/number) inside a hand-opened string would end up a
          // STRING (`"true"`) - drop the opening quote along with the closer.
          start -= 1;
          insertText = label;
        }
      } else {
        insertText = isJson && isString ? `"${label}"` : label;
      }
      // A completed value followed by another member needs its separator right away, or the
      // document is invalid the moment the popup closes. Cursor stays before the comma - the
      // value is what the reader was placing, the comma is just bookkeeping.
      const comma = isJson ? missingJsonComma(doc, end) : '';
      const anchor = start + insertText.length;
      view.dispatch({
        changes: { from: start, to: end, insert: insertText + comma },
        selection: { anchor },
      });
    },
  };
}

function buildValueOptions(resolvedSchema, isJson, quoteOpen) {
  if (Array.isArray(resolvedSchema.enum)) {
    return resolvedSchema.enum.map((value) => buildEnumOption(value, isJson, quoteOpen));
  }

  // A boolean has exactly two possible values, so it's just as pickable as a small enum - offer
  // both instead of leaving the reader to type "true"/"false" by hand.
  if (inferSchemaType(resolvedSchema) === 'boolean') {
    return [true, false].map((value) => buildEnumOption(value, isJson, quoteOpen));
  }

  return [];
}

// CompletionSource for the language's own `data.of({ autocomplete })` facet (see
// editorExtensions.js) - one instance per schemaScope, reused across keystrokes.
export function schemaCompletionSource(schemaScope) {
  const rootSchema = getScopedRootSchema(schemaScope);

  return (context) => {
    const { state, pos } = context;
    const isJson = detectIsJson(state);
    const info = isJson ? walkJson(state, pos, rootSchema) : walkYaml(state, pos, rootSchema);
    if (!info) {
      return null;
    }

    const resolved = resolveSchema(info.schema);
    let options;

    if (info.kind === 'key') {
      const used = new Set(info.usedKeys || []);
      options = Object.entries(resolved.properties || {})
        .filter(([name]) => !used.has(name) && name.startsWith(info.partialText))
        // Deprecated keys (e.g. the segment's v3 `properties` alias for `options`) stay out of
        // the popup - suggesting them would steer new configs onto them. Hover over one already
        // typed in a legacy config still works and shows the schema's own deprecation notice.
        // The `options` key itself only means something for a segment type whose schema branch
        // actually defines options (see mergeTypeBranch) - for a type without any, the resolved
        // schema is the base bare object, and suggesting the key would only lead the reader
        // into an empty mapping with an empty popup inside it.
        .filter(([name, propSchema]) => {
          const resolvedProp = resolveSchema(propSchema);
          if (resolvedProp.deprecated) {
            return false;
          }
          return name !== 'options' || Object.keys(resolvedProp.properties || {}).length > 0;
        })
        .map(([name, propSchema]) => buildPropertyOption(name, propSchema, isJson, !!info.quoteOpen));
    } else {
      options = buildValueOptions(resolved, isJson, !!info.quoteOpen)
        .filter((option) => option.label.startsWith(info.partialText));
    }

    if (!options.length) {
      return null;
    }

    // filter: false - the source already prefix-filtered against partialText, and CodeMirror's
    // own second filtering pass would re-match options against the raw doc text between `from`
    // and the caret, which breaks the bare-literal case: a seeded `false` under the caret would
    // filter `true` right back out of the popup the walker deliberately left it in for. No
    // validFor either - without CodeMirror's filter, narrowing as the reader types has to come
    // from re-running this source on every keystroke (a tree walk; cheap at typing speed).
    return { from: info.from, options, filter: false };
  };
}

// The hoverTooltip() source itself, kept separate from the extension it's wrapped into below so
// it can be exercised directly (it only ever reads `view.state`, so a plain `{ state }` stands in
// for a real EditorView in a test). Hovering a property KEY already in the doc. Reuses the same
// walker as completion: at the hover position the walker naturally resolves the ENCLOSING
// container's schema (it only advances past a key once `pos` is past its end, which isn't true
// while the mouse sits inside it), so `containerSchema.properties[keyToken.name]` is exactly this
// key's own definition.
export function schemaHoverSource(schemaScope) {
  const rootSchema = getScopedRootSchema(schemaScope);

  return (view, pos) => {
    const { state } = view;
    const isJson = detectIsJson(state);
    const info = isJson ? walkJson(state, pos, rootSchema) : walkYaml(state, pos, rootSchema);
    if (!info || info.kind !== 'key' || !info.keyToken) {
      return null;
    }

    const resolved = resolveSchema(info.schema);
    const propSchema = resolved.properties?.[info.keyToken.name];
    if (!propSchema) {
      return null;
    }

    const resolvedProp = resolveSchema(propSchema);
    const title = resolvedProp.title || info.keyToken.name;

    return {
      pos: info.keyToken.from,
      end: info.keyToken.to,
      above: true,
      create() {
        return { dom: buildHintDom(title, resolvedProp) };
      },
    };
  };
}

// hoverTooltip extension for the language Compartment (see editorExtensions.js).
export function schemaHoverTooltip(schemaScope) {
  return hoverTooltip(schemaHoverSource(schemaScope));
}
