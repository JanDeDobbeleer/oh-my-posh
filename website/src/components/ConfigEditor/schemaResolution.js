// Schema-resolution engine ported unchanged (behaviourally) from the pre-CodeMirror-6 editor's
// own completion.js (see `git show 04a7124a:website/src/components/ConfigEditor/completion.js` -
// the commit before this file was deleted in favour of a third-party schema-completion package,
// and the one before THAT deprecated it in favour of this hand-rolled walker again). It correctly
// resolves this schema's own `$ref`-with-sibling-keywords nodes (draft 2020-12 composition, not
// something that package's own json-schema-library v9 resolver ever supported - it threw
// "Mutiple typeIds matched" on several of schema.json's own segments) and its per-segment-type
// `if`/`then` branches. Kept free of CodeMirror imports so it can be reasoned about (and in
// principle unit tested) independently of the syntax-tree walker that now drives it
// (schemaCompletion.js) instead of the old char-by-char text scanners.
import schema from '../../../../themes/schema.json';

// Walks a schema node, resolving `$ref` and composing `anyOf`/`oneOf`/`allOf` into one plain
// object completion can read `.properties`/`.enum`/`.type` off of directly.
export function resolveSchema(node, root = schema) {
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

export function mergeSchema(base, incoming) {
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

export function getDefinitionSchema(ref, root = schema) {
  if (!ref || typeof ref !== 'string' || !ref.startsWith('#/')) {
    return {};
  }

  const target = ref.split('/').slice(1).reduce((acc, part) => acc?.[part], root);
  return resolveSchema(target, root);
}

export function deriveValueSchema(parentSchema, propertyName, root = schema) {
  const resolvedParent = resolveSchema(parentSchema, root);
  const propertySchema = resolvedParent.properties?.[propertyName];

  if (!propertySchema) {
    return {};
  }

  return resolveSchema(propertySchema, root);
}

// Only `definitions.segment` uses conditional `if.properties.type.const` / `then`
// branches today, but this stays generic so any similarly-shaped schema benefits.
export function mergeTypeBranch(schemaNode, typeValue) {
  const resolved = resolveSchema(schemaNode);
  if (!resolved.allOf || !typeValue) {
    return resolved;
  }

  const merged = { ...resolved };
  resolved.allOf.forEach((branch) => {
    if (branch.if?.properties?.type?.const === typeValue) {
      const thenSchema = resolveSchema(branch.then, schema);
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

export function getItemSchema(parentValueSchema) {
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
// Completion still needs a concrete type to know what to offer, so fall back to inferring it
// from shape - "properties" implies object, "items" implies array - before falling back
// further to a bare enum implying string.
export function inferSchemaType(resolvedSchema) {
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

// Schema.json's "examples" keyword, when present, is the most direct source of sample
// values for a tooltip. Absent that, a small enum is itself a good stand-in - but only
// when it's short enough to be a helpful hint rather than a wall of text (large enums,
// e.g. every segment type, are already fully browsable via the completion dropdown itself).
const MAX_ENUM_HINT_SIZE = 8;

export function buildPropertyHint(resolvedProp) {
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

// Which schema governs the outermost container of the document being edited. The studio edits
// a full theme config (schema fits directly); a segment doc page's sample editor only ever
// contains a single bare segment object (see website/src/components/Config.js), so completion
// there needs to start from #/definitions/segment instead - otherwise every top-level
// property/hover lookup resolves against the wrong schema entirely (root config fields like
// "final_space" instead of segment fields like "foreground").
export function getScopedRootSchema(schemaScope) {
  if (schemaScope === 'segment') {
    return getDefinitionSchema('#/definitions/segment', schema);
  }
  return schema;
}
