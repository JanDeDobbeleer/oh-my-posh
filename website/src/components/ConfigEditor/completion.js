import schema from '../../../../themes/schema.json';

const ROOT_SCHEMA = schema;

// The main scanner in getCompletionContext only walks text up to the cursor, so its
// per-frame usedKeys only ever sees sibling keys that appear BEFORE the cursor. When
// completing inside an already-populated object (the common "loaded an existing config"
// case), sibling keys placed AFTER the cursor are just as real and must be excluded too.
// This does a small forward-only scan from the cursor to the end of the current object
// (its matching closing `}`), collecting any key strings it passes at the same depth.
function collectForwardKeys(text, cursorOffset, insideOpenString) {
  const keys = [];
  let i = cursorOffset;

  // If the cursor sits inside a string we (or the user) already opened, find where that
  // string would close - but only treat it as a genuine closer if what follows it looks
  // like a key delimiter (`:`). Otherwise the "next quote" is really the start of an
  // unrelated sibling key sitting right next to our not-yet-closed string (e.g. the user
  // just typed a bare opening `"` for a brand new key immediately before an existing
  // one), and skipping past it would swallow that sibling's own opening quote.
  if (insideOpenString) {
    let closeAt = i;
    while (closeAt < text.length && text[closeAt] !== '"') {
      if (text[closeAt] === '\\') {
        closeAt += 1;
      }
      closeAt += 1;
    }

    if (closeAt < text.length) {
      let k = closeAt + 1;
      while (k < text.length && /\s/.test(text[k])) {
        k += 1;
      }
      if (text[k] === ':') {
        i = closeAt + 1;
      }
    }
  }

  let depth = 0;
  let inString = false;

  while (i < text.length) {
    const char = text[i];

    if (inString) {
      if (char === '\\') {
        i += 2;
        continue;
      }
      if (char === '"') {
        inString = false;
      }
      i += 1;
      continue;
    }

    if (char === '"') {
      const keyStart = i + 1;
      let j = keyStart;
      while (j < text.length && text[j] !== '"') {
        if (text[j] === '\\') {
          j += 1;
        }
        j += 1;
      }

      if (depth === 0) {
        // Only a string immediately followed by `:` (ignoring whitespace) is a key -
        // a plain string value at this depth must not be collected as one.
        let k = j + 1;
        while (k < text.length && /\s/.test(text[k])) {
          k += 1;
        }
        if (text[k] === ':') {
          keys.push(text.slice(keyStart, j));
        }
      }

      i = j + 1;
      continue;
    }

    if (char === '{' || char === '[') {
      depth += 1;
      i += 1;
      continue;
    }

    if (char === '}' || char === ']') {
      if (depth === 0) {
        break;
      }
      depth -= 1;
      i += 1;
      continue;
    }

    i += 1;
  }

  return keys;
}

function resolveSchema(node, root = ROOT_SCHEMA) {
  if (!node || typeof node !== 'object') {
    return {};
  }

  if (node.$ref) {
    const ref = node.$ref;
    if (ref.startsWith('#/')) {
      const target = ref.split('/').slice(1).reduce((acc, part) => acc?.[part], root);
      const resolved = resolveSchema(target, root);
      // Draft 2020-12 allows keywords alongside $ref; sibling keys (e.g. a
      // description overriding the target's) must win over the target's own.
      const siblings = { ...node };
      delete siblings.$ref;
      return { ...resolved, ...siblings };
    }
  }

  // anyOf/oneOf branches are alternatives (e.g. "enum or free string") rather than
  // required composition, so completion only needs the union of their enum values.
  if (node.anyOf || node.oneOf) {
    const branches = node.anyOf || node.oneOf;
    const merged = { ...node };
    delete merged.anyOf;
    delete merged.oneOf;

    branches.forEach((branch) => {
      const resolvedBranch = resolveSchema(branch, root);
      if (resolvedBranch.enum) {
        merged.enum = [...(merged.enum || []), ...resolvedBranch.enum];
      }
      if (resolvedBranch.type && !merged.type) {
        merged.type = resolvedBranch.type;
      }
    });

    return merged;
  }

  if (node.allOf) {
    return node.allOf.reduce((acc, child) => mergeSchema(acc, resolveSchema(child, root)), {
      ...node,
    });
  }

  return { ...node };
}

function mergeSchema(base, incoming) {
  const merged = { ...base };
  if (incoming.properties) {
    merged.properties = {
      ...(base.properties || {}),
      ...(incoming.properties || {}),
    };
  }

  if (incoming.type && !merged.type) {
    merged.type = incoming.type;
  }

  if (incoming.enum && !merged.enum) {
    merged.enum = incoming.enum;
  }

  if (incoming.default !== undefined && merged.default === undefined) {
    merged.default = incoming.default;
  }

  if (incoming.description && !merged.description) {
    merged.description = incoming.description;
  }

  if (incoming.title && !merged.title) {
    merged.title = incoming.title;
  }

  return merged;
}

function getDefinitionSchema(ref, root = ROOT_SCHEMA) {
  if (!ref || typeof ref !== 'string' || !ref.startsWith('#/')) {
    return {};
  }

  const target = ref.split('/').slice(1).reduce((acc, part) => acc?.[part], root);
  return resolveSchema(target, root);
}

function deriveValueSchema(parentSchema, propertyName, root = ROOT_SCHEMA) {
  const resolvedParent = resolveSchema(parentSchema, root);
  const propertySchema = resolvedParent.properties?.[propertyName];

  if (!propertySchema) {
    return {};
  }

  return resolveSchema(propertySchema, root);
}

// Only `definitions.segment` uses conditional `if.properties.type.const` / `then`
// branches today, but this stays generic so any similarly-shaped schema benefits.
function mergeTypeBranch(schema, typeValue) {
  const resolved = resolveSchema(schema);
  if (!resolved.allOf || !typeValue) {
    return resolved;
  }

  const merged = { ...resolved };
  resolved.allOf.forEach((branch) => {
    if (branch.if?.properties?.type?.const === typeValue) {
      const thenSchema = resolveSchema(branch.then, ROOT_SCHEMA);
      const nextProperties = { ...(merged.properties || {}) };

      Object.entries(thenSchema.properties || {}).forEach(([key, branchProperty]) => {
        const baseProperty = nextProperties[key];
        // A segment's own override (e.g. "options") replaces the base property's shape,
        // but shouldn't lose the base's description/title if the override doesn't set one.
        nextProperties[key] = baseProperty
          ? { description: baseProperty.description, title: baseProperty.title, ...branchProperty }
          : branchProperty;
      });

      merged.properties = nextProperties;
    }
  });

  return merged;
}

function getItemSchema(parentValueSchema) {
  const resolved = resolveSchema(parentValueSchema);
  if (!resolved.items) {
    return {};
  }

  return resolveSchema(resolved.items);
}

// A resolved schema's own "type" keyword is the most direct source, but per-segment-type
// conditionals (schema.json's "if type === X then properties: { options: {...} }" branches)
// routinely narrow a property down to just "properties"/"items"/"unevaluatedProperties"
// without repeating "type": "object" (JSON Schema doesn't require it for validation to work).
// Completion still needs a concrete type to know what to chain into, so fall back to
// inferring it from shape - "properties" implies object, "items" implies array - before
// falling back further to a bare enum implying string.
function inferSchemaType(resolvedSchema) {
  if (resolvedSchema.type) {
    return resolvedSchema.type;
  }

  if (resolvedSchema.properties) {
    return 'object';
  }

  if (resolvedSchema.items) {
    return 'array';
  }

  return resolvedSchema.enum ? 'string' : null;
}

function createObjectFrame(schema) {
  return {
    kind: 'object',
    schema: resolveSchema(schema),
    lastPropertyName: null,
    inValue: false,
    expectingKey: true,
    pendingValueSchema: null,
    usedKeys: [],
  };
}

function createArrayFrame(itemSchema) {
  return {
    kind: 'array',
    schema: itemSchema,
    lastPropertyName: null,
    inValue: false,
    expectingKey: false,
    pendingValueSchema: null,
    // Array items are never keys, so this never gains entries, but it must exist -
    // getCompletionContext spreads every frame's usedKeys regardless of frame kind.
    usedKeys: [],
  };
}

// Walks the raw text char-by-char (not a full parse — configs mid-edit are rarely
// valid JSON) tracking a stack of object/array frames so we know which schema
// applies at the cursor. Handles the "still typing, string not closed yet" case
// explicitly since that's the state completion is actually triggered in.
// Which schema governs the outermost `{` of the text being edited. The studio edits a full
// theme config (ROOT_SCHEMA fits directly); a segment doc page's sample editor only ever
// contains a single bare segment object (see website/src/components/Config.js), so completion
// there needs to start from #/definitions/segment instead - otherwise every top-level
// property/hover lookup resolves against the wrong schema entirely (root config fields like
// "final_space" instead of segment fields like "foreground").
function getScopedRootSchema(schemaScope) {
  if (schemaScope === 'segment') {
    return getDefinitionSchema('#/definitions/segment', ROOT_SCHEMA);
  }
  return ROOT_SCHEMA;
}

function getCompletionContext(text, cursorOffset, schemaScope = 'config') {
  const scopedRootSchema = getScopedRootSchema(schemaScope);
  const beforeCursor = text.slice(0, cursorOffset);
  const stack = [];
  let currentFrame = null;

  let inString = false;
  let escapeNext = false;
  let currentString = '';
  let stringStartedAsKey = false;

  for (let index = 0; index < beforeCursor.length; index += 1) {
    const char = beforeCursor[index];

    if (inString) {
      if (escapeNext) {
        escapeNext = false;
        currentString += char;
        continue;
      }

      if (char === '\\') {
        escapeNext = true;
        currentString += char;
        continue;
      }

      if (char === '"') {
        inString = false;

        if (stringStartedAsKey) {
          currentFrame.lastPropertyName = currentString;
          currentFrame.expectingKey = false;
          currentFrame.inValue = false;
          if (!currentFrame.usedKeys.includes(currentString)) {
            currentFrame.usedKeys.push(currentString);
          }
        } else if (currentFrame?.kind === 'object' && currentFrame.lastPropertyName === 'type') {
          currentFrame.schema = mergeTypeBranch(currentFrame.schema, currentString);
        }

        currentString = '';
      } else {
        currentString += char;
      }
      continue;
    }

    if (char === '"') {
      inString = true;
      currentString = '';
      stringStartedAsKey = !!currentFrame && currentFrame.kind === 'object' && currentFrame.expectingKey;
      continue;
    }

    if (char === '{') {
      // The very first `{` opens the config's own root object — there's no
      // enclosing property to derive a value schema from, so it uses ROOT_SCHEMA
      // directly instead of a (nonexistent) parent's pending value schema.
      let frameSchema;
      if (!currentFrame) {
        frameSchema = scopedRootSchema;
      } else if (currentFrame.kind === 'object') {
        frameSchema = currentFrame.pendingValueSchema || {};
      } else {
        frameSchema = currentFrame.schema;
      }

      const frame = createObjectFrame(frameSchema);
      stack.push(frame);
      currentFrame = frame;
      continue;
    }

    if (char === '[') {
      const parentValueSchema = !currentFrame
        ? {}
        : currentFrame.kind === 'object'
          ? currentFrame.pendingValueSchema
          : currentFrame.schema;
      const frame = createArrayFrame(getItemSchema(parentValueSchema));
      stack.push(frame);
      currentFrame = frame;
      continue;
    }

    if (char === '}' || char === ']') {
      if (stack.length > 0) {
        stack.pop();
        currentFrame = stack[stack.length - 1] || null;
      }
      continue;
    }

    if (char === ':') {
      if (currentFrame?.kind === 'object' && currentFrame.lastPropertyName) {
        currentFrame.pendingValueSchema = deriveValueSchema(currentFrame.schema, currentFrame.lastPropertyName);
        currentFrame.inValue = true;
        currentFrame.expectingKey = false;
      }
      continue;
    }

    if (char === ',') {
      if (currentFrame?.kind === 'object') {
        currentFrame.expectingKey = true;
        currentFrame.inValue = false;
        currentFrame.lastPropertyName = null;
        currentFrame.pendingValueSchema = null;
      }
      continue;
    }
  }

  if (!currentFrame) {
    return {
      schema: scopedRootSchema,
      currentPropertyName: null,
      partialText: inString ? currentString : '',
      inValue: false,
      insideOpenString: inString,
      usedKeys: [],
    };
  }

  const activeFrame = currentFrame;

  if (inString) {
    if (stringStartedAsKey) {
      return {
        schema: activeFrame.schema,
        currentPropertyName: null,
        partialText: currentString,
        inValue: false,
        insideOpenString: true,
        usedKeys: [...new Set([...activeFrame.usedKeys, ...collectForwardKeys(text, cursorOffset, true)])],
      };
    }

    return {
      schema: activeFrame.pendingValueSchema || {},
      currentPropertyName: activeFrame.lastPropertyName,
      partialText: currentString,
      inValue: true,
      insideOpenString: true,
      usedKeys: activeFrame.usedKeys,
    };
  }

  if (activeFrame.kind === 'object' && activeFrame.inValue) {
    // Unlike a string value (tracked char-by-char above via currentString), a bare literal
    // like a boolean's "true"/"false" leaves no trace in the walk above - so re-triggering
    // completion right after one (e.g. to swap a just-seeded default, see applyCompletion's
    // boolean chain) needs its own backward scan to find how much of it is replaceable.
    // Deliberately NOT reflected in partialText itself: unlike a string value, filtering
    // "true"/"false" by the literal that's already there would just filter the other one
    // out, defeating the point of offering both.
    const bareTokenMatch = /[^\s{}[\],"]+$/.exec(beforeCursor);
    return {
      schema: activeFrame.pendingValueSchema || {},
      currentPropertyName: activeFrame.lastPropertyName,
      partialText: '',
      replaceLength: bareTokenMatch ? bareTokenMatch[0].length : 0,
      insideOpenString: false,
      inValue: true,
      usedKeys: activeFrame.usedKeys,
    };
  }

  return {
    schema: activeFrame.schema,
    currentPropertyName: activeFrame.lastPropertyName,
    partialText: '',
    inValue: false,
    insideOpenString: false,
    usedKeys: [...new Set([...activeFrame.usedKeys, ...collectForwardKeys(text, cursorOffset, false)])],
  };
}

function normalizeCompletionItem(item) {
  return {
    label: item.label,
    kind: item.kind,
    // Raw text only — the caller decides whether to wrap it in quotes, since
    // that depends on whether the cursor is already inside an open string.
    insertText: item.insertText,
    needsQuotes: !!item.needsQuotes,
    detail: item.detail || '',
    // Only set on `property` items - tells the caller what kind of value slot follows,
    // so it can chain straight into the next completion cycle instead of waiting for
    // the user to type the opening character themselves. See applyCompletion (index.js).
    chainValueType: item.chainValueType || null,
    // Only meaningful when chainValueType === 'array' - the resolved type of the array's
    // own items, so the caller can seed a useful first element (e.g. `[""]` for an array
    // of strings) instead of leaving a bare `[]`.
    chainItemType: item.chainItemType || null,
    // Only set for scalar types (boolean/integer/number) that have a schema default -
    // seeds that literal instead of leaving the value slot empty/invalid.
    chainDefault: item.chainDefault !== undefined ? item.chainDefault : null,
    // Richer text for the hover tooltip (index.js) - kept separate from `detail` (the
    // short title already shown inline in the popup row) so the tooltip can show real
    // prose without repeating it. Always plain explanatory text - schema.json no longer
    // carries bare doc-link descriptions (see buildPropertyHint).
    description: item.description || '',
    // A short list of sample values worth showing under the description - either authored
    // directly in schema.json (the "examples" keyword) or, absent that, a small enum's own
    // values (see buildPropertyHint). Null when there's nothing worth showing.
    examples: item.examples || null,
  };
}

// Schema.json's "examples" keyword, when present, is the most direct source of sample
// values for a tooltip. Absent that, a small enum is itself a good stand-in - but only
// when it's short enough to be a helpful hint rather than a wall of text (large enums,
// e.g. every segment type, are already fully browsable via the completion dropdown itself).
const MAX_ENUM_HINT_SIZE = 8;

function buildPropertyHint(resolvedProp) {
  const description = resolvedProp.description && resolvedProp.description !== resolvedProp.title
    ? resolvedProp.description
    : '';

  let examples = Array.isArray(resolvedProp.examples) && resolvedProp.examples.length
    ? resolvedProp.examples
    : null;

  if (!examples) {
    const enumSource = resolvedProp.enum
      || (inferSchemaType(resolvedProp) === 'array' ? getItemSchema(resolvedProp).enum : null);
    if (Array.isArray(enumSource) && enumSource.length > 0 && enumSource.length <= MAX_ENUM_HINT_SIZE) {
      examples = enumSource;
    }
  }

  return { description, examples };
}


// Finds the property-key string (if any) whose quoted range spans `offset`, scanning the
// WHOLE text rather than just up to a cursor - unlike getCompletionContext, a hover target can
// sit anywhere in the document, including after the caret. Mirrors getCompletionContext's own
// object/array frame tracking, but only needs to know "is this string a key, and if so what
// object was it a key of" rather than resolve a full schema chain.
function findKeyTokenAt(text, offset) {
  const stack = [];
  let inString = false;
  let escapeNext = false;
  let stringStart = -1;

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];

    if (inString) {
      if (escapeNext) {
        escapeNext = false;
        continue;
      }
      if (char === '\\') {
        escapeNext = true;
        continue;
      }
      if (char === '"') {
        inString = false;
        const top = stack[stack.length - 1];
        if (top && top.kind === 'object' && top.expectingKey) {
          let k = i + 1;
          while (k < text.length && /\s/.test(text[k])) {
            k += 1;
          }
          if (text[k] === ':') {
            if (offset >= stringStart && offset <= i + 1) {
              return { keyName: text.slice(stringStart + 1, i), contextOffset: stringStart };
            }
            top.expectingKey = false;
          }
        }
      }
      continue;
    }

    if (char === '"') {
      inString = true;
      stringStart = i;
      continue;
    }

    if (char === '{') {
      stack.push({ kind: 'object', expectingKey: true });
      continue;
    }

    if (char === '[') {
      stack.push({ kind: 'array', expectingKey: false });
      continue;
    }

    if (char === '}' || char === ']') {
      stack.pop();
      continue;
    }

    if (char === ',') {
      const top = stack[stack.length - 1];
      if (top && top.kind === 'object') {
        top.expectingKey = true;
      }
      continue;
    }
  }

  return null;
}

// Powers the editor's own hover tooltip (index.js) - hovering directly over a key already
// typed in the config, not just an entry in the completion popup. Resolves the same schema
// info getPropertySuggestions would have offered for that key, using getCompletionContext at
// the offset right before the key's own opening quote so the surrounding object's schema
// still has that key's own definition.
export function getHoverInfo(text, format, offset, schemaScope = 'config') {
  if (format !== 'json') {
    return null;
  }

  const token = findKeyTokenAt(text, offset);
  if (!token) {
    return null;
  }

  const context = getCompletionContext(text, token.contextOffset, schemaScope);
  const resolvedSchema = resolveSchema(context.schema || getScopedRootSchema(schemaScope), ROOT_SCHEMA);
  const propSchema = resolvedSchema.properties?.[token.keyName];
  if (!propSchema) {
    return null;
  }

  const resolvedProp = resolveSchema(propSchema, ROOT_SCHEMA);
  const title = resolvedProp.title || token.keyName;
  const { description, examples } = buildPropertyHint(resolvedProp);

  return {
    title,
    text: description,
    examples,
  };
}

function getPropertySuggestions(context) {
  const resolvedSchema = resolveSchema(context.schema || ROOT_SCHEMA, ROOT_SCHEMA);
  if (!resolvedSchema.properties) {
    return [];
  }

  const usedKeys = context.usedKeys || [];

  return Object.entries(resolvedSchema.properties)
    .filter(([name]) => !usedKeys.includes(name))
    .map(([name, propSchema]) => {
    const resolvedProp = resolveSchema(propSchema, ROOT_SCHEMA);
    const detail = resolvedProp.title || resolvedProp.description || '';
    // Only worth surfacing in the tooltip when it says something detail doesn't already -
    // skip it when description IS what detail fell back to (bare title-less properties).
    const { description, examples } = detail === resolvedProp.description
      ? { description: '', examples: null }
      : buildPropertyHint(resolvedProp);
    // A string-typed enum (e.g. "style") still resolves with type: "string" today because
    // of how the schema expresses it (an anyOf branch with a bare "type": "string" fallback -
    // see resolveSchema's anyOf handling), but inferSchemaType covers any schema shape that
    // omits an explicit type (e.g. per-segment-type conditionals that only narrow
    // "properties"/"items" without repeating "type": "object"/"array").
    const chainValueType = inferSchemaType(resolvedProp);
    const chainItemType = chainValueType === 'array' ? inferSchemaType(getItemSchema(resolvedProp)) : null;
    // Every chainable value type needs *some* concrete literal to seed so the completion
    // always leaves valid JSON behind, even when the schema itself doesn't author a
    // "default" - false/0 are the same harmless placeholders the editor already relies on
    // elsewhere (e.g. an empty "" for strings, an empty {} for objects).
    const chainDefault = chainValueType === 'boolean'
      ? (resolvedProp.default !== undefined ? resolvedProp.default : false)
      : chainValueType === 'integer' || chainValueType === 'number'
        ? (resolvedProp.default !== undefined ? resolvedProp.default : 0)
        : undefined;
    return normalizeCompletionItem({
      label: name,
      kind: 'property',
      insertText: name,
      needsQuotes: true,
      detail,
      description,
      examples,
      chainValueType,
      chainItemType,
      chainDefault,
    });
  });
}

function getEnumSuggestions(context) {
  const resolvedSchema = resolveSchema(context.schema || ROOT_SCHEMA, ROOT_SCHEMA);
  if (resolvedSchema.enum) {
    const description = resolvedSchema.description && resolvedSchema.description !== resolvedSchema.title
      ? resolvedSchema.description
      : '';
    return resolvedSchema.enum.map((value) => normalizeCompletionItem({
      label: value,
      kind: 'value',
      insertText: typeof value === 'string' ? value : JSON.stringify(value),
      needsQuotes: typeof value === 'string',
      detail: `Enum value for ${resolvedSchema.title || 'property'}`,
      description,
    }));
  }

  // A boolean has exactly two possible values, so it's just as pickable as a small enum -
  // offer both instead of leaving the reader to type "true"/"false" by hand.
  if (inferSchemaType(resolvedSchema) === 'boolean') {
    const description = resolvedSchema.description && resolvedSchema.description !== resolvedSchema.title
      ? resolvedSchema.description
      : '';
    return [true, false].map((value) => normalizeCompletionItem({
      label: String(value),
      kind: 'value',
      insertText: String(value),
      needsQuotes: false,
      detail: `Boolean value for ${resolvedSchema.title || 'property'}`,
      description,
    }));
  }

  return [];
}

function getTypeSuggestions() {
  const segmentSchema = getDefinitionSchema('#/definitions/segment', ROOT_SCHEMA);
  const typeSchema = segmentSchema.properties?.type;
  return getEnumSuggestions({ schema: resolveSchema(typeSchema, ROOT_SCHEMA) });
}

function getSuggestionItems(context) {
  const partialText = context.partialText || '';

  if (context.inValue) {
    if (context.currentPropertyName === 'type') {
      return getTypeSuggestions().filter((item) => item.label.startsWith(partialText));
    }

    return getEnumSuggestions(context).filter((item) => item.label.startsWith(partialText));
  }

  return getPropertySuggestions(context).filter((item) => item.label.startsWith(partialText));
}

export function getCompletions(text, format, cursorOffset, schemaScope = 'config') {
  if (format !== 'json') {
    return [];
  }

  const context = getCompletionContext(text, cursorOffset, schemaScope);
  return getSuggestionItems(context);
}

// Companion to getCompletions() giving the caller precise replacement bounds: how many
// already-typed characters to replace, and whether the cursor sits inside an already-open
// string (so insertText for a needsQuotes item must NOT be re-wrapped in quotes). Normally
// that's just partialText.length, but a bare-literal value slot (see getCompletionContext's
// object/inValue branch) reports a separate replaceLength - the token to discard is there,
// but deliberately isn't the same text used to filter which suggestions to show.
export function getCompletionReplacement(text, format, cursorOffset, schemaScope = 'config') {
  if (format !== 'json') {
    return { start: cursorOffset, insideOpenString: false };
  }

  const context = getCompletionContext(text, cursorOffset, schemaScope);
  const partialText = context.partialText || '';
  const replaceLength = context.replaceLength !== undefined ? context.replaceLength : partialText.length;
  return {
    start: cursorOffset - replaceLength,
    insideOpenString: !!context.insideOpenString,
  };
}

