---
name: feedback_obsidian_workflow
description: "Always check Obsidian vault before implementing, and save learnings/decisions there"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 6d9b0ff3-71c5-41e7-bd5c-747a86a3fdfa
  modified: 2026-07-28T15:56:41.071Z
---

Before implementing anything in this project, read relevant notes in the Obsidian vault first (e.g. `Projects/JojoOnePieceSimulator2/`). After learning something new or making a decision, write it back to the vault (not just local memory).

**Why:** user wants vault, not just Claude Code memory, as source of truth for project knowledge — supports [[project_context]] and [[cicd_picture_pipeline]].
**How to apply:** at start of implementation tasks, search/read the vault folder for this project before acting. When a decision is made or something learned mid-task, update/create the relevant vault note (e.g. ADR.md) in the same turn, not deferred.
