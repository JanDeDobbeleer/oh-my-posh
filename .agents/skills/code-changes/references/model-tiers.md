# Model tiers and agent tooling

The workflow talks about capability tiers, not vendor names. Map whatever stack is in use onto
these four tiers. The model names below are a mid-2026 snapshot — they go stale; when a name no
longer exists, map its replacement by tier, not by nostalgia.

## The four tiers

| Tier        | Role                                       | Anthropic         | OpenAI            | Google            |
| ----------- | ------------------------------------------- | ----------------- | ----------------- | ----------------- |
| Escalation  | Judgment calls the coordinator flags        | Fable 5, Opus 4.8 | Sol, GPT-5 (high) | Gemini 2.5 Pro    |
| Coordinator | Resident; owns every phase by default       | Sonnet 5          | GPT-5, GPT-4.1    | Gemini 2.5 Flash  |
| Implementer | Pinned-spec work, often same tier as coordinator | Sonnet 5     | GPT-5, GPT-4.1    | Gemini 2.5 Flash  |
| Trivial     | Mechanical edits                            | Haiku 4.5         | GPT-4.1 mini      | Gemini Flash-Lite |

- **Escalation** — the strongest reasoning model available, called only when a trigger in
  [escalate.md](escalate.md) fires: unclear root cause, architectural risk, security sensitivity,
  an irreversible operation, repeated spec gaps, or low confidence in a review. Answers one
  specific question, then hands control back to the coordinator.
- **Coordinator** — a capable mid-tier model, resident for the whole task. Owns analysis,
  planning, supervision, and verification by default. It's capable enough for the large majority
  of work, and cheap enough relative to Escalation tier that most tasks never need to call up at
  all — that's the point of the split.
- **Implementer** — often literally the same model as the coordinator; the distinction is
  parallelism and workspace isolation, not capability. Falls back to a smaller model for batches
  of mechanical, unambiguous edits.
- **Trivial** — small, very fast models for mechanical, unambiguous edits: renames, config
  tweaks, typo fixes, doc touch-ups. Batch several to amortize the dispatch overhead. When in
  doubt between trivial and implementer, pick implementer — a wrong cheap edit costs more than
  the price difference.

## Tooling mappings

How to run the coordinator/escalation/implementer split in common agent tooling:

- **Claude Code** — run the session itself on a coordinator-tier model (e.g. Sonnet 5). Call
  `Agent` with `model: opus` (or `fable`) only at an escalation checkpoint, scoped to the one
  question that triggered it. Use worktree isolation for parallel implementer-tier tasks, and
  `model: haiku` for batched trivial edits.
- **GitHub Copilot** — run Copilot Chat on a coordinator-tier model for the whole flow; switch the
  model picker to the strongest reasoning model only for the specific question an escalation
  trigger flagged, then switch back. Standalone, well-specified tasks can still be assigned to the
  Copilot coding agent (assign the issue or PR to Copilot), which works on its own branch — the
  pinned spec becomes the issue body.
- **Codex** — orchestrate locally (CLI or chat) on a coordinator-tier model; dispatch each pinned
  spec as a Codex cloud task, one task per independent unit, and review the resulting diffs
  yourself, escalating to the frontier model only when a trigger fires.
- **Cursor and similar IDE agents** — one composer/agent session per task on a coordinator-tier
  model; background agents for the parallel implementer-tier ones; switch to the strongest model
  in-session only for an escalation checkpoint, then switch back.

**No subagent support at all?** Keep the phases, drop the parallelism: run the whole flow on a
coordinator-tier model, and switch the session model up to the strongest one only for the specific
question an escalation trigger flagged, then switch back down. The discipline transfers even when
the delegation mechanism does not.
