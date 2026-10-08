---
title: "TODO: única fuente de verdad de lo pendiente"
tags:
  - project
  - jojo-onepiece-simulator
  - todo
---

# TODO (auditado contra el código el 2026-10-08)

**Regla**: todo lo pendiente vive AQUÍ. Las notas por tema describen *qué se hizo y por qué*; si una nota
menciona trabajo futuro, se añade una línea aquí con enlace a la nota y en la nota se deja solo un
`→ ver [[TODO]]`. Al terminar algo: se borra de aquí (el historial lo guardan la nota y git), no se
deja un "pendiente" huérfano en la nota. Ver [[zettelkasten-workflow]].

Formato: `- [ ] item — contexto breve ([[nota]])`.

## 1. Por construir (confirmado: no existe en el código)

- [ ] Efectos visuales/sonoros por poder en el sorteo: registro `power-fx.ts`, `soundEffect`, `extraHoldMs`; faltan además los assets de audio ([[gameplay-power-fx]], [[gameplay-power-fx-stand-catalog]], [[gameplay-power-fx-devil-fruit-catalog]])
- [ ] Inventario de jugador: adapter de `ports.IInventory`, `AbilitySource=INVENTORY` (hoy `ErrInventoryNotSupported`), fase 2 de personajes, ventana de selección, gachapon ([[gameplay-versus-inventory-characters]], [[characters-content-shipped-2026-09-10]])
- [ ] Lock distribuido en Redis antes de la ventana de selección ([[gameplay-versus-inventory-characters]])
- [ ] Panel admin de usuarios `app/(app)/admin/users.tsx` (las rutas backend existen) ([[user-profile-feature]])
- [ ] Chip MVP + ronda de voto post-partida ([[game-victory-defeat-cinematic-2026-09-14]])
- [ ] Chip de recompensa "+XXX" — moneda/XP sin decidir ([[game-victory-defeat-cinematic-2026-09-14]])
- [ ] Avatares en la cinematic: `avatarThumb` en `ParticipantOutcomeResponse` ([[game-victory-defeat-cinematic-2026-09-14]])
- [ ] `avatarCard`/LQIP en `GameParticipantResponse`/`GameStageResponse` ([[entrega-imagenes-red-lenta-2026-09-07]])
- [ ] `Seq` en `services.GameEvent` para detectar drops del hub ([[gameplay-application-layer]], [[game-realtime-transport]])
- [ ] Progresión por runs entre rondas de Gauntlet (`GauntletMode.afterRound` es no-op) ([[gameplay-game-modes]])
- [ ] Tiebreak por LLM en lugar de moneda (intención futura) ([[gameplay-game-modes]])
- [ ] Lista de ban/kick persistente ([[game-invite-links-2026-09-17]])
- [ ] Previews enriquecidos (`og:`) al compartir invitación; universal/app link ([[game-invite-links-2026-09-17]])
- [ ] Cuenta atrás numérica del periodo de gracia por desconexión (falta campo de duración en el wire) ([[game-disconnect-grace-2026-09-15]])
- [ ] Guard "una partida activa por usuario" (`ActiveGameForUser` elige el primero) ([[game-disconnect-grace-2026-09-15]])
- [ ] `poolShortfalls` consciente de familias (requiere `evolvesFrom` en el DTO del catálogo) ([[gameplay-power-effects]])
- [ ] Aviso al editar en admin un poder cuyo nombre rompe una regla de efectos (hoy solo log al arrancar) ([[gameplay-power-effects]])
- [ ] Servicio `frontend-test` en `docker-compose.test.yml` ([[norma-verificacion-docker]])
- [ ] Hoja de instrucciones iOS "Compartir → Añadir a inicio" ([[pwa-movil-2026-10-08]])
- [ ] Regenerar swagger `apps/backend/docs` (y `-race` en CI para el test de concurrencia de game service) ([[pwa-movil-2026-10-08]], [[gameplay-application-layer]])

- [ ] **BUG** Drag-to-move no funciona (verificado solo en escritorio con eventos sintéticos; reproducir en móvil y escritorio real) ([[game-lobby-todo]])
- [ ] **BUG** Tooltips: se quedan flotando (p. ej. el de borrar un power); el scroll-hide no basta ([[norma-tooltips-y-ayuda-contextual]])
- [ ] **BUG** Config del lobby: los cambios no llegan a los demás usuarios; falta emitir evento WS de config actualizada y consumirlo en el cliente ([[game-lobby-frontend]], [[game-realtime-transport]])
- [ ] Migración que corrige el typo de fuente de Born This Way (Kei Nijimura) ([[catalog-seed-p7-p8-fruits-2026-10-08]])
- [ ] Dejar todo soportado para nativo (iOS/Android) aunque de momento solo se publique la PWA: no introducir APIs solo-web sin rama nativa ([[frontend-stack]])

## 2. Decisiones abiertas (necesitan al dueño)

- [ ] Pesos del evaluador de bots: ¿son definitivos los de 2026-10-04? ([[gameplay-domain-design]])
- [ ] Gachapon: duplicados, pity, `rarity` de Character sin cablear ([[gameplay-versus-inventory-characters]])
- [ ] Reglas Pika Pika → Hamon y grupo Battle IQ #11/#12 (propuestas no tomadas) ([[gameplay-power-effects]])
- [ ] Sección de ban de personajes en la config del lobby ([[gameplay-versus-inventory-characters]])

## 3. Por verificar a mano (no comprobable desde el código)

Pendiente de respuesta del dueño; se mueve a sección 1 si falla o se borra si OK.

- [ ] Cinematics: mute y reduced-motion sin probar (Versus 2 cuentas / Gauntlet caer y teclado ya OK salvo lo anterior)
- [ ] Pasada teclado-only completa (1-9, S, roving tabindex): la hace Claude con claude-in-chrome ([[game-vote-buttons-2026-08-26]])
- [ ] Re-subir a mano las imágenes borrosas anteriores al fix de `PICTURE_CARD_DIMENSION` (el backfill no re-transcodifica) ([[media-proxy-content-addressed]], [[entrega-imagenes-red-lenta-2026-09-07]])
- [ ] Rollback real de `cd.yml` (migración rota a propósito) ([[cicd-deployment]])
- [ ] Imágenes de los 29 Stands de Part 4 subidas en prod ([[catalog-seed-part4-stands]])
- [ ] Sorteo ronda 2 Versus con 2+ humanos; sospechosos `shouldReveal`/`assignmentSeq` ([[playtest-fixes-2026-10-03]])
- [ ] Auto-avance de 6 s + Skip en partida real ([[game-round-result-2026-08-28]])
- [ ] Tusk evolucionando en vivo con datos de prod (tras desplegar 00024) ([[gameplay-power-effects]])
- [ ] Foco con lector de pantalla en "Cargar más" en build nativo ([[catalogue-pagination]])

## 4. Limitaciones conocidas y aceptadas (no son tareas, no tocar sin decisión)

- Un solo backend: `GameEventHub` y `gameLocks` son de proceso; escalar exige pub/sub y lock en Redis ([[ADR]], [[game-lobby-persistence]])
- Access token sin revocación (15 min); sin `REDIS_URL` un reinicio desloguea a todos ([[session-token-storage-2026-09-05]])
- Duplicar pestaña copia `sessionStorage` y mata la familia de refresh token (dev) ([[dev-auth-bypass-2026-09-25]])
- Poderes se emparejan por NOMBRE; renombrar desactiva la regla en silencio ([[gameplay-power-effects]])
- `GoogleVerifier.Verify` no testeable offline ([[auth-hardening-2026-09-02]])
- Tooltips nativos sin centrar/clamp; polling de imágenes en nativo; `tamagui-web.css` 0 bytes ([[norma-tooltips-y-ayuda-contextual]], [[picture-events-sse]], [[frontend-responsive-frutiger-aero]])
- Test flaky bajo Docker con muchos workers (mitigado con `--maxWorkers=2`) ([[norma-verificacion-docker]])
