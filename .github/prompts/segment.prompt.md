---
description: 'Generate a new Oh My Posh segment (code, registration, docs, schema, sidebar)'
agent: 'agent'
model: 'Claude Sonnet 4'
tools: ['web/githubRepo', 'search/codebase', 'edit/createFile', 'edit/editFiles', 'read/problems', 'execute/getTerminalOutput', 'execute/runInTerminal', 'read/terminalLastCommand', 'read/terminalSelection', 'execute/createAndRunTask', 'execute/runTask', 'read/getTaskOutput', 'execute/runTests', 'search', 'execute/testFailure', 'search/usages']
---

# New Segment Prompt

Provide the inputs below. Keep this prompt minimal;
the detailed steps live in `.instructions/segment.md`.

Required

- Segment name
  - goTypeName (PascalCase, e.g., `New`)
  - id/slug (kebab/slug, e.g., `new`)
- Category (one of: cli, cloud, health, languages, music, scm, system, web)
- Title (e.g., `New`)
- One-line description

Optional

- Properties (list of objects with: key, type, title, description, default)
- Default template string (e.g., ` {{.Text}} `)
- Default style/background/foreground for docs sample
- Custom sample configuration block (optional)

Execute

- After inputs are provided and validated, run:
  - `.instructions/segment.md`
  - Pass the collected inputs (`goTypeName`, `id/slug`, `category`, `title`,
    `description`, `properties`, `template`).
