// Language + schema-completion wiring for the editor, kept free of React so it can be unit
// tested (in principle) and reasoned about independently from the mount/update lifecycle in
// index.js. json/yaml get schema-aware completion and hover from schemaCompletion.js's own
// syntax-tree walker (see that file's header for the schema-resolution library workaround it
// replaced); toml has no such support available, so it stays a plain syntax-only mode - it was
// never covered by the old hand-rolled completion.js engine either.
import { json, jsonLanguage } from '@codemirror/lang-json';
import { yaml, yamlLanguage } from '@codemirror/lang-yaml';
import { StreamLanguage } from '@codemirror/language';
import { toml } from '@codemirror/legacy-modes/mode/toml';
import { schemaCompletionSource, schemaHoverTooltip, schemaHintTheme } from './schemaCompletion';
import { getScopedRootSchema } from './schemaResolution';

// 'segment' (a segment doc's sample editor) completes against the bare segment shape; 'config'
// (the studio, and the default) completes against the full theme schema. Delegates to
// schemaResolution.js's getScopedRootSchema, which resolves the scoped shape against the schema's
// own `definitions` so its `#/definitions/...` refs still resolve regardless of scope.
export function getScopedSchema(schemaScope) {
  return getScopedRootSchema(schemaScope);
}

// One extension array per format, swapped into index.js's language Compartment whenever the
// format or schemaScope changes.
export function getLanguageExtensions(format, schemaScope) {
  switch (format) {
    case 'json':
      return [
        json(),
        jsonLanguage.data.of({ autocomplete: schemaCompletionSource(schemaScope) }),
        schemaHoverTooltip(schemaScope),
        schemaHintTheme,
      ];
    case 'yaml':
      return [
        yaml(),
        yamlLanguage.data.of({ autocomplete: schemaCompletionSource(schemaScope) }),
        schemaHoverTooltip(schemaScope),
        schemaHintTheme,
      ];
    case 'toml':
      return [StreamLanguage.define(toml)];
    default:
      return [];
  }
}
