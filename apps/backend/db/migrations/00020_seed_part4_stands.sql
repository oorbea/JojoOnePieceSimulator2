-- +goose Up
-- Seeds JoJo's Bizarre Adventure Part 4 (Diamond is Unbreakable) Stands.
-- Hand-authored new content (stats verified against jojowiki.com), not a
-- catalogsync prod-dump migration like 00017/00019 - see
-- ObsidianVault/catalog-seed-part4-stands.md. Runs in every environment
-- including prod (no ENVSUB guard): images are added later by an admin
-- through the normal picture-upload flow, so every row ships with the
-- powers.picture* defaults (picture_status = NONE).
--
-- Idempotent against a name an admin may have already created in prod:
-- every powers insert is ON CONFLICT DO NOTHING (id or name clash both
-- skip), and every dependent insert is gated on the powers row actually
-- existing with that id, so a skipped Stand never leaves an orphaned
-- stands/power_translations row behind.

-- Crazy Diamond (Josuke Higashikata)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0faf-7459-891a-aa738b67b7a3', 'STAND', 'Crazy Diamond', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0faf-7459-891a-aa738b67b7a3', 'A', 'A', 'D', 'B', 'B', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-7459-891a-aa738b67b7a3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-7459-891a-aa738b67b7a3', 'es-ES', 'Portador: Josuke Higashikata

Es un Stand cercano de fuerza y velocidad excepcionales cuya habilidad exclusiva es restaurar cualquier objeto u organismo a un estado anterior de su historia mediante el contacto, lo que le permite curar heridas, reparar objetos e incluso alterar la estructura de lo que toca. Josuke la combina con una potencia de combate brutal, aunque su alcance es prácticamente nulo.', ARRAY['Restauración: Devuelve cualquier objeto o ser vivo a un estado anterior de su historia con solo tocarlo, curando heridas o reparando daños.','Fuerza Sobrehumana: Golpea con una potencia devastadora capaz de derribar a varias personas de un solo puñetazo.','Velocidad Sobrehumana: Lanza puñetazos a más de 300 km/h, sorprendiendo incluso a Stands veloces en combate cercano.','Disparo de Precisión: Puede disparar un proyectil con la punta de los dedos con precisión quirúrgica a varias decenas de metros.','Restauración Táctica: Usa su poder de forma creativa para atrapar enemigos o fusionar objetos, más allá de la simple curación.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-7459-891a-aa738b67b7a3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-7459-891a-aa738b67b7a3', 'en-GB', 'User: Josuke Higashikata

This is a close-range Stand of exceptional strength and speed whose signature power is restoring any object or organism to an earlier state in its history through touch, letting it heal wounds, mend objects and even alter what it touches. Josuke pairs this with brutal combat power, though its range is next to nothing.', ARRAY['Restoration: Returns any object or living being to an earlier state in its history with a single touch, healing wounds or mending damage.','Superhuman Strength: Punches with devastating power capable of felling several people with a single blow.','Superhuman Speed: Throws punches at over 300 km/h, catching even fast Stands off guard in close combat.','Precision Shot: Can fire a projectile from its fingertip with surgical accuracy across several dozen metres.','Tactical Restoration: Uses its power creatively to trap enemies or fuse objects together, beyond simple healing.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-7459-891a-aa738b67b7a3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-7459-891a-aa738b67b7a3', 'ca-ES', 'Portador: Josuke Higashikata

És un Stand de combat proper amb una força i una velocitat excepcionals, l''habilitat exclusiva del qual és restaurar qualsevol objecte o organisme a un estat anterior de la seva història mitjançant el contacte, cosa que li permet curar ferides, reparar objectes i fins i tot alterar l''estructura del que toca. En Josuke la combina amb una potència de combat brutal, tot i que el seu abast és pràcticament nul.', ARRAY['Restauració: Retorna qualsevol objecte o ésser viu a un estat anterior de la seva història amb només tocar-lo, curant ferides o reparant danys.','Força Sobrehumana: Colpeja amb una potència devastadora capaç de tombar diverses persones d''un sol cop de puny.','Velocitat Sobrehumana: Llança cops de puny a més de 300 km/h, sorprenent fins i tot Stands veloços en combat proper.','Tret de Precisió: Pot disparar un projectil des de la punta dels dits amb precisió quirúrgica a diverses desenes de metres.','Restauració Tàctica: Fa servir el seu poder de manera creativa per atrapar enemics o fusionar objectes, més enllà de la simple curació.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-7459-891a-aa738b67b7a3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- The Hand (Okuyasu Nijimura)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0faf-79e1-aa68-c9c56cf130df', 'STAND', 'The Hand', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0faf-79e1-aa68-c9c56cf130df', 'B', 'B', 'D', 'C', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-79e1-aa68-c9c56cf130df')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-79e1-aa68-c9c56cf130df', 'es-ES', 'Portador: Okuyasu Nijimura

Es un Stand de combate cuerpo a cuerpo cuya habilidad exclusiva es borrar de la existencia cualquier cosa que su palma derecha roce con un movimiento de barrido, incluyendo el propio espacio. El vacío resultante se cierra al instante acercando lo que quedaba al otro lado, lo que Okuyasu aprende a usar tanto para atacar como para desplazarse.', ARRAY['Borrado Espacial: Elimina de la realidad cualquier objeto, material o porción de espacio que toque con un barrido de su palma derecha.','Teletransporte por Vacío: Borra la distancia entre dos puntos para acercarlos instantáneamente, permitiéndole desplazarse o atraer objetivos.','Cierre Instantáneo: El espacio borrado se recompone de inmediato uniendo lo que había alrededor, sin dejar rastro del corte.','Combate Cuerpo a Cuerpo: Posee una fuerza y velocidad notables que lo hacen peligroso incluso sin usar su borrado espacial.','Limitación del Barrido: Su poder solo funciona con un movimiento de barrido de la palma, por lo que puede bloquearse sujetando su brazo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-79e1-aa68-c9c56cf130df')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-79e1-aa68-c9c56cf130df', 'en-GB', 'User: Okuyasu Nijimura

This is a close-combat Stand whose signature power is erasing anything its right palm brushes with a sweeping motion, including space itself. The resulting gap closes instantly, pulling whatever lay beyond it closer, and Okuyasu learns to use this both to attack and to move around.', ARRAY['Space Erasure: Removes from reality any object, material or portion of space that its right palm sweeps across.','Void Teleportation: Erases the distance between two points to bring them instantly together, letting it move around or pull targets closer.','Instant Closure: The erased space stitches itself back together at once, joining what lay on either side without leaving any trace.','Close-Combat Prowess: Has notable strength and speed that make it dangerous even without relying on its erasure power.','Sweep Limitation: Its power only triggers with a sweeping motion of the palm, so it can be stopped by pinning its arm.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-79e1-aa68-c9c56cf130df')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-79e1-aa68-c9c56cf130df', 'ca-ES', 'Portador: Okuyasu Nijimura

És un Stand de combat cos a cos l''habilitat exclusiva del qual és esborrar de l''existència qualsevol cosa que la seva palma dreta fregui amb un moviment d''escombrada, incloent-hi el mateix espai. El buit resultant es tanca a l''instant apropant el que hi havia a l''altre costat, cosa que l''Okuyasu aprèn a fer servir tant per atacar com per desplaçar-se.', ARRAY['Esborrat Espacial: Elimina de la realitat qualsevol objecte, material o porció d''espai que toqui amb una escombrada de la seva palma dreta.','Teletransport per Buit: Esborra la distància entre dos punts per apropar-los a l''instant, cosa que li permet desplaçar-se o atreure objectius.','Tancament Instantani: L''espai esborrat es recompon de seguida unint el que hi havia a banda i banda, sense deixar rastre del tall.','Combat Cos a Cos: Té una força i una velocitat notables que el fan perillós fins i tot sense fer servir el seu esborrat espacial.','Limitació de l''Escombrada: El seu poder només funciona amb un moviment d''escombrada de la palma, per la qual cosa es pot bloquejar subjectant-li el braç.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-79e1-aa68-c9c56cf130df')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Echoes Act 0 (Koichi Hirose)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0faf-792f-97c3-2bd23f0f93a7', 'STAND', 'Echoes Act 0', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0faf-792f-97c3-2bd23f0f93a7', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-792f-97c3-2bd23f0f93a7', 'es-ES', 'Portador: Koichi Hirose

Es la forma latente de Echoes, un Stand encerrado en un caparazón parecido a un huevo con una cola que aún no ha eclosionado. Koichi ni siquiera sabe que lo posee ni cómo invocarlo conscientemente, por lo que ninguna de sus capacidades se ha manifestado todavía.', ARRAY['Forma de Huevo Latente: Echoes permanece dormido dentro de un caparazón inerte, sin habilidades activas ni control consciente por parte de Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-792f-97c3-2bd23f0f93a7', 'en-GB', 'User: Koichi Hirose

This is Echoes'' dormant form, a Stand sealed inside an egg-like shell with a tail that has yet to hatch. Koichi does not even know he has it or how to summon it on purpose, so none of its abilities have shown themselves yet.', ARRAY['Dormant Egg Form: Echoes remains asleep inside an inert shell, with no active abilities and no conscious control from Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0faf-792f-97c3-2bd23f0f93a7', 'ca-ES', 'Portador: Koichi Hirose

És la forma latent de l''Echoes, un Stand tancat dins d''una closca semblant a un ou amb una cua que encara no ha eclosionat. En Koichi ni tan sols sap que el té ni com invocar-lo conscientment, per la qual cosa cap de les seves capacitats s''ha manifestat encara.', ARRAY['Forma d''Ou Latent: L''Echoes roman adormit dins d''una closca inert, sense habilitats actives ni control conscient per part d''en Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Echoes Act 1 (Koichi Hirose)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-728e-9ede-c44c9f8af7b8', 'STAND', 'Echoes Act 1', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8', 'E', 'E', 'B', 'B', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8', 'es-ES', 'Portador: Koichi Hirose

Tras eclosionar, Echoes gana la capacidad de escribir onomatopeyas sobre superficies para generar el sonido correspondiente en la realidad, comunicándose con Koichi mediante caracteres que aparecen en su cuerpo. Es casi inútil en combate directo, pero su alcance y potencial son notables para un Stand tan joven.', ARRAY['Escritura de Onomatopeyas: Inscribe palabras que imitan sonidos sobre cualquier superficie, y ese sonido se produce físicamente al instante.','Mensajes Corporales: Puede escribir frases sobre el propio cuerpo de Echoes para comunicarse directamente con Koichi.','Alcance Amplio: Puede actuar a una distancia de hasta 50 metros, mucho más lejos que la mayoría de Stands de combate cercano.','Combate Casi Nulo: Carece de fuerza y velocidad ofensivas, siendo prácticamente inútil si se le obliga a pelear cuerpo a cuerpo.','Gran Potencial de Desarrollo: A pesar de su debilidad inicial, muestra un margen de crecimiento excepcional según evolucione Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8', 'en-GB', 'User: Koichi Hirose

Once hatched, Echoes gains the ability to write onomatopoeia on surfaces to produce the matching sound in reality, communicating with Koichi through characters that appear on its body. It is almost useless in direct combat, but its range and potential are remarkable for such a young Stand.', ARRAY['Onomatopoeia Writing: Inscribes words that mimic sounds onto any surface, and that sound is instantly produced in reality.','Body Messages: Can write phrases on Echoes'' own body to communicate directly with Koichi.','Wide Range: Can act at distances of up to 50 metres, far further than most close-combat Stands.','Near-Useless in Combat: Lacks offensive strength and speed, making it practically useless if forced into direct fighting.','Great Development Potential: Despite its early weakness, it shows an exceptional room for growth as Koichi develops it.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8', 'ca-ES', 'Portador: Koichi Hirose

Un cop eclosionat, l''Echoes guanya la capacitat d''escriure onomatopeies sobre superfícies per generar el so corresponent a la realitat, i es comunica amb en Koichi mitjançant caràcters que apareixen al seu cos. És gairebé inútil en combat directe, però el seu abast i potencial són notables per a un Stand tan jove.', ARRAY['Escriptura d''Onomatopeies: Inscriu paraules que imiten sons sobre qualsevol superfície, i aquest so es produeix físicament a l''instant.','Missatges Corporals: Pot escriure frases sobre el propi cos de l''Echoes per comunicar-se directament amb en Koichi.','Abast Ampli: Pot actuar a una distància de fins a 50 metres, molt més lluny que la majoria de Stands de combat proper.','Combat Gairebé Nul: Li manca força i velocitat ofensives, cosa que el fa pràcticament inútil si se''l força a lluitar cos a cos.','Gran Potencial de Desenvolupament: Malgrat la seva feblesa inicial, mostra un marge de creixement excepcional a mesura que en Koichi l''evoluciona.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Echoes Act 2 (Koichi Hirose)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d', 'STAND', 'Echoes Act 2', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d', 'C', 'D', 'B', 'B', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d', 'es-ES', 'Portador: Koichi Hirose

Al evolucionar, Echoes gana la capacidad de disparar efectos sonoros físicos desde su cola que se adhieren a los objetivos y desencadenan la sensación descrita (viento, calor, elasticidad) al contacto. También puede anular temporalmente uno de los cinco sentidos de su víctima, aunque solo puede mantener un efecto activo a la vez.', ARRAY['Efectos Sonoros Físicos: Dispara palabras escritas convertidas en proyectiles que provocan físicamente el efecto que describen al impactar.','Anulación de Sentidos: Puede bloquear temporalmente la vista, el oído u otro sentido del objetivo al que impacte con su ataque sonoro.','Mayor Movilidad: Gana velocidad respecto a su forma anterior, aunque sigue siendo modesta comparada con Stands de combate.','Alcance Efectivo: Puede activar sus efectos sobre objetivos a varios metros de distancia sin necesidad de contacto directo.','Un Solo Efecto Activo: Solo puede mantener un efecto sonoro funcionando a la vez, por lo que debe elegir bien su uso en combate.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d', 'en-GB', 'User: Koichi Hirose

Upon evolving, Echoes gains the ability to fire physical sound effects from its tail that stick to targets and trigger the described sensation (wind, heat, elasticity) on contact. It can also temporarily shut down one of a victim''s five senses, though it can only keep a single effect active at a time.', ARRAY['Physical Sound Effects: Fires written words turned into projectiles that physically inflict the effect they describe on impact.','Sense Cancellation: Can temporarily block sight, hearing or another sense in whichever target its sound attack strikes.','Improved Mobility: Gains speed compared to its earlier form, though still modest next to combat-focused Stands.','Effective Range: Can trigger its effects on targets several metres away without needing direct contact.','Single Active Effect: Can only keep one sound effect running at a time, so its use must be chosen carefully in a fight.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d', 'ca-ES', 'Portador: Koichi Hirose

En evolucionar, l''Echoes guanya la capacitat de disparar efectes sonors físics des de la cua que s''enganxen als objectius i desencadenen la sensació descrita (vent, calor, elasticitat) en tocar-los. També pot anul·lar temporalment un dels cinc sentits de la víctima, tot i que només pot mantenir un efecte actiu alhora.', ARRAY['Efectes Sonors Físics: Dispara paraules escrites convertides en projectils que provoquen físicament l''efecte que descriuen en impactar.','Anul·lació de Sentits: Pot bloquejar temporalment la vista, l''oïda o un altre sentit de l''objectiu que impacti amb el seu atac sonor.','Més Mobilitat: Guanya velocitat respecte a la seva forma anterior, tot i que continua sent modesta comparada amb Stands de combat.','Abast Efectiu: Pot activar els seus efectes sobre objectius a diversos metres de distància sense necessitat de contacte directe.','Un Sol Efecte Actiu: Només pot mantenir un efecte sonor funcionant alhora, per la qual cosa n''ha de triar bé l''ús en combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Echoes Act 3 (Koichi Hirose)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7f77-8c31-cab107b96727', 'STAND', 'Echoes Act 3', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7f77-8c31-cab107b96727', 'B', 'B', 'C', 'B', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7f77-8c31-cab107b96727')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7f77-8c31-cab107b96727', 'es-ES', 'Portador: Koichi Hirose

En su forma final, Echoes adopta una silueta más definida y casi corpórea, ganando fuerza y velocidad de combate directo muy superiores a las de sus formas anteriores. Su ataque insignia, 3 Freeze, incrementa drásticamente el peso de lo que golpea, inmovilizando objetivos con un solo puñetazo.', ARRAY['3 Freeze: Multiplica drásticamente el peso de cualquier objetivo golpeado, dejándolo clavado en el sitio al instante.','Forma Casi Corpórea: Adopta una silueta sólida y definida que le permite combatir cuerpo a cuerpo con garras y golpes directos.','Fuerza y Velocidad Mejoradas: Supera con claridad a sus formas anteriores en potencia y rapidez de movimiento.','Comunicación Fluida: Mantiene la capacidad de comunicarse con Koichi, ahora con un carácter propio más marcado y colaborativo.','Alcance Reducido: A cambio de su poder de combate, su radio de acción efectivo se limita a pocos metros alrededor de Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7f77-8c31-cab107b96727')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7f77-8c31-cab107b96727', 'en-GB', 'User: Koichi Hirose

In its final form, Echoes takes on a more defined, almost solid shape, gaining direct combat strength and speed far beyond its earlier forms. Its signature attack, 3 Freeze, drastically increases the weight of whatever it strikes, pinning targets in place with a single punch.', ARRAY['3 Freeze: Drastically multiplies the weight of any target it strikes, pinning it in place instantly.','Near-Corporeal Form: Takes on a solid, defined shape that lets it fight hand-to-hand with claws and direct blows.','Improved Strength and Speed: Clearly surpasses its earlier forms in power and speed of movement.','Fluent Communication: Keeps the ability to communicate with Koichi, now with a more distinct, cooperative personality.','Reduced Range: In exchange for its combat power, its effective radius shrinks to just a few metres around Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7f77-8c31-cab107b96727')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7f77-8c31-cab107b96727', 'ca-ES', 'Portador: Koichi Hirose

En la seva forma final, l''Echoes adopta una silueta més definida i gairebé corpòria, i guanya una força i una velocitat de combat directe molt superiors a les de les seves formes anteriors. El seu atac insígnia, 3 Freeze, incrementa dràsticament el pes del que colpeja, immobilitzant objectius d''un sol cop de puny.', ARRAY['3 Freeze: Multiplica dràsticament el pes de qualsevol objectiu que colpeja, deixant-lo clavat al lloc a l''instant.','Forma Gairebé Corpòria: Adopta una silueta sòlida i definida que li permet combatre cos a cos amb urpes i cops directes.','Força i Velocitat Millorades: Supera amb claredat les seves formes anteriors en potència i rapidesa de moviment.','Comunicació Fluida: Manté la capacitat de comunicar-se amb en Koichi, ara amb un caràcter propi més marcat i col·laboratiu.','Abast Reduït: A canvi del seu poder de combat, el seu radi d''acció efectiu es limita a pocs metres al voltant d''en Koichi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7f77-8c31-cab107b96727')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Heaven's Door (Rohan Kishibe)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-78d5-82f0-2611b3e0761d', 'STAND', 'Heaven''s Door', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-78d5-82f0-2611b3e0761d', 'D', 'B', 'B', 'B', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-78d5-82f0-2611b3e0761d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-78d5-82f0-2611b3e0761d', 'es-ES', 'Portador: Rohan Kishibe

Heaven''s Door convierte a cualquier ser vivo en un libro viviente, exponiendo sus recuerdos, pensamientos y datos físicos como páginas legibles. Rohan puede leer, borrar o reescribir esas páginas para alterar la memoria y el comportamiento de su víctima, aunque su poder ofensivo directo es escaso.', ARRAY['Transformación en Libro: convierte a cualquier persona tocada en un libro que revela su cuerpo y mente como páginas legibles.','Borrado de Memoria: arrancar una página elimina ese recuerdo específico y causa daño físico proporcional a la víctima.','Órdenes Escritas: escribir una instrucción en las páginas obliga a la víctima a cumplirla de forma literal e irresistible.','Candado de Seguridad: puede escribir la orden de no poder atacarle, haciéndolo casi inmune en combate directo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-78d5-82f0-2611b3e0761d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-78d5-82f0-2611b3e0761d', 'en-GB', 'User: Rohan Kishibe

Heaven''s Door turns any living being into a living book, exposing their memories, thoughts and physical data as readable pages. Rohan can read, erase or rewrite those pages to alter his victim''s memory and behaviour, though its direct offensive power is limited.', ARRAY['Book Transformation: turns anyone it touches into a book, revealing their body and mind as readable pages.','Memory Erasure: tearing out a page deletes that specific memory and inflicts proportional physical damage on the victim.','Written Commands: writing an instruction on the pages forces the victim to obey it literally and irresistibly.','Safety Lock: Rohan can write an order preventing anyone from attacking him, making him almost untouchable in direct combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-78d5-82f0-2611b3e0761d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-78d5-82f0-2611b3e0761d', 'ca-ES', 'Portador: Rohan Kishibe

Heaven''s Door converteix qualsevol ésser viu en un llibre vivent, exposant els seus records, pensaments i dades físiques com a pàgines llegibles. En Rohan pot llegir, esborrar o reescriure aquestes pàgines per alterar la memòria i el comportament de la víctima, tot i que el seu poder ofensiu directe és escàs.', ARRAY['Transformació en Llibre: converteix qualsevol persona que toca en un llibre que revela el seu cos i la seva ment com a pàgines llegibles.','Esborrament de Memòria: arrencar una pàgina elimina aquest record concret i causa un dany físic proporcional a la víctima.','Ordres Escrites: escriure una instrucció a les pàgines obliga la víctima a complir-la de manera literal i irresistible.','Cadenat de Seguretat: en Rohan pot escriure una ordre que impedeixi que ningú l''ataqui, cosa que el fa gairebé intocable en combat directe.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-78d5-82f0-2611b3e0761d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Killer Queen (Yoshikage Kira)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-759f-8574-e3f698f99e8c', 'STAND', 'Killer Queen', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-759f-8574-e3f698f99e8c', 'A', 'B', 'D', 'B', 'B', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-759f-8574-e3f698f99e8c', 'es-ES', 'Portador: Yoshikage Kira

Killer Queen es un Stand humanoide de combate cercano capaz de transformar en bomba cualquier objeto o persona que toca con solo un roce. Yoshikage Kira, un asesino en serie que vive obsesionado con el anonimato, la usa para eliminar sin dejar rastro a quien amenaza con descubrir su secreto.', ARRAY['Bomba por Contacto: cualquier cosa que Killer Queen toque queda cargada como explosivo, detonable a voluntad de Kira.','Detonación a Voluntad: Kira decide cuándo activar la bomba implantada, incluso a distancia y con retardo.','Fuerza de Combate Cercano: posee gran potencia física y velocidad que lo hacen letal en enfrentamientos directos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-759f-8574-e3f698f99e8c', 'en-GB', 'User: Yoshikage Kira

Killer Queen is a close-combat humanoid Stand able to turn any object or person it touches into a bomb with the lightest brush. Yoshikage Kira, a serial killer obsessed with anonymity, uses it to erase without a trace anyone who threatens to expose his secret.', ARRAY['Touch Bomb: anything Killer Queen touches becomes a charged explosive that Kira can detonate at will.','Detonation on Command: Kira decides when to trigger the implanted bomb, even remotely and after a delay.','Close-Combat Power: it has great physical strength and speed, making it lethal in direct confrontation.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-759f-8574-e3f698f99e8c', 'ca-ES', 'Portador: Yoshikage Kira

Killer Queen és un Stand humanoide de combat cos a cos capaç de transformar en bomba qualsevol objecte o persona que toca amb un simple contacte. En Yoshikage Kira, un assassí en sèrie obsessionat amb l''anonimat, l''utilitza per eliminar sense deixar rastre qui amenaça de descobrir el seu secret.', ARRAY['Bomba per Contacte: tot allò que Killer Queen toca queda carregat com a explosiu, que en Kira pot detonar quan vulgui.','Detonació a Voluntat: en Kira decideix quan activar la bomba implantada, fins i tot a distància i amb retard.','Força de Combat Cos a Cos: té una gran potència física i velocitat que el fan letal en enfrontaments directes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Killer Queen: Sheer Heart Attack (Yoshikage Kira)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7c16-a26c-5f11e8c2e580', 'STAND', 'Killer Queen: Sheer Heart Attack', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580', 'A', 'B', 'D', 'B', 'B', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580', 'es-ES', 'Portador: Yoshikage Kira

Tras dominar la bomba por contacto, Kira despierta la segunda habilidad de Killer Queen: Sheer Heart Attack, un tanque autónomo en forma de escarabajo que se separa de su mano izquierda. Persigue de forma independiente la fuente de calor más intensa hasta hacerla estallar, mientras Killer Queen conserva intacta su capacidad original de convertir en bomba todo lo que toca.', ARRAY['Bomba por Contacto (heredada): sigue convirtiendo en explosivo cualquier objeto o persona que toque directamente.','Sheer Heart Attack: un tanque-insecto autónomo se desprende de la mano de Killer Queen y actúa de forma independiente.','Persecución Térmica: Sheer Heart Attack rastrea automáticamente la fuente de calor más alta del entorno para atacarla.','Caparazón Indestructible: su coraza resiste ataques directos y perfora cualquier superficie para seguir avanzando hacia el objetivo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580', 'en-GB', 'User: Yoshikage Kira

Having mastered the touch bomb, Kira awakens Killer Queen''s second ability: Sheer Heart Attack, an autonomous beetle-shaped tank that detaches from its left hand. It independently hunts down the strongest heat source until it detonates it, while Killer Queen retains its original ability to turn anything it touches into a bomb.', ARRAY['Touch Bomb (retained): still turns any object or person it directly touches into an explosive.','Sheer Heart Attack: an autonomous insect-like tank detaches from Killer Queen''s hand and acts on its own.','Heat-Seeking Pursuit: Sheer Heart Attack automatically tracks the strongest heat source nearby to attack it.','Indestructible Shell: its armour withstands direct attacks and drills through any surface to keep closing on its target.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580', 'ca-ES', 'Portador: Yoshikage Kira

Després de dominar la bomba per contacte, en Kira desperta la segona habilitat de Killer Queen: Sheer Heart Attack, un tanc autònom en forma d''escarabat que se separa de la seva mà esquerra. Persegueix de manera independent la font de calor més intensa fins a fer-la esclatar, mentre Killer Queen conserva intacta la seva capacitat original de convertir en bomba tot allò que toca.', ARRAY['Bomba per Contacte (heretada): continua convertint en explosiu qualsevol objecte o persona que toqui directament.','Sheer Heart Attack: un tanc-insecte autònom es desprèn de la mà de Killer Queen i actua de manera independent.','Persecució Tèrmica: Sheer Heart Attack rastreja automàticament la font de calor més alta de l''entorn per atacar-la.','Cuirassa Indestructible: la seva cuirassa resisteix atacs directes i perfora qualsevol superfície per seguir avançant cap a l''objectiu.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Killer Queen: Bites the Dust (Yoshikage Kira)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7d53-92d2-1107869ba82b', 'STAND', 'Killer Queen: Bites the Dust', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7d53-92d2-1107869ba82b', 'A', 'B', 'D', 'B', 'B', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d53-92d2-1107869ba82b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d53-92d2-1107869ba82b', 'es-ES', 'Portador: Yoshikage Kira

Acorralado, Kira despierta la tercera y más peligrosa habilidad de Killer Queen: Bites the Dust, implantada en secreto en otra persona que actúa como bomba de tiempo. Si esa persona revela la identidad de Kira, la habilidad detona y reinicia los últimos sesenta minutos en un bucle temporal, borrando toda prueba en su contra, mientras la bomba por contacto y Sheer Heart Attack siguen disponibles.', ARRAY['Bomba por Contacto y Sheer Heart Attack (heredadas): Kira conserva ambas habilidades previas intactas y listas para usar.','Bomba Viviente Secreta: implanta la habilidad en otra persona sin que esta lo sepa, convirtiéndola en un detonante ambulante.','Bucle Temporal de una Hora: al activarse, el tiempo retrocede sesenta minutos y solo Kira y el portador de la bomba recuerdan lo ocurrido.','Borrado de Pruebas: cada repetición del bucle permite a Kira eliminar cualquier indicio que lo incrimine antes de que ocurra.','Vulnerabilidad del Portador: la persona que porta la bomba sufre heridas graves si intenta hablar sobre la identidad de Kira.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d53-92d2-1107869ba82b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d53-92d2-1107869ba82b', 'en-GB', 'User: Yoshikage Kira

Cornered, Kira awakens Killer Queen''s third and most dangerous ability: Bites the Dust, secretly implanted in another person who acts as a living time bomb. If that person reveals Kira''s identity, the ability detonates and resets the last sixty minutes in a time loop, erasing any evidence against him, while the touch bomb and Sheer Heart Attack remain available.', ARRAY['Touch Bomb and Sheer Heart Attack (retained): Kira keeps both previous abilities intact and ready to use.','Secret Living Bomb: implants the ability in another person without their knowledge, turning them into a walking trigger.','One-Hour Time Loop: once triggered, time rewinds sixty minutes and only Kira and the bomb''s carrier remember what happened.','Evidence Erasure: each loop lets Kira remove any clue that could incriminate him before it happens.','Carrier''s Vulnerability: the person carrying the bomb suffers severe injury if they try to speak about Kira''s identity.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d53-92d2-1107869ba82b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d53-92d2-1107869ba82b', 'ca-ES', 'Portador: Yoshikage Kira

Acorralat, en Kira desperta la tercera i més perillosa habilitat de Killer Queen: Bites the Dust, implantada en secret en una altra persona que actua com a bomba de rellotgeria. Si aquesta persona revela la identitat d''en Kira, l''habilitat detona i reinicia els últims seixanta minuts en un bucle temporal, esborrant qualsevol prova en contra seva, mentre la bomba per contacte i Sheer Heart Attack continuen disponibles.', ARRAY['Bomba per Contacte i Sheer Heart Attack (heretades): en Kira conserva les dues habilitats anteriors intactes i llestes per usar.','Bomba Vivent Secreta: implanta l''habilitat en una altra persona sense que aquesta ho sàpiga, convertint-la en un detonant ambulant.','Bucle Temporal d''una Hora: en activar-se, el temps retrocedeix seixanta minuts i només en Kira i el portador de la bomba recorden el que ha passat.','Esborrament de Proves: cada repetició del bucle permet a en Kira eliminar qualsevol indici que l''incrimini abans que passi.','Vulnerabilitat del Portador: la persona que porta la bomba pateix ferides greus si intenta parlar sobre la identitat d''en Kira.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d53-92d2-1107869ba82b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Aqua Necklace (Anjuro Katagiri)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7292-8762-f1cce1596977', 'STAND', 'Aqua Necklace', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7292-8762-f1cce1596977', 'C', 'C', 'A', 'A', 'C', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7292-8762-f1cce1596977')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7292-8762-f1cce1596977', 'es-ES', 'Portador: Anjuro Katagiri

Aqua Necklace es un Stand acuático capaz de adoptar la forma de cualquier líquido, desde agua hasta licor, para infiltrarse en su víctima a través de la boca. Una vez dentro, destroza los órganos internos y toma el control del cuerpo poseído, aunque su poder destructivo directo es limitado.', ARRAY['Transformación Líquida: puede adoptar el aspecto y las propiedades de cualquier líquido, incluidos agua, alcohol o vapor.','Infiltración Oral: se cuela en el cuerpo de la víctima al ser ingerido, evitando así los ataques físicos externos.','Destrucción Interna: una vez dentro, desgarra los órganos y controla los movimientos del cuerpo poseído.','Vulnerabilidad al Caucho: no puede atravesar materiales de goma, lo que le impide penetrar por esa vía.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7292-8762-f1cce1596977')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7292-8762-f1cce1596977', 'en-GB', 'User: Anjuro Katagiri

Aqua Necklace is a watery Stand able to take the form of any liquid, from water to spirits, to slip into its victim through the mouth. Once inside, it tears apart the internal organs and seizes control of the possessed body, though its direct destructive power is limited.', ARRAY['Liquid Transformation: it can take on the shape and properties of any liquid, including water, alcohol or vapour.','Oral Infiltration: it slips into the victim''s body once swallowed, bypassing external physical attacks.','Internal Destruction: once inside, it tears through organs and controls the possessed body''s movements.','Rubber Vulnerability: it cannot pass through rubber materials, blocking that route of entry.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7292-8762-f1cce1596977')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7292-8762-f1cce1596977', 'ca-ES', 'Portador: Anjuro Katagiri

Aqua Necklace és un Stand aquós capaç d''adoptar la forma de qualsevol líquid, des d''aigua fins a licor, per infiltrar-se en la víctima a través de la boca. Un cop a dins, destrossa els òrgans interns i pren el control del cos posseït, tot i que el seu poder destructiu directe és limitat.', ARRAY['Transformació Líquida: pot adoptar l''aspecte i les propietats de qualsevol líquid, inclosos aigua, alcohol o vapor.','Infiltració Oral: s''esmuny dins el cos de la víctima en ser ingerit, evitant així els atacs físics externs.','Destrucció Interna: un cop a dins, esquinça els òrgans i controla els moviments del cos posseït.','Vulnerabilitat al Cautxú: no pot travessar materials de goma, cosa que li impedeix penetrar per aquesta via.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7292-8762-f1cce1596977')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Bad Company (Keicho Nijimura)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7d27-abc8-c958d2e709a7', 'STAND', 'Bad Company', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7d27-abc8-c958d2e709a7', 'B', 'B', 'C', 'B', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d27-abc8-c958d2e709a7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d27-abc8-c958d2e709a7', 'es-ES', 'Portador: Keicho Nijimura

Bad Company es un Stand de colonia formado por un pequeño ejército de soldados, tanques y helicópteros en miniatura que actúan de forma coordinada. Keicho lo despliega para acorralar a un objetivo desde múltiples frentes con disparos, minas y misiles a escala reducida.', ARRAY['Infantería Miniatura: sesenta soldados armados con rifles y cuchillos atacan en formación desde múltiples ángulos a la vez.','Vehículos Blindados: siete tanques en miniatura disparan proyectiles capaces de agrietar paredes y causar heridas profundas.','Apoyo Aéreo: cuatro helicópteros armados sobrevuelan el campo de batalla disparando y lanzando misiles sobre el objetivo.','Muro de Acero: al disparar todas las unidades a la vez pueden interceptar proyectiles entrantes, aunque dejan expuesto al usuario.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d27-abc8-c958d2e709a7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d27-abc8-c958d2e709a7', 'en-GB', 'User: Keicho Nijimura

Bad Company is a colony Stand made up of a miniature army of soldiers, tanks and helicopters acting in coordination. Keicho deploys it to corner a target from multiple fronts with scaled-down gunfire, mines and missiles.', ARRAY['Miniature Infantry: sixty soldiers armed with rifles and knives attack in formation from multiple angles at once.','Armoured Vehicles: seven miniature tanks fire shells capable of cracking walls and inflicting deep wounds.','Air Support: four armed helicopters fly over the battlefield firing guns and launching missiles at the target.','Wall of Steel: firing every unit at once can intercept incoming projectiles, though it leaves the user exposed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d27-abc8-c958d2e709a7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7d27-abc8-c958d2e709a7', 'ca-ES', 'Portador: Keicho Nijimura

Bad Company és un Stand de colònia format per un petit exèrcit de soldats, tancs i helicòpters en miniatura que actuen de manera coordinada. En Keicho el desplega per acorralar un objectiu des de múltiples fronts amb trets, mines i míssils a escala reduïda.', ARRAY['Infanteria en Miniatura: seixanta soldats armats amb rifles i ganivets ataquen en formació des de múltiples angles alhora.','Vehicles Blindats: set tancs en miniatura disparen projectils capaços d''esquerdar parets i causar ferides profundes.','Suport Aeri: quatre helicòpters armats sobrevolen el camp de batalla disparant i llançant míssils sobre l''objectiu.','Mur d''Acer: en disparar totes les unitats alhora poden interceptar projectils entrants, tot i que deixen exposat l''usuari.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7d27-abc8-c958d2e709a7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- The Lock (Tamami Kobayashi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7864-9234-efa187cc315c', 'STAND', 'The Lock', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7864-9234-efa187cc315c', 'E', 'E', 'A', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7864-9234-efa187cc315c')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7864-9234-efa187cc315c', 'es-ES', 'Portador: Tamami Kobayashi

The Lock es un Stand automático sin poder de combate directo, pero terriblemente peligroso: se materializa como un gran candado que se ancla al pecho de la víctima y crece con cada culpa que esta siente. Tamami lo usa para chantajear y estafar, acumulando remordimientos reales o inventados hasta empujar a sus víctimas al suicidio.', ARRAY['Anclaje de Culpa: The Lock se fija al pecho de quien siente remordimiento y permanece ahí de forma indefinida.','Amplificación del Remordimiento: cada nueva culpa, real o inventada, hace crecer y pesar más el candado.','Inducción al Suicidio: si la culpa se vuelve insoportable, empuja a la víctima a quitarse la vida.','Detector de Mentiras: el candado se queda quieto ante una mentira y se mueve ante la verdad.','Multiobjetivo Silencioso: Tamami puede activarlo sobre varias personas a la vez sin que lo perciban.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7864-9234-efa187cc315c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7864-9234-efa187cc315c', 'en-GB', 'User: Tamami Kobayashi

The Lock is an automatic Stand with no direct combat power, yet dreadfully dangerous: it manifests as a large padlock that fastens onto the victim''s chest and grows heavier with every ounce of guilt they feel. Tamami uses it to blackmail and swindle people, piling up real or invented wrongdoings until the victim is driven to suicide.', ARRAY['Guilt Anchor: The Lock fastens to the chest of anyone feeling remorse and stays there indefinitely.','Guilt Amplification: every fresh accusation, real or invented, makes the padlock grow bigger and heavier.','Suicide Inducement: once guilt becomes unbearable, it drives the victim to take their own life.','Lie Detector: the padlock stays still when it senses a lie and moves when it senses honesty.','Silent Multi-Targeting: Tamami can activate it on several people at once without them noticing.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7864-9234-efa187cc315c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7864-9234-efa187cc315c', 'ca-ES', 'Portador: Tamami Kobayashi

The Lock és un Stand automàtic sense poder de combat directe, però terriblement perillós: es materialitza com un gran cadenat que s''ancora al pit de la víctima i creix amb cada culpa que aquesta sent. En Tamami l''utilitza per xantatjar i estafar, acumulant remordiments reals o inventats fins a empènyer les víctimes al suïcidi.', ARRAY['Ancoratge de Culpa: The Lock es fixa al pit de qui sent remordiment i hi roman de manera indefinida.','Amplificació del Remordiment: cada nova culpa, real o inventada, fa créixer i pesar més el cadenat.','Inducció al Suïcidi: si la culpa es torna insuportable, empeny la víctima a llevar-se la vida.','Detector de Mentides: el cadenat es queda quiet davant una mentida i es mou davant la veritat.','Multiobjectiu Silenciós: en Tamami el pot activar sobre diverses persones alhora sense que se n''adonin.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7864-9234-efa187cc315c')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Surface (Toshikazu Hazamada)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-70a1-887d-425d9f32061b', 'STAND', 'Surface', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-70a1-887d-425d9f32061b', 'B', 'B', 'C', 'B', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-70a1-887d-425d9f32061b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-70a1-887d-425d9f32061b', 'es-ES', 'Portador: Toshikazu Hazamada

Surface es un maniquí de madera que, al ser tocado por alguien, copia a la perfección su aspecto, voz y gestos, distinguible solo por un tornillo en la frente. Su núcleo de madera protege a Hazamada de cualquier daño que reciba la copia, y puede obligar al imitado a sincronizar sus movimientos con los suyos.', ARRAY['Mimetismo Perfecto: copia el aspecto, la voz y los gestos de la persona que toca, huellas incluidas.','Núcleo de Madera: el daño sobre Surface no afecta a Hazamada, y las partes cortadas revelan madera.','Sincronización Forzada: obliga al copiado a repetir sus movimientos, impidiéndole atacarlo de frente.','Réplica de Personalidad: adopta el comportamiento de la persona imitada, pero sigue siendo leal a su portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-70a1-887d-425d9f32061b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-70a1-887d-425d9f32061b', 'en-GB', 'User: Toshikazu Hazamada

Surface is a wooden mannequin that, once touched by someone, perfectly copies their appearance, voice and mannerisms, distinguishable only by a screw on its forehead. Its wooden core shields Hazamada from any damage the copy takes, and it can force the copied person to mirror its own movements.', ARRAY['Perfect Mimicry: copies the appearance, voice and mannerisms of whoever it touches, fingerprints included.','Wooden Core: damage dealt to Surface never harms Hazamada, and severed parts reveal plain wood.','Forced Synchronisation: compels the copied person to mirror its movements, stopping them attacking head-on.','Personality Replication: adopts the copied person''s behaviour while staying loyal to its own user.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-70a1-887d-425d9f32061b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-70a1-887d-425d9f32061b', 'ca-ES', 'Portador: Toshikazu Hazamada

Surface és un maniquí de fusta que, quan algú el toca, copia a la perfecció el seu aspecte, veu i gestos, distingible només per un cargol al front. El seu nucli de fusta protegeix l''Hazamada de qualsevol dany que rebi la còpia, i pot obligar la persona imitada a sincronitzar els moviments amb els seus.', ARRAY['Mimetisme Perfecte: copia l''aspecte, la veu i els gestos de qui el toca, empremtes incloses.','Nucli de Fusta: el dany sobre Surface no afecta l''Hazamada, i les parts tallades mostren fusta.','Sincronització Forçada: obliga la persona copiada a repetir els seus moviments, impedint-li atacar-lo de front.','Rèplica de Personalitat: adopta el comportament de la persona imitada, tot i que roman lleial al seu portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-70a1-887d-425d9f32061b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Love Deluxe (Yukako Yamagishi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb0-7c47-ba95-c4cb36d44649', 'STAND', 'Love Deluxe', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb0-7c47-ba95-c4cb36d44649', 'B', 'B', 'C', 'A', 'E', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c47-ba95-c4cb36d44649')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c47-ba95-c4cb36d44649', 'es-ES', 'Portadora: Yukako Yamagishi

Love Deluxe da vida al cabello de Yukako, permitiéndole alargarlo a longitudes extraordinarias y controlarlo con la mente como si fueran tentáculos. Con él puede inmovilizar objetivos, implantar mechones en el cuerpo de otros para controlarlos a distancia y golpear desde varios metros de rango.', ARRAY['Cabello Vivo: controla su propio pelo con la mente, dotándolo de fuerza y destreza notables.','Crecimiento Extremo: lo alarga hasta longitudes que pueden atravesar paredes enteras de una casa.','Implantación de Mechones: introduce hebras en el cuerpo de un objetivo para controlarlo a distancia.','Proyección a Distancia: dispara mechones de cabello para golpear e inmovilizar sin acercarse.','Interacción con el Entorno: extiende su pelo hasta el fuego u otros peligros para usarlos contra el rival.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c47-ba95-c4cb36d44649')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c47-ba95-c4cb36d44649', 'en-GB', 'User: Yukako Yamagishi

Love Deluxe brings Yukako''s hair to life, letting her grow it to extraordinary lengths and control it with her mind like a set of tentacles. With it she can immobilise targets, implant strands into other people''s bodies to control them remotely, and strike from several metres away.', ARRAY['Living Hair: controls her own hair with her mind, granting it remarkable strength and dexterity.','Extreme Growth: lengthens her hair enough to pierce through an entire house''s walls.','Strand Implantation: plants strands of hair inside a target''s body to control them remotely.','Ranged Projection: shoots strands of hair to strike and restrain targets without closing the distance.','Environmental Interaction: extends her hair towards fire or other hazards to turn them against a foe.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c47-ba95-c4cb36d44649')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb0-7c47-ba95-c4cb36d44649', 'ca-ES', 'Portadora: Yukako Yamagishi

Love Deluxe dona vida al cabell de na Yukako, cosa que li permet allargar-lo fins a longituds extraordinàries i controlar-lo amb la ment com si fossin tentacles. Amb ell pot immobilitzar objectius, implantar mechons al cos d''altres persones per controlar-les a distància i colpejar des de diversos metres de distància.', ARRAY['Cabell Viu: controla el seu propi cabell amb la ment, dotant-lo d''una força i destresa notables.','Creixement Extrem: l''allarga fins a longituds que poden travessar les parets senceres d''una casa.','Implantació de Mechons: introdueix fils al cos d''un objectiu per controlar-lo a distància.','Projecció a Distància: dispara fils de cabell per colpejar i immobilitzar sense apropar-s''hi.','Interacció amb l''Entorn: estén el cabell fins al foc o altres perills per fer-los servir contra el rival.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb0-7c47-ba95-c4cb36d44649')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Red Hot Chili Pepper (Akira Otoishi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-75f2-be6b-961f065a805d', 'STAND', 'Red Hot Chili Pepper', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-75f2-be6b-961f065a805d', 'A', 'A', 'A', 'A', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-75f2-be6b-961f065a805d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-75f2-be6b-961f065a805d', 'es-ES', 'Portador: Akira Otoishi

Red Hot Chili Pepper es un Stand basado en electricidad cuya peligrosidad se dispara cerca de cualquier fuente eléctrica. Puede absorber corriente para curarse y volverse más fuerte, viajar a través de cables y electrodomésticos, y convertir objetos o personas en electricidad pura para moverlos a voluntad.', ARRAY['Absorción Eléctrica: se alimenta de corriente cercana para curar heridas y aumentar su fuerza y velocidad.','Viaje por Cableado: se desplaza dentro de cables, baterías y aparatos eléctricos para ocultarse o moverse rápido.','Electrificación de Objetos: convierte en electricidad lo que toca para transportarlo o dañarlo a distancia.','Dependencia Energética: sin electricidad cercana se debilita, oxida y acaba amenazando la vida de Otoishi.','Vulnerabilidad al Agua: el contacto con el agua dispersa su electricidad y lo deja inutilizado.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-75f2-be6b-961f065a805d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-75f2-be6b-961f065a805d', 'en-GB', 'User: Akira Otoishi

Red Hot Chili Pepper is an electricity-based Stand whose danger spikes near any power source. It can absorb current to heal itself and grow stronger, travel through cables and appliances, and turn objects or people into pure electricity to move them at will.', ARRAY['Electrical Absorption: feeds on nearby current to heal wounds and boost its own strength and speed.','Cable Travel: moves through wires, batteries and electrical appliances to hide or travel rapidly.','Object Electrification: turns whatever it touches into electricity to carry or harm it at range.','Power Dependency: without a nearby power source it weakens, rusts, and eventually threatens Otoishi''s life.','Water Vulnerability: contact with water disperses its electricity and shuts it down completely.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-75f2-be6b-961f065a805d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-75f2-be6b-961f065a805d', 'ca-ES', 'Portador: Akira Otoishi

Red Hot Chili Pepper és un Stand basat en electricitat la perillositat del qual s''dispara a prop de qualsevol font elèctrica. Pot absorbir corrent per curar-se i tornar-se més fort, viatjar per cables i electrodomèstics, i convertir objectes o persones en electricitat pura per moure''ls a voluntat.', ARRAY['Absorció Elèctrica: s''alimenta de corrent propera per curar ferides i augmentar la seva força i velocitat.','Viatge pel Cablejat: es desplaça dins de cables, bateries i aparells elèctrics per amagar-se o moure''s ràpid.','Electrificació d''Objectes: converteix en electricitat allò que toca per transportar-lo o fer-lo mal a distància.','Dependència Energètica: sense electricitat propera s''afebleix, s''oxida i acaba amenaçant la vida de l''Otoishi.','Vulnerabilitat a l''Aigua: el contacte amb l''aigua dispersa la seva electricitat i el deixa inutilitzat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-75f2-be6b-961f065a805d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Ratt (Bug-Eaten)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-790a-a1e6-1ffcc7906302', 'STAND', 'Ratt', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-790a-a1e6-1ffcc7906302', 'B', 'C', 'D', 'B', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-790a-a1e6-1ffcc7906302')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-790a-a1e6-1ffcc7906302', 'es-ES', 'Portador: Bug-Eaten

Ratt es el Stand de una rata a la que Akira Otoishi disparó con su Arco y Flecha, dándole forma de pequeña torreta que dispara dardos corrosivos. Incapaz de moverse por sí mismo, compensa su nula movilidad con un tamaño diminuto que lo hace casi imposible de detectar antes de que dispare.', ARRAY['Dardos Corrosivos: dispara proyectiles que derriten carne y hueso al impactar, casi imposibles de curar.','Alcance Extendido: puede disparar hasta 60 metros de distancia y hacerlo en ráfagas.','Rebote Táctico: hace rebotar los dardos en obstáculos para alcanzar objetivos desde ángulos ciegos.','Fusión de Carne: los cuerpos alcanzados por varios dardos pueden fundirse entre sí en una masa gelatinosa.','Camuflaje por Tamaño: su forma de torreta diminuta le permite pasar desapercibido antes de atacar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-790a-a1e6-1ffcc7906302')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-790a-a1e6-1ffcc7906302', 'en-GB', 'User: Bug-Eaten

Ratt is the Stand of a rat shot by Akira Otoishi''s Bow and Arrow, taking the shape of a tiny turret that fires corrosive darts. Unable to move on its own, it makes up for its complete lack of mobility with a minuscule size that makes it nearly impossible to spot before it fires.', ARRAY['Corrosive Darts: fires projectiles that melt flesh and bone on impact, leaving wounds that are nearly impossible to heal.','Extended Range: can fire darts up to 60 metres away, including in bursts.','Tactical Ricochet: bounces darts off obstacles to hit targets from blind spots.','Flesh Fusion: bodies struck by several darts can melt and fuse together into a gelatinous mass.','Size Camouflage: its tiny turret shape lets it go unnoticed until it opens fire.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-790a-a1e6-1ffcc7906302')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-790a-a1e6-1ffcc7906302', 'ca-ES', 'Portador: Bug-Eaten

Ratt és el Stand d''una rata a qui l''Akira Otoishi va disparar amb el seu Arc i Fletxa, amb forma de petita torreta que dispara dards corrosius. Incapaç de moure''s per si mateix, compensa la seva nul·la mobilitat amb una mida diminuta que el fa gairebé impossible de detectar abans que disparí.', ARRAY['Dards Corrosius: dispara projectils que fonen carn i os en impactar, gairebé impossibles de curar.','Abast Estès: pot disparar fins a 60 metres de distància i fer-ho en ràfegues.','Rebot Tàctic: fa rebotar els dards en obstacles per encertar objectius des d''angles cecs.','Fusió de Carn: els cossos encertats per diversos dards es poden fondre entre ells en una massa gelatinosa.','Camuflatge per Mida: la seva forma de torreta diminuta li permet passar desapercebut abans d''atacar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-790a-a1e6-1ffcc7906302')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Harvest (Shigekiyo "Shigechi" Yangu)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7d57-97e6-c4efc955d0ca', 'STAND', 'Harvest', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca', 'E', 'B', 'A', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca', 'es-ES', 'Portador: Shigekiyo "Shigechi" Yangu

Harvest es un Stand de unidades múltiples formado por cerca de 500 pequeñas criaturas que pueden moverse libremente por todo Morioh. Individualmente son débiles, pero Shigechi las usa en enjambre para recolectar dinero, transportarlo a él mismo y atacar los puntos débiles de un rival.', ARRAY['Enjambre de Unidades: unas 500 criaturas diminutas actúan solas o coordinadas por toda la ciudad.','Recolección a Distancia: recoge monedas, cupones y objetos pequeños allá donde estén.','Ataque en Grupo: se agrupan sobre un objetivo para atacar puntos débiles como los ojos.','Inyección con Aguijón: cada unidad puede inyectar líquidos directamente en el torrente sanguíneo del rival.','Transporte y Camuflaje: pueden cargar a Shigechi a gran velocidad o camuflarlo con hojas y escombros.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca', 'en-GB', 'User: Shigekiyo "Shigechi" Yangu

Harvest is a multi-unit Stand made up of roughly 500 tiny creatures that can roam freely across Morioh. Individually weak, Shigechi uses them as a swarm to collect money, carry himself around, and strike at an opponent''s weak points.', ARRAY['Unit Swarm: around 500 tiny creatures act alone or together across the whole town.','Remote Collection: gathers coins, coupons and other small items wherever they are.','Group Attack: swarms a target to strike weak points such as the eyes.','Stinger Injection: each unit can inject liquid directly into an opponent''s bloodstream.','Transport and Camouflage: can carry Shigechi at speed or disguise him with leaves and debris.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca', 'ca-ES', 'Portador: Shigekiyo "Shigechi" Yangu

Harvest és un Stand de múltiples unitats format per unes 500 petites criatures que es poden moure lliurement per tot Morioh. Individualment són febles, però en Shigechi les fa servir en eixam per recollir diners, transportar-lo a ell mateix i atacar els punts febles d''un rival.', ARRAY['Eixam d''Unitats: unes 500 criatures diminutes actuen soles o coordinades per tota la ciutat.','Recol·lecció a Distància: recull monedes, cupons i objectes petits allà on siguin.','Atac en Grup: s''agrupen sobre un objectiu per atacar punts febles com els ulls.','Injecció amb Fibló: cada unitat pot injectar líquids directament al torrent sanguini del rival.','Transport i Camuflatge: poden carregar en Shigechi a gran velocitat o camuflar-lo amb fulles i runa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Atom Heart Father (Yoshihiro Kira)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7f5d-bf44-033c52faf367', 'STAND', 'Atom Heart Father', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7f5d-bf44-033c52faf367', 'E', 'E', 'NULL', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f5d-bf44-033c52faf367')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f5d-bf44-033c52faf367', 'es-ES', 'Portador: Yoshihiro Kira

Es un Stand ligado a una cámara Polaroid maldita que permite al fantasma de Yoshihiro Kira existir y moverse dentro de las fotografías que revela. Cualquiera puede usar la cámara, lo que lo convierte en un arma poderosa pero fácil de sabotear.', ARRAY['Fotografía Sobrenatural: La cámara revela y hace visible el fantasma de Yoshihiro únicamente dentro de las fotos que toma.','Interacción desde la Imagen: Yoshihiro puede tocar e interactuar con cualquier objeto o persona capturado dentro del encuadre fotografiado.','Muro Invisible: Crea una barrera indestructible en los bordes de la foto que atrapa a quien queda dentro de ella.','Persistencia tras la Cámara: Las fotografías reveladas conservan su efecto de forma indefinida aunque la cámara original sea destruida.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f5d-bf44-033c52faf367')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f5d-bf44-033c52faf367', 'en-GB', 'User: Yoshihiro Kira

It is a Stand bound to a cursed Polaroid camera that lets Yoshihiro Kira''s ghost exist and move within the photographs it develops. Anyone can operate the camera, which makes it a powerful weapon but easy to sabotage.', ARRAY['Supernatural Photography: The camera reveals and makes Yoshihiro''s ghost visible only inside the pictures it takes.','Interaction from the Image: Yoshihiro can touch and interact with anything or anyone captured within the photographed frame.','Invisible Wall: Creates an unbreakable barrier at the photo''s edges that traps whoever is caught inside it.','Persistence Beyond the Camera: Developed photographs keep their effect indefinitely even if the original camera is destroyed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f5d-bf44-033c52faf367')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f5d-bf44-033c52faf367', 'ca-ES', 'Portador: Yoshihiro Kira

És un Stand lligat a una càmera Polaroid maleïda que permet al fantasma d''en Yoshihiro Kira existir i moure''s dins les fotografies que revela. Qualsevol pot fer servir la càmera, cosa que el converteix en una arma poderosa però fàcil de sabotejar.', ARRAY['Fotografia Sobrenatural: La càmera revela i fa visible el fantasma d''en Yoshihiro únicament dins les fotos que fa.','Interacció des de la Imatge: En Yoshihiro pot tocar i interactuar amb qualsevol objecte o persona capturats dins l''enquadrament fotografiat.','Mur Invisible: Crea una barrera indestructible a les vores de la foto que atrapa qui hi queda a dins.','Persistència més enllà de la Càmera: Les fotografies revelades conserven el seu efecte de manera indefinida encara que es destrueixi la càmera original.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f5d-bf44-033c52faf367')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Boy II Man (Ken Oyanagi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-722d-a1c9-e6d482e59fc0', 'STAND', 'Boy II Man', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0', 'C', 'B', 'C', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0', 'es-ES', 'Portador: Ken Oyanagi

Es un Stand ligado a un ritual de piedra, papel o tijera: si Ken gana una partida, arrebata un tercio del poder del Stand rival y lo incorpora temporalmente al suyo. Su gran debilidad es que en el juego está permitido hacer trampas.', ARRAY['Robo de Poder por Manga China: Al ganar una ronda, extrae físicamente un tercio de la capacidad del Stand contrario.','Absorción Temporal: Ken puede usar durante un tiempo limitado las habilidades robadas como si fueran propias.','Reglas Permisivas: El juego admite trampas, lo que permite a un rival astuto recuperar la ventaja perdida.','Cuerpo Blindado: Su armadura remachada le da una resistencia física notable en combate cuerpo a cuerpo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0', 'en-GB', 'User: Ken Oyanagi

It is a Stand bound to a rock-paper-scissors ritual: if Ken wins a round, it seizes a third of the rival Stand''s power and folds it into his own, temporarily. Its major weakness is that cheating is allowed in the game.', ARRAY['Power Theft via Hand Sign: Winning a round physically extracts a third of the opposing Stand''s capability.','Temporary Absorption: Ken can wield the stolen abilities as his own for a limited time.','Permissive Rules: Cheating is allowed in the game, letting a cunning rival reclaim their lost edge.','Armoured Body: Its riveted armour gives it notable physical resilience in close combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0', 'ca-ES', 'Portador: Ken Oyanagi

És un Stand lligat a un ritual de pedra, paper o tisores: si en Ken guanya una partida, arrabassa un terç del poder de l''Stand rival i l''incorpora temporalment al seu. La seva gran feblesa és que en el joc es permet fer trampes.', ARRAY['Robatori de Poder pel Joc de Mans: En guanyar una ronda, extreu físicament un terç de la capacitat de l''Stand contrari.','Absorció Temporal: En Ken pot fer servir durant un temps limitat les habilitats robades com si fossin pròpies.','Regles Permissives: El joc admet trampes, cosa que permet a un rival astut recuperar l''avantatge perdut.','Cos Blindat: La seva armadura reblada li dona una resistència física notable en combat cos a cos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Highway Star (Yuya Fungami)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7c9b-b004-89f8abc9c16c', 'STAND', 'Highway Star', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c', 'C', 'B', 'A', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c', 'es-ES', 'Portador: Yuya Fungami

Es un Stand de largo alcance con forma humanoide oscura que persigue a sus víctimas por el olfato y actúa de forma casi autónoma. Absorbe nutrientes de quien toca y transfiere esa energía a Yuya para curarlo, mientras se desplaza dividido en fragmentos difíciles de esquivar.', ARRAY['Rastreo Olfativo Permanente: Memoriza el olor de un objetivo y puede seguirlo indefinidamente a cualquier distancia.','Fragmentación Corporal: Se divide en segmentos planos que se mueven de forma independiente para colarse por huecos y atacar desde varios ángulos.','Absorción de Nutrientes: Al tocar a la víctima, drena su energía vital y la transfiere directamente a Yuya para curarlo.','Proyección de Ilusiones: Genera visiones engañosas adaptadas a cada víctima para atraerla hacia una trampa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c', 'en-GB', 'User: Yuya Fungami

It is a long-range Stand with a dark humanoid form that hunts its victims by scent and acts almost autonomously. It absorbs nutrients from anyone it touches and transfers that energy to Yuya to heal him, while moving split into fragments that are hard to dodge.', ARRAY['Permanent Scent Tracking: Memorises a target''s scent and can follow it indefinitely across any distance.','Bodily Fragmentation: Splits into flat segments that move independently to slip through gaps and strike from several angles.','Nutrient Absorption: Drains a victim''s life energy on contact and transfers it straight to Yuya to heal him.','Illusion Projection: Generates deceptive visions tailored to each victim to lure them into a trap.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c', 'ca-ES', 'Portador: Yuya Fungami

És un Stand de llarg abast amb forma humanoide fosca que persegueix les seves víctimes per l''olfacte i actua de manera gairebé autònoma. Absorbeix nutrients de qui toca i transfereix aquesta energia a en Yuya per curar-lo, mentre es desplaça dividit en fragments difícils d''esquivar.', ARRAY['Rastreig Olfactiu Permanent: Memoritza l''olor d''un objectiu i el pot seguir indefinidament a qualsevol distància.','Fragmentació Corporal: Es divideix en segments plans que es mouen de manera independent per colar-se per buits i atacar des de diversos angles.','Absorció de Nutrients: En tocar la víctima, li drena l''energia vital i la transfereix directament a en Yuya per curar-lo.','Projecció d''Il·lusions: Genera visions enganyoses adaptades a cada víctima per atraure-la cap a un parany.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Super Fly (Toyohiro Kanedaichi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7967-8f01-02ccef87b43f', 'STAND', 'Super Fly', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7967-8f01-02ccef87b43f', 'E', 'E', 'NULL', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7967-8f01-02ccef87b43f')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7967-8f01-02ccef87b43f', 'es-ES', 'Portador: Toyohiro Kanedaichi

Es un Stand automático ligado a una enorme torre de transmisión de 38 metros, visible para cualquiera y capaz de actuar sin intervención directa de Toyohiro. Encierra a una persona en su interior y devuelve amplificado cualquier daño recibido contra su propia estructura.', ARRAY['Encierro en la Torre: Atrapa a una persona dentro de la estructura y la recubre de acero para impedir que escape.','Reflejo de Daño: Cualquier golpe o disparo contra la torre rebota devuelto contra quien lo lanzó.','Expulsión de Objetos: Todo objeto extraño insertado en la estructura es expulsado violentamente hacia fuera.','Persistencia sin Usuario: La torre sigue funcionando como Stand incluso si Toyohiro muere.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7967-8f01-02ccef87b43f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7967-8f01-02ccef87b43f', 'en-GB', 'User: Toyohiro Kanedaichi

It is an automatic Stand bound to a huge 38-metre transmission tower, visible to anyone and able to act without Toyohiro''s direct input. It traps a person inside and hurls back any damage done to its own structure, amplified.', ARRAY['Tower Confinement: Traps a person inside the structure and coats them in steel to stop them escaping.','Damage Reflection: Any blow or shot against the tower rebounds straight back at whoever caused it.','Object Ejection: Any foreign object inserted into the structure is violently thrown back out.','Persistence Without a User: The tower keeps functioning as a Stand even if Toyohiro dies.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7967-8f01-02ccef87b43f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7967-8f01-02ccef87b43f', 'ca-ES', 'Portador: Toyohiro Kanedaichi

És un Stand automàtic lligat a una enorme torre de transmissió de 38 metres, visible per a tothom i capaç d''actuar sense la intervenció directa d''en Toyohiro. Empresona una persona al seu interior i retorna amplificat qualsevol dany rebut contra la seva pròpia estructura.', ARRAY['Empresonament a la Torre: Atrapa una persona dins l''estructura i la recobreix d''acer per impedir que fugi.','Reflex de Dany: Qualsevol cop o tret contra la torre rebota i torna cap a qui l''ha llançat.','Expulsió d''Objectes: Tot objecte estrany inserit a l''estructura n''és expulsat violentament cap enfora.','Persistència sense Usuari: La torre continua funcionant com a Stand fins i tot si en Toyohiro mor.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7967-8f01-02ccef87b43f')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Enigma (Terunosuke Miyamoto)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7ee8-bd4b-af61be09f301', 'STAND', 'Enigma', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7ee8-bd4b-af61be09f301', 'E', 'E', 'C', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7ee8-bd4b-af61be09f301')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7ee8-bd4b-af61be09f301', 'es-ES', 'Portador: Terunosuke Miyamoto

Es un Stand capaz de convertir en papel a objetos y personas, incluido su propio usuario, sellándolos en el interior de una hoja doblada. Para atrapar a alguien, Terunosuke debe reconocer primero el gesto involuntario que esa persona hace cuando siente miedo.', ARRAY['Sellado en Papel: Convierte en una hoja de papel a cualquier objeto o persona cuyo gesto de miedo haya identificado.','Liberación por Desdoblado: Lo sellado recupera su forma original en el instante en que se desdobla el papel.','Daño Compartido: Cualquier daño hecho al papel se refleja de igual manera en lo que contiene.','Transformación Propia: Terunosuke puede convertirse él mismo en papel al instante para esquivar ataques.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7ee8-bd4b-af61be09f301')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7ee8-bd4b-af61be09f301', 'en-GB', 'User: Terunosuke Miyamoto

It is a Stand able to turn objects and people, including its own user, into paper, sealing them inside a folded sheet. To trap someone, Terunosuke must first recognise the involuntary gesture that person makes when afraid.', ARRAY['Paper Sealing: Turns any object or person whose fear tell has been identified into a sheet of paper.','Release by Unfolding: Whatever is sealed regains its original form the instant the paper is unfolded.','Shared Damage: Any damage done to the paper is reflected equally onto whatever it contains.','Self Transformation: Terunosuke can instantly turn himself into paper to dodge attacks.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7ee8-bd4b-af61be09f301')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7ee8-bd4b-af61be09f301', 'ca-ES', 'Portador: Terunosuke Miyamoto

És un Stand capaç de convertir en paper objectes i persones, inclòs el seu propi usuari, segellant-los dins un full doblegat. Per atrapar algú, en Terunosuke ha de reconèixer primer el gest involuntari que fa aquesta persona quan sent por.', ARRAY['Segellat en Paper: Converteix en un full de paper qualsevol objecte o persona el gest de por del qual s''hagi identificat.','Alliberament en Desplegar-lo: Allò segellat recupera la seva forma original en l''instant que es desplega el paper.','Dany Compartit: Qualsevol dany fet al paper es reflecteix igualment en allò que conté.','Transformació Pròpia: En Terunosuke pot convertir-se ell mateix en paper a l''instant per esquivar atacs.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7ee8-bd4b-af61be09f301')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Cheap Trick (Masazo Kinoto)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-71d5-af5f-e2bae8cc4879', 'STAND', 'Cheap Trick', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879', 'E', 'E', 'E', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879', 'es-ES', 'Portador: Masazo Kinoto

Es un Stand parásito que se aferra a la espalda de su portador y solo se vuelve visible para él, incapaz de ser retirado sin arrancarle la piel. Si otra persona llega a ver la espalda, el portador muere y el Stand pasa a poseer al testigo, como le ocurrió a Rohan Kishibe.', ARRAY['Adherencia Parasitaria: Se fija a la espalda del portador con dedos de ventosa imposibles de despegar sin matarlo.','Transferencia Letal: Si alguien más ve la espalda del portador, este muere y el Stand se traslada al testigo.','Habla Universal: Puede conversar con cualquiera, incluidos animales y personas sin Stand.','Acoso Psicológico: Hostiga constantemente a su portador para desgastar su mente y provocar que otros vean su espalda.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879', 'en-GB', 'User: Masazo Kinoto

It is a parasitic Stand that clings to its user''s back and is only visible to him, unable to be removed without tearing off his skin. If someone else sees the back, the user dies and the Stand goes on to possess the witness, as happened to Rohan Kishibe.', ARRAY['Parasitic Grip: Fixes itself to the user''s back with suction-cup fingers that cannot be peeled off without killing him.','Lethal Transfer: If anyone else sees the user''s back, he dies and the Stand moves on to the witness.','Universal Speech: Able to converse with anyone, including animals and people without a Stand.','Psychological Harassment: Constantly torments its user to wear down his mind and provoke others into seeing his back.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879', 'ca-ES', 'Portador: Masazo Kinoto

És un Stand paràsit que s''enganxa a l''esquena del seu portador i només es fa visible per a ell, impossible de retirar sense arrencar-li la pell. Si algú altre arriba a veure-li l''esquena, el portador mor i l''Stand passa a posseir el testimoni, com li va passar a en Rohan Kishibe.', ARRAY['Adherència Paràsita: Es fixa a l''esquena del portador amb dits de ventosa impossibles de desenganxar sense matar-lo.','Transferència Letal: Si algú altre veu l''esquena del portador, aquest mor i l''Stand es trasllada al testimoni.','Parla Universal: Pot conversar amb qualsevol, inclosos animals i persones sense Stand.','Assetjament Psicològic: Assetja constantment el portador per desgastar-li la ment i provocar que d''altres li vegin l''esquena.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Pearl Jam (Tonio Trussardi)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8', 'STAND', 'Pearl Jam', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8', 'E', 'C', 'B', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8', 'es-ES', 'Portador: Tonio Trussardi

Pearl Jam es un Stand colonia formado por diminutas criaturas con forma de tomate y perla, invisible mientras está unido a los platos de Tonio. Se infunde en su alta cocina para curar cualquier dolencia del comensal según los ingredientes usados.', ARRAY['Diagnóstico por Lectura de Manos: Tonio identifica la dolencia exacta del cliente con solo tocarle la mano antes de cocinar.','Cocina Curativa: Infunde Pearl Jam en el plato para sanar una enfermedad concreta al ser ingerido.','Especificidad del Remedio: Cada plato solo afecta a quien padece la dolencia para la que fue preparado, sin efecto en los demás comensales.','Extracción de Órganos Dañados: En casos graves, expulsa del cuerpo el órgano enfermo y lo sustituye por uno sano usando los nutrientes del plato.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8', 'en-GB', 'User: Tonio Trussardi

Pearl Jam is a colony Stand made up of tiny tomato-and-pearl-onion creatures, invisible while bound to Tonio''s dishes. He infuses it into his haute cuisine to cure a diner''s ailment depending on the ingredients used.', ARRAY['Diagnosis by Palm Reading: Tonio identifies a customer''s exact ailment just by touching their hand before cooking.','Curative Cooking: Infuses Pearl Jam into a dish to heal a specific illness once eaten.','Remedy Specificity: Each dish only affects someone suffering from the ailment it was prepared for, leaving other diners unaffected.','Diseased Organ Extraction: In severe cases, it expels the sick organ from the body and replaces it with a healthy one using the dish''s nutrients.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8', 'ca-ES', 'Portador: Tonio Trussardi

Pearl Jam és un Stand colònia format per petites criatures amb forma de tomàquet i perla, invisible mentre està unit als plats d''en Tonio. S''infon en la seva alta cuina per curar qualsevol malaltia del comensal segons els ingredients usats.', ARRAY['Diagnòstic per Lectura de Mans: En Tonio identifica la malaltia exacta del client només tocant-li la mà abans de cuinar.','Cuina Curativa: Infon Pearl Jam al plat per guarir una malaltia concreta en ser ingerit.','Especificitat del Remei: Cada plat només afecta qui pateix la malaltia per a la qual va ser preparat, sense efecte en la resta de comensals.','Extracció d''Òrgans Malalts: En casos greus, expulsa del cos l''òrgan malalt i el substitueix per un de sa fent servir els nutrients del plat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Achtung Baby (Shizuka Joestar)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-799e-95d4-314231b37640', 'STAND', 'Achtung Baby', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-799e-95d4-314231b37640', 'E', 'E', 'NULL', 'A', 'E', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-799e-95d4-314231b37640')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-799e-95d4-314231b37640', 'es-ES', 'Portadora: Shizuka Joestar

Achtung Baby está ligado al propio cuerpo de la bebé Shizuka y la vuelve invisible junto con su entorno inmediato. Su alcance crece de forma involuntaria cuando la niña se asusta o se altera.', ARRAY['Invisibilidad Corporal: Vuelve invisible a Shizuka mientras permanece tranquila.','Expansión por Estrés: El área de invisibilidad se agranda automáticamente si la bebé se altera o llora.','Persistencia del Efecto: Los objetos o personas afectados siguen invisibles incluso al salir del alcance del Stand.','Vulnerabilidad Sensorial: Pese a la invisibilidad, el sonido o el olor aún permiten detectar lo oculto.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-799e-95d4-314231b37640')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-799e-95d4-314231b37640', 'en-GB', 'User: Shizuka Joestar

Achtung Baby is bound to baby Shizuka''s own body and turns her invisible along with her immediate surroundings. Its range grows involuntarily whenever the infant gets startled or upset.', ARRAY['Bodily Invisibility: Turns Shizuka invisible while she stays calm.','Stress Expansion: The invisibility field automatically widens if the baby gets upset or cries.','Lingering Effect: Affected objects or people stay invisible even after leaving the Stand''s range.','Sensory Vulnerability: Despite the invisibility, sound or smell can still reveal what''s hidden.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-799e-95d4-314231b37640')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-799e-95d4-314231b37640', 'ca-ES', 'Portadora: Shizuka Joestar

Achtung Baby està lligat al propi cos de la nadó Shizuka i la torna invisible juntament amb el seu entorn immediat. El seu abast creix involuntàriament quan la nena s''espanta o s''altera.', ARRAY['Invisibilitat Corporal: Torna invisible la Shizuka mentre es manté tranquil·la.','Expansió per Estrès: L''àrea d''invisibilitat s''engrandeix automàticament si la nadó s''altera o plora.','Persistència de l''Efecte: Els objectes o persones afectats continuen invisibles fins i tot en sortir de l''abast del Stand.','Vulnerabilitat Sensorial: Malgrat la invisibilitat, el so o l''olor encara permeten detectar allò amagat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-799e-95d4-314231b37640')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Earth Wind and Fire (Mikitaka Hazekura)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7d90-99c1-3d58d5912a1b', 'STAND', 'Earth Wind and Fire', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b', 'C', 'C', 'NULL', 'A', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b', 'es-ES', 'Portador: Mikitaka Hazekura

Earth Wind and Fire descompone el cuerpo de Mikitaka en tiras de una sustancia desconocida para recomponerlo como cualquier objeto que elija, replicando su aspecto, peso y textura. Su origen como Stand o habilidad alienígena nunca queda del todo claro.', ARRAY['Transformación en Objetos: Se descompone y recompone como cualquier objeto, imitando su apariencia, peso y textura.','Réplica de Propiedades: Al transformarse adquiere cualidades físicas del objeto imitado, como el frío de un helado.','Transformación Parcial: Puede cambiar solo una parte del cuerpo o revertir la transformación de forma incompleta.','Límite de Complejidad: No logra replicar mecanismos complejos ni rasgos humanos con precisión.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b', 'en-GB', 'User: Mikitaka Hazekura

Earth Wind and Fire breaks Mikitaka''s body down into strips of an unknown substance, then reassembles it as any object he chooses, replicating its look, weight and texture. Whether it''s truly a Stand or an alien ability is never fully settled.', ARRAY['Object Transformation: Breaks down and reforms as any object, mimicking its appearance, weight and texture.','Property Replication: Gains the physical qualities of the object copied, such as the cold of an ice cream.','Partial Transformation: Can change just one body part or leave a transformation incomplete.','Complexity Limit: Cannot accurately replicate complex machinery or human faces.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b', 'ca-ES', 'Portador: Mikitaka Hazekura

Earth Wind and Fire descompon el cos d''en Mikitaka en tires d''una substància desconeguda per recompondre''l com qualsevol objecte que triï, replicant-ne l''aspecte, el pes i la textura. El seu origen com a Stand o habilitat alienígena mai no queda del tot clar.', ARRAY['Transformació en Objectes: Es descompon i es recompon com qualsevol objecte, imitant-ne l''aparença, el pes i la textura.','Rèplica de Propietats: En transformar-se adquireix qualitats físiques de l''objecte imitat, com el fred d''un gelat.','Transformació Parcial: Pot canviar només una part del cos o revertir la transformació de manera incompleta.','Límit de Complexitat: No aconsegueix replicar mecanismes complexos ni trets humans amb precisió.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Cinderella (Aya Tsuji)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7a67-8560-0adda4af5b56', 'STAND', 'Cinderella', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7a67-8560-0adda4af5b56', 'D', 'C', 'C', 'C', 'A', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7a67-8560-0adda4af5b56')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7a67-8560-0adda4af5b56', 'es-ES', 'Portadora: Aya Tsuji

Cinderella tiene forma de maniquí robótico femenino y remodela temporalmente el rostro o el cuerpo de quien lo desee, mejorando su atractivo. El efecto solo se vuelve permanente si el cliente aplica un lápiz de labios especial cada media hora durante un día entero.', ARRAY['Remodelación Temporal: Genera rasgos faciales o corporales nuevos que duran treinta minutos tras cada aplicación.','Fijación Permanente: El cambio se vuelve definitivo si se reaplica el lápiz de labios especial cada media hora durante veinticuatro horas.','Memoria de Rostros: Recuerda cada rostro que ha embellecido y puede reproducirlo de forma idéntica más tarde.','Trasplante de Rasgos: Puede extraer una parte del cuerpo de una persona e injertarla en otra para crear una identidad falsa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7a67-8560-0adda4af5b56')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7a67-8560-0adda4af5b56', 'en-GB', 'User: Aya Tsuji

Cinderella takes the form of a feminine robotic mannequin and temporarily reshapes a client''s face or body to enhance their looks. The change only becomes permanent if the client applies a special lipstick every half hour for a full day.', ARRAY['Temporary Remodelling: Creates new facial or body features that last thirty minutes after each application.','Permanent Fixing: The change becomes lasting if the special lipstick is reapplied every half hour for twenty-four hours.','Face Memory: Remembers every face it has beautified and can reproduce it identically later.','Feature Transplant: Can remove a body part from one person and graft it onto another to create a false identity.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7a67-8560-0adda4af5b56')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7a67-8560-0adda4af5b56', 'ca-ES', 'Portadora: Aya Tsuji

Cinderella té forma de maniquí robòtic femení i remodela temporalment el rostre o el cos de qui ho vulgui, millorant-ne l''atractiu. L''efecte només esdevé permanent si la clienta aplica un pintallavis especial cada mitja hora durant un dia sencer.', ARRAY['Remodelació Temporal: Genera trets facials o corporals nous que duren trenta minuts després de cada aplicació.','Fixació Permanent: El canvi esdevé definitiu si es reaplica el pintallavis especial cada mitja hora durant vint-i-quatre hores.','Memòria de Rostres: Recorda cada rostre que ha embellit i el pot reproduir de manera idèntica més endavant.','Trasplantament de Trets: Pot extreure una part del cos d''una persona i empeltar-la a una altra per crear una identitat falsa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7a67-8560-0adda4af5b56')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Stray Cat (Tama)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e70c-0fb1-7f3f-9b50-0171b6314fef', 'STAND', 'Stray Cat', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e70c-0fb1-7f3f-9b50-0171b6314fef', 'B', 'E', 'NULL', 'A', 'E', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f3f-9b50-0171b6314fef')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f3f-9b50-0171b6314fef', 'es-ES', 'Portador: Tama

Tama era un gato callejero que murió alcanzado por una Flecha del Stand y renació como planta con alma felina. Stray Cat controla el aire a su alrededor mediante sus bigotes, disparando burbujas de aire comprimido capaces de herir o inmovilizar.', ARRAY['Reencarnación Vegetal: Tras morir, el alma de Tama revive en un cuerpo de planta que conserva su memoria y personalidad.','Control del Aire: Manipula el aire circundante con sus bigotes para formar burbujas de presión.','Burbujas de Combate: Dispara burbujas invisibles capaces de dañar, atrapar o detonar a distancia.','Dependencia de la Luz: Su poder se debilita o queda inactivo sin exposición solar directa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f3f-9b50-0171b6314fef')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f3f-9b50-0171b6314fef', 'en-GB', 'User: Tama

Tama was a stray cat killed by a Stand Arrow and reborn as a plant carrying a feline soul. Stray Cat controls the surrounding air through its whiskers, firing compressed air bubbles that can wound or immobilise.', ARRAY['Plant Reincarnation: After dying, Tama''s soul lives on in a plant body that keeps its memory and personality.','Air Control: Manipulates the surrounding air through its whiskers to form pressurised bubbles.','Combat Bubbles: Fires invisible bubbles capable of harming, trapping or detonating targets at a distance.','Sunlight Dependency: Its power weakens or shuts down without direct sunlight.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f3f-9b50-0171b6314fef')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e70c-0fb1-7f3f-9b50-0171b6314fef', 'ca-ES', 'Portador: Tama

En Tama era un gat de carrer mort per una Fletxa de Stand i renascut com una planta amb ànima felina. Stray Cat controla l''aire del seu voltant mitjançant els bigotis i dispara bombolles d''aire comprimit capaces de ferir o immobilitzar.', ARRAY['Reencarnació Vegetal: Després de morir, l''ànima d''en Tama reviu en un cos de planta que conserva la seva memòria i personalitat.','Control de l''Aire: Manipula l''aire del voltant amb els bigotis per formar bombolles de pressió.','Bombolles de Combat: Dispara bombolles invisibles capaces de fer mal, atrapar o detonar a distància.','Dependència de la Llum: El seu poder s''afebleix o queda inactiu sense exposició solar directa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e70c-0fb1-7f3f-9b50-0171b6314fef')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Evolution chains (Echoes Act 0->1->2->3, Killer Queen ->+SHA ->+BtD).
UPDATE stands SET evolves_from_id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7' WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e70c-0faf-792f-97c3-2bd23f0f93a7');
UPDATE stands SET evolves_from_id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8' WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8');
UPDATE stands SET evolves_from_id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d' WHERE id = '01a0e70c-0fb0-7f77-8c31-cab107b96727' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d');
UPDATE stands SET evolves_from_id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c' WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e70c-0fb0-759f-8574-e3f698f99e8c');
UPDATE stands SET evolves_from_id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580' WHERE id = '01a0e70c-0fb0-7d53-92d2-1107869ba82b' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580');

-- +goose Down
-- Cascades to stands + power_translations rows via ON DELETE CASCADE.
DELETE FROM powers WHERE id IN (
  '01a0e70c-0faf-7459-891a-aa738b67b7a3',
  '01a0e70c-0faf-79e1-aa68-c9c56cf130df',
  '01a0e70c-0faf-792f-97c3-2bd23f0f93a7',
  '01a0e70c-0fb0-728e-9ede-c44c9f8af7b8',
  '01a0e70c-0fb0-7cef-9dca-95ba97d8fe5d',
  '01a0e70c-0fb0-7f77-8c31-cab107b96727',
  '01a0e70c-0fb0-78d5-82f0-2611b3e0761d',
  '01a0e70c-0fb0-759f-8574-e3f698f99e8c',
  '01a0e70c-0fb0-7c16-a26c-5f11e8c2e580',
  '01a0e70c-0fb0-7d53-92d2-1107869ba82b',
  '01a0e70c-0fb0-7292-8762-f1cce1596977',
  '01a0e70c-0fb0-7d27-abc8-c958d2e709a7',
  '01a0e70c-0fb0-7864-9234-efa187cc315c',
  '01a0e70c-0fb0-70a1-887d-425d9f32061b',
  '01a0e70c-0fb0-7c47-ba95-c4cb36d44649',
  '01a0e70c-0fb1-75f2-be6b-961f065a805d',
  '01a0e70c-0fb1-790a-a1e6-1ffcc7906302',
  '01a0e70c-0fb1-7d57-97e6-c4efc955d0ca',
  '01a0e70c-0fb1-7f5d-bf44-033c52faf367',
  '01a0e70c-0fb1-722d-a1c9-e6d482e59fc0',
  '01a0e70c-0fb1-7c9b-b004-89f8abc9c16c',
  '01a0e70c-0fb1-7967-8f01-02ccef87b43f',
  '01a0e70c-0fb1-7ee8-bd4b-af61be09f301',
  '01a0e70c-0fb1-71d5-af5f-e2bae8cc4879',
  '01a0e70c-0fb1-7f9e-beb0-bebb1aad8ce8',
  '01a0e70c-0fb1-799e-95d4-314231b37640',
  '01a0e70c-0fb1-7d90-99c1-3d58d5912a1b',
  '01a0e70c-0fb1-7a67-8560-0adda4af5b56',
  '01a0e70c-0fb1-7f3f-9b50-0171b6314fef'
);
