---
description: 'Generate a new Oh My Posh segment (code, registration, docs, schema, sidebar)'
agent: 'agent'
model: 'Claude Sonnet 4'
tools: ['execute/testFailure', 'execute/getTerminalOutput', 'execute/runTask', 'execute/createAndRunTask', 'execute/runInTerminal', 'execute/runTests', 'read/problems', 'read/readFile', 'read/terminalSelection', 'read/terminalLastCommand', 'read/getTaskOutput', 'edit/createFile', 'edit/editFiles', 'search', 'web/githubRepo', 'agent']
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

Documentation

- Use the `runSubagent` tool with `agentName: "Segment Documentation"` to delegate
  documentation creation and updates.
- Pass the segment details (name, category, title, description, properties, template)
  in the prompt parameter for the subagent to generate or update the MDX documentation file.
