---
title: "Per-power FX: curated Devil Fruit catalog (planned, 2026-09-03)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - planned
  - design
---

# Per-power FX: curated Devil Fruit catalog (planned, 2026-09-03)

**Status: planning only, nothing here is built.** Sibling to [[gameplay-power-fx-stand-catalog]] -
same deal, different power family. Read [[gameplay-power-fx]] first for the shape this feeds
(`PowerFx`, `cardEffect`/`avatarEffect`, `RevealFxMaxMs`, the `.web`/`.native` split, reduced-motion
rules). This note only adds *content* to that registry, not new mechanism. Closes the gap that
[[gameplay-power-fx-stand-catalog]]'s "Devil Fruits — not decided yet" section flagged.

## Sound: every entry here also gets one

**Owner decision (2026-09-03)**: same as the Stand catalog - every Devil Fruit below gets a sound
effect alongside its visual, asset TBD, provided by the owner when this is actually built. Unlike
Stands (which each get their own war-cry), the owner hasn't said these need to be a "cry" per se -
just "a sound" - so treat the sound *type* per entry as undecided too, not only the asset. `PowerFx`
still only needs the one optional `soundEffect` field either way (see [[gameplay-power-fx]]'s Seed
catalog section).

## The catalog

| Devil Fruit | Visual effect (owner's idea) | Sound |
|---|---|---|
| Gomu Gomu no Mi / Hito Hito no Mi, Model: Nika | Rubbery bounce effect, screen background fills with white clouds and the moon behind | Sound (TBD asset) |
| Mera Mera no Mi | Flame effect | Sound (TBD asset) |
| Suna Suna no Mi | Profile picture decomposes into sand | Sound (TBD asset) |
| Goro Goro no Mi | Electrocution effect | Sound (TBD asset) |
| Hie Hie no Mi | Screen freezes over | Sound (TBD asset) |
| Yami Yami no Mi | Darkness invades the screen | Sound (TBD asset) |
| Pika Pika no Mi | Profile picture turns to light and zig-zags across the screen | Sound (TBD asset) |
| Magu Magu no Mi | Lava bursts across the screen | Sound (TBD asset) |
| Mori Mori no Mi | Trees grow around the picture | Sound (TBD asset) |
| Ito Ito no Mi | Taut threads stretch across the screen | Sound (TBD asset) |
| Hana Hana no Mi | Two arms sprout from the sides | Sound (TBD asset) |
| Yomi Yomi no Mi | Soul leaves the body | Sound (TBD asset) |
| Nikyu Nikyu no Mi | A bubble comes out of the profile picture | Sound (TBD asset) |
| Ope Ope no Mi | "Room" effect expanding outward | Sound (TBD asset) |
| Gura Gura no Mi | Earthquake effect | Sound (TBD asset) |
| Zushi Zushi no Mi | Meteor effect | Sound (TBD asset) |
| Soru Soru no Mi | Souls spin around the profile picture | Sound (TBD asset) |
| Jiki Jiki no Mi | Electromagnetism with lightning bolts | Sound (TBD asset) |
| Toshi Toshi no Mi | Profile picture shrinks, then grows back | Sound (TBD asset) |
| Neko Neko no Mi, Model: Leopard | Claw marks across the screen | Sound (TBD asset) |
| **Every Ancient Zoan** (group entry, not per-fruit) | Jurassic effect with a dinosaur roar, screen background becomes a Cretaceous jungle | Sound (TBD asset) |
| Tori Tori no Mi, Model: ... (Phoenix) | Blue flame effect | Sound (TBD asset) |
| Hito Hito no Mi, Model: Daibutsu | Radiant/glowing effect around the profile picture | Sound (TBD asset) |
| Inu Inu no Mi, Model: Okuchi no Makami | Howl effect, "cloud scarf" wraps around the profile picture, screen atmosphere turns icy | Sound (TBD asset) |
| Hebi Hebi no Mi, Model: Yamata no Orochi | Profile picture duplicates into 8 | Sound (TBD asset) |
| Uo Uo no Mi, Model: Seiryu | Oriental dragon effect, cloudy background | Sound (TBD asset) |
| **Every other Mythical Zoan** (group entry, not per-fruit; everything above already named is excluded) | Mythological-themed aura across the screen, profile picture glows faintly | Sound (TBD asset) |

## Group entries: resolved via a Zoan-subtype fallback tier

**Decided with the owner (2026-09-03)**: the two group rows above ("every Ancient Zoan", "every
other Mythical Zoan") are a new `ZOAN_SUBTYPE_FX_FALLBACK` tier in [[gameplay-power-fx]]'s sketch,
resolved between `POWER_FX` (curated, per-fruit) and `FRUIT_TYPE_FX_FALLBACK.ZOAN` (regular Zoan) -
see that note's Shape sketch section for the exact resolution order. Rejected alternative: expanding
`POWER_FX` with one generated entry per Ancient/Mythical fruit at registry-build time - kept as a
real fallback tier instead, since it's a group rule, not a per-fruit curated choice.

Related: [[gameplay-power-fx]], [[gameplay-power-fx-stand-catalog]], [[game-match-assignment-frontend]],
[[gameplay-game-modes]], [[gameplay-domain-design]].
