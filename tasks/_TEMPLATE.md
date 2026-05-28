<!--
Per-task handoff file. Copy to tasks/<id>-<slug>.md to start a task (e.g. tasks/T-001-db-schema.md).
Self-contained: a fresh agent on another machine must be able to resume from this file alone.
Keep terse. Delete this file when the task ships (fold the outcome into tasks/progress.md → Project State).
-->

# T-NNN · <task title>

Status: planned <!-- planned | in progress | blocked | deferred | in review -->
Owner/Machine: <who@machine>
Started: YYYY-MM-DD · Updated: YYYY-MM-DD
Branch: <git branch> · Dashboard: [progress.md](./progress.md)

## Goal

What "done" looks like — the acceptance bar. 1–3 sentences.

## Context

Why this task exists, constraints, and links to specs / ADRs / docs. Enough for a cold start.

## Plan / Checklist

- [ ] step
- [ ] step
- [x] done step _(commit <sha>)_

## Current state

Where things stand right now, what the last session left, and the **exact next action**.
This is the resume anchor — keep it accurate.

## Decisions (task-local)

Transient choices for THIS task only. Durable cross-cutting decisions go to `docs/decisions/` (link the ADR here).

- Chose X over Y because Z.

## Blockers / open questions

- none

## Verification

How to prove the task works end-to-end (commands + expected output).

```
<command>
# expected: <result>
```

## Key files

- `path/to/file` — role
