---
name: segment-create
description: >
  Checklist for adding a new Oh My Posh segment. Use when asked to add or create a new segment:
  lists every file to touch, the existing segment to copy, and how to verify.
---

## Template

Copy `src/segments/taskwarrior.go` and `src/segments/taskwarrior_test.go`: two options, a
`Template()`/`Enabled()` pair, and every shell call going through the mocked environment.

- Embed `Base` first in the struct; its exported fields are the template properties.
- Declare options as `options.Option` constants; read them with `s.options.String(...)`,
  `s.options.Bool(...)`, `s.options.KeyValueMap(...)` and so on, passing the default there.
- Reach the OS, shell, files and network only through `s.env` so tests can mock it.
- Override `Activation()` only for a cheap precondition such as a project file (see
  `src/segments/dvc.go`); `Base` defaults to always active.
- Tests are table-driven: a `mock.Environment` from `src/runtime/mock`, options as an
  `options.Map`, then `Init(props, env)` and assertions on `Enabled()` and the exported fields.

## Checklist

Keep every list sorted alphabetically.

1. `src/segments/<id>.go` and `src/segments/<id>_test.go`.
1. `src/config/segment_types.go`: add the constant `<ID> SegmentType = "<id>"`.
1. `src/config/segment_registry.go`: add `gob.Register(&segments.<GoType>{})` to `init()` and
   `<ID>: func() SegmentWriter { return &segments.<GoType>{} }` to `Segments`. A missing map
   entry fails silently at runtime: the segment never renders.
1. `themes/schema.json`: add `<id>` to `definitions.segment.properties.type.enum` and an
   `if`/`then` block to `definitions.segment.allOf` declaring every option. Copy the
   `taskwarrior` block; its `unevaluatedProperties: false` rejects undeclared options.
1. `website/docs/segments/<category>/<id>.mdx`, where category is one of agents, cli, cloud,
   health, languages, music, scm, system or web. Follow the `segment-docs` skill, and keep a
   `## What` paragraph: the segment catalog plugin reads it.
1. `website/sidebars.js`: add `"segments/<category>/<id>"` to that category's `items`.
1. `website/segment_data.json`: add sample data under `<id>`, shaped like the `taskwarrior` entry.
1. `website/plugins/segments/registry.json`: generated. The first `go test ./config/...` run
   rewrites it and fails; rerun to confirm, then keep the file.
1. Credentials needed? Add the doc id to `AUTH_TIERS` in `website/plugins/segments/data.js`.

## Verify

From `src/`:

```shell
go test ./segments/... -run Test<GoType>
go test ./config/...
```

Run `npm run build` in `website/` only when docs changed.
