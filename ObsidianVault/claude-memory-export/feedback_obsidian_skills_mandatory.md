---
name: feedback-obsidian-skills-mandatory
description: "All operations on the project's local ObsidianVault must go through the user-level obsidian-skills, never the mcp__obsidian-vault__* MCP tools"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: c5104002-4051-4d64-8e7d-db876428d83c
  modified: 2026-08-31T11:43:53.468Z
---

For ANY operation on this project's local vault (`ObsidianVault/` in the repo) — read, search, create, edit, link — always use the `obsidian:*` skills (obsidian-cli, obsidian-markdown, obsidian-bases, json-canvas, defuddle), installed at user level. Never use the `mcp__obsidian-vault__*` MCP tools for this project.

**Why:** the `mcp__obsidian-vault__*` MCP server points at a different vault entirely (BrainTrust project's vault, not this repo's `ObsidianVault/`) — confirmed 2026-08-31 when `search_vault`/`list_tasks` returned BrainTrust content instead of this project's notes. Using it silently gives wrong-vault results.

**How to apply:** before touching the vault, load the relevant `obsidian:*` skill via Skill tool instead of reaching for `mcp__obsidian-vault__*`. Read/search still works fine with plain Read/Grep/Glob on `ObsidianVault/` directly (as done historically per [[feedback_obsidian_workflow]]) — the skills matter most for writes (create/edit notes, frontmatter, wikilinks) so format stays Obsidian-correct.
