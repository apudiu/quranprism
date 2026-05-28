# Product Requirements

What quranprism must do and **why** — the product/business source of truth.

- **HOW** we build it → `docs/decisions/` (engineering ADRs).
- **What's being worked on now** → `tasks/`.

## Vision

quranprism plays Quran recitation interleaved with translations in multiple languages, in the order and combination the listener chooses. One source text, refracted into every language a listener wants — verse by verse.

## Tiers (canonical gating table)

_Free vs paid breakdown to be defined. Add the canonical tier matrix here once pricing decisions are made._

## Domains

_Per-domain requirement files to be added (e.g. Recitation, Translation, UserPrefs, Subscription, Privacy, Compliance) with stable requirement IDs._

| Domain | File | IDs | Scope |
|---|---|---|---|
| _(none yet)_ | | | |

## Conventions

- Requirement IDs are **stable** — never renumber. Cite them from code comments, task files, and ADRs (e.g. "implements `REC-3`").
- To retire a requirement: strike it (`~~REC-9~~ deprecated`) briefly, then delete.
- Keep each file **minimal** — requirements, not prose. Prune the obsolete.
- Product/business rules live here, not in code or `CLAUDE.md`. Engineering choices go to `docs/decisions/`.
- Add a domain by copying [`_TEMPLATE.md`](./_TEMPLATE.md).
