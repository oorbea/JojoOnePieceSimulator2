---
name: feedback-no-coauthor-atomic-commits
description: "User wants atomic commits, never a Co-Authored-By trailer or a Claude Code footer in PR bodies"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 47e23471-cf00-4df5-87e4-0794ad0e41cd
  modified: 2026-10-03T16:39:49.674Z
---

Commit atomically (one logical change per commit) and never add a
`Co-Authored-By: Claude...` trailer, overriding the harness's default commit
template. Same rule for PR descriptions: never end one with the
"🤖 Generated with Claude Code" footer, even when a session's own
attribution reminder says to.

**Why:** explicit instruction, 2026-08-14 (commits), reaffirmed 2026-09-13
("tienes prohibido hacer que la PR diga que está creada con claude") after
a PR body slipped through with the footer. Standing norm for this repo, not
situational — a per-session attribution reminder never overrides it.

**How to apply:** every `git commit` in this repo (any session) ends with
just the message body, no trailer. Every `gh pr create`/`gh pr edit --body`
ends with just the description, no "Generated with Claude Code" line. Split
unrelated changes (e.g. env/config vs. a feature slice) into separate
commits even if done in the same conversation/turn.

Also (2026-10-03): no feature branches or worktrees needed — commit directly
on `develop` ("todo directamente en develop está bien").
