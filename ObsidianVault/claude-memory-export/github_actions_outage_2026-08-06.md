---
name: github-actions-outage-2026-08-06
description: GitHub Actions incident on 2026-08-06 caused CI to silently not trigger on push/PR for this repo — not a repo/workflow config bug
metadata: 
  node_type: memory
  type: project
  originSessionId: 61265a5b-22f6-40a4-b8cc-8a4c4b03ffb8
  modified: 2026-08-06T21:54:42.680Z
---

On 2026-08-06, pushes to `develop` and PR#9 (develop→main) triggered zero Actions runs — no check-runs, no queued/pending runs, nothing in the runs API at all. Ruled out: workflow file changes, `[skip ci]`, paths-filters, repo disabled/archived, Actions permissions (all confirmed fine via `gh api`). Close/reopen PR and an empty-commit push both failed to trigger a run too.

Root cause: GitHub-wide Actions incident — https://www.githubstatus.com/incidents/qcvjkzcs7j74

**Why:** worth remembering so a future "CI not starting" report isn't immediately treated as a pipeline/config bug — check githubstatus.com first if all the `gh api` checks (workflow state, permissions, trigger config, recent runs for the commit) come back clean.

**How to apply:** when CI doesn't trigger and `.github/workflows/ci.yml` triggers/config look correct and Actions permissions are enabled, check https://www.githubstatus.com/ before deep-diagnosing the repo. See [[cicd_picture_pipeline]] for the pipeline's actual job structure (unrelated to this incident).
