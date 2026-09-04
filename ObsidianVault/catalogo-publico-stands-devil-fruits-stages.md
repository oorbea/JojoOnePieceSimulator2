---
title: "Feature: public read-only catalogue for Stands/Devil Fruits/Stages (2026-09-03)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
---

# Public read-only catalogue (2026-09-03)

Before this, Stands/Devil Fruits/Stages were only browsable at `/admin/*`, gated to `ADMIN`. A
regular player had no way to look at the catalogue. Owner's instructions: any logged-in user
should be able to browse it read-only, with a "card brought to the foreground" detail view (also
usable from admin, with an Edit button there), and no create/edit/delete anywhere outside admin.

## Status

Done. See [[nav-indicador-deslizante]] for the unrelated top-nav change shipped in the same pass,
and [[game-lobby-frontend]] for the home-screen channel grid this unlocked.

## Reuse strategy: `readOnly` as a discriminated union, not a copy

The three admin screens (`stands-screen.tsx`/`devil-fruits-screen.tsx`/`stages-screen.tsx`) were
already a strict container → screen → card triple with the entire write surface isolated into two
prop bundles (`form`, `deleteConfirm`) plus `onCreateNew`/`onEdit`/`onDelete`/`openingEditId` - see
[[admin-panel-crud-ux-fixes]]. That made the read-only variant a type-level flag instead of a
duplicated screen:

```ts
type BaseProps = { /* data, loading/error, search, filters, detail-modal state */ }
type WritableProps = { readOnly?: false; onCreateNew; onEdit; onDelete; openingEditId; form; deleteConfirm }
type ReadOnlyProps = { readOnly: true }
type Props = BaseProps & (WritableProps | ReadOnlyProps)
```

Destructuring the union props object directly doesn't narrow on later property access - the
working pattern is to keep the raw `props` parameter (not destructure the writable-only fields)
and gate every writable-only branch on `props.readOnly` directly (`props.readOnly ? null :
<GlossButton onPress={props.onCreateNew} .../>`), which TS does narrow correctly across a ternary.

Cards (`stand-card.tsx` etc.) got the same treatment: `onOpenDetail: () => void` (required, opens
the new detail modal) plus `readOnly?: boolean` with `onEdit`/`onDelete` now optional, hidden
behind `{readOnly ? null : <actions row/>}`.

New thin containers (`CatalogStandsContainer` etc., one per feature, next to the admin container)
duplicate only the read side of the admin container: search + debounce, filter state, the
`hasActiveFilters`-branched query. **Everything write-only is deliberately not copied**:

- `use-*-mutations` hooks and `openCreate`/`openEdit`/`onSubmit`/`onConfirmDelete`.
- `clearEtags()` on mutation success.
- The picture `refetchInterval` backoff for `PENDING` - owner's call: catalogue shows the
  placeholder until a refresh, no polling added for it. (Native still gets it for free anyway,
  since `useStands`/`useDevilFruits`/`useStages` bake that polling into the hook itself,
  unconditional on caller - see [[picture-events-sse]]. Not worth forking the hook over.)
- The `FAILED`-picture toast watcher (`useRef<Set<string>>`).
- `GET .../translations` - **admin-only, would 403**. The catalogue only ever reads the public,
  locale-resolved shape (see [[i18n-multi-language]]'s hybrid API contract).
- Stand's unfiltered `useStands()` for the create form's `evolvesFrom` picker - but the *filter's*
  own unfiltered roster (for `evolvesFrom` filter options) is still needed, same trap as
  [[admin-search-and-filters]] documents: deriving a filter's own option list from the filtered
  grid would make applying any filter narrow that very picker.

Query keys are untouched (`standKeys` etc.) - catalogue and admin share one cache entry per
locale, which is correct and desirable.

## Routing: authenticated, not public

Routes: `/catalog/stands`, `/catalog/devil-fruits`, `/catalog/stages`, under
`app/(app)/catalog/` with a pass-through `_layout.tsx` (`<Slot/>`, no extra guard - `(app)`'s own
layout already requires a session). Deliberately **not** outside the auth group: owner wants any
*logged-in* user, not a public unauthenticated page. Unlike `/admin/*`, there is no role check -
every session, `REGULAR` or `ADMIN`, gets the exact same read-only screen at `/catalog/*`; an
admin who wants to edit still goes to `/admin/*`.

## `DetailModal`: shared shell, per-feature body

`shared/components/presentational/detail-modal.tsx` is the "card in the foreground" chrome - same
`Modal` + dimmed-backdrop + centered `GlassPanel` recipe as `ConfirmSheet`/`*FormModal` (title, X
close, Esc-to-close on web, optional `footer`, scrollable body). `StandDetail`/`DevilFruitDetail`/
`StageDetail` (one per feature, next to the card) render everything the public DTO carries -
`description`, `skills` (as a vertical chip stack, `skills-field.tsx`'s pattern, not wrapping
pills - see [[admin-panel-crud-ux-fixes]] §"pills clip"), the stat grid for Stands - and reuse
`ImageLightbox` for the picture. None of `id`/`pictureStatus`/`translations` ever reach it.

Wired identically on both sides: the card's body (not its thumbnail) opens the modal via
`onOpenDetail`; in admin, the modal's footer gets an Edit button that closes the modal and calls
the same `onEdit` the card's own edit button uses. In the catalogue, the footer is omitted
entirely (`props.readOnly` gates it) - strictly see-only, as instructed.

## Things to watch

- `useStands`/`useDevilFruits`/`useStages`' native picture-poll runs even for the read-only
  containers (it's baked into the hook, not opt-in) - harmless, arguably a plus, but worth knowing
  if "no polling in catalogue" is ever taken literally on native.
- The card's body `Pressable` (opens detail) and the thumbnail's own `Pressable` (opens lightbox)
  are siblings, not nested - nesting would risk the outer one swallowing the thumb's tap. Confirmed
  both fire independently.
