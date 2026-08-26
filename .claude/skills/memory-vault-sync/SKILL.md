---
name: memory-vault-sync
description: Push this machine's Claude Code memory (project norms, workflows, feedback) to the repo's git-tracked ObsidianVault/ folder, or pull it down on a new machine to rebuild memory there. Use when the user says "sync memory to vault", "load memory from vault", "set up Claude on another machine", or when the Stop hook flags unsynced memory changes.
---

# /memory-vault-sync — cross-machine memory via the repo's ObsidianVault/

This project's Claude Code memory lives at a machine-local path (not in git):
`~/.claude/projects/<project-slug>/memory/*.md`.

The sync target is `ObsidianVault/claude-memory-export/` **inside this repo** — a plain folder,
git-tracked, not the `mcp__obsidian-vault__*` MCP tools (that MCP server is connected to a
different, personal vault on this machine and has no content for this project — confirmed
2026-08-26). Use `Read`/`Write`/`Glob`, never the MCP obsidian tools, for this skill.

Because the target is inside the repo, "syncing to another machine" is just committing + pushing
these files and `git pull`-ing on the other machine — no MCP round-trip needed.

## 0. Locate the local memory dir

```bash
win_path=$(pwd -W 2>/dev/null || pwd)
slug=$(printf '%s' "$win_path" | sed 's/[\/:]/-/g')
userprofile_fixed=$(printf '%s' "$USERPROFILE" | tr '\\' '/')
mem_dir="$userprofile_fixed/.claude/projects/$slug/memory"
ls "$mem_dir"
```

Confirm `MEMORY.md` is present before continuing — if not, the slug guess is wrong; ask the user
for the correct memory path instead of guessing further.

## 1. Push (local memory → ObsidianVault/claude-memory-export/)

For every `*.md` file in `$mem_dir` except `MEMORY.md`:

1. Read it (frontmatter + body untouched — this is the whole point, don't paraphrase).
2. Write it verbatim to `ObsidianVault/claude-memory-export/<filename>` in the repo.
3. Skip a file if the destination already has identical content (no-op, no diff noise).

Then write/update `ObsidianVault/claude-memory-export/INDEX.md` with the current local
`MEMORY.md` content, so a reader sees the same one-line-per-memory index without opening every
file.

Before pushing, scan each file for anything that looks like a secret (API key, token, password,
connection string with credentials) despite the memory-type conventions ruling that out — refuse
to push that file and tell the user instead.

Report which files were created/updated/skipped-as-identical, then remind the user these are
**untracked/modified files in the repo** — nothing is synced to another machine until they're
committed and pushed (never do that automatically; per project norm only commit when asked).

## 2. Pull (ObsidianVault/claude-memory-export/ → local memory, new machine)

1. `git pull` first if the user hasn't already, so the export folder is current.
2. Glob `ObsidianVault/claude-memory-export/*.md`, skip `INDEX.md`.
3. For each file: read it, and write it verbatim (frontmatter intact) to `$mem_dir/<filename>` —
   only if it doesn't already exist locally, or its content differs (then overwrite; the repo copy
   is the sync source of truth for this flow).
4. For each file written/updated, ensure a matching one-line entry exists in local `MEMORY.md`
   (append if missing — never duplicate an existing entry for the same file).
5. After loading, tell the user which memories were loaded and flag any that name a specific
   file/function/flag worth re-verifying against current repo state before relying on it (per the
   "before recommending from memory" rule already in the memory system prompt).

## Which mode to run

- User says "sync to vault", "push memory", or the Stop-hook reminder fired → run Push (§1).
- User says "load memory", "pull from vault", "set up on this machine" → run Pull (§2).
- Ambiguous → ask.
