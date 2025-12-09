---
description: 'Automate scaffolding for a new Oh My Posh segment: code, registration, docs, schema, sidebar updates'
applyTo: ['src/segments/*.go', 'website/docs/**/*.mdx', 'themes/schema.json', 'website/sidebars.js']
---

# Segment scaffolding instructions

Goal

- Given user inputs (segment id, Go type name, title, category, description,
  properties, template), generate all required code and docs to add a new
  segment end-to-end, following repo conventions.

Inputs

- id: kebab/slug used as `type` and docs filename (e.g., `new`)
- goType: PascalCase name for the Go struct (e.g., `New`)
- title: human readable (e.g., `New`)
- category: one of cli|cloud|health|languages|music|scm|system|web
- description: one-line description
- properties: list of { key, type, title, description, default }
- template: default template string (e.g., ` {{.Text}} `)

Contract

- Idempotent: do not duplicate registrations, constants, map entries, sidebar
  links, or schema entries.
- Alphabetical insertions where applicable.
- Compile-ready Go code, formatted.
- Docs lint clean according to `.github/instructions/markdown.md`.

Implementation steps

1. Create Go writer file: `src/segments/<id>.go`

- If file exists, skip creation.
- Use this template; include property consts for each property key.

```go
package segments

import (
    "github.com/jandedobbeleer/oh-my-posh/src/segments/options"
    "github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type {{goType}} struct {
    Base

    // computed fields used in template
    Text string
}

// properties
const (
{{#each properties}}
    // {{this.title}}: {{this.description}}
    {{ pascalCase this.key }} options.Property = "{{this.key}}"
{{/each}})

func (s *{{goType}}) Enabled() bool {
    // set up data for the template, using defaults from properties
    {{#if (propExists properties "text")}}
    s.Text = s.props.GetString({{ pascalCase "text" }}, {{ defaultFor "text" }})
    {{else}}
    s.Text = s.props.GetString({{ pascalCase (firstKey properties) }}, "")
    {{/if}}
    return true
}

func (s *{{goType}}) Template() string {
    return {{ printf "%q" template }}
}
```

1. Register in `src/config/segment_types.go`

- Ensure in `init()` there is `gob.Register(&segments.{{goType}}{})` exactly
  once.
- Add constant: `{{ upper id }} SegmentType = "{{id}}"` in the alphabetical
  block.
- Add to `var Segments = map[SegmentType]func() SegmentWriter{}` with key
  `{{ upper id }}` mapping to `&segments.{{goType}}{}`.
- Keep lists alphabetically sorted. If not sorted, insert at correct position.

1. Documentation file

- Path: `website/docs/segments/{{category}}/{{id}}.mdx`.
- If file exists, skip. Else create with this template:

````mdx
---
id: {{id}}
title: {{title}}
sidebar_label: {{title}}
---

## What

{{description}}

## Sample Configuration

import Config from '@site/src/components/Config.js';

<Config data={{
  "type": "{{id}}",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "properties": {
{{#each properties}}
    "{{this.key}}": {{ json this.default }},
{{/each}}
  }
}}/>

## Properties

| Name | Type | Description | Default |
| ---- | ---- | ----------- | ------- |
{{#each properties}}| `{{this.key}}` | `{{this.type}}` | {{this.description}} | `{{ stringify this.default }}` |
{{/each}}
````

1. Sidebar

- Update `website/sidebars.js` under the correct category array to include
  `"segments/{{category}}/{{id}}"`.
- Insert alphabetically; if already present, do nothing.

1. JSON Schema

- File: `themes/schema.json`.
- Add `"{{id}}"` to `#/definitions/segment/properties/type/enum` if missing.
- Add an `allOf` entry for this segment guarded by
  `{ properties: { type: { const: "{{id}}" } } }` that declares each property
  as defined by inputs. Use appropriate JSON Schema types and include title,
  description, default.
- Keep the `allOf` array in a stable order by type name if feasible; otherwise
  append if not present.

Validation

- After changes, run `go build` (task: build oh-my-posh). Ensure no compile
  errors.
- Check markdown formatting; respect 120-char line length and fenced blocks with
  language.

Notes

- Use UTF-32 escapes (e.g., "\uEFF1") for icon defaults in docs and code.
- Keep code minimal. Complex logic should be added by maintainers after
  scaffold if needed.

Optional

1. Tests

- Create a minimal test file at `src/segments/{{id}}_test.go` using the
  table-driven style. Include at least a happy-path test that asserts
  `Enabled()` returns true and the template renders expected output with default
  options.
