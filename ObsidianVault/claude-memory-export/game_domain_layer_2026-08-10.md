---
name: game_domain_layer_2026-08-10
description: "Gauntlet/Versus game domain layer built (backend), domain-only, docs in vault"
metadata: 
  node_type: memory
  type: project
  originSessionId: 6e921e5d-7825-4270-9066-aaeb29870548
  modified: 2026-08-11T11:52:24.173Z
---

Built the full domain layer (State/Strategy/Template Method, no infra/routes/migrations) for the
Gauntlet (cooperative) and Versus (2-team) game modes at `apps/backend/internal/domain/entities/game/`,
plus 12 new enums and 6 new ports (`IStageCatalog`, `IGamePowerPool`, `IAssignmentWeights`,
`ITiebreaker`, `IGameHistory`, `IInventory`). Deleted dead `user.Player`. All rules were nailed down
via extensive Q&A with the user before writing any code, then written to
`C:\code\JojoOnePieceSimulator2\ObsidianVault\gameplay-game-modes.md` (rules) and
`gameplay-domain-design.md` (technical design/patterns/known debt) — read those before touching
this feature again.

**Why:** user wants gameplay rules preserved with precision as the "final goal" reference, per
[[feedback_obsidian_workflow]]. Follow-up tandas since this memory was written (all done, see
[[stages_admin_crud_2026-08-11]] for the latest): Redis lobby storage, websocket transport,
application services/DTOs/routes, stage catalog content + per-viewer locale, and stage admin CRUD
(backend tests + full frontend). Still-open port adapters as of 2026-08-11: assignment weights,
tiebreaker, history, inventory.

**How to apply:** before extending game logic, read the two vault notes above in full — they encode
non-obvious invariants (fruit-before-mastery coupling, the 3 named Stands forcing spin=INFINITE,
early-close voting on all-humans-voted, host reassignment on disconnect, bot-fill Versus-only) that
aren't discoverable from a quick code skim.
