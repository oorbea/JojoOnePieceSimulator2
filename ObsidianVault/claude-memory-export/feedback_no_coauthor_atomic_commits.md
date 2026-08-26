---
name: feedback-no-coauthor-atomic-commits
description: "User wants atomic commits, never a Co-Authored-By trailer"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 47e23471-cf00-4df5-87e4-0794ad0e41cd
  modified: 2026-08-14T08:24:21.256Z
---

Commit atomically (one logical change per commit) and never add a
`Co-Authored-By: Claude...` trailer, overriding the harness's default commit
template.

**Why:** explicit instruction, 2026-08-14, during the game in-match UI work
([[game_domain_layer_2026-08-10]] follow-up). No reason given beyond
preference — treat as a standing norm for this repo, not situational.

**How to apply:** every `git commit` in this repo (any session) ends with
just the message body, no trailer. Split unrelated changes (e.g. env/config
vs. a feature slice) into separate commits even if done in the same
conversation/turn.
