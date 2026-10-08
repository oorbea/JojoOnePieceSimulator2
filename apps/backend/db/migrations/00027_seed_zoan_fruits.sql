-- +goose Up
-- Seeds the manga-canon Zoan / Mythical Zoan Devil Fruits missing from the catalogue.
-- Hand-authored new content (wiki-sourced), not a catalogsync dump; see
-- ObsidianVault/catalog-seed-p7-p8-fruits-2026-10-08.md. Runs in every
-- environment including prod (no ENVSUB guard); images are added later by an
-- admin, so rows keep the powers.picture* defaults. Idempotent against a fruit an
-- admin already created: powers inserts are ON CONFLICT DO NOTHING (id or name),
-- dependent inserts are gated on that powers row existing.

-- Uma Uma no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bded-7a49-a17b-127f0c89d0b6', 'DEVIL_FRUIT', 'Uma Uma no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bded-7a49-a17b-127f0c89d0b6', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-7a49-a17b-127f0c89d0b6')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-7a49-a17b-127f0c89d0b6', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de caballo o en un caballo completo.', ARRAY['Transformación en caballo: El usuario puede adoptar la forma de un caballo completo, ganando la velocidad y la resistencia del animal y pudiendo cargar con jinetes.','Forma híbrida equina: Una forma intermedia que combina rasgos de caballo y de humano, conservando la movilidad del usuario y aumentando su fuerza física.','Mordisco: En forma de caballo o híbrida, el usuario puede morder a sus oponentes, un ataque sencillo pero eficaz a corta distancia.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-7a49-a17b-127f0c89d0b6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-7a49-a17b-127f0c89d0b6', 'en-GB', 'Allows the user to transform into a horse hybrid or a full horse at will.', ARRAY['Horse Transformation: The user can take the form of a full horse, gaining the speed and stamina of the animal while remaining able to carry riders.','Horse Hybrid Form: A half-way form that blends horse and human traits, keeping the user''s mobility while adding enhanced physical strength.','Biting Strike: In horse or hybrid form the user can bite opponents, a simple but effective close-range attack.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-7a49-a17b-127f0c89d0b6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-7a49-a17b-127f0c89d0b6', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de cavall o en un cavall complet.', ARRAY['Transformació en cavall: L''usuari pot adoptar la forma d''un cavall complet, guanyant la velocitat i la resistència de l''animal i podent carregar amb genets.','Forma híbrida equina: Una forma intermèdia que combina trets de cavall i d''humà, conservant la mobilitat de l''usuari i augmentant-ne la força física.','Mossegada: En forma de cavall o híbrida, l''usuari pot mossegar els oponents, un atac senzill però eficaç a curta distància.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-7a49-a17b-127f0c89d0b6')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Rako Rako no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdee-789b-a71b-a70749422cb9', 'DEVIL_FRUIT', 'Rako Rako no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdee-789b-a71b-a70749422cb9', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-789b-a71b-a70749422cb9')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-789b-a71b-a70749422cb9', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de nutria marina o en una nutria marina completa.', ARRAY['Transformación en nutria marina: El usuario puede convertirse en una nutria marina completa, una de las dos formas que otorga la fruta junto a la híbrida.','Forma híbrida de nutria: Un híbrido de nutria marina y humano con capacidades físicas mejoradas, capaz de empuñar armas y blandirlas con gran fuerza.','Rakko Hammer: En forma híbrida, el usuario descarga una pesada maza sobre su oponente con un potente golpe desde arriba.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-789b-a71b-a70749422cb9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-789b-a71b-a70749422cb9', 'en-GB', 'Allows the user to transform into a sea otter hybrid or a full sea otter at will.', ARRAY['Sea Otter Transformation: The user can become a full sea otter, one of the two forms this fruit grants alongside the hybrid.','Otter Hybrid Form: A sea otter-human hybrid with enhanced physical capabilities, able to wield weapons and swing them with great force.','Rakko Hammer: In hybrid form the user brings a heavy maul crashing down on an opponent in a powerful overhead strike.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-789b-a71b-a70749422cb9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-789b-a71b-a70749422cb9', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de llúdriga marina o en una llúdriga marina completa.', ARRAY['Transformació en llúdriga marina: L''usuari pot convertir-se en una llúdriga marina completa, una de les dues formes que atorga la fruita al costat de la híbrida.','Forma híbrida de llúdriga: Un híbrid de llúdriga marina i humà amb capacitats físiques millorades, capaç d''empunyar armes i brandar-les amb gran força.','Rakko Hammer: En forma híbrida, l''usuari deixa caure una pesada maça sobre l''oponent amb un potent cop des de dalt.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-789b-a71b-a70749422cb9')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Tori Tori no mi: Model Falcon
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdef-721d-9d97-30054adc6ffa', 'DEVIL_FRUIT', 'Tori Tori no mi: Model Falcon', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdef-721d-9d97-30054adc6ffa', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-721d-9d97-30054adc6ffa')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-721d-9d97-30054adc6ffa', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de halcón o en un halcón gigante, otorgando vuelo y los agudos sentidos de un ave rapaz.', ARRAY['Vuelo de halcón: En ambas formas el usuario puede volar, una habilidad poco común entre las Frutas del Diablo, y en su forma completa puede cargar a otras personas a su espalda.','Sentidos y fuerza de depredador: La transformación aumenta la velocidad, la fuerza y los sentidos, lo que lo hace más apto para el combate físico que la mayoría de usuarios Zoan.','Tobizume: En forma híbrida, el usuario se lanza en picado a una velocidad cegadora y desgarra al oponente con sus garras.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-721d-9d97-30054adc6ffa')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-721d-9d97-30054adc6ffa', 'en-GB', 'Allows the user to transform into a falcon hybrid or a giant falcon at will, granting flight and the keen senses of a bird of prey.', ARRAY['Falcon Flight: In either form the user can fly, a rare ability among Devil Fruits, and in full falcon form can carry other people on their back.','Predator Senses and Strength: The transformation boosts speed, strength and senses, making the user better suited to physical combat than most Zoan users.','Tobizume: In hybrid form the user swoops down at blinding speed and slashes the opponent with their talons.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-721d-9d97-30054adc6ffa')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-721d-9d97-30054adc6ffa', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de falcó o en un falcó gegant, atorgant vol i els aguts sentits d''un ocell rapinyaire.', ARRAY['Vol de falcó: En ambdues formes l''usuari pot volar, una habilitat poc comuna entre les Fruites del Diable, i en la forma completa pot carregar altres persones a l''esquena.','Sentits i força de depredador: La transformació augmenta la velocitat, la força i els sentits, cosa que el fa més apte per al combat físic que la majoria d''usuaris Zoan.','Tobizume: En forma híbrida, l''usuari es llança en picat a una velocitat enlluernadora i esquinça l''oponent amb les urpes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-721d-9d97-30054adc6ffa')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Tori Tori no mi: Model Albatross
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf0-78c3-b124-091e1f790f58', 'DEVIL_FRUIT', 'Tori Tori no mi: Model Albatross', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf0-78c3-b124-091e1f790f58', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-78c3-b124-091e1f790f58')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-78c3-b124-091e1f790f58', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de albatros o en un albatros completo.', ARRAY['Transformación en albatros: El usuario puede adoptar la forma de un albatros completo, una de las dos formas que otorga la fruta junto a la híbrida.','Forma híbrida de albatros: Un híbrido de ave y humano con capacidades físicas mejoradas que puede mantenerse activo de forma permanente, pues el usuario puede conservarlo en todo momento.','Alas sin vuelo: Pese al animal en que se basa, este poder no permite volar, por lo que su valor reside en la mejora física de la forma híbrida.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-78c3-b124-091e1f790f58')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-78c3-b124-091e1f790f58', 'en-GB', 'Allows the user to transform into an albatross hybrid or a full albatross at will.', ARRAY['Albatross Transformation: The user can take the form of a full albatross, one of the two forms this fruit grants alongside the hybrid.','Albatross Hybrid Form: A bird-human hybrid with enhanced physical capabilities that can be kept active indefinitely, as the user may hold it at all times.','Grounded Wings: Despite the animal it is based on, this power does not allow its user to fly, so its value lies in the hybrid form''s physical enhancement.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-78c3-b124-091e1f790f58')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-78c3-b124-091e1f790f58', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid d''albatros o en un albatros complet.', ARRAY['Transformació en albatros: L''usuari pot adoptar la forma d''un albatros complet, una de les dues formes que atorga la fruita al costat de la híbrida.','Forma híbrida d''albatros: Un híbrid d''ocell i humà amb capacitats físiques millorades que es pot mantenir activa de manera permanent, ja que l''usuari pot conservar-la en tot moment.','Ales sense vol: Malgrat l''animal en què es basa, aquest poder no permet volar, de manera que el seu valor rau en la millora física de la forma híbrida.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-78c3-b124-091e1f790f58')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Inu Inu no mi: Model Tanuki
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2', 'DEVIL_FRUIT', 'Inu Inu no mi: Model Tanuki', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de tanuki o en un tanuki completo, el cánido parecido a un mapache originario del País de Wano.', ARRAY['Transformación en tanuki: El usuario puede convertirse en un tanuki completo, un cánido parecido a un mapache, una de las dos formas que otorga la fruta junto a la híbrida.','Forma híbrida de tanuki: Un híbrido de tanuki y humano que da al usuario un cuerpo más móvil y erguido que la forma animal.','Compañero con vida: La fruta puede ser ingerida por un objeto inanimado, que cobra vida como mascota doméstica, aunque el poder en sí no ofrece capacidades de combate.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2', 'en-GB', 'Allows the user to transform into a tanuki hybrid or a full tanuki, the raccoon-like canine native to Wano Country.', ARRAY['Tanuki Transformation: The user can become a full tanuki, a raccoon-like canine, as one of the two forms this fruit grants alongside the hybrid.','Tanuki Hybrid Form: A tanuki-human hybrid that gives the user a more mobile, upright body than the animal form.','Living Companion: The fruit can be eaten by an inanimate object, which comes to life as a household pet, though the power itself offers no combat capabilities.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de tanuki o en un tanuki complet, el cànid semblant a un os rentador originari del País de Wano.', ARRAY['Transformació en tanuki: L''usuari pot convertir-se en un tanuki complet, un cànid semblant a un os rentador, una de les dues formes que atorga la fruita al costat de la híbrida.','Forma híbrida de tanuki: Un híbrid de tanuki i humà que dona a l''usuari un cos més mòbil i dret que la forma animal.','Company amb vida: La fruita pot ser ingerida per un objecte inanimat, que cobra vida com a mascota domèstica, tot i que el poder en si no ofereix capacitats de combat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf1-7cb5-94d2-1a72d72c83b2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Inu Inu no mi: Model Hound
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf2-7ca2-a6aa-befb7392c582', 'DEVIL_FRUIT', 'Inu Inu no mi: Model Hound', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf2-7ca2-a6aa-befb7392c582', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf2-7ca2-a6aa-befb7392c582')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf2-7ca2-a6aa-befb7392c582', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de perro de caza o en un perro de caza completo, ganando la fuerza y los sentidos de un sabueso.', ARRAY['Transformación en sabueso: El usuario puede convertirse en un perro de caza completo, un animal ideal para la velocidad y la persecución.','Forma híbrida de sabueso: Un híbrido de sabueso y humano que combina rasgos caninos con un cuerpo humanoide para un combate cuerpo a cuerpo versátil.','Instintos de depredador: La transformación aumenta la fuerza física, la velocidad y los sentidos, haciendo al usuario más apto para el combate físico que los usuarios Zoan herbívoros.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf2-7ca2-a6aa-befb7392c582')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf2-7ca2-a6aa-befb7392c582', 'en-GB', 'Allows the user to transform into a hound hybrid or a full hound at will, gaining the strength and senses of a hunting dog.', ARRAY['Hound Transformation: The user can become a full hound, a hunting dog built for speed and pursuit.','Hound Hybrid Form: A hound-human hybrid that blends canine traits with a humanoid body for versatile close combat.','Predator Instincts: The transformation increases physical strength, speed and senses, making the user better suited to physical combat than herbivorous Zoan users.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf2-7ca2-a6aa-befb7392c582')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf2-7ca2-a6aa-befb7392c582', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de gos de caça o en un gos de caça complet, guanyant la força i els sentits d''un gos coniller.', ARRAY['Transformació en gos de caça: L''usuari pot convertir-se en un gos de caça complet, un animal ideal per a la velocitat i la persecució.','Forma híbrida de gos de caça: Un híbrid de gos de caça i humà que combina trets canins amb un cos humanoide per a un combat cos a cos versàtil.','Instints de depredador: La transformació augmenta la força física, la velocitat i els sentits, fent l''usuari més apte per al combat físic que els usuaris Zoan herbívors.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf2-7ca2-a6aa-befb7392c582')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Batto Batto no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf3-7df4-bf27-cf96ba7addf9', 'DEVIL_FRUIT', 'Batto Batto no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9', 'ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de murciélago o en un murciélago completo.', ARRAY['Transformación en murciélago: El usuario puede convertirse en un murciélago completo, una de las dos formas que otorga la fruta junto a la híbrida.','Alas de murciélago: En forma híbrida, al usuario le crecen alas de murciélago en la espalda, lo que favorece el movimiento sigiloso en combate.','Succión de sangre: Al morder la piel de un oponente y succionar su sangre, el usuario lo incapacita rápidamente y lo duerme, incluso frente a luchadores poderosos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9', 'en-GB', 'Allows the user to transform into a bat hybrid or a full bat at will.', ARRAY['Bat Transformation: The user can become a full bat, one of the two forms this fruit grants alongside the hybrid.','Bat Wings: In hybrid form the user grows bat-like wings from their back, supporting stealthy movement in combat.','Blood Drain: By biting into an opponent''s skin and sucking their blood, the user quickly incapacitates them and puts them to sleep, even against powerful fighters.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de ratpenat o en un ratpenat complet.', ARRAY['Transformació en ratpenat: L''usuari pot convertir-se en un ratpenat complet, una de les dues formes que atorga la fruita al costat de la híbrida.','Ales de ratpenat: En forma híbrida, a l''usuari li creixen ales de ratpenat a l''esquena, cosa que afavoreix el moviment sigil·lós en combat.','Succió de sang: En mossegar la pell d''un oponent i xuclar-ne la sang, l''usuari l''incapacita ràpidament i l''adorm, fins i tot davant de lluitadors poderosos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf3-7df4-bf27-cf96ba7addf9')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Uma Uma no mi: Model Pegasus
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf4-7c3d-bc55-8bebabf6c601', 'DEVIL_FRUIT', 'Uma Uma no mi: Model Pegasus', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601', 'MYTHICAL_ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601', 'es-ES', 'Permite al usuario transformarse a voluntad en un híbrido de Pegaso o en un Pegaso completo, otorgando el poder de volar.', ARRAY['Transformación en Pegaso: El usuario puede adoptar la forma de un caballo alado, el mítico Pegaso, en su versión completa o híbrida.','Vuelo alado: Las alas otorgan el poder de volar, permitiendo al usuario desplazarse por el aire a voluntad.','Montura aérea: En el aire, el usuario puede cargar pasajeros a su espalda, actuando como corcel volador para sus aliados.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601', 'en-GB', 'Allows the user to transform into a Pegasus hybrid or a full Pegasus at will, gaining the power of flight.', ARRAY['Pegasus Transformation: The user can take the form of a winged horse, the mythical Pegasus, in full or hybrid form.','Winged Flight: The wings grant the power of flight, allowing the user to travel through the air at will.','Aerial Mount: In the air the user can carry passengers on their back, working as a flying steed for allies.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601', 'ca-ES', 'Permet a l''usuari transformar-se a voluntat en un híbrid de Pegàs o en un Pegàs complet, atorgant el poder de volar.', ARRAY['Transformació en Pegàs: L''usuari pot adoptar la forma d''un cavall alat, el mític Pegàs, en la versió completa o híbrida.','Vol alat: Les ales atorguen el poder de volar, permetent a l''usuari desplaçar-se per l''aire a voluntat.','Muntura aèria: A l''aire, l''usuari pot carregar passatgers a l''esquena, actuant com a corser volador per als seus aliats.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf4-7c3d-bc55-8bebabf6c601')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Hito Hito no mi: Model Onyudo
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf5-71df-95ae-55a7eeba81a4', 'DEVIL_FRUIT', 'Hito Hito no mi: Model Onyudo', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf5-71df-95ae-55a7eeba81a4', 'MYTHICAL_ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf5-71df-95ae-55a7eeba81a4')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf5-71df-95ae-55a7eeba81a4', 'es-ES', 'Permite al usuario transformarse en un onyudo, un yokai con forma de monje gigante del folclore japonés, obteniendo inteligencia humana y habla.', ARRAY['Transformación en onyudo: El usuario se convierte en un monje colosal de aspecto humano de más de cuatro metros, con una fuerza y una velocidad enormes.','Intelecto humano: Al ser una fruta Hito Hito, la transformación concede a un usuario no humano inteligencia humana y la capacidad de hablar.','Armas blancas: La forma incluye una naginata y una espada al cinto, que el usuario empuña en combate y que desaparecen al recuperar su forma original.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf5-71df-95ae-55a7eeba81a4')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf5-71df-95ae-55a7eeba81a4', 'en-GB', 'Allows the user to transform into an onyudo, a giant monk yokai of Japanese legend, gaining human intelligence and speech.', ARRAY['Onyudo Transformation: The user grows into a towering, human-like monk of over four metres, with enormous strength and speed to match.','Human Intellect: As a Hito Hito fruit, the transformation grants a non-human user human intelligence and the ability to speak.','Bladed Weaponry: The form comes with a naginata and a sword at the hip, which the user wields in combat and which vanish when the user changes back.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf5-71df-95ae-55a7eeba81a4')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf5-71df-95ae-55a7eeba81a4', 'ca-ES', 'Permet a l''usuari transformar-se en un onyudo, un yokai amb forma de monjo gegant del folklore japonès, obtenint intel·ligència humana i parla.', ARRAY['Transformació en onyudo: L''usuari es converteix en un monjo colossal d''aspecte humà de més de quatre metres, amb una força i una velocitat enormes.','Intel·lecte humà: En ser una fruita Hito Hito, la transformació concedeix a un usuari no humà intel·ligència humana i la capacitat de parlar.','Armes blanques: La forma inclou una naginata i una espasa al cinturó, que l''usuari empunya en combat i que desapareixen en recuperar la forma original.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf5-71df-95ae-55a7eeba81a4')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Ryu Ryu no mi: Model Kirin
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf6-7f51-802a-2a5c31dd8a49', 'DEVIL_FRUIT', 'Ryu Ryu no mi: Model Kirin', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49', 'MYTHICAL_ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49', 'es-ES', 'Permite al usuario transformarse en un qilin o en un híbrido de qilin, una bestia mítica capaz de dormir a los demás y hacer realidad sus sueños.', ARRAY['Rayos del sueño: El usuario dispara rayos de energía anillados que duermen a quien alcanzan, incluso a gran distancia, y puede elegir entre un sueño profundo y una parálisis en la que la víctima sigue consciente.','Materialización de sueños: Cuando un objetivo duerme, el usuario puede extraer objetos y seres de la nube de sueño sobre su cabeza y darles forma física.','Monstruos de pesadilla: El usuario puede conjurar criaturas nacidas de miedos soñados que atacan por sí solas, y desatar muchas a la vez mediante los Nightmare Holes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49', 'en-GB', 'Allows the user to transform into a qilin or a half-qilin hybrid, a mythical beast able to put others to sleep and make their dreams real.', ARRAY['Sleep Beams: The user fires ringed energy beams that put anyone they hit to sleep, even from a long distance, and can choose between deep sleep and a conscious state of paralysis.','Dream Manifestation: Once a target sleeps, the user can pull objects and beings out of the dream cloud above their head and give them physical form.','Nightmare Monsters: The user can conjure creatures born from dreamed fears that attack on their own, and unleash many at once through Nightmare Holes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49', 'ca-ES', 'Permet a l''usuari transformar-se en un qilin o en un híbrid de qilin, una bèstia mítica capaç d''adormir els altres i fer realitat els seus somnis.', ARRAY['Raigs del son: L''usuari dispara raigs d''energia anellats que adormen qui toquen, fins i tot a gran distància, i pot triar entre un son profund i una paràlisi en què la víctima segueix conscient.','Materialització de somnis: Quan un objectiu dorm, l''usuari pot extreure objectes i éssers del núvol de somni sobre el seu cap i donar-los forma física.','Monstres de malson: L''usuari pot conjurar criatures nascudes de pors somiades que ataquen per si soles, i deixar-ne anar moltes alhora mitjançant els Nightmare Holes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf6-7f51-802a-2a5c31dd8a49')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Risu Risu no mi: Model Ratatoskr
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf7-7519-b3cd-e20c1a421dcd', 'DEVIL_FRUIT', 'Risu Risu no mi: Model Ratatoskr', 'LEGENDARY')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd', 'MYTHICAL_ZOAN' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd', 'es-ES', 'Otorga el poder de transformarse en Ratatoskr, una legendaria ardilla de hielo, y de producir hielo y nieblas gélidas.', ARRAY['Congelación instantánea: El usuario genera hielo y nieblas gélidas que congelan al instante a sus objetivos, que no pueden descongelarse hasta que el usuario lo decide.','Rayo de hielo: Al hacer chocar cristales de hielo entre sí se genera y amplifica electricidad, por lo que el poder combina bien con otras habilidades basadas en el rayo.','Niflheim: Un golpe de martillo que impacta de lleno en un enemigo y lo encierra al instante en una gruesa capa de hielo, utilizable incluso sin transformarse.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd', 'en-GB', 'Grants the power to transform into Ratatoskr, a legendary ice squirrel, and to produce ice and freezing mists.', ARRAY['Flash Freeze: The user generates ice and freezing mists that instantly freeze targets, who cannot thaw until the user wills it.','Ice Lightning: Making ice crystals collide generates and amplifies lightning, so the power pairs well with other lightning-based abilities.','Niflheim: A hammer blow that lands squarely on an enemy and instantly encases them in a thick layer of ice, usable even when the user is not transformed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd', 'ca-ES', 'Atorga el poder de transformar-se en Ratatoskr, un llegendari esquirol de gel, i de produir gel i boires gèlides.', ARRAY['Congelació instantània: L''usuari genera gel i boires gèlides que congelen a l''instant els objectius, que no es poden descongelar fins que l''usuari ho decideix.','Llamp de gel: En fer xocar cristalls de gel entre ells es genera i s''amplifica electricitat, de manera que el poder combina bé amb altres habilitats basades en el llamp.','Niflheim: Un cop de martell que impacta de ple en un enemic i l''envolta a l''instant d''una gruixuda capa de gel, utilitzable fins i tot sense transformar-se.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf7-7519-b3cd-e20c1a421dcd')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- The game-only 'Model: Vampire' (Unlimited World Red) is not manga canon: the canon
-- Batto Batto no mi (Stussy) is seeded above. Deleting cascades to its
-- devil_fruits/power_translations rows.
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Batto Batto no mi: Model Vampire';

-- +goose Down
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Uma Uma no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Rako Rako no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Tori Tori no mi: Model Falcon';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Tori Tori no mi: Model Albatross';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Inu Inu no mi: Model Tanuki';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Inu Inu no mi: Model Hound';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Batto Batto no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Uma Uma no mi: Model Pegasus';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Hito Hito no mi: Model Onyudo';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Ryu Ryu no mi: Model Kirin';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Risu Risu no mi: Model Ratatoskr';
