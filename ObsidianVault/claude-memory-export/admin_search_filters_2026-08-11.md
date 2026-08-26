---
name: admin_search_filters_2026-08-11
description: Server-side search (?q=) + filter bars added to Stands/Devil Fruits/Stages admin screens
metadata: 
  node_type: memory
  type: project
  originSessionId: 6e088623-0619-4a53-9e7b-424c0cdf31be
  modified: 2026-08-11T16:23:58.482Z
---

Added search+filters to Stands and Devil Fruits admin screens (dead `StandFilters`/
`DevilFruitFilters` wiring finally used), and mid-tanda extended it to a new `?q=`
backend search param used by all three entities including Stages, which migrated off
its client-side text filter. Full writeup: `ObsidianVault/admin-search-and-filters.md`.

**Why:** user asked to add search/filters to Stands/Devil Fruits "like Stages has";
once it turned out Stages' search wasn't backend-backed either, user redirected to
make it server-side everywhere instead of shipping two different search strategies.

**How to apply:** two traps if touching these filters again — [[game_domain_layer_2026-08-10]]-era code:
- Cache keys (`internal/infrastructure/cache/keys.go` `standFilterKey`/`devilFruitFilterKey`)
  are hand-serialized field lists, NOT derived from the struct — adding a filter field
  without adding it to the key string means two different filter values share one Redis
  cache slot (wrong/stale results, no error). Backend has a test guarding this now.
- `FilterStandRows`' WHERE lives inside a recursive CTE (`base`) because of the
  evolvesFrom ancestor chain — adding a join there must not add columns to `base`'s
  SELECT list, or the `UNION` with the recursive branch breaks.

Related: [[stages_admin_crud_2026-08-11]].
