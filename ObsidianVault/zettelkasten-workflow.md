---
title: "Workflow: Obsidian Zettelkasten updates"
tags:
  - project
  - jojo-onepiece-simulator
  - workflow
  - meta
---

# Workflow — Obsidian Zettelkasten Updates

## Rule

Before implementing any task, review the ObsidianVault (`ObsidianVault/`) for project context. After any of these events, create or update vault notes following Zettelkasten method (atomic notes, interlinked):

- Decision made
- New learning or gotcha discovered
- Functionality added, changed, or removed

## Why

Owner wants all project knowledge centralized and interconnected in Obsidian. Not just code — decisions, reasoning, and context must be preserved for future reference.

## How to apply

1. Read relevant vault notes before starting work.
2. After changes: create atomic note (one idea per note) with frontmatter (title, tags), body, and `[[wikilinks]]` to related notes.
3. Keep existing notes updated if context changes.
4. **Pending work lives only in [[TODO]].** Never leave a bare "pendiente"/"not built yet" in a topic note: add a line to TODO.md and link it (`→ ver [[TODO]]`). When finishing something, delete it from TODO.md in the same commit. Notes with stale "not built yet" text were the cause of the 2026-10-08 audit.

Related: [[ADR]], [[overview]]
