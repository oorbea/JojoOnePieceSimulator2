-- +goose Up
-- Seeds JoJo's Bizarre Adventure Part 5 (Vento Aureo/Golden Wind) Stands.
-- Hand-authored new content (stats verified against jojowiki.com), not a
-- catalogsync prod-dump migration like 00017/00019 - see
-- ObsidianVault/catalog-seed-part5-stands.md. Runs in every environment
-- including prod (no ENVSUB guard): images are added later by an admin
-- through the normal picture-upload flow, so every row ships with the
-- powers.picture* defaults (picture_status = NONE).
--
-- Idempotent against a name an admin may have already created in prod:
-- every powers insert is ON CONFLICT DO NOTHING (id or name clash both
-- skip), and every dependent insert is gated on the powers row actually
-- existing with that id, so a skipped Stand never leaves an orphaned
-- stands/power_translations row behind.
--
-- Also fixes house-style violations in the existing Chariot Requiem row
-- (from 00017/00019): missing/wrong header line, paragraph-length skills,
-- ca-ES typos. Guarded the same way 00019 guards admin-edit-wins updates:
-- only touches the row if it still has the pre-fix updated_at timestamp.

-- Gold Experience (Giorno Giovanna)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7c80-900f-5802c64c17e6', 'STAND', 'Gold Experience', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7c80-900f-5802c64c17e6', 'C', 'A', 'C', 'D', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c80-900f-5802c64c17e6')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c80-900f-5802c64c17e6', 'es-ES', 'Portador: Giorno Giovanna

Es un Stand de combate cercano cuya habilidad exclusiva es infundir vida propia a cualquier objeto o ser sin vida que toca, transformándolo en un organismo vivo funcional. Su velocidad y potencial son sobresalientes, aunque su fuerza bruta es modesta comparada con otros Stands de su categoría.', ARRAY['Creación de Vida: Convierte cualquier objeto u organismo muerto que toca en un ser vivo funcional.','Golpe Vital: Sus puñetazos inyectan vida orgánica descontrolada en el rival, causando mutaciones internas letales.','Sentido de la Verdad: Percibe mentiras y siente lo que sienten las plantas y criaturas creadas por su poder.','Regeneración Acelerada: Puede curar heridas propias transformando la lesión en tejido vivo nuevo.','Reflejos de Combate: Su velocidad y potencial le permiten adaptarse sobre la marcha a rivales impredecibles.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c80-900f-5802c64c17e6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c80-900f-5802c64c17e6', 'en-GB', 'User: Giorno Giovanna

This is a close-range Stand whose signature power is infusing its own life into any object or lifeless being it touches, turning it into a fully functional living organism. Its speed and potential are outstanding, though its raw power is modest next to other Stands in its class.', ARRAY['Life Creation: Turns any object or dead organism it touches into a functioning living being.','Vital Strike: Its punches inject uncontrolled organic life into the target, causing lethal internal mutations.','Sense of Truth: Detects lies and can feel what plants and creatures made by its power feel.','Accelerated Healing: Can heal its own wounds by turning the injury into fresh living tissue.','Combat Reflexes: Its speed and potential let it adapt on the fly to unpredictable opponents.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c80-900f-5802c64c17e6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c80-900f-5802c64c17e6', 'ca-ES', 'Portador: Giorno Giovanna

És un Stand de combat proper l''habilitat exclusiva del qual és infondre vida pròpia a qualsevol objecte o ésser sense vida que toca, transformant-lo en un organisme viu funcional. La seva velocitat i potencial són excel·lents, tot i que la seva força bruta és modesta comparada amb altres Stands de la seva categoria.', ARRAY['Creació de Vida: Converteix qualsevol objecte o organisme mort que toca en un ésser viu funcional.','Cop Vital: Els seus cops de puny injecten vida orgànica descontrolada al rival, causant mutacions internes letals.','Sentit de la Veritat: Percep mentides i sent el que senten les plantes i criatures creades pel seu poder.','Regeneració Accelerada: Pot curar ferides pròpies transformant la lesió en teixit viu nou.','Reflexos de Combat: La seva velocitat i potencial li permeten adaptar-se sobre la marxa a rivals imprevisibles.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c80-900f-5802c64c17e6')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Gold Experience: Requiem (Giorno Giovanna)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7198-a825-c53ec83cb393', 'STAND', 'Gold Experience: Requiem', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7198-a825-c53ec83cb393', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7198-a825-c53ec83cb393')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7198-a825-c53ec83cb393', 'es-ES', 'Portador: Giorno Giovanna

Es la forma evolucionada de Gold Experience, alcanzada tras "matar" al propio Stand original. Sus estadísticas escapan a cualquier escala de medición conocida. Su habilidad definitiva anula la causalidad: borra el concepto de la muerte para su objetivo, condenándolo a repetir para siempre el instante previo a morir, y deshace cualquier "hetappu" (reversión temporal) que se le aplique.', ARRAY['Retorno a Cero: Anula las acciones y la voluntad de quien se le opone antes de que se hagan realidad.','Negación de la Muerte: Borra el concepto de muerte del objetivo, atrapándolo en un bucle eterno de morir.','Anulación de Hetappu: Deshace cualquier reversión temporal o "vuelta atrás" usada en su contra.','Percepción Absoluta: Conoce de antemano cualquier ataque o plan dirigido contra Giorno.','Autonomía Total: Actúa por voluntad propia incluso sin la intervención consciente de su portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7198-a825-c53ec83cb393')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7198-a825-c53ec83cb393', 'en-GB', 'User: Giorno Giovanna

This is Gold Experience''s evolved form, reached after "killing" the original Stand. Its stats escape any known measuring scale. Its ultimate power negates causality: it erases the concept of death for its target, condemning them to endlessly repeat the moment before dying, and undoes any "hetappu" (time reversal) used against it.', ARRAY['Return to Zero: Nullifies the actions and will of anyone opposing it before they become reality.','Denial of Death: Erases the target''s concept of death, trapping them in an endless loop of dying.','Hetappu Negation: Undoes any time reversal or "rewind" used against it.','Absolute Perception: Foreknows any attack or plan aimed at Giorno.','Total Autonomy: Acts of its own will even without its user''s conscious intervention.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7198-a825-c53ec83cb393')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7198-a825-c53ec83cb393', 'ca-ES', 'Portador: Giorno Giovanna

És la forma evolucionada de Gold Experience, assolida després de "matar" el mateix Stand original. Les seves estadístiques escapen a qualsevol escala de mesura coneguda. La seva habilitat definitiva anul·la la causalitat: esborra el concepte de la mort per al seu objectiu, condemnant-lo a repetir per sempre l''instant previ a morir, i desfà qualsevol "hetappu" (reversió temporal) que se li apliqui.', ARRAY['Retorn a Zero: Anul·la les accions i la voluntat de qui se li oposa abans que es facin realitat.','Negació de la Mort: Esborra el concepte de mort de l''objectiu, atrapant-lo en un bucle etern de morir.','Anul·lació d''Hetappu: Desfà qualsevol reversió temporal o "marxa enrere" usada en contra seva.','Percepció Absoluta: Coneix per endavant qualsevol atac o pla dirigit contra en Giorno.','Autonomia Total: Actua per voluntat pròpia fins i tot sense la intervenció conscient del seu portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7198-a825-c53ec83cb393')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Sticky Fingers (Bruno Bucciarati)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7317-98c9-8361101ec2a6', 'STAND', 'Sticky Fingers', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7317-98c9-8361101ec2a6', 'A', 'A', 'C', 'D', 'C', 'D' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7317-98c9-8361101ec2a6')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7317-98c9-8361101ec2a6', 'es-ES', 'Portador: Bruno Bucciarati

Es un Stand de combate cercano de gran fuerza y velocidad cuya habilidad exclusiva es crear cremalleras en cualquier superficie que toca, incluyendo cuerpos y objetos, permitiendo abrirlos, separarlos o unirlos a voluntad. Su alcance es corto y su resistencia limitada, pero su potencia de golpe es devastadora.', ARRAY['Creación de Cremalleras: Genera una cremallera funcional en cualquier superficie con solo tocarla.','División de Objetos: Puede separar completamente cualquier cosa a lo largo de la cremallera creada.','Camuflaje Espacial: Se esconde a sí mismo o a otros dentro de espacios creados con sus cremalleras.','Fuerza de Combate: Golpea con una potencia brutal capaz de perforar cuerpos con facilidad.','Trampas Tácticas: Usa las cremalleras para inmovilizar extremidades o crear vías de escape improvisadas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7317-98c9-8361101ec2a6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7317-98c9-8361101ec2a6', 'en-GB', 'User: Bruno Bucciarati

This is a close-range Stand of great strength and speed whose signature power is creating zippers on any surface it touches, including bodies and objects, letting it open, separate or join them at will. Its range is short and its stamina limited, but its striking power is devastating.', ARRAY['Zipper Creation: Generates a functional zipper on any surface with a single touch.','Object Splitting: Can fully separate anything along a zipper it has created.','Spatial Camouflage: Hides itself or others inside spaces made with its zippers.','Combat Strength: Strikes with brutal power capable of piercing bodies with ease.','Tactical Traps: Uses zippers to immobilise limbs or create improvised escape routes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7317-98c9-8361101ec2a6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7317-98c9-8361101ec2a6', 'ca-ES', 'Portador: Bruno Bucciarati

És un Stand de combat proper de gran força i velocitat l''habilitat exclusiva del qual és crear cremalleres a qualsevol superfície que toca, incloent-hi cossos i objectes, cosa que li permet obrir-los, separar-los o unir-los a voluntat. El seu abast és curt i la seva resistència limitada, però la seva potència de cop és devastadora.', ARRAY['Creació de Cremalleres: Genera una cremallera funcional a qualsevol superfície amb només tocar-la.','Divisió d''Objectes: Pot separar completament qualsevol cosa al llarg de la cremallera creada.','Camuflatge Espacial: S''amaga a si mateix o a altres dins d''espais creats amb les seves cremalleres.','Força de Combat: Colpeja amb una potència brutal capaç de perforar cossos amb facilitat.','Trampes Tàctiques: Fa servir les cremalleres per immobilitzar extremitats o crear vies de fugida improvisades.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7317-98c9-8361101ec2a6')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Moody Blues (Leone Abbacchio)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7987-8c9a-b1925f1a790b', 'STAND', 'Moody Blues', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7987-8c9a-b1925f1a790b', 'C', 'C', 'A', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7987-8c9a-b1925f1a790b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7987-8c9a-b1925f1a790b', 'es-ES', 'Portador: Leone Abbacchio

Es un Stand investigador cuya habilidad exclusiva es materializar una réplica exacta de una persona u objeto para reproducir un momento pasado que Abbacchio haya presenciado, incluyendo diálogos y acciones. Su alcance se dispara durante estas repeticiones, aunque su fuerza de combate directa es discreta.', ARRAY['Repetición del Pasado: Recrea físicamente una escena presenciada, reproduciendo palabras y gestos exactos.','Investigación a Distancia: Envía réplicas a explorar lugares lejanos mientras Abbacchio permanece a salvo.','Interrogatorio Fantasma: Las réplicas pueden ser interrogadas por Abbacchio para extraer información oculta.','Resistencia Prolongada: Mantiene la repetición activa durante largos periodos sin fatigarse.','Camuflaje Perceptivo: Las réplicas suelen pasar desapercibidas para quienes no conocen su naturaleza.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7987-8c9a-b1925f1a790b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7987-8c9a-b1925f1a790b', 'en-GB', 'User: Leone Abbacchio

This is an investigative Stand whose signature power is materialising an exact replica of a person or object to replay a past moment Abbacchio witnessed, including dialogue and actions. Its range spikes during these replays, though its direct combat strength is modest.', ARRAY['Past Replay: Physically recreates a witnessed scene, reproducing exact words and gestures.','Remote Investigation: Sends replicas to scout distant locations while Abbacchio stays safe.','Ghost Interrogation: The replicas can be questioned by Abbacchio to extract hidden information.','Extended Stamina: Keeps the replay active for long stretches without tiring.','Perceptual Camouflage: The replicas usually go unnoticed by those unaware of their nature.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7987-8c9a-b1925f1a790b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7987-8c9a-b1925f1a790b', 'ca-ES', 'Portador: Leone Abbacchio

És un Stand investigador l''habilitat exclusiva del qual és materialitzar una rèplica exacta d''una persona o objecte per reproduir un moment passat que l''Abbacchio hagi presenciat, incloent-hi diàlegs i accions. El seu abast s''enfila durant aquestes repeticions, tot i que la seva força de combat directa és discreta.', ARRAY['Repetició del Passat: Recrea físicament una escena presenciada, reproduint paraules i gestos exactes.','Investigació a Distància: Envia rèpliques a explorar llocs llunyans mentre l''Abbacchio resta a resguard.','Interrogatori Fantasma: Les rèpliques poden ser interrogades per l''Abbacchio per extreure informació oculta.','Resistència Perllongada: Manté la repetició activa durant llargs períodes sense fatigar-se.','Camuflatge Perceptiu: Les rèpliques solen passar desapercebudes per a qui no en coneix la naturalesa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7987-8c9a-b1925f1a790b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Sex Pistols (Guido Mista)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7d2c-ad6e-2779521a0c46', 'STAND', 'Sex Pistols', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7d2c-ad6e-2779521a0c46', 'E', 'C', 'B', 'A', 'A', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d2c-ad6e-2779521a0c46')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d2c-ad6e-2779521a0c46', 'es-ES', 'Portador: Guido Mista

Es un Stand compuesto por seis pequeños seres humanoides que, por sí solos, apenas causan daño, pero pueden montarse sobre las balas disparadas por Mista para guiarlas, acelerarlas, detenerlas en el aire o hacerlas rebotar con precisión milimétrica. Su alcance depende por completo de la distancia que la bala pueda recorrer.', ARRAY['Guiado de Balas: Monta cada uno de los seis miembros sobre una bala para dirigir su trayectoria a voluntad.','Frenado en el Aire: Puede detener una bala en pleno vuelo o hacerla rebotar en cualquier dirección.','Aceleración de Disparo: Aumenta la velocidad de una bala ya disparada para sorprender al objetivo.','Trabajo en Equipo: Los seis miembros colaboran para ejecutar maniobras imposibles para una bala normal.','Resistencia Física: Cada miembro aguanta golpes y caídas considerables sin desactivarse.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d2c-ad6e-2779521a0c46')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d2c-ad6e-2779521a0c46', 'en-GB', 'User: Guido Mista

This is a Stand made of six tiny humanoid beings that barely do damage on their own, but can ride the bullets Mista fires to steer, accelerate, stop mid-air or bounce them with pinpoint precision. Its range depends entirely on how far a bullet can travel.', ARRAY['Bullet Guidance: Each of the six members rides a bullet to steer its path at will.','Mid-Air Braking: Can stop a bullet in flight or bounce it off in any direction.','Shot Acceleration: Speeds up an already-fired bullet to catch the target off guard.','Teamwork: The six members cooperate to pull off manoeuvres no normal bullet could manage.','Physical Resilience: Each member withstands considerable hits and falls without shutting down.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d2c-ad6e-2779521a0c46')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d2c-ad6e-2779521a0c46', 'ca-ES', 'Portador: Guido Mista

És un Stand format per sis petits éssers humanoides que, tot sols, tot just fan mal, però poden pujar sobre les bales que dispara en Mista per guiar-les, accelerar-les, aturar-les a l''aire o fer-les rebotar amb precisió mil·limètrica. El seu abast depèn completament de la distància que la bala pugui recórrer.', ARRAY['Guiatge de Bales: Cada un dels sis membres puja sobre una bala per dirigir-ne la trajectòria a voluntat.','Frenada a l''Aire: Pot aturar una bala en ple vol o fer-la rebotar en qualsevol direcció.','Acceleració de Tret: Augmenta la velocitat d''una bala ja disparada per sorprendre l''objectiu.','Treball en Equip: Els sis membres col·laboren per executar maniobres impossibles per a una bala normal.','Resistència Física: Cada membre aguanta cops i caigudes considerables sense desactivar-se.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d2c-ad6e-2779521a0c46')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Aerosmith (Narancia Ghirga)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7ecc-9964-e2b28b7c0432', 'STAND', 'Aerosmith', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7ecc-9964-e2b28b7c0432', 'B', 'B', 'B', 'C', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ecc-9964-e2b28b7c0432')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ecc-9964-e2b28b7c0432', 'es-ES', 'Portador: Narancia Ghirga

Es un Stand no humanoide con forma de pequeño avión teledirigido que dispara ametralladoras y bombas de gran alcance para labores de reconocimiento y ataque a distancia. Su potencia y alcance son notables, pero su escasa precisión obliga a Narancia a compensarla desatando toda su munición sin contemplaciones.', ARRAY['Vuelo Autónomo: Sobrevuela grandes distancias para explorar el terreno y localizar objetivos ocultos.','Ametralladoras Gemelas: Dispara ráfagas continuas capaces de acribillar un área amplia.','Bombas de Racimo: Lanza pequeñas bombas explosivas que detonan al impacto o al detectar movimiento.','Radar de Sonido: Detecta a los enemigos por el eco de sus propios disparos y explosiones.','Ataque Sin Cuartel: Compensa su falta de puntería descargando todo su arsenal de una sola vez.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ecc-9964-e2b28b7c0432')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ecc-9964-e2b28b7c0432', 'en-GB', 'User: Narancia Ghirga

This is a non-humanoid Stand shaped like a small remote-controlled plane that fires machine guns and long-range bombs for reconnaissance and ranged attacks. Its power and range are notable, but its poor precision forces Narancia to compensate by unleashing all its firepower without restraint.', ARRAY['Autonomous Flight: Flies over great distances to scout terrain and locate hidden targets.','Twin Machine Guns: Fires continuous bursts capable of riddling a wide area.','Cluster Bombs: Drops small explosive bombs that detonate on impact or on detecting movement.','Sound Radar: Detects enemies through the echo of its own gunfire and explosions.','All-Out Assault: Compensates for poor aim by discharging its whole arsenal at once.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ecc-9964-e2b28b7c0432')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ecc-9964-e2b28b7c0432', 'ca-ES', 'Portador: Narancia Ghirga

És un Stand no humanoide amb forma de petit avió teledirigit que dispara metralladores i bombes de llarg abast per a tasques de reconeixement i atac a distància. La seva potència i abast són notables, però la seva escassa precisió obliga en Narancia a compensar-ho desfermant tota la seva munició sense contemplacions.', ARRAY['Vol Autònom: Sobrevola grans distàncies per explorar el terreny i localitzar objectius amagats.','Metralladores Bessones: Dispara ràfegues continues capaces de foradar una àrea àmplia.','Bombes de Dispersió: Llança petites bombes explosives que detonen a l''impacte o en detectar moviment.','Radar de So: Detecta els enemics per l''eco dels seus propis trets i explosions.','Atac Sense Contemplacions: Compensa la seva manca de punteria descarregant tot el seu arsenal d''un sol cop.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ecc-9964-e2b28b7c0432')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Purple Haze (Pannacotta Fugo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7bb1-879d-68a1f7639d35', 'STAND', 'Purple Haze', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7bb1-879d-68a1f7639d35', 'A', 'B', 'C', 'E', 'E', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7bb1-879d-68a1f7639d35')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7bb1-879d-68a1f7639d35', 'es-ES', 'Portador: Pannacotta Fugo

Es un Stand de combate cercano con forma humanoide feroz, cuya arma principal son unas cápsulas alojadas en sus manos que contienen un virus devorador de carne. Combina una fuerza brutal capaz de romper muros con los puños con una letal capacidad biológica, aunque su alcance es limitado y el virus se debilita con la luz.', ARRAY['Fuerza Devastadora: Golpea con tal potencia que puede derribar un muro de ladrillo de un solo puñetazo.','Virus Devorador de Carne: Libera un patógeno mortal que descompone la carne en segundos, afectando también a otros Stands.','Cápsulas Proyectil: Puede disparar las cápsulas del virus a distancia además de detonarlas por contacto directo.','Vulnerabilidad a la Luz: El virus se neutraliza por completo al exponerse a la luz solar, lo que limita su uso táctico.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7bb1-879d-68a1f7639d35')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7bb1-879d-68a1f7639d35', 'en-GB', 'User: Pannacotta Fugo

This is a close-range Stand with a fierce humanoid form whose main weapon is a set of capsules on its hands holding a flesh-eating virus. It pairs brutal strength capable of shattering brick walls with a lethal biological attack, though its range is limited and the virus is weakened by light.', ARRAY['Devastating Strength: Punches hard enough to break through a brick wall in a single blow.','Flesh-Eating Virus: Releases a deadly pathogen that decomposes flesh within seconds, affecting other Stands too.','Capsule Projectiles: Can fire the virus capsules at range in addition to detonating them on contact.','Weakness to Light: The virus is fully neutralised by sunlight, limiting its tactical use.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7bb1-879d-68a1f7639d35')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7bb1-879d-68a1f7639d35', 'ca-ES', 'Portador: Pannacotta Fugo

És un Stand de combat proper amb forma humanoide ferotge, l''arma principal del qual són unes càpsules allotjades a les mans que contenen un virus devorador de carn. Combina una força brutal capaç de trencar murs d''un sol cop de puny amb una capacitat biològica letal, tot i que el seu abast és limitat i el virus es debilita amb la llum.', ARRAY['Força Devastadora: Colpeja amb tanta potència que pot esfondrar un mur de maó d''un sol cop de puny.','Virus Devorador de Carn: Allibera un patogen mortal que descompon la carn en segons, i afecta també altres Stands.','Càpsules Projectil: Pot disparar les càpsules del virus a distància, a més de detonar-les per contacte directe.','Vulnerabilitat a la Llum: El virus queda neutralitzat del tot en exposar-se a la llum solar, cosa que en limita l''ús tàctic.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7bb1-879d-68a1f7639d35')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Spice Girl (Trish Una)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7dd5-97c6-4e5797545490', 'STAND', 'Spice Girl', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7dd5-97c6-4e5797545490', 'A', 'A', 'C', 'B', 'D', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7dd5-97c6-4e5797545490')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7dd5-97c6-4e5797545490', 'es-ES', 'Portadora: Trish Una

Es un Stand humanoide de combate cercano con conciencia propia, capaz de ablandar cualquier material con su golpe hasta convertirlo en algo elástico y gomoso. Sus impactos son extremadamente potentes y precisos a corta distancia, y el efecto puede mantenerse incluso tras romper el contacto.', ARRAY['Ablandamiento: Convierte cualquier objeto que golpea en un material elástico y flexible, incluso metal o piedra.','Resiliencia Reforzada: Lo ablandado se vuelve casi indestructible, absorbiendo golpes y proyectiles sin romperse.','Retorno Elástico: El material ablandado conserva energía cinética y puede rebotar con fuerza al recuperar su forma.','Control Selectivo: Puede activar o desactivar el ablandamiento a voluntad, incluso a distancia del objetivo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7dd5-97c6-4e5797545490')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7dd5-97c6-4e5797545490', 'en-GB', 'User: Trish Una

This is a close-range humanoid Stand with a will of its own, able to soften any material it strikes into something elastic and rubbery. Its blows are extremely powerful and precise at short range, and the effect can persist even after contact breaks.', ARRAY['Softening: Turns anything it hits into an elastic, pliable material, even metal or stone.','Reinforced Resilience: Softened matter becomes nearly indestructible, absorbing blows and projectiles without breaking.','Elastic Rebound: Softened material keeps kinetic energy and can snap back violently as it returns to shape.','Selective Control: Can turn the softening effect on or off at will, even from a distance.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7dd5-97c6-4e5797545490')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7dd5-97c6-4e5797545490', 'ca-ES', 'Portadora: Trish Una

És un Stand humanoide de combat proper amb consciència pròpia, capaç d''estovar qualsevol material amb el seu cop fins a convertir-lo en una cosa elàstica i gomosa. Els seus impactes són extremament potents i precisos a curta distància, i l''efecte es pot mantenir fins i tot després de trencar el contacte.', ARRAY['Estovament: Converteix qualsevol objecte que colpeja en un material elàstic i flexible, fins i tot metall o pedra.','Resiliència Reforçada: El que ha estat estovat es torna gairebé indestructible, i absorbeix cops i projectils sense trencar-se.','Retorn Elàstic: El material estovat conserva energia cinètica i pot rebotar amb força en recuperar la seva forma.','Control Selectiu: Pot activar o desactivar l''estovament a voluntat, fins i tot a distància de l''objectiu.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7dd5-97c6-4e5797545490')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Mr. President (Coco Jumbo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7ad7-a7ed-9be5b1e4e682', 'STAND', 'Mr. President', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682', 'E', 'E', 'E', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682', 'es-ES', 'Portador: Coco Jumbo

Es un Stand sin capacidad de combate cuyo portador, la tortuga Coco Jumbo, oculta en su interior una habitación secreta accesible mediante una llave especial. El espacio funciona como refugio móvil, con muebles, televisión y nevera, y ofrece cierta protección frente a amenazas externas.', ARRAY['Refugio Oculto: Genera una habitación interior accesible al insertar la llave en el caparazón de Coco Jumbo.','Acceso Controlado: Solo se puede entrar o salir con la llave; retirarla expulsa a quien esté dentro.','Visión Distorsionada: La joya transparente de la llave permite ver el exterior o el interior, aunque con distorsión de perspectiva.','Sin Capacidad de Combate: Carece de cualquier habilidad ofensiva, aunque algunos poderes externos pueden penetrar el refugio.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682', 'en-GB', 'User: Coco Jumbo

This is a Stand with no combat ability whose user, the turtle Coco Jumbo, hides a secret room inside itself, accessible via a special key. The space acts as a mobile hideout, complete with furniture, a television and a fridge, offering some protection from outside threats.', ARRAY['Hidden Refuge: Opens an interior room accessible by inserting the key into Coco Jumbo''s shell.','Controlled Access: Only the key can let someone in or out; pulling it out ejects anyone inside.','Distorted Viewing: The key''s transparent gem lets its holder see in or out, though with perspective distortion.','No Combat Ability: Has no offensive power at all, and some outside Stand effects can still reach inside.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682', 'ca-ES', 'Portador: Coco Jumbo

És un Stand sense capacitat de combat el portador del qual, la tortuga Coco Jumbo, amaga al seu interior una habitació secreta accessible mitjançant una clau especial. L''espai funciona com un refugi mòbil, amb mobles, televisor i nevera, i ofereix certa protecció davant d''amenaces externes.', ARRAY['Refugi Amagat: Genera una habitació interior accessible en inserir la clau a la closca de Coco Jumbo.','Accés Controlat: Només es pot entrar o sortir amb la clau; retirar-la expulsa qui hi hagi a dins.','Visió Distorsionada: La joia transparent de la clau permet veure l''exterior o l''interior, tot i que amb distorsió de perspectiva.','Sense Capacitat de Combat: No té cap habilitat ofensiva, tot i que alguns poders externs poden penetrar el refugi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- King Crimson (Diavolo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-772d-b39f-6d66d8b6a1bc', 'STAND', 'King Crimson', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-772d-b39f-6d66d8b6a1bc', 'A', 'A', 'E', 'E', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-772d-b39f-6d66d8b6a1bc')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-772d-b39f-6d66d8b6a1bc', 'es-ES', 'Portador: Diavolo (también usado por Vinegar Doppio con su permiso)

Es un Stand humanoide de fuerza y velocidad devastadoras cuya habilidad definitiva es borrar un fragmento de tiempo de hasta diez segundos, durante el cual solo Diavolo permanece consciente y puede reposicionarse a su favor. Además, mediante su sub-habilidad Epitaph, puede vislumbrar el resultado final de los sucesos con varios segundos de antelación.', ARRAY['Borrado de Tiempo: Elimina de la existencia un intervalo de hasta diez segundos, durante el cual nadie más percibe ni recuerda lo ocurrido.','Reposicionamiento Ventajoso: Aprovecha el tiempo borrado para moverse o prepararse, aunque no puede atacar durante ese lapso.','Fuerza y Velocidad Superiores: Golpea y se mueve con una potencia y rapidez devastadoras en combate cercano.','Epitaph (Precognición): Proyecta el resultado final de los próximos segundos como una imagen, sin mostrar los pasos intermedios.','Uso Compartido: Vinegar Doppio puede invocar el Stand con el permiso de Diavolo, incluyendo su sub-habilidad Epitaph.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-772d-b39f-6d66d8b6a1bc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-772d-b39f-6d66d8b6a1bc', 'en-GB', 'User: Diavolo (also used by Vinegar Doppio with his permission)

This is a humanoid Stand of devastating strength and speed whose ultimate power is erasing a stretch of time of up to ten seconds, during which only Diavolo stays conscious and can reposition himself to his advantage. Through its sub-ability Epitaph, it can also glimpse the final outcome of events several seconds in advance.', ARRAY['Time Erasure: Wipes out an interval of up to ten seconds from existence, during which no one else perceives or remembers what happened.','Advantageous Repositioning: Uses the erased time to move or prepare, though it cannot attack during that span.','Superior Strength and Speed: Strikes and moves with devastating power and speed in close combat.','Epitaph (Precognition): Projects the final outcome of the next few seconds as an image, without showing the steps in between.','Shared Use: Vinegar Doppio can summon the Stand with Diavolo''s permission, including its Epitaph sub-ability.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-772d-b39f-6d66d8b6a1bc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-772d-b39f-6d66d8b6a1bc', 'ca-ES', 'Portador: Diavolo (també usat per en Vinegar Doppio amb el seu permís)

És un Stand humanoide de força i velocitat devastadores l''habilitat definitiva del qual és esborrar un fragment de temps de fins a deu segons, durant el qual només en Diavolo roman conscient i es pot reposicionar al seu favor. A més, mitjançant la seva subhabilitat Epitaph, pot entreveure el resultat final dels esdeveniments amb uns segons d''antelació.', ARRAY['Esborrament del Temps: Elimina de l''existència un interval de fins a deu segons, durant el qual ningú més no percep ni recorda el que ha passat.','Reposicionament Avantatjós: Aprofita el temps esborrat per moure''s o preparar-se, tot i que no pot atacar durant aquest lapse.','Força i Velocitat Superiors: Colpeja i es mou amb una potència i rapidesa devastadores en combat proper.','Epitaph (Precognició): Projecta el resultat final dels propers segons com una imatge, sense mostrar els passos intermedis.','Ús Compartit: En Vinegar Doppio pot invocar el Stand amb el permís d''en Diavolo, incloent-hi la seva subhabilitat Epitaph.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-772d-b39f-6d66d8b6a1bc')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Black Sabbath (Polpo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7082-a957-9bac4ad82f49', 'STAND', 'Black Sabbath', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7082-a957-9bac4ad82f49', 'E', 'A', 'A', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7082-a957-9bac4ad82f49')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7082-a957-9bac4ad82f49', 'es-ES', 'Portador: Polpo

Es un Stand que habita en las sombras, capaz de moverse a gran velocidad de una sombra a otra y de arrastrar a sus víctimas hacia la oscuridad para arrancarles el alma. Actúa de forma completamente autónoma sin voluntad propia, activándose solo cuando su portador enciende un mechero, y guarda en su interior una Flecha capaz de conceder Stands.', ARRAY['Manipulación de Sombras: Se desplaza a velocidad extrema por cualquier sombra conectada, pero es vulnerable a la luz solar directa.','Extracción de Almas: Arrastra a la víctima por su propia sombra para arrancarle el alma del cuerpo.','Custodia de la Flecha: Guarda en su boca una Flecha capaz de otorgar el poder de Stand a quien sobrevive a su impacto.','Funcionamiento Autónomo: Actúa según órdenes preestablecidas sin personalidad propia, activándose al reencender el mechero de Polpo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7082-a957-9bac4ad82f49')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7082-a957-9bac4ad82f49', 'en-GB', 'User: Polpo

This is a Stand that lives within shadows, able to move at tremendous speed from one shadow to another and drag its victims into the darkness to tear out their souls. It acts entirely on its own, with no will of its own, activating only when its user relights a lighter, and it carries an Arrow inside capable of granting Stand powers.', ARRAY['Shadow Manipulation: Travels at extreme speed through any connected shadow, but is vulnerable to direct sunlight.','Soul Extraction: Drags a victim through their own shadow to tear the soul from their body.','Arrow Custody: Keeps an Arrow in its mouth capable of granting Stand power to anyone who survives being struck by it.','Autonomous Operation: Follows preset commands with no personality of its own, activating when Polpo relights his lighter.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7082-a957-9bac4ad82f49')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7082-a957-9bac4ad82f49', 'ca-ES', 'Portador: Polpo

És un Stand que habita a les ombres, capaç de moure''s a gran velocitat d''una ombra a una altra i d''arrossegar les seves víctimes cap a la foscor per arrencar-los l''ànima. Actua de forma completament autònoma sense voluntat pròpia, activant-se només quan el seu portador encén un encenedor, i guarda al seu interior una Fletxa capaç de concedir poders de Stand.', ARRAY['Manipulació d''Ombres: Es desplaça a velocitat extrema per qualsevol ombra connectada, però és vulnerable a la llum solar directa.','Extracció d''Ànimes: Arrossega la víctima per la seva pròpia ombra per arrencar-li l''ànima del cos.','Custòdia de la Fletxa: Guarda a la boca una Fletxa capaç d''atorgar el poder de Stand a qui sobreviu al seu impacte.','Funcionament Autònom: Actua segons ordres preestablertes sense personalitat pròpia, activant-se en tornar a encendre l''encenedor d''en Polpo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7082-a957-9bac4ad82f49')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Soft Machine (Mario Zucchero)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7c2f-ab40-cf18590074bd', 'STAND', 'Soft Machine', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7c2f-ab40-cf18590074bd', 'A', 'C', 'E', 'A', 'D', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c2f-ab40-cf18590074bd')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c2f-ab40-cf18590074bd', 'es-ES', 'Portador: Mario Zucchero

Es un Stand de combate cercano que, al clavar su fino estoque en un objetivo, lo deshincha como si fuera un globo, dejándolo flácido y elástico mientras permanece con vida. Zucchero lo usa para arrastrar a sus rehenes por espacios diminutos y para camuflarse entre objetos idénticos ya deshinchados.', ARRAY['Deshinchado: Al clavar su estoque, convierte a personas u objetos en formas flácidas y elásticas sin matarlos.','Transporte por Rendijas: Arrastra a sus víctimas deshinchadas a través de huecos y rendijas minúsculas.','Camuflaje entre Objetos: Se oculta apilando objetos deshinchados idénticos entre sí para pasar desapercibido.','Autodeshinchado: Puede aplicar su propia habilidad sobre sí mismo para esconderse o escapar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c2f-ab40-cf18590074bd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c2f-ab40-cf18590074bd', 'en-GB', 'User: Mario Zucchero

This is a close-range Stand that, by stabbing a target with its thin rapier, deflates them like a balloon, leaving them limp and rubbery while still alive. Zucchero uses it to drag hostages through tiny spaces and to camouflage himself among identical deflated objects.', ARRAY['Deflation: Stabbing with its rapier turns people or objects into limp, rubbery, deflated shapes without killing them.','Transport Through Gaps: Drags its deflated victims through tiny holes and crevices.','Camouflage Among Objects: Hides by stacking identical deflated objects to avoid detection.','Self-Deflation: Can use its own ability on itself to hide or escape.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c2f-ab40-cf18590074bd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7c2f-ab40-cf18590074bd', 'ca-ES', 'Portador: Mario Zucchero

És un Stand de combat proper que, en clavar el seu fi estoc en un objectiu, el desinfla com si fos un globus, deixant-lo flàccid i elàstic mentre continua amb vida. En Zucchero l''utilitza per arrossegar els seus ostatges per espais minúsculs i per camuflar-se entre objectes idèntics ja desinflats.', ARRAY['Desinflament: En clavar el seu estoc, converteix persones o objectes en formes flàccides i elàstiques sense matar-los.','Transport per Escletxes: Arrossega les seves víctimes desinflades a través de forats i escletxes minúscules.','Camuflatge entre Objectes: S''amaga apilant objectes desinflats idèntics entre si per passar desapercebut.','Autodesinflament: Pot aplicar la seva pròpia habilitat sobre si mateix per amagar-se o escapar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7c2f-ab40-cf18590074bd')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kraft Work (Sale)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-788b-b4b6-8bf1dd4dcebb', 'STAND', 'Kraft Work', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb', 'A', 'A', 'E', 'C', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb', 'es-ES', 'Portador: Sale

Es un Stand capaz de eliminar la energía cinética de cualquier objeto que toca, dejándolo completamente fijo en el espacio o adherido a otra superficie. Sale acumula esa energía con toques repetidos y la libera de golpe, convirtiendo objetos cotidianos en proyectiles con la fuerza de un disparo.', ARRAY['Fijación de Objetos: Elimina la energía cinética al contacto, inmovilizando objetos en el aire o adhiriéndolos entre sí.','Acumulación de Energía: Toques repetidos sobre un objeto fijado acumulan fuerza cinética latente.','Liberación Explosiva: Libera la energía acumulada de golpe, lanzando el objeto como un proyectil con potencia de arma de fuego.','Defensa Instantánea: Puede inmovilizar proyectiles en pleno vuelo, incluso balas dirigidas contra él.','Alcance Limitado: El efecto se disipa si Sale se aleja demasiado del objetivo fijado.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb', 'en-GB', 'User: Sale

This is a Stand able to strip away the kinetic energy of anything it touches, leaving it completely fixed in space or stuck to another surface. Sale builds up that energy through repeated touches and then releases it all at once, turning everyday objects into projectiles with the force of gunfire.', ARRAY['Object Locking: Removes kinetic energy on contact, freezing objects mid-air or fusing them to one another.','Energy Accumulation: Repeated touches on a locked object build up latent kinetic force.','Explosive Release: Unleashes the stored energy at once, firing the object off like a gun-powered projectile.','Instant Defence: Can freeze projectiles mid-flight, including bullets aimed at him.','Range Limit: The effect fades if Sale moves too far from the locked target.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb', 'ca-ES', 'Portador: Sale

És un Stand capaç d''eliminar l''energia cinètica de qualsevol objecte que toca, deixant-lo completament fixat a l''espai o enganxat a una altra superfície. En Sale acumula aquesta energia amb tocs repetits i l''allibera de cop, convertint objectes quotidians en projectils amb la força d''un tret.', ARRAY['Fixació d''Objectes: Elimina l''energia cinètica en tocar-los, immobilitzant objectes a l''aire o enganxant-los entre si.','Acumulació d''Energia: Tocs repetits sobre un objecte fixat acumulen força cinètica latent.','Alliberament Explosiu: Allibera l''energia acumulada de cop, disparant l''objecte com un projectil amb potència d''arma de foc.','Defensa Instantània: Pot immobilitzar projectils en ple vol, fins i tot bales dirigides contra ell.','Abast Limitat: L''efecte es dissipa si en Sale s''allunya massa de l''objectiu fixat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Metallica (Risotto Nero)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-719f-9a06-c37e9e7ea3c2', 'STAND', 'Metallica', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-719f-9a06-c37e9e7ea3c2', 'C', 'C', 'C', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-719f-9a06-c37e9e7ea3c2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-719f-9a06-c37e9e7ea3c2', 'es-ES', 'Portador: Risotto Nero

Es un enjambre de pequeños seres metálicos que habita dentro del cuerpo de Risotto Nero y le permite manipular el hierro mediante magnetismo, tanto en el entorno como en el interior de otros cuerpos vivos. Con él extrae el hierro de la sangre de sus víctimas para forjar cuchillas y agujas que las hieren desde dentro, y también recubre su propia piel con partículas de hierro para camuflarse.', ARRAY['Manipulación del Hierro: Controla mediante magnetismo cualquier hierro presente en el entorno o en el cuerpo de un objetivo.','Armas Internas: Transforma el hierro de la hemoglobina en cuchillas, agujas o tijeras que dañan a la víctima desde dentro.','Drenaje de Hemoglobina: Extrae el hierro de la sangre, reduciendo la capacidad del cuerpo para transportar oxígeno.','Camuflaje Metálico: Cubre el cuerpo de Risotto con partículas de hierro que reflejan la luz para ocultarlo.','Reconstrucción con Grapas: Puede reunir y fijar miembros seccionados mediante grapas de hierro improvisadas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-719f-9a06-c37e9e7ea3c2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-719f-9a06-c37e9e7ea3c2', 'en-GB', 'User: Risotto Nero

This is a swarm of small metallic beings living inside Risotto Nero''s body that lets him manipulate iron through magnetism, both in the surrounding environment and inside other living bodies. He uses it to pull iron from his victims'' blood and forge it into blades and needles that wound them from within, and also coats his own skin with iron particles to camouflage himself.', ARRAY['Iron Manipulation: Magnetically controls any iron present in the environment or inside a target''s body.','Internal Weapons: Turns the iron in haemoglobin into blades, needles or scissors that damage the victim from within.','Haemoglobin Drain: Pulls iron out of the blood, reducing the body''s ability to carry oxygen.','Metallic Camouflage: Coats Risotto''s body in iron particles that reflect light to hide him.','Staple Repair: Can gather and fix severed limbs back on using improvised iron staples.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-719f-9a06-c37e9e7ea3c2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-719f-9a06-c37e9e7ea3c2', 'ca-ES', 'Portador: Risotto Nero

És un eixam de petits éssers metàl·lics que habita dins el cos d''en Risotto Nero i li permet manipular el ferro mitjançant el magnetisme, tant a l''entorn com a l''interior d''altres cossos vius. Amb ell extreu el ferro de la sang de les seves víctimes per forjar fulles i agulles que les fereixen des de dins, i també recobreix la seva pròpia pell amb partícules de ferro per camuflar-se.', ARRAY['Manipulació del Ferro: Controla mitjançant magnetisme qualsevol ferro present a l''entorn o al cos d''un objectiu.','Armes Internes: Transforma el ferro de l''hemoglobina en fulles, agulles o tisores que danyen la víctima des de dins.','Drenatge d''Hemoglobina: Extreu el ferro de la sang, reduint la capacitat del cos per transportar oxigen.','Camuflatge Metàl·lic: Cobreix el cos d''en Risotto amb partícules de ferro que reflecteixen la llum per amagar-lo.','Reconstrucció amb Grapes: Pot reunir i fixar membres seccionats mitjançant grapes de ferro improvisades.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-719f-9a06-c37e9e7ea3c2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Little Feet (Formaggio)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-781d-8b1d-60dffe9073d3', 'STAND', 'Little Feet', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-781d-8b1d-60dffe9073d3', 'D', 'B', 'E', 'A', 'D', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-781d-8b1d-60dffe9073d3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-781d-8b1d-60dffe9073d3', 'es-ES', 'Portador: Formaggio

Es un Stand que, al cortar a su objetivo con el filo de su dedo índice, provoca un encogimiento progresivo que puede reducir a una persona adulta a apenas unos centímetros en cuestión de minutos. El efecto persiste sin importar la distancia y Formaggio puede detenerlo, revertirlo de golpe o aplicarlo sobre sí mismo de forma instantánea.', ARRAY['Corte Encogedor: Su dedo índice afilado inicia un encogimiento progresivo en cualquier ser vivo u objeto que corta.','Efecto a Distancia: El encogimiento continúa avanzando aunque el objetivo se aleje de Little Feet.','Control Total: Formaggio puede detener, ralentizar o revertir violentamente el encogimiento a voluntad.','Autoencogimiento Inmediato: Puede reducir su propio tamaño de forma instantánea, sin necesidad de corte previo.','Interrupción por Objetos: Si la víctima suelta algo que llevaba encima, el encogimiento deja de afectar a ese objeto.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-781d-8b1d-60dffe9073d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-781d-8b1d-60dffe9073d3', 'en-GB', 'User: Formaggio

This is a Stand that, by cutting its target with the blade of its index finger, triggers a gradual shrinking that can reduce a grown adult to just a few centimetres within minutes. The effect keeps working regardless of distance, and Formaggio can pause it, violently reverse it, or apply it to himself instantly.', ARRAY['Shrinking Cut: Its sharp index finger starts a gradual shrinking effect on any living being or object it cuts.','Effect at a Distance: The shrinking keeps progressing even if the target moves away from Little Feet.','Full Control: Formaggio can stop, slow down or violently reverse the shrinking at will.','Instant Self-Shrinking: Can shrink his own body instantly, with no need for a prior cut.','Dropped-Item Break: If a victim drops something they were carrying, the shrinking stops affecting that item.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-781d-8b1d-60dffe9073d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-781d-8b1d-60dffe9073d3', 'ca-ES', 'Portador: Formaggio

És un Stand que, en tallar el seu objectiu amb el tall del seu dit índex, provoca un encongiment progressiu que pot reduir una persona adulta a només uns centímetres en qüestió de minuts. L''efecte persisteix sense importar la distància i en Formaggio pot aturar-lo, revertir-lo de cop o aplicar-lo sobre si mateix de forma instantània.', ARRAY['Tall Encongidor: El seu dit índex afilat inicia un encongiment progressiu en qualsevol ésser viu o objecte que talla.','Efecte a Distància: L''encongiment continua avançant encara que l''objectiu s''allunyi d''en Little Feet.','Control Total: En Formaggio pot aturar, alentir o revertir violentament l''encongiment a voluntat.','Autoencongiment Immediat: Pot reduir la seva pròpia mida de forma instantània, sense necessitat d''un tall previ.','Interrupció per Objectes: Si la víctima deixa anar alguna cosa que portava, l''encongiment deixa d''afectar aquell objecte.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-781d-8b1d-60dffe9073d3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Man in the Mirror (Illuso)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7d06-b716-46149209ca0d', 'STAND', 'Man in the Mirror', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7d06-b716-46149209ca0d', 'C', 'C', 'B', 'D', 'C', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d06-b716-46149209ca0d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d06-b716-46149209ca0d', 'es-ES', 'Portador: Illuso

Es un Stand capaz de abrir cualquier superficie reflectante como puerta a un mundo dentro de los espejos, al que arrastra a sus víctimas dejándolas invisibles e inaudibles para el exterior. Illuso puede elegir separar a un usuario de su propio Stand al arrastrarlo, y solo él puede manipular objetos y personas mientras permanecen atrapados dentro.', ARRAY['Mundo Especular: Convierte cualquier superficie reflectante en una puerta hacia una dimensión dentro de los espejos.','Arrastre Selectivo: Puede elegir separar a un Stand de su usuario al arrastrarlos al mundo del espejo, dejando a la víctima indefensa.','Aislamiento Total: Quien queda atrapado en el mundo especular se vuelve invisible e inaudible para el exterior.','Multiplicación de Portales: Romper un espejo con gente atrapada dentro crea nuevos portales en lugar de liberarlas.','Vínculo Vital: Si Illuso muere, el mundo del espejo se destruye y todos sus prisioneros quedan libres.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d06-b716-46149209ca0d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d06-b716-46149209ca0d', 'en-GB', 'User: Illuso

This is a Stand able to turn any reflective surface into a doorway to a world inside mirrors, dragging victims in and leaving them invisible and inaudible to the outside. Illuso can choose to separate a Stand user from their own Stand when pulling them in, and only he can manipulate objects and people while they remain trapped inside.', ARRAY['Mirror World: Turns any reflective surface into a doorway to a dimension inside mirrors.','Selective Pull: Can choose to separate a Stand from its user when dragging them into the mirror world, leaving the victim defenceless.','Total Isolation: Anyone trapped in the mirror world becomes invisible and inaudible to the outside.','Portal Multiplication: Breaking a mirror with people trapped inside creates new portals instead of freeing them.','Life Link: If Illuso dies, the mirror world collapses and all its prisoners are freed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d06-b716-46149209ca0d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7d06-b716-46149209ca0d', 'ca-ES', 'Portador: Illuso

És un Stand capaç d''obrir qualsevol superfície reflectant com a porta a un món dins dels miralls, on arrossega les seves víctimes deixant-les invisibles i inaudibles per a l''exterior. L''Illuso pot triar separar un usuari del seu propi Stand en arrossegar-lo, i només ell pot manipular objectes i persones mentre romanen atrapats a dins.', ARRAY['Món Especular: Converteix qualsevol superfície reflectant en una porta cap a una dimensió dins dels miralls.','Arrossegament Selectiu: Pot triar separar un Stand del seu usuari en arrossegar-los cap al món del mirall, deixant la víctima indefensa.','Aïllament Total: Qui queda atrapat al món especular esdevé invisible i inaudible per a l''exterior.','Multiplicació de Portals: Trencar un mirall amb gent atrapada a dins crea nous portals en lloc d''alliberar-los.','Vincle Vital: Si l''Illuso mor, el món del mirall es destrueix i tots els seus presoners queden lliures.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7d06-b716-46149209ca0d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Beach Boy (Pesci)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7620-8083-58437b47e8e9', 'STAND', 'Beach Boy', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7620-8083-58437b47e8e9', 'C', 'B', 'B', 'C', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7620-8083-58437b47e8e9')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7620-8083-58437b47e8e9', 'es-ES', 'Portador: Pesci

Es un Stand de largo alcance con forma de caña de pesca que Pesci empuña como arma de reconocimiento y combate. Su anzuelo y su sedal son intangibles y pueden atravesar paredes y cuerpos sin ser detectados, permitiéndole engancharse a un objetivo a distancia y percibir a través de las vibraciones el número de personas, su latido y la posición exacta de sus órganos. Cualquier daño causado al objeto o persona enganchados se refleja sobre ellos en lugar de dañar el sedal, que resulta prácticamente indestructible una vez clavado.', ARRAY['Sedal Intangible: El anzuelo y el hilo atraviesan paredes y cuerpos sin ser percibidos por el objetivo.','Detección por Vibración: Analiza las vibraciones del sedal para determinar el número, posición y latido de las personas cercanas.','Reflejo de Daño: El sedal es indestructible una vez enganchado; cualquier daño que reciba se transmite al enganchado.','Precisión Quirúrgica: Localiza órganos y puntos vitales al centímetro guiándose solo por el tacto del hilo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7620-8083-58437b47e8e9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7620-8083-58437b47e8e9', 'en-GB', 'User: Pesci

This is a long-range Stand shaped like a fishing rod that Pesci wields as both a reconnaissance and combat tool. Its hook and line are intangible and can pass through walls and bodies undetected, letting him snag a target from a distance and read vibrations to sense how many people are nearby, their heartbeat, and the exact position of their organs. Any damage inflicted on the hooked object or person is reflected onto them instead of harming the line, which becomes essentially indestructible once it catches something.', ARRAY['Intangible Line: The hook and line pass through walls and bodies without the target noticing.','Vibration Detection: Reads the line''s vibrations to work out the number, position and heartbeat of nearby people.','Damage Reflection: The line becomes indestructible once hooked; any damage it takes is transferred to the hooked target.','Surgical Precision: Pinpoints organs and vital spots to the centimetre using only the line''s feedback.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7620-8083-58437b47e8e9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7620-8083-58437b47e8e9', 'ca-ES', 'Portador: Pesci

És un Stand d''abast llarg amb forma de canya de pescar que en Pesci fa servir com a eina de reconeixement i combat. El seu ham i el seu fil són intangibles i poden travessar parets i cossos sense ser detectats, cosa que li permet enganxar-se a un objectiu a distància i percebre mitjançant les vibracions el nombre de persones, el seu batec i la posició exacta dels seus òrgans. Qualsevol dany causat a l''objecte o persona enganxats es reflecteix sobre ells en lloc de malmetre el fil, que esdevé pràcticament indestructible un cop clavat.', ARRAY['Fil Intangible: L''ham i el fil travessen parets i cossos sense que l''objectiu se n''adoni.','Detecció per Vibració: Analitza les vibracions del fil per determinar el nombre, la posició i el batec de les persones properes.','Reflex de Dany: El fil esdevé indestructible un cop enganxat; qualsevol dany que rebi es transmet a l''enganxat.','Precisió Quirúrgica: Localitza òrgans i punts vitals al centímetre guiant-se només pel tacte del fil.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7620-8083-58437b47e8e9')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- The Grateful Dead (Prosciutto)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-74e4-bdb4-dc862f97afa2', 'STAND', 'The Grateful Dead', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-74e4-bdb4-dc862f97afa2', 'B', 'E', 'B', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74e4-bdb4-dc862f97afa2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74e4-bdb4-dc862f97afa2', 'es-ES', 'Portador: Prosciutto

Es un Stand de largo alcance sin extremidades inferiores que emite una bruma capaz de envejecer aceleradamente cualquier organismo vivo. Cuanto más caliente esté el cuerpo de la víctima, más rápido envejece, hasta perder dientes, encogerse y perder facultades mentales; el frío intenso, como el hielo, puede revertir el proceso. Prosciutto es inmune a su propio poder y puede activarlo o desactivarlo en sí mismo a voluntad, lo que le permite infiltrarse disfrazado de anciano o de niño. A pesar de su alcance devastador, es un Stand lento y de escasa precisión.', ARRAY['Bruma Envejecedora: Emite una niebla que envejece de forma acelerada a cualquier ser vivo que la respire o toque.','Envejecimiento por Contacto: El contacto directo acelera el envejecimiento mucho más rápido que la simple exposición a la bruma.','Dependencia Térmica: La velocidad del envejecimiento depende del calor corporal de la víctima; el frío extremo revierte el efecto.','Inmunidad Selectiva: Prosciutto es inmune a su propio poder y puede activarlo o desactivarlo en sí mismo para disfrazarse.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74e4-bdb4-dc862f97afa2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74e4-bdb4-dc862f97afa2', 'en-GB', 'User: Prosciutto

This is a long-range Stand with no legs that emits a mist capable of rapidly aging any living organism. The hotter the victim''s body, the faster they age, losing teeth, shrinking and losing their mental faculties; intense cold, such as ice, can reverse the process. Prosciutto is immune to his own power and can switch it on or off in himself at will, letting him disguise himself as an old man or a child. Despite its devastating range, it''s a slow Stand with poor precision.', ARRAY['Aging Mist: Emits a mist that rapidly ages any living being that breathes or touches it.','Contact Aging: Direct contact accelerates aging far faster than mere exposure to the mist.','Heat Dependency: The aging rate depends on the victim''s body heat; extreme cold reverses the effect.','Selective Immunity: Prosciutto is immune to his own power and can toggle it on himself to disguise his age.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74e4-bdb4-dc862f97afa2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74e4-bdb4-dc862f97afa2', 'ca-ES', 'Portador: Prosciutto

És un Stand d''abast llarg sense extremitats inferiors que emet una boira capaç d''envellir de manera accelerada qualsevol organisme viu. Com més calent estigui el cos de la víctima, més ràpid envelleix, fins a perdre dents, encongir-se i perdre facultats mentals; el fred intens, com el gel, pot revertir el procés. En Prosciutto és immune al seu propi poder i pot activar-lo o desactivar-lo en si mateix a voluntat, cosa que li permet infiltrar-se disfressat d''ancià o de nen. Malgrat el seu abast devastador, és un Stand lent i de poca precisió.', ARRAY['Boira Envellidora: Emet una boira que envelleix de manera accelerada qualsevol ésser viu que la respiri o toqui.','Envelliment per Contacte: El contacte directe accelera l''envelliment molt més ràpid que la simple exposició a la boira.','Dependència Tèrmica: La velocitat de l''envelliment depèn de la calor corporal de la víctima; el fred extrem reverteix l''efecte.','Immunitat Selectiva: En Prosciutto és immune al seu propi poder i pot activar-lo o desactivar-lo en si mateix per disfressar-se.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74e4-bdb4-dc862f97afa2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Baby Face (Melone)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-74f4-b4bd-59c36098341d', 'STAND', 'Baby Face', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-74f4-b4bd-59c36098341d', 'A', 'B', 'A', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74f4-b4bd-59c36098341d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74f4-b4bd-59c36098341d', 'es-ES', 'Portador: Melone

Es un Stand único con forma de ordenador portátil capaz de generar, a partir de una muestra de ADN, un segundo Stand independiente y sensible llamado "Junior". Tras analizar la compatibilidad genética de una posible "madre", Melone inicia el proceso de gestación, que se completa en apenas tres minutos y produce una criatura humanoide capaz de fragmentar la materia en cubos sin dañar la vida que contienen. El Junior crece, aprende y gana experiencia por su cuenta, aunque su independencia lo hace propenso a desobedecer si la voluntad de la "madre" era muy fuerte.', ARRAY['Generación de Homúnculo: Crea un segundo Stand sensible, el Junior, a partir de una muestra de ADN analizada por el ordenador.','Gestación Rápida: El proceso de creación del Junior se completa en apenas tres minutos.','Fragmentación Cúbica: El Junior descompone la materia en segmentos cúbicos sin dañar los tejidos vivos, pudiendo reordenarlos o extraer partes.','Crecimiento Autónomo: El Junior se desarrolla física y mentalmente con el tiempo, aprendiendo de lo que observa.','Comunicación Remota: El ordenador mantiene contacto con el Junior mediante cámara y micrófono para guiarlo a distancia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74f4-b4bd-59c36098341d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74f4-b4bd-59c36098341d', 'en-GB', 'User: Melone

This is a unique laptop-shaped Stand able to generate, from a DNA sample, a second independent and sentient Stand called the "Junior". After analysing a potential "mother''s" genetic compatibility, Melone starts the gestation process, which finishes in barely three minutes and produces a humanoid creature able to fragment matter into cubes without harming any life inside them. The Junior grows, learns and gains experience on its own, though its independence makes it prone to disobedience if its "mother''s" willpower was especially strong.', ARRAY['Homunculus Generation: Creates a second sentient Stand, the Junior, from a DNA sample analysed by the computer.','Rapid Gestation: The Junior''s creation process finishes in barely three minutes.','Cubic Fragmentation: The Junior breaks matter into cubic segments without harming living tissue, able to rearrange them or extract parts.','Autonomous Growth: The Junior develops physically and mentally over time, learning from what it observes.','Remote Communication: The computer keeps contact with the Junior via camera and microphone to guide it from afar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74f4-b4bd-59c36098341d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-74f4-b4bd-59c36098341d', 'ca-ES', 'Portador: Melone

És un Stand únic amb forma d''ordinador portàtil capaç de generar, a partir d''una mostra d''ADN, un segon Stand independent i sensible anomenat "Junior". Després d''analitzar la compatibilitat genètica d''una possible "mare", en Melone inicia el procés de gestació, que es completa en tot just tres minuts i produeix una criatura humanoide capaç de fragmentar la matèria en cubs sense malmetre la vida que contenen. El Junior creix, aprèn i guanya experiència pel seu compte, tot i que la seva independència el fa propens a desobeir si la voluntat de la "mare" era molt forta.', ARRAY['Generació d''Homuncle: Crea un segon Stand sensible, el Junior, a partir d''una mostra d''ADN analitzada per l''ordinador.','Gestació Ràpida: El procés de creació del Junior es completa en tot just tres minuts.','Fragmentació Cúbica: El Junior descompon la matèria en segments cúbics sense malmetre els teixits vius, i pot reordenar-los o extreure''n parts.','Creixement Autònom: El Junior es desenvolupa físicament i mentalment amb el temps, aprenent del que observa.','Comunicació Remota: L''ordinador manté contacte amb el Junior mitjançant càmera i micròfon per guiar-lo a distància.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-74f4-b4bd-59c36098341d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- White Album (Ghiaccio)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7325-94e6-5b25b92dddd7', 'STAND', 'White Album', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7325-94e6-5b25b92dddd7', 'A', 'C', 'C', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7325-94e6-5b25b92dddd7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7325-94e6-5b25b92dddd7', 'es-ES', 'Portador: Ghiaccio

Es un Stand con forma de traje que envuelve casi todo el cuerpo de Ghiaccio y le permite generar frío extremo al contacto, alcanzando temperaturas de hasta -100°C capaces de congelar gasolina o agua de mar en segundos. Su técnica más letal, "White Album Gently Weeps", desciende hasta los -210°C y congela el aire circundante en láminas invisibles que desvían proyectiles, aunque agota enormemente su resistencia. El traje también puede solidificar la humedad del entorno en una armadura de hielo capaz de resistir balas, aunque cuenta con un punto débil crítico: una rejilla de ventilación en la nuca.', ARRAY['Congelación al Contacto: Reduce la temperatura de cualquier cosa que toca hasta unos -100°C en cuestión de segundos.','White Album Gently Weeps: Congela el aire circundante a -210°C creando láminas invisibles que desvían ataques a distancia.','Armadura de Hielo: Solidifica la humedad ambiental en una coraza protectora que resiste golpes y disparos.','Punto Débil Sellable: Puede congelar el aire dentro de la rejilla de su nuca para protegerse ese punto vulnerable.','Fuerza y Velocidad Sobrehumanas: Rompe piedra de un puñetazo y patina a más de 80 km/h.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7325-94e6-5b25b92dddd7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7325-94e6-5b25b92dddd7', 'en-GB', 'User: Ghiaccio

This is a suit-shaped Stand that covers almost all of Ghiaccio''s body and lets him generate extreme cold on contact, reaching temperatures as low as -100°C, capable of freezing petrol or seawater within seconds. Its deadliest technique, "White Album Gently Weeps", drops the temperature to -210°C and freezes the surrounding air into invisible sheets that deflect projectiles, though it drains his stamina heavily. The suit can also solidify ambient moisture into ice armour able to withstand bullets, but it has one critical weak point: a vent at the back of his neck.', ARRAY['Contact Freezing: Drops the temperature of anything it touches to around -100°C within seconds.','White Album Gently Weeps: Freezes the surrounding air to -210°C, creating invisible sheets that deflect ranged attacks.','Ice Armour: Solidifies ambient moisture into protective plating that withstands strikes and gunfire.','Sealable Weak Point: Can freeze the air inside his neck vent to shield that one vulnerable spot.','Superhuman Strength and Speed: Shatters stone with a single punch and skates at over 80 km/h.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7325-94e6-5b25b92dddd7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7325-94e6-5b25b92dddd7', 'ca-ES', 'Portador: Ghiaccio

És un Stand amb forma de vestit que embolcalla gairebé tot el cos d''en Ghiaccio i li permet generar fred extrem en tocar, arribant a temperatures d''uns -100°C capaces de congelar gasolina o aigua de mar en segons. La seva tècnica més letal, "White Album Gently Weeps", baixa fins als -210°C i congela l''aire del voltant en làmines invisibles que desvien projectils, tot i que li esgota enormement la resistència. El vestit també pot solidificar la humitat de l''entorn en una armadura de gel capaç de resistir bales, però compta amb un punt feble crític: una reixeta de ventilació al clatell.', ARRAY['Congelació en Contacte: Redueix la temperatura de qualsevol cosa que toca fins a uns -100°C en qüestió de segons.','White Album Gently Weeps: Congela l''aire del voltant a -210°C creant làmines invisibles que desvien atacs a distància.','Armadura de Gel: Solidifica la humitat ambiental en una cuirassa protectora que resisteix cops i trets.','Punt Feble Segellable: Pot congelar l''aire dins la reixeta del clatell per protegir aquest punt vulnerable.','Força i Velocitat Sobrehumanes: Trenca pedra d''un cop de puny i patina a més de 80 km/h.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7325-94e6-5b25b92dddd7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Clash (Squalo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7fdd-bf02-dba712bb3761', 'STAND', 'Clash', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7fdd-bf02-dba712bb3761', 'D', 'A', 'B', 'A', 'A', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7fdd-bf02-dba712bb3761')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7fdd-bf02-dba712bb3761', 'es-ES', 'Portador: Squalo

Es un Stand de largo alcance con forma de tiburón mecánico que solo puede existir dentro de líquidos. Squalo lo utiliza para teletransportarse instantáneamente entre cualquier masa de líquido cercana —agua, vino, sopa o incluso lágrimas— y emerger de forma repentina para atacar con violencia antes de sumergirse de nuevo. Su tamaño se adapta al volumen del líquido que ocupa, desde uno diminuto en una copa de vino hasta uno mayor en un charco. No puede penetrar líquidos sellados y su radio de teletransporte es limitado, además de perder velocidad si resulta herido.', ARRAY['Salto Líquido: Se teletransporta al instante entre masas de líquido cercanas para atacar por sorpresa.','Adaptación de Tamaño: Ajusta sus dimensiones según el volumen del líquido que ocupa en cada momento.','Emboscada Letal: Emerge brevemente del líquido para arrancar partes del cuerpo del objetivo antes de sumergirse de nuevo.','Limitación de Alcance: No puede entrar en líquidos sellados y su salto se restringe a apenas dos o tres metros.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7fdd-bf02-dba712bb3761')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7fdd-bf02-dba712bb3761', 'en-GB', 'User: Squalo

This is a long-range Stand shaped like a mechanical shark that can only exist inside liquids. Squalo uses it to instantly teleport between any nearby body of liquid — water, wine, soup or even tears — and suddenly emerge to attack viciously before submerging again. Its size adapts to the volume of liquid it occupies, from tiny inside a wine glass to larger in a puddle. It cannot enter sealed liquids, its teleport range is limited, and it loses speed if injured.', ARRAY['Liquid Warp: Instantly teleports between nearby bodies of liquid to strike by surprise.','Size Adaptation: Adjusts its size to match the volume of liquid it currently occupies.','Lethal Ambush: Briefly surfaces from the liquid to tear off parts of the target''s body before submerging again.','Range Limitation: Cannot enter sealed liquids and its warp range is restricted to just two or three metres.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7fdd-bf02-dba712bb3761')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7fdd-bf02-dba712bb3761', 'ca-ES', 'Portador: Squalo

És un Stand d''abast llarg amb forma de tauró mecànic que només pot existir dins de líquids. L''Squalo l''utilitza per teletransportar-se a l''instant entre qualsevol massa de líquid propera —aigua, vi, sopa o fins i tot llàgrimes— i emergir de sobte per atacar amb violència abans de tornar a submergir-se. La seva mida s''adapta al volum del líquid que ocupa, des d''una de minúscula en una copa de vi fins a una de més gran en un bassal. No pot penetrar líquids segellats i el seu radi de teletransport és limitat, a més de perdre velocitat si resulta ferit.', ARRAY['Salt Líquid: Es teletransporta a l''instant entre masses de líquid properes per atacar per sorpresa.','Adaptació de Mida: Ajusta les seves dimensions segons el volum del líquid que ocupa en cada moment.','Emboscada Letal: Emergeix breument del líquid per arrencar parts del cos de l''objectiu abans de tornar a submergir-se.','Limitació d''Abast: No pot entrar en líquids segellats i el seu salt es restringeix a tot just dos o tres metres.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7fdd-bf02-dba712bb3761')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Talking Head (Tizzano)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-793e-b653-5c436ef206ac', 'STAND', 'Talking Head', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-793e-b653-5c436ef206ac', 'E', 'E', 'B', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-793e-b653-5c436ef206ac')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-793e-b653-5c436ef206ac', 'es-ES', 'Portador: Tizzano

Es un Stand de largo alcance pero débil en combate directo, de tamaño similar a una lengua humana, cuya única función es adherirse a la lengua de una víctima para obligarla a mentir. Mientras permanece adherido, la víctima se ve forzada a decir mentiras al hablar, escribir o gesticular, siendo plenamente consciente de ello sin poder evitarlo. Tizzano puede alternar el efecto para permitir que diga la verdad en momentos estratégicos, y también ejerce un control físico menor sobre el anfitrión, pudiendo forzarle a asentir, señalar o incluso alargarle la lengua para manipular objetos.', ARRAY['Mentira Forzada: Obliga a la víctima a mentir en todo lo que dice, escribe o gesticula mientras está adherido a su lengua.','Alternancia de Veracidad: Permite activar o desactivar el efecto para que la víctima diga la verdad en un momento concreto.','Control Corporal Menor: Fuerza pequeños movimientos en el anfitrión, como asentir o señalar objetos.','Manipulación por Lengua: Puede alargar la lengua de la víctima para alcanzar y usar objetos como si fueran extremidades.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-793e-b653-5c436ef206ac')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-793e-b653-5c436ef206ac', 'en-GB', 'User: Tizzano

This is a long-range but weak-in-combat Stand, about the size of a human tongue, whose sole function is to latch onto a victim''s tongue and force them to lie. While attached, the victim is compelled to lie in whatever they say, write or gesture, fully aware of it but unable to stop. Tizzano can toggle the effect to let them tell the truth at strategic moments, and also exerts minor physical control over the host, able to force them to nod, point, or even stretch their tongue out to manipulate objects.', ARRAY['Forced Lying: Forces the victim to lie in everything they say, write or gesture while attached to their tongue.','Truth Toggle: Can switch the effect on or off to let the victim tell the truth at a chosen moment.','Minor Body Control: Forces small movements on the host, such as nodding or pointing at objects.','Tongue Manipulation: Can stretch the victim''s tongue out to reach and use objects like an extra limb.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-793e-b653-5c436ef206ac')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-793e-b653-5c436ef206ac', 'ca-ES', 'Portador: Tizzano

És un Stand d''abast llarg però feble en combat directe, de mida semblant a una llengua humana, la funció única del qual és enganxar-se a la llengua d''una víctima per obligar-la a mentir. Mentre hi roman enganxat, la víctima es veu forçada a mentir en parlar, escriure o gesticular, plenament conscient d''això sense poder-ho evitar. En Tizzano pot alternar l''efecte per permetre-li dir la veritat en moments estratègics, i també exerceix un control físic menor sobre l''amfitrió, podent forçar-lo a assentir, assenyalar o fins i tot allargar-li la llengua per manipular objectes.', ARRAY['Mentida Forçada: Obliga la víctima a mentir en tot el que diu, escriu o gesticula mentre està enganxada a la seva llengua.','Alternança de Veracitat: Permet activar o desactivar l''efecte perquè la víctima digui la veritat en un moment concret.','Control Corporal Menor: Força petits moviments en l''amfitrió, com assentir o assenyalar objectes.','Manipulació per Llengua: Pot allargar la llengua de la víctima per abastar i usar objectes com si fossin una extremitat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-793e-b653-5c436ef206ac')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Notorious B.I.G. (Carne)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-70d4-ac72-6df470a8f79b', 'STAND', 'Notorious B.I.G.', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-70d4-ac72-6df470a8f79b', 'A', 'INFINITE', 'INFINITE', 'INFINITE', 'E', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-70d4-ac72-6df470a8f79b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-70d4-ac72-6df470a8f79b', 'es-ES', 'Portador: Carne

Es un Stand automático que solo despierta plenamente tras la muerte de su portador, convirtiéndose en una masa de carne que devora cualquier forma de energía —incluido el poder de otros Stands o la energía mecánica— para crecer sin límite. Aunque no puede ver, detecta el movimiento a su alrededor y persigue siempre al objeto más rápido de su entorno, acelerando hasta superarlo. Su ceguera lo hace errático, pero su capacidad de regenerarse a partir del más mínimo fragmento lo vuelve prácticamente indestructible.', ARRAY['Devoración Energética: Consume cualquier tipo de energía cercana, incluida la de otros Stands, para aumentar su tamaño y poder.','Persecución del Más Rápido: Detecta y persigue automáticamente el objeto más veloz en su radio de acción, acelerando sin cesar hasta alcanzarlo.','Regeneración Total: Cualquier fragmento superviviente, por pequeño que sea, puede reconstruir el Stand entero si encuentra energía suficiente.','Detección Ciega: Pese a no tener visión funcional, percibe el movimiento circundante mediante otros sentidos, lo que lo hace casi imposible de emboscar.','Crecimiento Imparable: Cuanto más devora, más grande y peligroso se vuelve, sin un límite superior conocido a su expansión.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-70d4-ac72-6df470a8f79b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-70d4-ac72-6df470a8f79b', 'en-GB', 'User: Carne

This is an automatic Stand that only fully awakens after its user''s death, becoming a mass of flesh that devours any form of energy — including other Stands'' power or mechanical energy — to grow without limit. Though unable to see, it detects movement around itself and always chases the fastest object nearby, accelerating until it overtakes it. Its blindness makes it erratic, but its ability to regenerate from even the smallest fragment makes it nearly indestructible.', ARRAY['Energy Devouring: Consumes any nearby form of energy, including that of other Stands, to grow in size and power.','Fastest-Target Pursuit: Automatically detects and chases the fastest-moving object within range, endlessly accelerating until it catches up.','Total Regeneration: Any surviving fragment, however small, can rebuild the entire Stand if it finds enough energy.','Blind Detection: Despite lacking functional sight, it senses surrounding movement through other means, making ambushes nearly impossible.','Unstoppable Growth: The more it devours, the bigger and more dangerous it becomes, with no known upper limit to its expansion.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-70d4-ac72-6df470a8f79b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-70d4-ac72-6df470a8f79b', 'ca-ES', 'Portador: Carne

És un Stand automàtic que només desperta plenament després de la mort del seu portador, convertint-se en una massa de carn que devora qualsevol forma d''energia —inclòs el poder d''altres Stands o l''energia mecànica— per créixer sense límit. Tot i que no hi veu, detecta el moviment al seu voltant i persegueix sempre l''objecte més ràpid del seu entorn, accelerant fins a superar-lo. La seva ceguesa el fa erràtic, però la seva capacitat de regenerar-se a partir del fragment més mínim el fa pràcticament indestructible.', ARRAY['Devoració Energètica: Consumeix qualsevol tipus d''energia propera, inclosa la d''altres Stands, per augmentar la seva mida i poder.','Persecució del Més Ràpid: Detecta i persegueix automàticament l''objecte més veloç dins el seu radi d''acció, accelerant sense parar fins a atrapar-lo.','Regeneració Total: Qualsevol fragment supervivent, per petit que sigui, pot reconstruir el Stand sencer si troba prou energia.','Detecció Cega: Tot i no tenir visió funcional, percep el moviment del voltant mitjançant altres sentits, cosa que el fa gairebé impossible d''emboscar.','Creixement Imparable: Com més devora, més gran i perillós es torna, sense un límit superior conegut a la seva expansió.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-70d4-ac72-6df470a8f79b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Green Day (Cioccolata)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7568-8196-483037810bd8', 'STAND', 'Green Day', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7568-8196-483037810bd8', 'A', 'C', 'A', 'A', 'E', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7568-8196-483037810bd8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7568-8196-483037810bd8', 'es-ES', 'Portador: Cioccolata

Es un Stand humanoide capaz de generar un moho letal que se activa cuando la altitud de la víctima desciende. El moho consume cualquier organismo vivo que toca, propagándose de una víctima a otra a gran velocidad, aunque es incapaz de dañar objetos inertes o a los propios Stands. Cioccolata también puede controlar miembros amputados infectados con esporas de moho, usándolos como trampas ambulantes para tender emboscadas desde varias direcciones a la vez.', ARRAY['Moho Letal: Genera un moho que se activa al descender la altitud de la víctima, devorando cualquier tejido vivo con el que entra en contacto.','Propagación en Cadena: El moho salta de un organismo infectado a otro cercano, expandiéndose con rapidez si no se contiene a tiempo.','Inmunidad a lo Inerte: Su moho no afecta a objetos ni a Stands, solo a seres vivos, lo que lo hace altamente selectivo.','Extremidades Controladas: Puede manejar a distancia miembros amputados sembrados de esporas para atacar por sorpresa.','Sinergia con Oasis: Combina su moho con el poder de Secco para hundir a las víctimas bajo tierra y acelerar la infección.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7568-8196-483037810bd8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7568-8196-483037810bd8', 'en-GB', 'User: Cioccolata

This is a humanoid Stand able to generate a lethal mold that triggers whenever the victim''s altitude decreases. The mold consumes any living organism it touches, spreading from victim to victim at high speed, though it cannot harm inanimate objects or Stands themselves. Cioccolata can also control severed limbs infected with mold spores, using them as walking traps to ambush from several directions at once.', ARRAY['Lethal Mold: Generates a mold that activates as the victim''s altitude drops, devouring any living tissue it touches.','Chain Spread: The mold leaps from an infected organism to another nearby one, expanding rapidly if left unchecked.','Immunity to the Inert: Its mold does not affect objects or Stands, only living beings, making it highly selective.','Controlled Limbs: Can remotely operate severed limbs seeded with spores to strike by surprise.','Synergy with Oasis: Combines its mold with Secco''s power to sink victims underground and accelerate the infection.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7568-8196-483037810bd8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7568-8196-483037810bd8', 'ca-ES', 'Portador: Cioccolata

És un Stand humanoide capaç de generar una floridura letal que s''activa quan l''altitud de la víctima descendeix. La floridura consumeix qualsevol organisme viu que toca, propagant-se d''una víctima a una altra a gran velocitat, tot i que és incapaç de fer mal a objectes inerts o als mateixos Stands. En Cioccolata també pot controlar membres amputats infectats amb espores de floridura, fent-los servir com a trampes ambulants per parar emboscades des de diverses direccions alhora.', ARRAY['Floridura Letal: Genera una floridura que s''activa en descendir l''altitud de la víctima, devorant qualsevol teixit viu amb què entra en contacte.','Propagació en Cadena: La floridura salta d''un organisme infectat a un altre de proper, expandint-se ràpidament si no es conté a temps.','Immunitat a l''Inert: La seva floridura no afecta objectes ni Stands, només éssers vius, cosa que la fa altament selectiva.','Extremitats Controlades: Pot manejar a distància membres amputats sembrats d''espores per atacar per sorpresa.','Sinergia amb Oasis: Combina la seva floridura amb el poder d''en Secco per enfonsar les víctimes sota terra i accelerar la infecció.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7568-8196-483037810bd8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Oasis (Secco)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-75ca-a23c-bdd5e6f86f5d', 'STAND', 'Oasis', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d', 'A', 'A', 'B', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d', 'es-ES', 'Portador: Secco

Es un Stand que se manifiesta como un traje que cubre casi todo el cuerpo de Secco, dejando solo los ojos, una hendidura en el torso y las muñecas al descubierto. Le permite licuar el suelo y cualquier material sólido con el que entra en contacto, convirtiéndolo en una sustancia fangosa a través de la cual puede nadar y desplazarse bajo tierra a gran velocidad mientras permanece oculto. También potencia su fuerza de combate cuerpo a cuerpo y le permite disparar proyectiles endurecidos escupiendo el material licuado.', ARRAY['Licuefacción del Terreno: Convierte tierra, piedra y otros materiales sólidos en una sustancia fangosa al tocarlos.','Natación Subterránea: Se desplaza a gran velocidad bajo tierra a través del material licuado, permaneciendo oculto de sus enemigos.','Golpe Elástico: Aprovecha la elasticidad del fango para amplificar la fuerza de sus puñetazos.','Proyectil de Barro: Escupe material licuado endurecido a modo de proyectil contundente.','Detección por Vibración: Localiza enemigos mediante las vibraciones que se transmiten a través del terreno licuado.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d', 'en-GB', 'User: Secco

This is a Stand that manifests as a suit covering nearly all of Secco''s body, leaving only his eyes, a slit across his torso, and his wrists exposed. It lets him liquefy the ground and any solid material it touches, turning it into a mud-like substance he can swim and travel through underground at high speed while staying hidden. It also boosts his close-combat strength and lets him spit hardened projectiles made from the liquefied material.', ARRAY['Ground Liquefaction: Turns earth, stone and other solid materials into a mud-like substance on contact.','Underground Swimming: Travels at high speed underground through the liquefied material, staying hidden from enemies.','Elastic Strike: Uses the mud''s elasticity to amplify the force of his punches.','Mud Projectile: Spits hardened liquefied material as a blunt-force projectile.','Vibration Detection: Locates enemies through vibrations carried by the liquefied ground.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d', 'ca-ES', 'Portador: Secco

És un Stand que es manifesta com un vestit que cobreix gairebé tot el cos d''en Secco, deixant només els ulls, una escletxa al tors i els canells al descobert. Li permet liquar el terra i qualsevol material sòlid amb què entra en contacte, convertint-lo en una substància fangosa a través de la qual pot nedar i desplaçar-se sota terra a gran velocitat mentre roman ocult. També potencia la seva força de combat cos a cos i li permet disparar projectils endurits escopint el material liquat.', ARRAY['Liqüefacció del Terreny: Converteix terra, pedra i altres materials sòlids en una substància fangosa en tocar-los.','Natació Subterrània: Es desplaça a gran velocitat sota terra a través del material liquat, romanent ocult dels seus enemics.','Cop Elàstic: Aprofita l''elasticitat del fang per amplificar la força dels seus cops de puny.','Projectil de Fang: Escup material liquat endurit a mode de projectil contundent.','Detecció per Vibració: Localitza enemics mitjançant les vibracions que es transmeten a través del terreny liquat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Rolling Stones (Scolippi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e72f-3090-7e8b-80c8-da5c17f08e29', 'STAND', 'Rolling Stones', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e72f-3090-7e8b-80c8-da5c17f08e29', 'NULL', 'B', 'A', 'A', 'E', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7e8b-80c8-da5c17f08e29')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7e8b-80c8-da5c17f08e29', 'es-ES', 'Portador: Scolippi

Es un Stand completamente automático que despertó en Scolippi durante su infancia y que actúa por su cuenta, sin que él pueda controlar cuándo aparece o desaparece. Adopta la forma de una piedra o escultura enroscada marcada con el kanji de la desgracia y persigue a quienes tienen un destino de muerte inminente, transformándose poco a poco hasta reflejar la manera exacta en que van a morir. Al tocar a su objetivo le concede una muerte serena y sin heridas visibles, y aunque se le destruya, sus fragmentos se recomponen para revelar igualmente el destino marcado.', ARRAY['Predicción Mortal: Detecta y persigue a quienes están destinados a morir en un futuro cercano, sin que nada pueda impedirlo.','Transformación Reveladora: Se va moldeando poco a poco hasta mostrar la forma exacta de la muerte que le espera al objetivo.','Muerte Serena: Al contacto, provoca el fallecimiento del objetivo sin dolor ni marcas visibles, como si el destino simplemente se cumpliera.','Movimiento Subterráneo: Puede desplazarse a través de tierra y muros sin ser detectado mientras sigue su rastro.','Inmunidad de los No Destinados: Cualquiera que no tenga la muerte marcada puede tocarlo sin sufrir ningún daño.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7e8b-80c8-da5c17f08e29')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7e8b-80c8-da5c17f08e29', 'en-GB', 'User: Scolippi

This is a fully automatic Stand that awakened in Scolippi during his childhood and acts entirely on its own, beyond his control over when it appears or vanishes. It takes the form of a coiled stone or sculpture marked with the kanji for misfortune, and pursues those fated to die soon, gradually reshaping itself to reflect the exact manner of their coming death. On contact with its target it grants a peaceful death with no visible wounds, and even when destroyed, its fragments reassemble to reveal the same marked fate.', ARRAY['Death Prediction: Detects and pursues those fated to die in the near future, with nothing able to stop it.','Revealing Transformation: Gradually reshapes itself to show the exact form of the death awaiting its target.','Peaceful Death: On contact, causes the target''s death painlessly and without visible marks, as if fate were simply fulfilled.','Underground Movement: Can travel through earth and walls undetected while tracking its target.','Immunity of the Unmarked: Anyone not fated to die can touch it without suffering any harm.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7e8b-80c8-da5c17f08e29')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e72f-3090-7e8b-80c8-da5c17f08e29', 'ca-ES', 'Portador: Scolippi

És un Stand completament automàtic que va despertar en Scolippi durant la seva infància i que actua pel seu compte, sense que ell pugui controlar quan apareix o desapareix. Adopta la forma d''una pedra o escultura cargolada marcada amb el kanji de la desgràcia i persegueix aquells que tenen un destí de mort imminent, transformant-se a poc a poc fins a reflectir la manera exacta en què moriran. En tocar el seu objectiu li concedeix una mort serena i sense ferides visibles, i encara que es destrueixi, els seus fragments es recomponen per revelar igualment el destí marcat.', ARRAY['Predicció Mortal: Detecta i persegueix aquells que estan destinats a morir en un futur proper, sense que res ho pugui impedir.','Transformació Reveladora: Es va modelant a poc a poc fins a mostrar la forma exacta de la mort que espera l''objectiu.','Mort Serena: En contacte, provoca la mort de l''objectiu sense dolor ni marques visibles, com si el destí simplement es complís.','Moviment Subterrani: Es pot desplaçar a través de terra i murs sense ser detectat mentre segueix el seu rastre.','Immunitat dels No Marcats: Qualsevol que no tingui la mort marcada el pot tocar sense patir cap dany.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e72f-3090-7e8b-80c8-da5c17f08e29')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Evolution chain (Gold Experience -> Gold Experience: Requiem).
UPDATE stands SET evolves_from_id = '01a0e72f-3090-7c80-900f-5802c64c17e6' WHERE id = '01a0e72f-3090-7198-a825-c53ec83cb393' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e72f-3090-7c80-900f-5802c64c17e6');

-- Chariot Requiem: house-style fix only (stats/rarity/evolution untouched).
-- Guarded like 00019's admin-edit-wins pattern: only rewrites the row if it
-- still has the exact updated_at timestamp from the original 00017/00019 seed
-- (i.e. no admin has touched it since).
UPDATE power_translations SET description = 'Portador: Ninguno (se separa de Jean Pierre Polnareff)

Es un Stand que actúa de forma completamente autónoma tras la muerte de su portador original, sin necesitar a ningún humano para manifestarse. Su habilidad principal consiste en intercambiar y manipular almas entre cuerpos, además de reflejar y controlar temporalmente el poder de otros Stands.', skills = ARRAY['Manipulación de Almas: Extrae e intercambia almas entre distintos cuerpos, incluso entre personas y animales.','Control y Reflejo de Stands: Puede tomar el control temporal de otros Stands o reflejar sus habilidades contra su propio usuario.','Autonomía Completa: Actúa por sí mismo sin necesitar un portador humano, tras separarse de Jean Pierre Polnareff.']::text[]
  WHERE power_id = '019fea51-b58d-789a-ab81-361abf533ec1' AND locale = 'es-ES' AND EXISTS (SELECT 1 FROM powers WHERE id = '019fea51-b58d-789a-ab81-361abf533ec1' AND updated_at = '2026-09-27T14:53:37.97178Z'::timestamptz);
UPDATE power_translations SET description = 'User: None (separated from Jean Pierre Polnareff)

This Stand acts entirely on its own after its original user''s death, needing no human to manifest. Its main power is swapping and manipulating souls between bodies, alongside reflecting and temporarily controlling other Stands'' powers.', skills = ARRAY['Soul Manipulation: Extracts and swaps souls between different bodies, even between people and animals.','Stand Control and Reflection: Can temporarily seize control of other Stands or reflect their abilities back at their own user.','Complete Autonomy: Acts entirely on its own with no human user needed, after separating from Jean Pierre Polnareff.']::text[]
  WHERE power_id = '019fea51-b58d-789a-ab81-361abf533ec1' AND locale = 'en-GB' AND EXISTS (SELECT 1 FROM powers WHERE id = '019fea51-b58d-789a-ab81-361abf533ec1' AND updated_at = '2026-09-27T14:53:37.97178Z'::timestamptz);
UPDATE power_translations SET description = 'Portador: Ningú (se separa d''en Jean Pierre Polnareff)

Aquest Stand actua de manera completament autònoma després de la mort del seu portador original, sense necessitar cap humà per manifestar-se. La seva habilitat principal consisteix a intercanviar i manipular ànimes entre cossos, a més de reflectir i controlar temporalment el poder d''altres Stands.', skills = ARRAY['Manipulació d''Ànimes: Extreu i intercanvia ànimes entre diferents cossos, fins i tot entre persones i animals.','Control i Reflex d''Stands: Pot prendre el control temporal d''altres Stands o reflectir-ne les habilitats contra el seu propi usuari.','Autonomia Completa: Actua tot sol sense necessitar cap portador humà, després de separar-se d''en Jean Pierre Polnareff.']::text[]
  WHERE power_id = '019fea51-b58d-789a-ab81-361abf533ec1' AND locale = 'ca-ES' AND EXISTS (SELECT 1 FROM powers WHERE id = '019fea51-b58d-789a-ab81-361abf533ec1' AND updated_at = '2026-09-27T14:53:37.97178Z'::timestamptz);

-- +goose Down
-- Cascades to stands + power_translations rows via ON DELETE CASCADE.
-- Chariot Requiem is NOT deleted (pre-existing row, only its translations
-- were touched above) and its translation fix is not reverted here - the
-- old text was the buggy one, there is nothing worth rolling back to.
DELETE FROM powers WHERE id IN (
  '01a0e72f-3090-7c80-900f-5802c64c17e6',
  '01a0e72f-3090-7198-a825-c53ec83cb393',
  '01a0e72f-3090-7317-98c9-8361101ec2a6',
  '01a0e72f-3090-7987-8c9a-b1925f1a790b',
  '01a0e72f-3090-7d2c-ad6e-2779521a0c46',
  '01a0e72f-3090-7ecc-9964-e2b28b7c0432',
  '01a0e72f-3090-7bb1-879d-68a1f7639d35',
  '01a0e72f-3090-7dd5-97c6-4e5797545490',
  '01a0e72f-3090-7ad7-a7ed-9be5b1e4e682',
  '01a0e72f-3090-772d-b39f-6d66d8b6a1bc',
  '01a0e72f-3090-7082-a957-9bac4ad82f49',
  '01a0e72f-3090-7c2f-ab40-cf18590074bd',
  '01a0e72f-3090-788b-b4b6-8bf1dd4dcebb',
  '01a0e72f-3090-719f-9a06-c37e9e7ea3c2',
  '01a0e72f-3090-781d-8b1d-60dffe9073d3',
  '01a0e72f-3090-7d06-b716-46149209ca0d',
  '01a0e72f-3090-7620-8083-58437b47e8e9',
  '01a0e72f-3090-74e4-bdb4-dc862f97afa2',
  '01a0e72f-3090-74f4-b4bd-59c36098341d',
  '01a0e72f-3090-7325-94e6-5b25b92dddd7',
  '01a0e72f-3090-7fdd-bf02-dba712bb3761',
  '01a0e72f-3090-793e-b653-5c436ef206ac',
  '01a0e72f-3090-70d4-ac72-6df470a8f79b',
  '01a0e72f-3090-7568-8196-483037810bd8',
  '01a0e72f-3090-75ca-a23c-bdd5e6f86f5d',
  '01a0e72f-3090-7e8b-80c8-da5c17f08e29'
);
