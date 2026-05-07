---
name: vale-user-facing-text
description: >
  Enforce Vale prose checks on any user-facing text the agent creates or edits —
  comments, documentation, status messages, release notes, READMEs, or any
  explanatory content that users will read. Always use this skill when producing
  or changing prose, even when Vale is not explicitly requested.
---

# Vale Validation For User-Facing Text

This skill ships its own Vale configuration as a bundled artifact so it works
consistently regardless of whether the project has its own `.vale.ini`.

## Bundled Config

The embedded config lives at `config/.vale.ini` inside this skill's directory.
It declares the style packages and targets all Markdown files. Vale downloads
the style packages into `config/.styles/` next to it, keeping everything
self-contained.

Always pass `--config <SKILL_DIR>/config/.vale.ini` to every Vale invocation.
Never rely on the project's own `.vale.ini` or omit the flag.

## Resolving `<SKILL_DIR>`

`<SKILL_DIR>` is the directory that contains this `SKILL.md` file. Determine
it at runtime from the path the skill was loaded from. Examples:

- Local skill: `skills/vale-user-facing-text`
- APM module: `apm_modules/<vendor>/agentic/skills/vale-user-facing-text`

## When To Apply

Apply this skill whenever creating or editing user-facing text, including:

- Markdown documentation (`.md`, `.mdx`)
- Code comments and docstrings that explain behavior to humans
- Changelogs, release notes, migration notes, and onboarding guides
- Any generated informational text intended for users

## First-Run Setup

Before running Vale, check whether `config/.styles/` exists under `<SKILL_DIR>`.
If that folder is missing, run sync first. Do not run validation commands until
sync has completed successfully.

When `config/.styles/` does not exist:

```shell
vale sync --config <SKILL_DIR>/config/.vale.ini
```

After sync, the downloaded styles are stored in `<SKILL_DIR>/config/.styles/`.
For later runs, skip sync if that folder already exists.

## Validation Workflow

1. Identify all touched files containing user-facing prose. This includes:
   - `.md` and `.mdx` files edited or created in this change
   - Source files (`.go`, `.py`, `.ts`, etc.) that contain comments or docstrings
     with human-readable explanations rather than code identifiers
2. Ensure style packages are present: if `<SKILL_DIR>/config/.styles/` is absent,
   run `vale sync --config <SKILL_DIR>/config/.vale.ini` before proceeding.
3. Run Vale against each edited Markdown file:

```shell
vale --config <SKILL_DIR>/config/.vale.ini <edited-file>
```

4. For user-facing prose in non-Markdown locations (such as inline code
  comments), copy the prose into a temporary `.md` file and validate it.
  Name the temp file `_vale_tmp.md` and place it next to the source file.
  Delete it once validation is complete, never commit it.

```shell
vale --config <SKILL_DIR>/config/.vale.ini _vale_tmp.md
```

5. Apply the wording fixes back to the source while preserving technical
  accuracy and intent.
6. Re-run Vale until remaining findings are resolved or intentionally kept with
  a documented reason.

## Handling Findings

- Fix clarity, tone, and wording issues first.
- Prefer concise, direct phrasing over verbose language.
- Keep domain-specific terms and code identifiers unchanged when required.
- If a suggestion conflicts with correctness, keep the correct wording and
  briefly note the reason in your response.

## Completion Checklist

- [ ] Style packages are present in `config/.styles/`
- [ ] Every edited user-facing text section was validated with the bundled config
- [ ] Vale findings were resolved or explicitly justified
- [ ] Final prose remains technically accurate and audience-appropriate
