-- +goose Up
-- Seeds JoJo's Bizarre Adventure Part 6 (Stone Ocean) Stands.
-- Hand-authored new content (stats verified against jojowiki.com), not a
-- catalogsync prod-dump migration like 00017/00019 - see
-- ObsidianVault/catalog-seed-part6-stands.md. Runs in every environment
-- including prod (no ENVSUB guard): images are added later by an admin
-- through the normal picture-upload flow, so every row ships with the
-- powers.picture* defaults (picture_status = NONE).
--
-- Idempotent against a name an admin may have already created in prod:
-- every powers insert is ON CONFLICT DO NOTHING (id or name clash both
-- skip), and every dependent insert is gated on the powers row actually
-- existing with that id, so a skipped Stand never leaves an orphaned
-- stands/power_translations row behind.

-- Stone Free (Jolyne Cujoh)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80e-7439-9ade-39ef1ad32169', 'STAND', 'Stone Free', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80e-7439-9ade-39ef1ad32169', 'A', 'B', 'C', 'A', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80e-7439-9ade-39ef1ad32169')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80e-7439-9ade-39ef1ad32169', 'es-ES', 'Portador: Jolyne Cujoh

Es un Stand de combate cercano cuya habilidad exclusiva es deshilachar su propio cuerpo (y cualquier otra cosa que toque) en hilos de cuerda, permitiéndole extenderse, atarse a superficies, tejer objetos e incluso reconstruirse tras sufrir daño grave. Con el paso de la historia su potencia de golpe aumenta drásticamente, hasta volverse comparable al impacto de un pequeño meteorito.', ARRAY['Deshilachado Corporal: Convierte su propio cuerpo o el de un objeto tocado en hilos de cuerda, permitiendo alargarse o dividirse.','Tejido Táctico: Usa los hilos para atrapar enemigos, tender trampas o fabricar herramientas improvisadas en pleno combate.','Regeneración por Hilos: Reconstruye partes de su cuerpo deshilachadas, recuperándose de heridas que serían letales para un humano normal.','Golpe Reforzado: Su puñetazo, potenciado más adelante por Survivor, alcanza una fuerza destructiva colosal pese a su corto alcance.','Combate Cercano Versátil: Combina fuerza y improvisación para adaptarse a rivales de todo tipo en distancias muy reducidas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80e-7439-9ade-39ef1ad32169')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80e-7439-9ade-39ef1ad32169', 'en-GB', 'User: Jolyne Cujoh

This is a close-range combat Stand whose signature power is unravelling its own body (and anything else it touches) into strings, letting it stretch, tether itself to surfaces, weave objects, and even rebuild itself after severe damage. Its punching power increases dramatically as the story progresses, eventually becoming comparable to the impact of a small meteorite.', ARRAY['Body Unravelling: Turns its own body or a touched object into strings, allowing it to stretch or split apart.','Tactical Weaving: Uses the strings to trap enemies, set traps, or craft improvised tools mid-fight.','String Regeneration: Rebuilds unravelled parts of its body, recovering from wounds that would be lethal to a normal human.','Enhanced Punch: Its punch, later boosted by Survivor, reaches colossal destructive force despite its short range.','Versatile Close Combat: Combines strength and improvisation to adapt to all kinds of opponents at very short distance.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80e-7439-9ade-39ef1ad32169')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80e-7439-9ade-39ef1ad32169', 'ca-ES', 'Portador: Jolyne Cujoh

És un Stand de combat proper l''habilitat exclusiva del qual és desfilar el seu propi cos (i qualsevol altra cosa que toqui) en fils de corda, cosa que li permet estirar-se, lligar-se a superfícies, teixir objectes i fins i tot reconstruir-se després de patir un dany greu. Amb el pas de la història la seva potència de cop augmenta dràsticament, fins a esdevenir comparable a l''impacte d''un petit meteorit.', ARRAY['Desfilat Corporal: Converteix el seu propi cos o el d''un objecte tocat en fils de corda, cosa que li permet allargar-se o dividir-se.','Teixit Tàctic: Fa servir els fils per atrapar enemics, parar trampes o fabricar eines improvisades en ple combat.','Regeneració per Fils: Reconstrueix parts del seu cos desfilades, recuperant-se de ferides que serien letals per a un humà normal.','Cop Reforçat: El seu cop de puny, potenciat més endavant per Survivor, assoleix una força destructiva colossal malgrat el seu curt abast.','Combat Proper Versàtil: Combina força i improvisació per adaptar-se a rivals de tota mena a distàncies molt reduïdes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80e-7439-9ade-39ef1ad32169')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Burning Down the House (Emporio Alnino)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7314-b376-08684b35ecc9', 'STAND', 'Burning Down the House', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7314-b376-08684b35ecc9', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7314-b376-08684b35ecc9')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7314-b376-08684b35ecc9', 'es-ES', 'Portador: Emporio Alnino

Es un Stand sin forma física propia, incapaz de combatir directamente, cuya habilidad exclusiva es materializar el fantasma de una sala de música oculta tras una grieta invisible en una pared. Dentro de ese espacio Emporio puede esconderse por completo de la vista y de la percepción de otros Stands, usándolo como refugio seguro más que como arma.', ARRAY['Materialización de la Sala: Invoca una grieta invisible en una pared que da acceso a una sala de música fantasma escondida.','Refugio Imperceptible: Quien entra en la sala queda oculto por completo, indetectable incluso para Stands especializados en rastreo.','Ocultación Sensorial: Dentro de la sala no se puede ver, oír ni sentir lo que ocurre fuera, aislando por completo a su ocupante.','Uso Táctico como Escondite: Emporio la emplea para espiar con seguridad o para protegerse durante situaciones de combate ajenas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7314-b376-08684b35ecc9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7314-b376-08684b35ecc9', 'en-GB', 'User: Emporio Alnino

This Stand has no physical form and cannot fight directly; its signature power is materialising the ghost of a music room hidden behind an invisible crack in a wall. Inside that space Emporio can hide completely from sight and from other Stands'' perception, using it as a safe refuge rather than a weapon.', ARRAY['Room Materialisation: Summons an invisible crack in a wall leading into a hidden ghost music room.','Imperceptible Refuge: Anyone who enters the room becomes completely hidden, undetectable even to tracking-focused Stands.','Sensory Isolation: Inside the room nothing outside can be seen, heard or felt, fully isolating its occupant.','Tactical Hideout Use: Emporio uses it to spy safely or to shelter himself during combat happening around him.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7314-b376-08684b35ecc9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7314-b376-08684b35ecc9', 'ca-ES', 'Portador: Emporio Alnino

És un Stand sense forma física pròpia, incapaç de combatre directament, l''habilitat exclusiva del qual és materialitzar el fantasma d''una sala de música amagada darrere d''una esquerda invisible en una paret. Dins d''aquest espai l''Emporio es pot amagar completament de la vista i de la percepció d''altres Stands, fent-lo servir com a refugi segur més que com a arma.', ARRAY['Materialització de la Sala: Invoca una esquerda invisible en una paret que dona accés a una sala de música fantasma amagada.','Refugi Imperceptible: Qui entra a la sala queda ocult completament, indetectable fins i tot per a Stands especialitzats en rastreig.','Ocultació Sensorial: Dins la sala no es pot veure, sentir ni percebre el que passa a fora, aïllant completament el seu ocupant.','Ús Tàctic com a Amagatall: L''Emporio la fa servir per espiar amb seguretat o per protegir-se durant situacions de combat alienes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7314-b376-08684b35ecc9')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kiss (Ermes Costello)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7656-8d29-3dcfdb798513', 'STAND', 'Kiss', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7656-8d29-3dcfdb798513', 'A', 'A', 'A', 'A', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7656-8d29-3dcfdb798513')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7656-8d29-3dcfdb798513', 'es-ES', 'Portador: Ermes Costello

Es un Stand humanoide de combate cercano cuya habilidad exclusiva es crear pegatinas con forma de labios que, al pegarse sobre un objeto o ser vivo, generan una copia idéntica de este; al despegar la pegatina, ambas copias se fusionan de nuevo liberando una fuerza destructiva. Ermes compensa así su corto alcance con una potencia, velocidad y resistencia notables.', ARRAY['Pegatina Duplicadora: Crea una pegatina con forma de labios que, pegada a un objeto o persona, genera una copia exacta de este.','Fusión Destructiva: Al retirar la pegatina, el original y la copia se recombinan violentamente, causando un daño severo.','Fuerza y Velocidad Físicas: Golpea con una potencia y rapidez sobrehumanas propias de un Stand de combate cercano de primer nivel.','Resistencia Notable: Aguanta impactos considerables gracias a una resistencia física fuera de lo común.','Uso Creativo de Copias: Emplea las duplicaciones con ingenio táctico, más allá del simple combate directo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7656-8d29-3dcfdb798513')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7656-8d29-3dcfdb798513', 'en-GB', 'User: Ermes Costello

This is a close-range humanoid Stand whose signature power is creating lip-shaped stickers that, once stuck onto an object or living being, generate an identical copy of it; when the sticker is peeled off, both copies merge back together, unleashing destructive force. Ermes makes up for its short range with notable power, speed and endurance.', ARRAY['Duplicating Sticker: Creates a lip-shaped sticker that, when stuck onto an object or person, generates an exact copy of it.','Destructive Fusion: Removing the sticker violently recombines the original and the copy, causing severe damage.','Physical Strength and Speed: Strikes with superhuman power and speed befitting a top-tier close-range Stand.','Notable Endurance: Withstands considerable impacts thanks to unusually high physical resilience.','Creative Use of Copies: Uses the duplications with tactical ingenuity, beyond simple direct combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7656-8d29-3dcfdb798513')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7656-8d29-3dcfdb798513', 'ca-ES', 'Portador: Ermes Costello

És un Stand humanoide de combat proper l''habilitat exclusiva del qual és crear enganxines amb forma de llavis que, en enganxar-se sobre un objecte o ésser viu, en generen una còpia idèntica; en despegar l''enganxina, ambdues còpies es fusionen de nou alliberant una força destructiva. L''Ermes compensa així el seu curt abast amb una potència, velocitat i resistència notables.', ARRAY['Enganxina Duplicadora: Crea una enganxina amb forma de llavis que, enganxada a un objecte o persona, en genera una còpia exacta.','Fusió Destructiva: En retirar l''enganxina, l''original i la còpia es recombinen violentament, causant un dany sever.','Força i Velocitat Físiques: Colpeja amb una potència i rapidesa sobrehumanes pròpies d''un Stand de combat proper de primer nivell.','Resistència Notable: Aguanta impactes considerables gràcies a una resistència física fora del comú.','Ús Creatiu de Còpies: Fa servir les duplicacions amb enginy tàctic, més enllà del simple combat directe.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7656-8d29-3dcfdb798513')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Foo Fighters (F.F.)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-77f3-bc44-db5294ed4574', 'STAND', 'Foo Fighters', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-77f3-bc44-db5294ed4574', 'B', 'A', 'C', 'A', 'C', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77f3-bc44-db5294ed4574')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77f3-bc44-db5294ed4574', 'es-ES', 'Portador: F.F. (la propia colonia de plancton, que es a la vez el Stand y su usuaria)

Es un caso único entre los Stands: una colonia de plancton que alcanzó consciencia gracias a un disco de Stand, siendo simultáneamente el Stand y su propio usuario. Puede invadir cuerpos de peces o de personas fallecidas para moverse, disolverse en agua para desplazarse con libertad y reconstruirse tras sufrir daño gracias a su naturaleza distribuida sin un punto vital único.', ARRAY['Invasión Corporal: Toma el control de un cuerpo de pez o de una persona reciente fallecida para moverse e interactuar con el entorno.','Disolución Acuática: Se descompone en su forma de plancton y se desplaza libremente a través del agua.','Regeneración Distribuida: Al no depender de un único punto vital, se recompone tras sufrir daños que serían letales para otros Stands.','Sellado de Heridas: Puede introducirse en el cuerpo de un aliado para taponar heridas internas y sostenerlo con vida.','Dependencia del Agua: Necesita hidratación constante para mantener su forma, debilitándose si se seca.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77f3-bc44-db5294ed4574')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77f3-bc44-db5294ed4574', 'en-GB', 'User: F.F. (the plankton colony itself, which is both the Stand and its own user)

This is a unique case among Stands: a plankton colony that gained consciousness through a Stand DISC, being simultaneously the Stand and its own user. It can invade fish bodies or recently deceased people to move around, dissolve into water to travel freely, and rebuild itself after damage thanks to its distributed nature with no single vital point.', ARRAY['Body Invasion: Takes control of a fish''s body or a recently deceased person to move and interact with its surroundings.','Aquatic Dissolution: Breaks down into its plankton form and travels freely through water.','Distributed Regeneration: Having no single vital point, it reassembles after damage that would be lethal to other Stands.','Wound Sealing: Can enter an ally''s body to plug internal wounds and keep them alive.','Water Dependency: Needs a steady supply of water to keep its form, weakening if it dries out.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77f3-bc44-db5294ed4574')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77f3-bc44-db5294ed4574', 'ca-ES', 'Portador: F.F. (la mateixa colònia de plàncton, que és alhora l''Stand i la seva usuària)

És un cas únic entre els Stands: una colònia de plàncton que va assolir consciència gràcies a un disc d''Stand, sent alhora l''Stand i la seva pròpia usuària. Pot envair cossos de peixos o de persones mortes recentment per moure''s, dissoldre''s en aigua per desplaçar-se amb llibertat i reconstruir-se després de patir danys gràcies a la seva naturalesa distribuïda sense un punt vital únic.', ARRAY['Invasió Corporal: Pren el control d''un cos de peix o d''una persona morta recentment per moure''s i interactuar amb l''entorn.','Dissolució Aquàtica: Es descompon en la seva forma de plàncton i es desplaça lliurement a través de l''aigua.','Regeneració Distribuïda: En no dependre d''un únic punt vital, es recomposa després de patir danys que serien letals per a altres Stands.','Segellat de Ferides: Es pot introduir dins el cos d''un aliat per tapar ferides internes i mantenir-lo amb vida.','Dependència de l''Aigua: Necessita hidratació constant per mantenir la seva forma, i s''afebleix si s''asseca.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77f3-bc44-db5294ed4574')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Diver Down (Narciso Anasui)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7762-a95f-f4e65d73a1a2', 'STAND', 'Diver Down', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7762-a95f-f4e65d73a1a2', 'A', 'A', 'E', 'C', 'B', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7762-a95f-f4e65d73a1a2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7762-a95f-f4e65d73a1a2', 'es-ES', 'Portador: Narciso Anasui

Es un Stand de combate muy cercano, capaz de decapitar a un rival de un solo golpe, cuya habilidad exclusiva es atravesar objetos sólidos para reorganizar su composición interna sin romper su superficie externa. Anasui también puede almacenar la fuerza de un golpe dentro de un objeto y liberarla después, multiplicando el efecto de sus impactos.', ARRAY['Fuerza Devastadora: Golpea con tal potencia que puede decapitar a una persona de un solo corte de mano.','Fase Sólida: Atraviesa objetos sólidos y reorganiza su interior sin dañar la superficie exterior, ocultando cosas dentro de ellos.','Almacenamiento de Impacto: Guarda la fuerza de un golpe dentro de un objeto para liberarla más tarde con efecto acumulado.','Velocidad de Combate: Actúa con una rapidez que le permite sorprender a rivales en distancias extremadamente cortas.','Reorganización Táctica: Usa su fase a través de la materia para esconder objetos o personas dentro de superficies sólidas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7762-a95f-f4e65d73a1a2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7762-a95f-f4e65d73a1a2', 'en-GB', 'User: Narciso Anasui

This is a very close-range combat Stand, capable of decapitating an opponent in a single blow, whose signature power is passing through solid objects to rearrange their internal composition without breaking their outer surface. Anasui can also store the force of a strike inside an object and release it later, multiplying the effect of his hits.', ARRAY['Devastating Strength: Hits with such power that it can decapitate a person with a single chop.','Solid Phasing: Passes through solid objects and rearranges their interior without damaging the outer surface, hiding things inside them.','Impact Storage: Stores the force of a strike inside an object to release it later with cumulative effect.','Combat Speed: Moves with a swiftness that lets it catch opponents off guard at extremely close range.','Tactical Rearranging: Uses its phasing through matter to hide objects or people inside solid surfaces.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7762-a95f-f4e65d73a1a2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7762-a95f-f4e65d73a1a2', 'ca-ES', 'Portador: Narciso Anasui

És un Stand de combat molt proper, capaç de decapitar un rival d''un sol cop, l''habilitat exclusiva del qual és travessar objectes sòlids per reorganitzar-ne la composició interna sense trencar-ne la superfície externa. L''Anasui també pot emmagatzemar la força d''un cop dins d''un objecte i alliberar-la després, multiplicant l''efecte dels seus impactes.', ARRAY['Força Devastadora: Colpeja amb tanta potència que pot decapitar una persona d''un sol cop de mà.','Fase Sòlida: Travessa objectes sòlids i en reorganitza l''interior sense malmetre''n la superfície exterior, amagant-hi coses a dins.','Emmagatzematge d''Impacte: Guarda la força d''un cop dins d''un objecte per alliberar-la més tard amb efecte acumulat.','Velocitat de Combat: Actua amb una rapidesa que li permet sorprendre rivals a distàncies extremadament curtes.','Reorganització Tàctica: Fa servir la seva fase a través de la matèria per amagar objectes o persones dins de superfícies sòlides.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7762-a95f-f4e65d73a1a2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Weather Report (Weather Report)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-77af-8a40-056b85e046bb', 'STAND', 'Weather Report', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-77af-8a40-056b85e046bb', 'A', 'B', 'C', 'A', 'E', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77af-8a40-056b85e046bb')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77af-8a40-056b85e046bb', 'es-ES', 'Portador: Weather Report

Es un Stand de rango extraordinario capaz de controlar el clima a escala planetaria, generando lluvia, viento, rayos y niebla a voluntad para dominar el campo de batalla. Su fuerza y resistencia son notables y su alcance directo es corto, pero compensa con un dominio meteorológico que se extiende kilómetros a la redonda, aunque su precisión en tareas delicadas es sorprendentemente pobre.', ARRAY['Control Climático: Manipula el clima a voluntad en un radio de kilómetros, generando lluvia, viento o niebla según su conveniencia.','Ojo del Huracán: Crea zonas de calma absoluta rodeadas de tormenta para ocultar su posición o la de sus aliados.','Rayos y Relámpagos: Invoca descargas eléctricas capaces de electrocutar o cegar a sus enemigos a distancia.','Niebla Táctica: Genera bancos de niebla densa que reducen drásticamente la visibilidad del enemigo.','Fuerza de Combate: Posee una potencia física considerable en distancias cortas pese a ser un Stand de rango.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77af-8a40-056b85e046bb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77af-8a40-056b85e046bb', 'en-GB', 'User: Weather Report

This is an extraordinary-range Stand able to control the weather on a planetary scale, generating rain, wind, lightning and fog at will to dominate the battlefield. Its strength and endurance are notable and its direct reach is short, but it makes up for that with weather mastery spanning kilometres, though its precision at delicate tasks is surprisingly poor.', ARRAY['Weather Control: Manipulates the weather at will across a radius of kilometres, generating rain, wind or fog as it sees fit.','Eye of the Hurricane: Creates zones of absolute calm surrounded by storms to hide its own or its allies'' position.','Thunder and Lightning: Summons electrical discharges capable of electrocuting or blinding enemies at range.','Tactical Fog: Generates dense fog banks that drastically reduce enemy visibility.','Combat Strength: Possesses considerable physical power at short range despite being a range-type Stand.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77af-8a40-056b85e046bb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-77af-8a40-056b85e046bb', 'ca-ES', 'Portador: Weather Report

És un Stand d''abast extraordinari capaç de controlar el clima a escala planetària, generant pluja, vent, llamps i boira a voluntat per dominar el camp de batalla. La seva força i resistència són notables i el seu abast directe és curt, però ho compensa amb un domini meteorològic que s''estén quilòmetres a la rodona, tot i que la seva precisió en tasques delicades és sorprenentment pobra.', ARRAY['Control Climàtic: Manipula el clima a voluntat en un radi de quilòmetres, generant pluja, vent o boira segons li convingui.','Ull de l''Huracà: Crea zones de calma absoluta envoltades de tempesta per amagar la seva posició o la dels seus aliats.','Llamps i Trons: Invoca descàrregues elèctriques capaces d''electrocutar o encegar els enemics a distància.','Boira Tàctica: Genera bancs de boira densa que redueixen dràsticament la visibilitat de l''enemic.','Força de Combat: Posseeix una potència física considerable a distàncies curtes malgrat ser un Stand d''abast.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-77af-8a40-056b85e046bb')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Weather Report: Heavy Weather (Weather Report)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7f2e-9906-43249b2bb2b4', 'STAND', 'Weather Report: Heavy Weather', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7f2e-9906-43249b2bb2b4', 'A', 'B', 'C', 'A', 'E', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7f2e-9906-43249b2bb2b4')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7f2e-9906-43249b2bb2b4', 'es-ES', 'Portador: Weather Report

Es la forma desatada de Weather Report, un estado que surge de forma automática cuando la rabia de Weather contra la humanidad se desborda, construyendo sobre su ya inmenso control del clima. En este estado genera arcoíris que transforman en caracoles a cualquiera que los toque, criaturas que se reproducen y multiplican de forma exponencial hasta invadir la zona por completo.', ARRAY['Arcoíris Mutante: Genera arcoíris que convierten en caracoles a quien los toca, salvo al propio Weather Report.','Reproducción Exponencial: Los caracoles creados se multiplican a gran velocidad gracias a su hermafroditismo.','Control Climático Desatado: Conserva y amplifica todo el dominio meteorológico de su forma base sin restricciones.','Rabia Automática: Se activa involuntariamente por el odio acumulado de Weather Report, sin necesidad de orden consciente.','Dominio Territorial: Convierte una zona entera en un ecosistema hostil e imparable en muy poco tiempo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7f2e-9906-43249b2bb2b4')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7f2e-9906-43249b2bb2b4', 'en-GB', 'User: Weather Report

This is Weather Report''s unleashed form, a state that arises automatically when Weather''s rage against humanity boils over, building on his already immense weather control. In this state it generates rainbows that turn anyone who touches them into snails, creatures that reproduce and multiply exponentially until they overrun the area entirely.', ARRAY['Mutant Rainbows: Generates rainbows that turn whoever touches them into snails, sparing only Weather Report himself.','Exponential Reproduction: The created snails multiply at great speed thanks to their hermaphroditism.','Unleashed Weather Control: Retains and amplifies the full weather mastery of its base form without restriction.','Automatic Rage: Triggers involuntarily from Weather Report''s accumulated hatred, with no need for conscious command.','Territorial Dominance: Turns an entire area into a hostile, unstoppable ecosystem in very little time.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7f2e-9906-43249b2bb2b4')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7f2e-9906-43249b2bb2b4', 'ca-ES', 'Portador: Weather Report

És la forma deslligada d''en Weather Report, un estat que sorgeix de manera automàtica quan la ràbia d''en Weather contra la humanitat es desborda, construint sobre el seu ja immens control del clima. En aquest estat genera arcs de Sant Martí que transformen en cargols qualsevol que els toqui, criatures que es reprodueixen i multipliquen de manera exponencial fins a envair la zona per complet.', ARRAY['Arcs de Sant Martí Mutants: Genera arcs de Sant Martí que converteixen en cargols qui els toca, excepte en Weather Report mateix.','Reproducció Exponencial: Els cargols creats es multipliquen a gran velocitat gràcies al seu hermafroditisme.','Control Climàtic Deslligat: Conserva i amplifica tot el domini meteorològic de la seva forma base sense restriccions.','Ràbia Automàtica: S''activa involuntàriament per l''odi acumulat d''en Weather Report, sense necessitat d''ordre conscient.','Domini Territorial: Converteix una zona sencera en un ecosistema hostil i imparable en molt poc temps.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7f2e-9906-43249b2bb2b4')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Whitesnake (Enrico Pucci)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-776e-94c5-141b2b6d4824', 'STAND', 'Whitesnake', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-776e-94c5-141b2b6d4824', 'C', 'D', 'NULL', 'A', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-776e-94c5-141b2b6d4824')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-776e-94c5-141b2b6d4824', 'es-ES', 'Portador: Enrico Pucci

Es un Stand de largo alcance capaz de extraer el alma y la memoria de una persona en forma de DISC, permitiendo robar Stands ajenos o borrar recuerdos concretos e insertarlos en otra víctima. Actúa a unos veinte metros de distancia, lo que permite a Pucci moverlo sin ser detectado, y aunque su ataque directo no es su punto fuerte, su resistencia es notable.', ARRAY['Extracción de DISC de Stand: Saca el Stand de una persona en forma de disco físico para robarlo o entregarlo a otro.','Extracción de DISC de Memoria: Extrae recuerdos específicos de la mente de alguien, dejando lagunas que pueden manipularse.','Inserción de DISC: Implanta un DISC de Stand o de memoria en otra persona, dándole el poder o los recuerdos ajenos.','Control a Distancia: Opera hasta veinte metros lejos de Pucci, permitiéndole actuar en la sombra sin ser descubierto.','Resistencia Notable: Aguanta golpes considerables gracias a una durabilidad superior a la de muchos Stands de su rango.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-776e-94c5-141b2b6d4824')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-776e-94c5-141b2b6d4824', 'en-GB', 'User: Enrico Pucci

This is a long-range Stand able to extract a person''s soul and memory as a physical DISC, letting it steal other people''s Stands or erase specific memories and insert them into another victim. It operates roughly twenty metres away, letting Pucci move it undetected, and while its direct attack isn''t its strong suit, its endurance is notable.', ARRAY['Stand DISC Extraction: Pulls a person''s Stand out as a physical disc so it can be stolen or handed to someone else.','Memory DISC Extraction: Extracts specific memories from someone''s mind, leaving gaps that can be manipulated.','DISC Insertion: Implants a Stand or memory DISC into another person, granting them the stolen power or memories.','Remote Control: Operates up to twenty metres away from Pucci, letting him act from the shadows undetected.','Notable Endurance: Withstands considerable blows thanks to durability higher than many Stands of its range.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-776e-94c5-141b2b6d4824')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-776e-94c5-141b2b6d4824', 'ca-ES', 'Portador: Enrico Pucci

És un Stand d''abast llarg capaç d''extreure l''ànima i la memòria d''una persona en forma de DISC, cosa que li permet robar Stands aliens o esborrar records concrets i inserir-los en una altra víctima. Actua a uns vint metres de distància, cosa que permet a en Pucci moure''l sense ser detectat, i tot i que el seu atac directe no és el seu punt fort, la seva resistència és notable.', ARRAY['Extracció de DISC d''Stand: Treu l''Stand d''una persona en forma de disc físic per robar-lo o lliurar-lo a algú altre.','Extracció de DISC de Memòria: Extreu records específics de la ment d''algú, deixant llacunes que es poden manipular.','Inserció de DISC: Implanta un DISC d''Stand o de memòria en una altra persona, donant-li el poder o els records aliens.','Control a Distància: Opera fins a vint metres lluny d''en Pucci, permetent-li actuar des de l''ombra sense ser descobert.','Resistència Notable: Aguanta cops considerables gràcies a una durabilitat superior a la de molts Stands del seu abast.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-776e-94c5-141b2b6d4824')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- C-MOON (Enrico Pucci)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7093-a7c0-2728828eda74', 'STAND', 'C-MOON', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7093-a7c0-2728828eda74', 'NULL', 'B', 'B', 'NULL', 'NULL', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7093-a7c0-2728828eda74', 'es-ES', 'Portador: Enrico Pucci

Es la forma evolucionada de Whitesnake, nacida de la fusión de Pucci con el Bebé Verde, que sustituye la extracción de DISCs por el control de la gravedad. Invierte de dentro afuera cualquier cosa que golpea, desde objetos hasta cuerpos vivos, y aunque su potencia de impacto directo apenas puede clasificarse, su velocidad y control gravitacional lo convierten en una amenaza devastadora a varios kilómetros de distancia.', ARRAY['Inversión Gravitacional: Da la vuelta a cualquier cosa que toca, del interior al exterior, con resultados letales.','Manipulación de la Gravedad: Controla campos gravitatorios en un radio de kilómetros para desestabilizar a sus enemigos.','Anulación de Golpes: Invierte la trayectoria o el efecto de ataques entrantes, neutralizando casi cualquier ofensiva.','Control a Distancia: Hereda el alcance remoto de Whitesnake, permitiendo a Pucci dirigirlo mientras se esconde.','Agilidad de Combate: Se mueve con una velocidad notable en distancias cortas pese a operar principalmente a distancia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7093-a7c0-2728828eda74', 'en-GB', 'User: Enrico Pucci

This is Whitesnake''s evolved form, born from Pucci''s fusion with the Green Baby, which replaces DISC extraction with gravity control. It inverts inside-out anything it strikes, from objects to living bodies, and though its direct impact power barely fits a rating, its speed and gravitational control make it a devastating threat across several kilometres.', ARRAY['Gravitational Inversion: Turns anything it touches inside out, from the inside to the outside, with lethal results.','Gravity Manipulation: Controls gravitational fields across a radius of kilometres to destabilise its enemies.','Strike Nullification: Reverses the trajectory or effect of incoming attacks, neutralising almost any offense.','Remote Control: Inherits Whitesnake''s remote range, letting Pucci direct it while staying hidden.','Combat Agility: Moves with notable speed at close range despite operating mostly from a distance.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7093-a7c0-2728828eda74', 'ca-ES', 'Portador: Enrico Pucci

És la forma evolucionada de Whitesnake, nascuda de la fusió d''en Pucci amb el Nadó Verd, que substitueix l''extracció de DISCs pel control de la gravetat. Inverteix de dins cap enfora qualsevol cosa que colpeja, des d''objectes fins a cossos vius, i tot i que la seva potència d''impacte directe amb prou feines es pot classificar, la seva velocitat i control gravitacional el converteixen en una amenaça devastadora a diversos quilòmetres de distància.', ARRAY['Inversió Gravitacional: Gira qualsevol cosa que toca, de l''interior a l''exterior, amb resultats letals.','Manipulació de la Gravetat: Controla camps gravitatoris en un radi de quilòmetres per desestabilitzar els enemics.','Anul·lació de Cops: Inverteix la trajectòria o l''efecte d''atacs entrants, neutralitzant gairebé qualsevol ofensiva.','Control a Distància: Hereta l''abast remot de Whitesnake, permetent a en Pucci dirigir-lo mentre s''amaga.','Agilitat de Combat: Es mou amb una velocitat notable a distàncies curtes tot i operar principalment a distància.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Made in Heaven (Enrico Pucci)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-73fe-bb6e-4d44f1a91aa3', 'STAND', 'Made in Heaven', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3', 'B', 'INFINITE', 'C', 'A', 'C', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3', 'es-ES', 'Portador: Enrico Pucci

Es la evolución final del Stand de Pucci, alcanzada tras el gobierno de C-MOON sobre la gravedad, que ahora se dedica a acelerar el tiempo mismo hasta cotas prácticamente infinitas. Con este poder Pucci puede precipitar el fin y el reinicio del universo entero, moviéndose y actuando a una velocidad que ningún rival puede llegar siquiera a percibir.', ARRAY['Aceleración Temporal: Acelera el paso del tiempo de forma progresiva hasta alcanzar una velocidad prácticamente infinita.','Velocidad Infinita: Se mueve y ataca tan rápido que resulta virtualmente imposible de esquivar o siquiera ver.','Fin del Universo: Lleva el tiempo hasta su límite para provocar el colapso y reinicio de la existencia entera.','Bucle de la Reencarnación: Da lugar a un nuevo ciclo del universo donde los destinos pueden cambiar ligeramente.','Resistencia Superior: Su cuerpo aguanta el propio esfuerzo de manipular el tiempo a semejante escala sin quebrarse.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3', 'en-GB', 'User: Enrico Pucci

This is the final evolution of Pucci''s Stand, reached after C-MOON''s mastery over gravity, now devoted to accelerating time itself to near-infinite levels. With this power Pucci can hasten the end and rebirth of the entire universe, moving and acting at a speed no rival can even perceive.', ARRAY['Time Acceleration: Speeds up the passage of time progressively until reaching a near-infinite velocity.','Infinite Speed: Moves and attacks so fast that it is virtually impossible to dodge or even see.','End of the Universe: Pushes time to its limit to trigger the collapse and restart of all existence.','Reincarnation Loop: Ushers in a new cycle of the universe where fates can shift slightly.','Superior Endurance: Its body withstands the strain of manipulating time on such a scale without breaking.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3', 'ca-ES', 'Portador: Enrico Pucci

És l''evolució final de l''Stand d''en Pucci, assolida després del domini de C-MOON sobre la gravetat, que ara es dedica a accelerar el temps mateix fins a cotes pràcticament infinites. Amb aquest poder en Pucci pot precipitar la fi i el reinici de l''univers sencer, movent-se i actuant a una velocitat que cap rival pot ni tan sols percebre.', ARRAY['Acceleració Temporal: Accelera el pas del temps de manera progressiva fins a assolir una velocitat pràcticament infinita.','Velocitat Infinita: Es mou i ataca tan ràpid que resulta virtualment impossible d''esquivar o fins i tot de veure.','Fi de l''Univers: Porta el temps fins al seu límit per provocar el col·lapse i reinici de tota l''existència.','Bucle de la Reencarnació: Dona lloc a un nou cicle de l''univers on els destins poden canviar lleugerament.','Resistència Superior: El seu cos aguanta l''esforç mateix de manipular el temps a aquesta escala sense trencar-se.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Goo Goo Dolls (Gwess)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7221-866b-28cd22971b47', 'STAND', 'Goo Goo Dolls', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7221-866b-28cd22971b47', 'D', 'C', 'B', 'D', 'B', 'B' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7221-866b-28cd22971b47')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7221-866b-28cd22971b47', 'es-ES', 'Portador: Gwess

Es un Stand pequeño y robótico, apenas del tamaño de un ratón, cuya habilidad exclusiva es reducir instantáneamente el tamaño de cualquier objetivo dentro de su alcance, manteniéndolo diminuto mientras no se aleje demasiado de Gwess. Su ataque directo es casi inexistente, pero compensa con un alcance y una precisión notables que le permiten intimidar y controlar a sus víctimas.', ARRAY['Reducción Instantánea: Encoge a cualquier persona u objeto dentro de su alcance al tamaño de un ratón con solo un toque de su aguja.','Alcance Extendido: Actúa sobre objetivos a entre 20 y 30 metros, aunque el efecto se debilita cuanto más lejos está la víctima.','Control por Cercanía: Mantiene reducida a la víctima solo mientras permanezca cerca de Gwess; alejarse demasiado revierte el encogimiento.','Precisión Quirúrgica: Localiza con exactitud el punto de contacto necesario para activar su poder incluso en combate confuso.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7221-866b-28cd22971b47')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7221-866b-28cd22971b47', 'en-GB', 'User: Gwess

This is a tiny, robotic Stand no bigger than a mouse whose signature power is instantly shrinking any target within its range, keeping them tiny for as long as they don''t stray too far from Gwess. Its direct attack power is next to nothing, but it makes up for this with notable range and precision that let it intimidate and control its victims.', ARRAY['Instant Shrinking: Shrinks any person or object within range down to mouse size with a single touch of its needle.','Extended Range: Can act on targets up to 20-30 metres away, though the effect weakens the farther the victim strays.','Proximity Control: Keeps the victim shrunk only while they stay near Gwess; moving too far away reverses the shrinking.','Surgical Precision: Pinpoints the exact contact spot needed to trigger its power even in chaotic combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7221-866b-28cd22971b47')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7221-866b-28cd22971b47', 'ca-ES', 'Portador: Gwess

És un Stand petit i robòtic, amb prou feines de la mida d''un ratolí, l''habilitat exclusiva del qual és reduir instantàniament la mida de qualsevol objectiu dins del seu abast, mantenint-lo diminut mentre no s''allunyi massa d''en Gwess. El seu atac directe és gairebé inexistent, però ho compensa amb un abast i una precisió notables que li permeten intimidar i controlar les seves víctimes.', ARRAY['Reducció Instantània: Encongeix qualsevol persona o objecte dins del seu abast a la mida d''un ratolí amb només un toc de la seva agulla.','Abast Ampliat: Actua sobre objectius situats entre 20 i 30 metres, tot i que l''efecte s''afebleix com més lluny és la víctima.','Control per Proximitat: Manté reduïda la víctima només mentre romangui a prop d''en Gwess; allunyar-se massa reverteix l''encongiment.','Precisió Quirúrgica: Localitza amb exactitud el punt de contacte necessari per activar el seu poder fins i tot en combat confús.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7221-866b-28cd22971b47')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Manhattan Transfer (Johngalli A.)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-741a-a5c0-96d5f45103e6', 'STAND', 'Manhattan Transfer', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-741a-a5c0-96d5f45103e6', 'E', 'E', 'A', 'A', 'A', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-741a-a5c0-96d5f45103e6')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-741a-a5c0-96d5f45103e6', 'es-ES', 'Portador: Johngalli A.

Es un pequeño Stand no humanoide, similar a un satélite, que Johngalli A. envía a distancia para detectar los movimientos del aire a su alrededor. Carece por completo de capacidad de combate propia, pero su lectura precisa de las corrientes de aire permite a su portador visualizar mentalmente el entorno y localizar objetivos ocultos para disparar con precisión letal.', ARRAY['Lectura de Corrientes: Detecta con precisión los movimientos del aire alrededor del Stand y los transmite a Johngalli A.','Visión Mental del Entorno: Reconstruye mentalmente posición y silueta de los objetivos a partir del flujo de aire.','Apoyo a Larga Distancia: Complementa la puntería de francotirador de Johngalli A., permitiéndole disparar a través de puntos ciegos.','Vulnerabilidad a Gases: Introducir capas de gases distintos distorsiona su lectura del aire y provoca errores de cálculo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-741a-a5c0-96d5f45103e6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-741a-a5c0-96d5f45103e6', 'en-GB', 'User: Johngalli A.

This is a small, non-humanoid, satellite-like Stand that Johngalli A. sends out at a distance to detect the movements of the air around it. It has no combat ability of its own whatsoever, but its precise reading of air currents lets its user mentally visualise the surroundings and pinpoint hidden targets for lethally accurate shots.', ARRAY['Current Reading: Precisely detects the movements of the air around the Stand and relays them to Johngalli A.','Mental Visualisation: Lets Johngalli A. mentally reconstruct the position and rough silhouette of targets from the airflow.','Long-Range Support: Complements Johngalli A.''s sniping accuracy, letting him fire through blind spots.','Vulnerability to Gases: Its air reading can be distorted if different gas layers are introduced, leading Johngalli to misjudge his shots.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-741a-a5c0-96d5f45103e6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-741a-a5c0-96d5f45103e6', 'ca-ES', 'Portador: Johngalli A.

És un Stand petit no humanoide, semblant a un satèl·lit, que en Johngalli A. envia a distància per detectar els moviments de l''aire al seu voltant. No té cap capacitat de combat pròpia, però la seva lectura precisa dels corrents d''aire permet al seu portador visualitzar mentalment l''entorn i localitzar objectius amagats per disparar amb precisió letal.', ARRAY['Lectura de Corrents: Detecta amb precisió els moviments de l''aire al voltant del Stand i els transmet a en Johngalli A.','Visió Mental de l''Entorn: Reconstrueix mentalment la posició i la silueta dels objectius a partir del flux d''aire.','Suport a Llarga Distància: Complementa la punteria de franctirador d''en Johngalli A., permetent-li disparar a través de punts cecs.','Vulnerabilitat als Gasos: Introduir capes de gasos diferents distorsiona la seva lectura de l''aire i provoca errors de càlcul.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-741a-a5c0-96d5f45103e6')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Highway to Hell (Thunder McQueen)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7afa-86fc-a93b5956a99a', 'STAND', 'Highway to Hell', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7afa-86fc-a93b5956a99a', 'C', 'C', 'A', 'C', 'C', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7afa-86fc-a93b5956a99a')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7afa-86fc-a93b5956a99a', 'es-ES', 'Portador: Thunder McQueen

Es un Stand sin capacidad de combate propia que se manifiesta como cuatro protuberancias con pequeñas hélices, ligado a la naturaleza suicida de su portador. Cuando McQueen intenta quitarse la vida de cualquier forma, Highway to Hell reproduce la misma herida sobre una víctima elegida, compartiendo el daño letal sin que el propio McQueen llegue a morir.', ARRAY['Daño Compartido: Reproduce sobre la víctima elegida la misma herida que McQueen se inflige a sí mismo al intentar suicidarse.','Multiplicidad de Métodos: Se adapta a cualquier método de suicidio, ahorcamiento u ahogamiento, y lo manifiesta sobre el objetivo.','Alcance Amplio: Puede proyectar el daño compartido sobre víctimas relativamente alejadas de McQueen.','Autopreservación Paradójica: McQueen sobrevive a sus propios intentos porque el daño real recae sobre la víctima a la que se transfiere.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7afa-86fc-a93b5956a99a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7afa-86fc-a93b5956a99a', 'en-GB', 'User: Thunder McQueen

This is a Stand with no combat ability of its own, manifesting as four protrusions with tiny propellers, bound to its user''s suicidal nature. Whenever McQueen attempts to take his own life by any method, Highway to Hell reproduces the same injury on a chosen victim, sharing the lethal damage while McQueen himself never actually dies.', ARRAY['Shared Damage: Reproduces on the chosen victim the exact same injury McQueen inflicts on himself while attempting suicide.','Method Versatility: Adapts to any suicide method, such as hanging or drowning, manifesting the matching effect on the target.','Wide Range: Can project the shared damage onto victims positioned relatively far from McQueen.','Paradoxical Self-Preservation: McQueen survives his own attempts because the real damage lands on the transferred victim instead.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7afa-86fc-a93b5956a99a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7afa-86fc-a93b5956a99a', 'ca-ES', 'Portador: Thunder McQueen

És un Stand sense cap capacitat de combat pròpia que es manifesta com quatre protuberàncies amb petites hèlixs, lligat a la naturalesa suïcida del seu portador. Quan en McQueen intenta llevar-se la vida de qualsevol manera, Highway to Hell reprodueix la mateixa ferida sobre una víctima escollida, compartint el dany letal sense que en McQueen arribi a morir.', ARRAY['Dany Compartit: Reprodueix sobre la víctima escollida la mateixa ferida que en McQueen s''infligeix a si mateix en intentar suïcidar-se.','Multiplicitat de Mètodes: S''adapta a qualsevol mètode de suïcidi, penjament o ofegament, i el manifesta sobre l''objectiu.','Abast Ampli: Pot projectar el dany compartit sobre víctimes relativament allunyades d''en McQueen.','Autopreservació Paradoxal: En McQueen sobreviu als seus propis intents perquè el dany real recau sobre la víctima a qui es transfereix.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7afa-86fc-a93b5956a99a')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Marilyn Manson (Miraschon)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7741-a094-54b6af436969', 'STAND', 'Marilyn Manson', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7741-a094-54b6af436969', 'E', 'A', 'A', 'A', 'A', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7741-a094-54b6af436969')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7741-a094-54b6af436969', 'es-ES', 'Portador: Miraschon

Es un Stand humanoide con aspecto de siniestro cobrador de deudas, entregado a Miraschon mediante un DISC por Enrico Pucci. Se autodenomina ''la sombra en el corazón del perdedor'' y por ello es intangible ante cualquier ataque del perdedor o sus aliados, aunque a cambio detecta de forma infalible cualquier intento de hacer trampa en la apuesta, considerándolo derrota automática.', ARRAY['Intangibilidad del Perdedor: No puede ser golpeado por quien ha perdido la apuesta ni por sus aliados, ya que cualquier ataque lo atraviesa.','Detección de Trampas: Reconoce cualquier intento de hacer trampa, incluso pensar en agredir a Miraschon, y lo castiga como derrota inmediata.','Cobro Implacable: Aparece para reclamar el pago de la apuesta perdida con una precisión y velocidad sobrehumanas.','Aparición desde las Sombras: Se materializa y desaparece a través de las sombras de las personas implicadas en la apuesta.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7741-a094-54b6af436969')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7741-a094-54b6af436969', 'en-GB', 'User: Miraschon

This is a humanoid Stand styled like an eerie debt collector, handed to Miraschon on a DISC by Enrico Pucci. It calls itself ''the shadow within the loser''s heart'', which makes it intangible to any attack from the loser or their allies, though in exchange it infallibly detects any attempt at cheating in the wager and treats it as an automatic loss.', ARRAY['Loser''s Intangibility: Cannot be struck by whoever lost the wager or by their allies, since any attack simply passes through it.','Cheat Detection: Recognises any attempt to cheat, even just thinking of attacking Miraschon, and punishes it as an instant loss.','Relentless Collection: Appears to claim payment for the lost wager with superhuman precision and speed.','Shadow Manifestation: Materialises and vanishes through the shadows of the people involved in the wager.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7741-a094-54b6af436969')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7741-a094-54b6af436969', 'ca-ES', 'Portador: Miraschon

És un Stand humanoide amb aparença de sinistre cobrador de deutes, lliurat a la Miraschon mitjançant un DISC per Enrico Pucci. S''autoanomena ''l''ombra dins el cor del perdedor'' i per això és intangible davant qualsevol atac del perdedor o dels seus aliats, tot i que a canvi detecta de manera infal·lible qualsevol intent de fer trampa en l''aposta, considerant-ho una derrota automàtica.', ARRAY['Intangibilitat del Perdedor: No pot ser colpejat per qui ha perdut l''aposta ni pels seus aliats, ja que qualsevol atac el travessa.','Detecció de Trampes: Reconeix qualsevol intent de fer trampa, fins i tot pensar a agredir la Miraschon, i ho castiga com una derrota immediata.','Cobrament Implacable: Apareix per reclamar el pagament de l''aposta perduda amb una precisió i velocitat sobrehumanes.','Aparició des de les Ombres: Es materialitza i desapareix a través de les ombres de les persones implicades en l''aposta.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7741-a094-54b6af436969')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Jumpin' Jack Flash (Lang Rangler)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7e9b-afec-88b99aac115d', 'STAND', 'Jumpin'' Jack Flash', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7e9b-afec-88b99aac115d', 'B', 'C', 'B', 'A', 'D', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e9b-afec-88b99aac115d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e9b-afec-88b99aac115d', 'es-ES', 'Portador: Lang Rangler

Es un Stand humanoide de combate cercano cuya habilidad exclusiva es convertir en el centro de un campo de gravedad cero a cualquier víctima sobre la que escupe. Cuenta además con centrifugadoras integradas en las muñecas capaces de lanzar proyectiles a gran velocidad, combinando control gravitacional con ataques a distancia corta.', ARRAY['Gravedad Cero por Contacto: Su saliva convierte al objetivo en el centro de una zona de gravedad nula.','Centrifugadoras de Muñeca: Lanza proyectiles a alta velocidad gracias a mecanismos centrífugos incorporados en sus brazos.','Resistencia Notable: Posee una gran capacidad de aguante que le permite sostener combates prolongados.','Precisión Limitada: Su puntería y su capacidad de desarrollo son sus puntos más débiles, lo que dificulta los ataques certeros a distancia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e9b-afec-88b99aac115d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e9b-afec-88b99aac115d', 'en-GB', 'User: Lang Rangler

This is a close-range humanoid Stand whose signature power is turning any victim it spits on into the centre of a zero-gravity field. It also has built-in centrifuges on its wrists capable of launching projectiles at high speed, combining gravity control with close-range attacks.', ARRAY['Zero Gravity on Contact: Turns whoever is hit by its spit into the centre of a zero-gravity zone, making everything around them float.','Wrist Centrifuges: Fires projectiles at high velocity thanks to centrifugal mechanisms built into its arms.','Notable Endurance: Has considerable staying power that lets it hold up in prolonged fights.','Limited Precision: Its aim and growth capacity are its weakest points, making accurate ranged attacks difficult.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e9b-afec-88b99aac115d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e9b-afec-88b99aac115d', 'ca-ES', 'Portador: Lang Rangler

És un Stand humanoide de combat proper l''habilitat exclusiva del qual és convertir en el centre d''un camp de gravetat zero qualsevol víctima sobre la qual escup. També compta amb centrifugadores integrades als canells capaces de llançar projectils a gran velocitat, combinant control gravitacional amb atacs a curta distància.', ARRAY['Gravetat Zero per Contacte: Converteix qui rep la seva saliva en el centre d''una zona de gravetat nul·la, fent flotar tot al seu voltant.','Centrifugadores de Canell: Llança projectils a alta velocitat gràcies a mecanismes centrífugs incorporats als seus braços.','Resistència Notable: Té una gran capacitat d''aguant que li permet sostenir combats prolongats.','Precisió Limitada: La punteria i la capacitat de desenvolupament són els seus punts més febles.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e9b-afec-88b99aac115d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Limp Bizkit (Sports Maxx)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-78ff-90ba-44cdb6f0a03e', 'STAND', 'Limp Bizkit', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-78ff-90ba-44cdb6f0a03e', 'NULL', 'B', 'B', 'A', 'C', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-78ff-90ba-44cdb6f0a03e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-78ff-90ba-44cdb6f0a03e', 'es-ES', 'Portador: Sports Maxx

Es un Stand de habilidad pura, sin poder de combate directo, que permite a Sports Maxx reanimar cadáveres y restos humanos parciales convirtiéndolos en zombis completamente invisibles bajo su control. Estos zombis actúan como espías o asesinos silenciosos, revelando su posición solo si quedan cubiertos por algún líquido o sustancia.', ARRAY['Reanimación de Cadáveres: Transforma cuerpos muertos, enteros o en pedazos, en zombis obedientes que siguen sus órdenes.','Invisibilidad Total: Los zombis creados son invisibles a simple vista, salvo que algo los recubra y delate su forma.','Red de Vigilancia: Emplea a sus zombis para espiar, emboscar o recopilar información sin levantar sospechas.','Resistencia Sostenida: Carece de poder destructivo, pero su gran resistencia mantiene activos a numerosos zombis mucho tiempo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-78ff-90ba-44cdb6f0a03e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-78ff-90ba-44cdb6f0a03e', 'en-GB', 'User: Sports Maxx

This is a pure ability Stand with no direct combat power that lets Sports Maxx reanimate corpses and partial human remains, turning them into completely invisible zombies under his control. These zombies act as silent spies or assassins, only giving away their position if some liquid or substance coats them.', ARRAY['Corpse Reanimation: Turns dead bodies, whole or in pieces, into obedient zombies that follow his orders.','Total Invisibility: The zombies it creates are invisible to the naked eye unless something coats them and reveals their shape.','Surveillance Network: Uses his zombies to spy, ambush or gather information without raising suspicion.','Sustained Endurance: Though it lacks destructive power, its great endurance lets him keep numerous zombies active for a long time.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-78ff-90ba-44cdb6f0a03e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-78ff-90ba-44cdb6f0a03e', 'ca-ES', 'Portador: Sports Maxx

És un Stand d''habilitat pura, sense poder de combat directe, que permet a en Sports Maxx reanimar cadàvers i restes humanes parcials convertint-los en zombis completament invisibles sota el seu control. Aquests zombis actuen com a espies o assassins silenciosos, i només revelen la seva posició si queden coberts per algun líquid o substància.', ARRAY['Reanimació de Cadàvers: Transforma cossos morts, sencers o a trossos, en zombis obedients que segueixen les seves ordres.','Invisibilitat Total: Els zombis creats són invisibles a simple vista, tret que alguna cosa els cobreixi i en delati la forma.','Xarxa de Vigilància: Fa servir els seus zombis per espiar, emboscar o recopilar informació sense aixecar sospites.','Resistència Sostinguda: Manca de poder destructiu, però la seva gran resistència manté actius nombrosos zombis molt de temps.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-78ff-90ba-44cdb6f0a03e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Survivor (Guccio)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7e4b-95f5-5ccf64d489cc', 'STAND', 'Survivor', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7e4b-95f5-5ccf64d489cc', 'E', 'E', 'E', 'C', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e4b-95f5-5ccf64d489cc')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e4b-95f5-5ccf64d489cc', 'es-ES', 'Portador: Guccio

Es un Stand débil formado por un enjambre de pequeñas entidades que se desplazan a través del agua o líquidos similares, emitiendo una leve señal eléctrica que multiplica la agresividad de cualquiera que entre en contacto con ella. Quienes son afectados terminan enzarzados en peleas violentas sin control, sin que Guccio necesite intervenir directamente.', ARRAY['Enjambre Acuático: Se desplaza en pequeños grupos a través del agua, la sangre o cualquier líquido corporal para alcanzar a sus víctimas.','Amplificación de Agresividad: Emite una señal eléctrica que multiplica la ira y el instinto violento de quien la recibe.','Incitación al Caos: Provoca peleas entre las víctimas afectadas, haciendo que se ataquen entre ellas sin motivo aparente.','Bajo Poder Individual: Se considera uno de los Stands más débiles conocidos, sin capacidad ofensiva ni defensiva propia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e4b-95f5-5ccf64d489cc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e4b-95f5-5ccf64d489cc', 'en-GB', 'User: Guccio

This is a weak Stand made up of a swarm of tiny entities that travel through water or similar liquids, giving off a faint electrical signal that multiplies the aggression of anyone it touches. Those affected end up locked in uncontrolled, violent fights without Guccio needing to step in directly.', ARRAY['Aquatic Swarm: Travels in small groups through water, blood or any bodily fluid to reach its victims.','Aggression Amplification: Gives off an electrical signal that multiplies the anger and violent instincts of whoever receives it.','Inciting Chaos: Triggers fights among affected victims, making them attack each other for no apparent reason.','Low Individual Power: Considered one of the weakest known Stands, with no offensive or defensive capability of its own.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e4b-95f5-5ccf64d489cc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e4b-95f5-5ccf64d489cc', 'ca-ES', 'Portador: Guccio

És un Stand feble format per un eixam de petites entitats que es desplacen a través de l''aigua o líquids similars, i que emeten un lleu senyal elèctric que multiplica l''agressivitat de qui hi entra en contacte. Els afectats acaben embolicats en baralles violentes i descontrolades, sense que en Guccio hagi d''intervenir directament.', ARRAY['Eixam Aquàtic: Es desplaça en petits grups a través de l''aigua, la sang o qualsevol líquid corporal per arribar a les víctimes.','Amplificació de l''Agressivitat: Emet un senyal elèctric que multiplica la ràbia i l''instint violent de qui el rep.','Incitació al Caos: Provoca baralles entre les víctimes afectades, fent que s''ataquin entre elles sense motiu aparent.','Poder Individual Baix: Es considera un dels Stands més febles coneguts, sense capacitat ofensiva ni defensiva pròpia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e4b-95f5-5ccf64d489cc')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Planet Waves (Viviano Westwood)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-716a-99e8-1ac9ac8ff7ff', 'STAND', 'Planet Waves', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff', 'A', 'B', 'A', 'A', 'E', 'E' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff', 'es-ES', 'Portador: Viviano Westwood

Es un Stand con forma humanoide desollada que atrae pequeños meteoritos desde el espacio hacia la posición de su portador, convirtiéndolos en proyectiles mortales contra cualquiera que se encuentre cerca. Las rocas caen a gran velocidad y alcanzan temperaturas de miles de grados durante la reentrada, por lo que resulta extremadamente peligroso enfrentarse a Westwood a campo abierto.', ARRAY['Atracción de Meteoritos: Convoca pequeñas rocas espaciales que caen directamente hacia su portador y su entorno inmediato.','Impacto Silencioso: Los meteoritos llegan a tal velocidad que no pueden oírse hasta el momento del impacto.','Calor Extremo: Las rocas se calientan hasta unos 3000 °C durante la caída, aumentando su letalidad al golpear.','Amenaza a Campo Abierto: Su poder resulta especialmente peligroso en espacios exteriores sin cobertura frente al cielo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff', 'en-GB', 'User: Viviano Westwood

This is a Stand with a flayed humanoid form that pulls small meteorites down from space toward its user''s position, turning them into deadly projectiles against anyone nearby. The rocks fall at great speed and reach temperatures of thousands of degrees on reentry, making it extremely dangerous to face Westwood out in the open.', ARRAY['Meteorite Attraction: Summons small space rocks that fall directly toward its user and the immediate surroundings.','Silent Impact: The meteorites travel so fast they can''t be heard until the moment of impact.','Extreme Heat: The rocks heat up to around 3,000°C during the fall, increasing their lethality on impact.','Open-Air Threat: Its power is especially dangerous in outdoor spaces with no cover from the sky above.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff', 'ca-ES', 'Portador: Viviano Westwood

És un Stand amb forma humanoide escorxada que atrau petits meteorits des de l''espai cap a la posició del seu portador, convertint-los en projectils mortals contra qualsevol que hi hagi a prop. Les roques cauen a gran velocitat i assoleixen temperatures de milers de graus durant la reentrada, cosa que fa extremadament perillós enfrontar-se a en Westwood a camp obert.', ARRAY['Atracció de Meteorits: Convoca petites roques espacials que cauen directament cap al seu portador i l''entorn immediat.','Impacte Silenciós: Els meteorits arriben a tanta velocitat que no es poden sentir fins al moment de l''impacte.','Calor Extrema: Les roques s''escalfen fins a uns 3000 °C durant la caiguda, augmentant-ne la letalitat en colpejar.','Amenaça a Camp Obert: El seu poder és especialment perillós en espais exteriors sense cobertura respecte al cel.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Dragon's Dream (Kenzou)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7a83-ae22-04b33914881e', 'STAND', 'Dragon''s Dream', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7a83-ae22-04b33914881e', 'NULL', 'NULL', 'NULL', 'A', 'NULL', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7a83-ae22-04b33914881e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7a83-ae22-04b33914881e', 'es-ES', 'Portador: Kenzou

Es un Stand sentiente e independiente que carece de capacidad de ataque directo, dedicado por completo a analizar el feng shui del entorno para determinar los ángulos y posiciones más favorables o desfavorables en un combate. Actúa casi como un árbitro imparcial, guiando a Kenzou hacia el golpe perfecto mientras le advierte de las posiciones que deben evitarse.', ARRAY['Análisis de Feng Shui: Estudia el entorno para identificar los puntos de máxima fortuna y máximo peligro en cualquier situación.','Guía de Asesinato: Señala a Kenzou el ángulo y el momento exactos para asestar un golpe letal.','Esquiva Predictiva: Permite anticipar y evitar ataques enemigos gracias al análisis constante de las posiciones desfavorables.','Imparcialidad Sentiente: Actúa con voluntad propia y una ética estricta, negándose a favorecer injustamente a su portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7a83-ae22-04b33914881e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7a83-ae22-04b33914881e', 'en-GB', 'User: Kenzou

This is a sentient, independent Stand with no direct attacking ability, devoted entirely to reading the feng shui of its surroundings to work out the most favourable or unfavourable angles and positions in a fight. It acts almost like an impartial referee, guiding Kenzou toward the perfect strike while warning him of positions to avoid.', ARRAY['Feng Shui Analysis: Studies its surroundings to pinpoint the luckiest and most dangerous spots in any situation.','Assassination Guidance: Points Kenzou to the exact angle and moment for a lethal strike.','Predictive Dodging: Lets him anticipate and avoid enemy attacks through constant analysis of unfavourable positions.','Sentient Impartiality: Acts with a will of its own and a strict code of fairness, refusing to unjustly favour its user.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7a83-ae22-04b33914881e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7a83-ae22-04b33914881e', 'ca-ES', 'Portador: Kenzou

És un Stand sensible i independent que no té capacitat d''atac directe, dedicat completament a analitzar el feng shui de l''entorn per determinar els angles i les posicions més favorables o desfavorables en un combat. Actua gairebé com un àrbitre imparcial, guiant en Kenzou cap al cop perfecte mentre l''avisa de les posicions que cal evitar.', ARRAY['Anàlisi de Feng Shui: Estudia l''entorn per identificar els punts de màxima fortuna i màxim perill en qualsevol situació.','Guia d''Assassinat: Assenyala a en Kenzou l''angle i el moment exactes per assestar un cop letal.','Esquiva Predictiva: Permet anticipar i evitar atacs enemics gràcies a l''anàlisi constant de les posicions desfavorables.','Imparcialitat Sensible: Actua amb voluntat pròpia i una ètica estricta, i es nega a afavorir injustament el seu portador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7a83-ae22-04b33914881e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Yo-Yo Ma (D an G)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7464-933a-5baebc7e633f', 'STAND', 'Yo-Yo Ma', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7464-933a-5baebc7e633f', 'C', 'D', 'A', 'A', 'D', 'C' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7464-933a-5baebc7e633f')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7464-933a-5baebc7e633f', 'es-ES', 'Portador: D an G

Es un Stand automático y sentiente con forma de enano robusto, prácticamente indestructible, que actúa como un guardián autónomo. Cuando no es observado por más de una persona a la vez, segrega una saliva ácida capaz de disolver cualquier parte del cuerpo de su objetivo, mientras que en presencia de varios testigos se comporta como un sirviente dócil que obedece órdenes.', ARRAY['Saliva Ácida: Disuelve la carne y los tejidos de su objetivo al entrar en contacto con su baba corrosiva.','Actuación Encubierta: Solo ataca cuando cree que no hay más de un testigo observándolo.','Invulnerabilidad: Resiste disparos y golpes directos sin sufrir el más mínimo daño.','Autonomía Total: Actúa sin necesidad de órdenes constantes, siguiendo su propio criterio para proteger a D an G.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7464-933a-5baebc7e633f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7464-933a-5baebc7e633f', 'en-GB', 'User: D an G

This is an automatic, sentient Stand shaped like a stocky dwarf, virtually indestructible, that acts as an autonomous guardian. When it isn''t being watched by more than one person at once, it secretes an acidic saliva able to dissolve any part of its target''s body, while in front of several witnesses it behaves like a docile servant that follows orders.', ARRAY['Acidic Saliva: Dissolves its target''s flesh and tissue on contact with its corrosive drool.','Covert Action: Only attacks when it believes no more than one witness is watching.','Invulnerability: Withstands gunshots and direct blows without taking the slightest damage.','Full Autonomy: Acts without needing constant orders, following its own judgement to protect D an G.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7464-933a-5baebc7e633f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7464-933a-5baebc7e633f', 'ca-ES', 'Portador: D an G

És un Stand automàtic i sensible amb forma de nan robust, pràcticament indestructible, que actua com un guardià autònom. Quan no l''observa més d''una persona alhora, segrega una saliva àcida capaç de dissoldre qualsevol part del cos del seu objectiu, mentre que davant de diversos testimonis es comporta com un servent dòcil que obeeix ordres.', ARRAY['Saliva Àcida: Dissol la carn i els teixits del seu objectiu en contacte amb la seva bava corrosiva.','Actuació Encoberta: Només ataca quan creu que no hi ha més d''un testimoni observant-lo.','Invulnerabilitat: Resisteix trets i cops directes sense patir el més mínim dany.','Autonomia Total: Actua sense necessitat d''ordres constants, seguint el seu propi criteri per protegir en D an G.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7464-933a-5baebc7e633f')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Green, Green Grass of Home (Green Baby)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7b07-b05d-9293cb7752f5', 'STAND', 'Green, Green Grass of Home', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7b07-b05d-9293cb7752f5', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7b07-b05d-9293cb7752f5')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7b07-b05d-9293cb7752f5', 'es-ES', 'Portador: Green Baby

Es un Stand humanoide pequeño y enmascarado que protege instintivamente al Green Baby encogiendo el tamaño de cualquier persona u objeto que se le acerque, cuanto más próxima esté la amenaza. El efecto puede continuar de forma indefinida, llegando a hacer que el objetivo se reduzca hasta desaparecer por completo.', ARRAY['Encogimiento Progresivo: Reduce proporcionalmente el tamaño de quien se aproxima al bebé, cuanto más cerca está.','Amenaza Silenciosa: Actúa de forma pasiva y constante, sin necesidad de que el bebé sea consciente de su propio poder.','Reducción sin Límite: Puede llegar a encoger a su objetivo hasta un punto en el que prácticamente desaparece.','Protección Instintiva: Funciona como un mecanismo de defensa automático en torno al Green Baby.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7b07-b05d-9293cb7752f5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7b07-b05d-9293cb7752f5', 'en-GB', 'User: Green Baby

This is a small, masked humanoid Stand that instinctively protects the Green Baby by shrinking anyone or anything that approaches, the closer the threat gets. The effect can go on indefinitely, eventually shrinking the target until it all but vanishes.', ARRAY['Progressive Shrinking: Proportionally shrinks whoever approaches the baby, the closer they get.','Silent Threat: Acts passively and constantly, without the baby needing to be aware of its own power.','Unlimited Reduction: Can shrink its target down to the point where it virtually disappears.','Instinctive Protection: Works as an automatic defence mechanism around the Green Baby.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7b07-b05d-9293cb7752f5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7b07-b05d-9293cb7752f5', 'ca-ES', 'Portador: Green Baby

És un Stand humanoide petit i emmascarat que protegeix instintivament el Green Baby encongint la mida de qualsevol persona o objecte que se li acosti, com més a prop hi ha l''amenaça. L''efecte pot continuar de manera indefinida, fins al punt de fer que l''objectiu es redueixi fins a desaparèixer completament.', ARRAY['Encongiment Progressiu: Redueix proporcionalment la mida de qui s''apropa al nadó, com més a prop hi és.','Amenaça Silenciosa: Actua de manera passiva i constant, sense que el nadó hagi de ser conscient del seu propi poder.','Reducció sense Límit: Pot arribar a encongir el seu objectiu fins a un punt en què pràcticament desapareix.','Protecció Instintiva: Funciona com un mecanisme de defensa automàtic al voltant del Green Baby.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7b07-b05d-9293cb7752f5')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Jail House Lock (Miuccia Miuller)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-74b5-af83-f39958eb6520', 'STAND', 'Jail House Lock', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-74b5-af83-f39958eb6520', 'NULL', 'C', 'B', 'A', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74b5-af83-f39958eb6520')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74b5-af83-f39958eb6520', 'es-ES', 'Portador: Miuccia Miuller

Es un Stand no combativo ligado a los muros de la prisión de G.D., cuya habilidad exclusiva es "encerrar" mentalmente a cualquiera que intente escapar de sus límites, degradando su memoria hasta dejarlo incapaz de idear una fuga coherente. Miuccia lo usa como guardián psicológico del recinto, compensando la nula capacidad ofensiva del Stand con un control absoluto sobre quien cruce el perímetro.', ARRAY['Bloqueo Carcelario: Ata mentalmente a quien intente traspasar los muros de la prisión, activándose al cruzar el límite.','Amnesia Selectiva: La víctima solo recuerda tres cosas a la vez; al intentar una cuarta, olvida la primera.','Aislamiento Mental: Explota la pérdida de memoria para desorientar al objetivo e impedirle planear una fuga con sentido.','Cuerpo No Combativo: Carece de manos y de fuerza ofensiva directa, dependiendo por completo de esta manipulación psicológica.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74b5-af83-f39958eb6520')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74b5-af83-f39958eb6520', 'en-GB', 'User: Miuccia Miuller

This is a non-combative Stand bound to the walls of G.D. Prison, whose signature power is mentally "locking" anyone who tries to escape its boundaries, degrading their memory until they can no longer plan a coherent escape. Miuccia uses it as the facility''s psychological warden, making up for the Stand''s total lack of offensive capability with absolute control over anyone who crosses the perimeter.', ARRAY['Prison Lock: Mentally binds anyone who tries to cross the walls of G.D. Prison, triggering the instant the boundary is crossed.','Selective Amnesia: The victim can only remember three things at a time; trying to recall a fourth makes them automatically forget the first.','Mental Isolation: Exploits the memory loss to disorient the target and stop them from planning a coherent escape.','Non-Combat Body: Lacks hands and any direct offensive strength, relying entirely on this psychological manipulation.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74b5-af83-f39958eb6520')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74b5-af83-f39958eb6520', 'ca-ES', 'Portador: Miuccia Miuller

És un Stand no combatiu lligat als murs de la presó de G.D., l''habilitat exclusiva del qual és "tancar" mentalment qualsevol que intenti escapar dels seus límits, degradant-li la memòria fins a deixar-lo incapaç d''idear una fugida coherent. Na Miuccia el fa servir com a guardià psicològic del recinte, compensant la nul·la capacitat ofensiva del Stand amb un control absolut sobre qui travessi el perímetre.', ARRAY['Bloqueig Carcerari: Lliga mentalment qualsevol que intenti travessar els murs de la presó G.D., activant-se tan bon punt es creua el límit.','Amnèsia Selectiva: La víctima només pot recordar tres coses alhora; quan intenta recordar-ne una quarta, oblida automàticament la primera.','Aïllament Mental: Explota la pèrdua de memòria per desorientar l''objectiu i impedir-li planejar una fugida amb sentit.','Cos No Combatiu: No té mans ni força ofensiva directa, i depèn totalment d''aquesta manipulació psicològica.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74b5-af83-f39958eb6520')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Bohemian Rhapsody (Ungalo)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-70fd-b825-e5dfc5c1cdc7', 'STAND', 'Bohemian Rhapsody', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7', 'NULL', 'NULL', 'INFINITE', 'A', 'NULL', 'A' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7', 'es-ES', 'Portador: Ungalo

Es un fenómeno más que un Stand con forma física, capaz de convertir en realidad a cualquier personaje de ficción que aparezca en libros, dibujos o ilustraciones dentro de su alcance. Cuando alguien contempla la imagen de un personaje que le apasiona, su alma es arrastrada a ese papel, adoptando su apariencia y su destino narrativo. Al no tener cuerpo ni límite de distancia, resulta prácticamente imposible de combatir por medios convencionales.', ARRAY['Encarnación de Personajes: Convierte a quien admira un personaje de ficción en ese personaje, con su apariencia y destino.','Alcance Global: Actúa sobre el mundo entero de manera simultánea, sin límite práctico de distancia.','Fenómeno Sin Forma: No posee cuerpo físico ni personalidad propia, lo que lo hace virtualmente invulnerable al combate directo.','Reescritura de la Realidad: Fusiona ficción y realidad, dando existencia física a criaturas y objetos imaginarios.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7', 'en-GB', 'User: Ungalo

This is more a phenomenon than a Stand with a physical form, able to make any fictional character from a book, drawing or illustration real within its reach. When someone gazes at the image of a character they love, their soul is pulled into that role, taking on its appearance and narrative fate. With no body and no practical distance limit, it is next to impossible to fight through conventional means.', ARRAY['Character Incarnation: Turns whoever admires the image of a fictional character into that very character, appearance and fate included.','Global Reach: Acts on the entire world simultaneously, with no practical distance limit.','Formless Phenomenon: Has no physical body or personality of its own, making it virtually invulnerable to direct combat.','Reality Rewrite: Merges fiction and reality, giving physical existence to imaginary creatures and objects.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7', 'ca-ES', 'Portador: Ungalo

És més aviat un fenomen que no pas un Stand amb forma física, capaç de convertir en realitat qualsevol personatge de ficció que aparegui en llibres, dibuixos o il·lustracions dins el seu abast. Quan algú contempla la imatge d''un personatge que l''apassiona, la seva ànima és arrossegada a aquell paper, n''adopta l''aparença i el destí narratiu. Com que no té cos ni límit de distància, resulta pràcticament impossible de combatre per mitjans convencionals.', ARRAY['Encarnació de Personatges: Converteix qui admira un personatge de ficció en aquell personatge, amb la seva aparença i destí.','Abast Global: Actua sobre tot el món simultàniament, sense límit pràctic de distància.','Fenomen Sense Forma: No té cos físic ni personalitat pròpia, cosa que el fa virtualment invulnerable al combat directe.','Reescriptura de la Realitat: Fusiona ficció i realitat, i dona existència física a criatures i objectes imaginaris.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Sky High (Rikiel)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7ff6-a691-ef06758e3381', 'STAND', 'Sky High', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7ff6-a691-ef06758e3381', 'NULL', 'NULL', 'B', 'C', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7ff6-a691-ef06758e3381')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7ff6-a691-ef06758e3381', 'es-ES', 'Portador: Rikiel

Es un Stand de rango que permite controlar una colonia invisible de criaturas voladoras llamadas Rods, capaces de succionar el calor corporal de sus víctimas y provocar ceguera al atacar sus párpados. Rikiel, un joven enfermizo e inseguro, ve cómo la fuerza de este Stand fluctúa con su propia confianza en sí mismo, volviéndose mucho más peligroso cuando logra sobreponerse a sus miedos.', ARRAY['Control de las Rods: Dirige una colonia de criaturas voladoras invisibles para atacar a distancia sin ser detectado.','Absorción de Calor: Las Rods succionan el calor corporal de sus víctimas, pudiendo provocarles hipotermia progresiva.','Ceguera Focalizada: Puede ordenar a las Rods que ataquen los párpados de un objetivo para dejarlo ciego.','Poder Ligado a la Confianza: La eficacia del Stand crece o mengua según el estado emocional y la seguridad de Rikiel.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7ff6-a691-ef06758e3381')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7ff6-a691-ef06758e3381', 'en-GB', 'User: Rikiel

This is a ranged Stand that lets its user control an invisible colony of flying creatures called Rods, able to suck the body heat from their victims and cause blindness by attacking their eyelids. Rikiel, a sickly and insecure young man, finds this Stand''s strength fluctuates with his own self-confidence, becoming far more dangerous once he manages to overcome his fears.', ARRAY['Rod Control: Directs an invisible colony of flying creatures to attack from a distance undetected.','Heat Absorption: The Rods suck the body heat from their victims, causing progressive hypothermia.','Focused Blindness: Can order the Rods to attack a target''s eyelids to blind them.','Confidence-Bound Power: The Stand''s effectiveness rises or falls with Rikiel''s emotional state and self-assurance.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7ff6-a691-ef06758e3381')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7ff6-a691-ef06758e3381', 'ca-ES', 'Portador: Rikiel

És un Stand d''abast que permet controlar una colònia invisible de criatures voladores anomenades Rods, capaces de succionar la calor corporal de les seves víctimes i provocar ceguesa atacant-los les parpelles. En Rikiel, un jove malaltís i insegur, veu com la força d''aquest Stand fluctua amb la seva pròpia confiança, i es torna molt més perillós quan aconsegueix superar les seves pors.', ARRAY['Control de les Rods: Dirigeix una colònia de criatures voladores invisibles per atacar a distància sense ser detectat.','Absorció de Calor: Les Rods succionen la calor corporal de les víctimes i els poden provocar hipotèrmia progressiva.','Ceguesa Focalitzada: Pot ordenar a les Rods que ataquin les parpelles d''un objectiu per deixar-lo cec.','Poder Lligat a la Confiança: L''eficàcia del Stand augmenta o minva segons l''estat emocional i la seguretat d''en Rikiel.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7ff6-a691-ef06758e3381')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Under World (Donatello Versus)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-74ea-83b4-dd465b93e481', 'STAND', 'Under World', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-74ea-83b4-dd465b93e481', 'NULL', 'C', 'A', 'C', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74ea-83b4-dd465b93e481')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74ea-83b4-dd465b93e481', 'es-ES', 'Portador: Donatello Versus

Es un Stand no combativo capaz de excavar la tierra para desenterrar la "memoria" de cualquier suceso ocurrido sobre el terreno de Orlando, Florida, sin límite alguno de antigüedad. Versus lo usa como herramienta de información e investigación, compensando su total falta de poder ofensivo con la capacidad de sacar a la luz pistas, objetos y sucesos del pasado que le dan ventaja frente a sus enemigos.', ARRAY['Excavación de Recuerdos: Escarba la tierra para desenterrar la memoria de cualquier suceso pasado ocurrido en Orlando.','Sin Límite Temporal: Recupera eventos de cualquier época, por antigua que sea, siempre que haya ocurrido en la zona.','Herramienta No Combativa: Carece de capacidad ofensiva directa, dependiendo por completo de la información desenterrada.','Reconstrucción de Escenas: Permite extraer objetos y pistas de sucesos pasados para usarlos a su favor en el presente.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74ea-83b4-dd465b93e481')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74ea-83b4-dd465b93e481', 'en-GB', 'User: Donatello Versus

This is a non-combative Stand able to dig through the earth to unearth the "memory" of any event that took place on the ground of Orlando, Florida, with no limit on how far back it can reach. Versus uses it as an information-gathering tool, making up for its total lack of offensive power with the ability to bring past clues, objects and events to light that give him the edge over his enemies.', ARRAY['Memory Excavation: Digs through the earth to unearth the memory of any past event that took place in Orlando.','No Time Limit: Retrieves events from any era, however old, as long as they happened in the area.','Non-Combat Tool: Has no direct offensive capability, relying entirely on the information it unearths.','Scene Reconstruction: Lets him extract objects and clues from past events to use to his advantage in the present.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74ea-83b4-dd465b93e481')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-74ea-83b4-dd465b93e481', 'ca-ES', 'Portador: Donatello Versus

És un Stand no combatiu capaç d''excavar la terra per desenterrar la "memòria" de qualsevol succés ocorregut sobre el terreny d''Orlando, Florida, sense cap límit d''antiguitat. En Versus el fa servir com a eina d''informació i investigació, compensant la seva absoluta manca de poder ofensiu amb la capacitat de treure a la llum pistes, objectes i successos del passat que li donen avantatge davant els seus enemics.', ARRAY['Excavació de Records: Escarba la terra per desenterrar la memòria de qualsevol succés passat ocorregut a Orlando.','Sense Límit Temporal: Recupera esdeveniments de qualsevol època, per antiga que sigui, sempre que hagi passat a la zona.','Eina No Combativa: No té capacitat ofensiva directa i depèn totalment de la informació desenterrada.','Reconstrucció d''Escenes: Li permet extreure objectes i pistes de successos passats per fer-los servir al seu favor en el present.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-74ea-83b4-dd465b93e481')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Boiling Water Stand (Enrico Pucci)
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a0e753-d80f-7e2d-81c7-c6d65da01169', 'STAND', 'Boiling Water Stand', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO stands (id, attack_power, speed, attack_range, endurance, "precision", potential)
  SELECT '01a0e753-d80f-7e2d-81c7-c6d65da01169', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL', 'NULL' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e2d-81c7-c6d65da01169')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e2d-81c7-c6d65da01169', 'es-ES', 'Portador: Enrico Pucci

Es un Stand sin nombre propio en el canon, cuyo DISC original fue robado por White Snake y cuyo dueño legítimo nunca se revela. Pucci lo emplea insertando el DISC en un cadáver o cuerpo cercano para activar su única habilidad: hacer hervir instantáneamente cualquier agua con la que entre en contacto la víctima, convirtiéndolo en un arma letal contra oponentes hechos de agua.', ARRAY['Ebullición Instantánea: Hace hervir de golpe cualquier agua con la que entra en contacto la víctima.','Arma Letal Contra Seres Acuáticos: Resulta devastador contra objetivos hechos de agua o líquidos, como F.F.','Uso Mediante DISC: Pucci lo activa insertando el DISC robado en un cuerpo, ya que no es su Stand original.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e2d-81c7-c6d65da01169')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e2d-81c7-c6d65da01169', 'en-GB', 'User: Enrico Pucci

This is a Stand with no proper name in canon, whose original DISC was stolen by White Snake and whose rightful owner is never revealed. Pucci uses it by inserting the DISC into a nearby body or corpse to trigger its one power: instantly boiling any water the victim comes into contact with, making it a lethal weapon against opponents made of water.', ARRAY['Instant Boiling: Instantly boils any water the victim comes into contact with.','Lethal Weapon Against Water-Based Beings: Devastating against targets made of water or liquid, such as F.F.','Used Through a DISC: Pucci activates it by inserting the stolen DISC into a body, since it is not his own original Stand.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e2d-81c7-c6d65da01169')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a0e753-d80f-7e2d-81c7-c6d65da01169', 'ca-ES', 'Portador: Enrico Pucci

És un Stand sense nom propi en el cànon, el DISC original del qual va ser robat per White Snake i el propietari legítim del qual mai no es revela. En Pucci el fa servir inserint el DISC en un cadàver o cos proper per activar-ne l''única habilitat: fer bullir a l''instant qualsevol aigua amb què entri en contacte la víctima, cosa que el converteix en una arma letal contra oponents fets d''aigua.', ARRAY['Ebullició Instantània: Fa bullir de cop qualsevol aigua amb què entra en contacte la víctima.','Arma Letal Contra Éssers Aquàtics: Resulta devastador contra objectius fets d''aigua o líquids, com F.F.','Ús Mitjançant DISC: En Pucci l''activa inserint el DISC robat en un cos, ja que no és el seu Stand original.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a0e753-d80f-7e2d-81c7-c6d65da01169')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Evolution chains: Weather Report -> Weather Report: Heavy Weather;
-- Whitesnake -> C-MOON -> Made in Heaven (parent-first order matters
-- only for readability here - each UPDATE is independent and idempotent).
UPDATE stands SET evolves_from_id = '01a0e753-d80f-77af-8a40-056b85e046bb' WHERE id = '01a0e753-d80f-7f2e-9906-43249b2bb2b4' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e753-d80f-77af-8a40-056b85e046bb');
UPDATE stands SET evolves_from_id = '01a0e753-d80f-776e-94c5-141b2b6d4824' WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e753-d80f-776e-94c5-141b2b6d4824');
UPDATE stands SET evolves_from_id = '01a0e753-d80f-7093-a7c0-2728828eda74' WHERE id = '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3' AND evolves_from_id IS NULL AND EXISTS (SELECT 1 FROM stands WHERE id = '01a0e753-d80f-7093-a7c0-2728828eda74');

-- +goose Down
-- Cascades to stands + power_translations rows via ON DELETE CASCADE.
DELETE FROM powers WHERE id IN (
  '01a0e753-d80e-7439-9ade-39ef1ad32169',
  '01a0e753-d80f-7314-b376-08684b35ecc9',
  '01a0e753-d80f-7656-8d29-3dcfdb798513',
  '01a0e753-d80f-77f3-bc44-db5294ed4574',
  '01a0e753-d80f-7762-a95f-f4e65d73a1a2',
  '01a0e753-d80f-77af-8a40-056b85e046bb',
  '01a0e753-d80f-7f2e-9906-43249b2bb2b4',
  '01a0e753-d80f-776e-94c5-141b2b6d4824',
  '01a0e753-d80f-7093-a7c0-2728828eda74',
  '01a0e753-d80f-73fe-bb6e-4d44f1a91aa3',
  '01a0e753-d80f-7221-866b-28cd22971b47',
  '01a0e753-d80f-741a-a5c0-96d5f45103e6',
  '01a0e753-d80f-7afa-86fc-a93b5956a99a',
  '01a0e753-d80f-7741-a094-54b6af436969',
  '01a0e753-d80f-7e9b-afec-88b99aac115d',
  '01a0e753-d80f-78ff-90ba-44cdb6f0a03e',
  '01a0e753-d80f-7e4b-95f5-5ccf64d489cc',
  '01a0e753-d80f-716a-99e8-1ac9ac8ff7ff',
  '01a0e753-d80f-7a83-ae22-04b33914881e',
  '01a0e753-d80f-7464-933a-5baebc7e633f',
  '01a0e753-d80f-7b07-b05d-9293cb7752f5',
  '01a0e753-d80f-74b5-af83-f39958eb6520',
  '01a0e753-d80f-70fd-b825-e5dfc5c1cdc7',
  '01a0e753-d80f-7ff6-a691-ef06758e3381',
  '01a0e753-d80f-74ea-83b4-dd465b93e481',
  '01a0e753-d80f-7e2d-81c7-c6d65da01169'
);
