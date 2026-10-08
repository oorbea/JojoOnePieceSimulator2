-- +goose Up
-- Seeds the 67 manga-canon Paramecia Devil Fruits missing from the catalogue.
-- Hand-authored new content (wiki-sourced), not a catalogsync dump; see
-- ObsidianVault/catalog-seed-p7-p8-fruits-2026-10-08.md. Runs in every
-- environment including prod (no ENVSUB guard); images are added later by an
-- admin, so rows keep the powers.picture* defaults. Idempotent against a fruit an
-- admin already created: powers inserts are ON CONFLICT DO NOTHING (id or name),
-- dependent inserts are gated on that powers row existing.

-- Bane Bane no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdae-7349-847a-ed112e2a7606', 'DEVIL_FRUIT', 'Bane Bane no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdae-7349-847a-ed112e2a7606', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdae-7349-847a-ed112e2a7606')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdae-7349-847a-ed112e2a7606', 'es-ES', 'Permite al usuario convertir sus extremidades en muelles, lanzándose a sí mismo y sus golpes con un rebote explosivo.', ARRAY['Piernas de muelle: El usuario convierte sus piernas en muelles para dispararse a una velocidad enorme, a menudo demasiado rápido para que el rival lo siga (como el Spring Hopper de Bellamy).','Impacto de rebote: La fuerza del rebote contra paredes y suelos puede destrozar la superficie golpeada e incluso partir un barco en dos.','Brazos de muelle: Aplicar el mismo poder a los brazos genera puñetazos potentes y de largo alcance, aunque el usuario solo puede desplazarse en línea recta y de forma predecible.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdae-7349-847a-ed112e2a7606')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdae-7349-847a-ed112e2a7606', 'en-GB', 'Allows the user to turn their limbs into springs, launching themselves and their blows with explosive bounce.', ARRAY['Spring Legs: The user turns their legs into springs to rocket themselves at enormous speed, often too fast for opponents to follow (as with Bellamy''s Spring Hopper).','Bouncing Impact: The rebound force from bouncing off walls and floors can shatter the surface struck, and can even split a ship in half.','Spring Arms: Applying the same power to the arms produces long-reaching, powerful punches, although the user can only travel in a straight, predictable line.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdae-7349-847a-ed112e2a7606')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdae-7349-847a-ed112e2a7606', 'ca-ES', 'Permet a l''usuari convertir les seves extremitats en molles, llançant-se a si mateix i els seus cops amb un rebot explosiu.', ARRAY['Cames de molla: L''usuari converteix les cames en molles per disparar-se a una velocitat enorme, sovint massa de pressa perquè el rival el segueixi (com el Spring Hopper de Bellamy).','Impacte de rebot: La força del rebot contra parets i terra pot destrossar la superfície colpejada i fins i tot partir un vaixell en dos.','Braços de molla: Aplicar el mateix poder als braços genera cops de puny potents i de llarg abast, tot i que l''usuari només es pot desplaçar en línia recta i de manera previsible.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdae-7349-847a-ed112e2a7606')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Noro Noro no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdaf-70e9-9adb-479b3b349dd7', 'DEVIL_FRUIT', 'Noro Noro no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdaf-70e9-9adb-479b3b349dd7', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdaf-70e9-9adb-479b3b349dd7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdaf-70e9-9adb-479b3b349dd7', 'es-ES', 'Permite al usuario disparar rayos de luz que ralentizan durante treinta segundos todo lo que alcanzan.', ARRAY['Fotones lentos: El usuario emite rayos morados de partículas especiales que ralentizan a personas, objetos o incluso el propio espacio durante treinta segundos (como el Noro Noro Beam de Foxy).','Impacto almacenado: Los golpes recibidos por un objetivo ralentizado liberan toda su fuerza cinética de golpe cuando recupera su velocidad normal, haciéndolos mucho más dañinos.','Plataformas y trampas ralentizadas: Las balas de cañón ralentizadas sirven de plataformas flotantes o trampas con retardo, pero los rayos se reflejan en espejos y también ralentizan al usuario.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdaf-70e9-9adb-479b3b349dd7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdaf-70e9-9adb-479b3b349dd7', 'en-GB', 'Allows the user to fire beams of light that slow down whatever they hit for thirty seconds.', ARRAY['Slow Photons: The user emits purple beams of special particles that slow down people, objects or even space itself for thirty seconds (as with Foxy''s Noro Noro Beam).','Stored Impact: Strikes landed on a slowed target release all their kinetic force at once when it returns to normal speed, making the blows far more damaging.','Slowed Platforms and Traps: Slowed cannonballs can serve as floating platforms or timed traps, though the beams can be reflected by mirrors and also slow the user.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdaf-70e9-9adb-479b3b349dd7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdaf-70e9-9adb-479b3b349dd7', 'ca-ES', 'Permet a l''usuari disparar rajos de llum que alenteixen durant trenta segons tot el que toquen.', ARRAY['Fotons lents: L''usuari emet rajos morats de partícules especials que alenteixen persones, objectes o fins i tot l''espai mateix durant trenta segons (com el Noro Noro Beam de Foxy).','Impacte emmagatzemat: Els cops rebuts per un objectiu alentit alliberen tota la seva força cinètica de cop quan recupera la velocitat normal, fent-los molt més dolorosos.','Plataformes i paranys alentits: Les bales de canó alentides serveixen de plataformes flotants o paranys amb retard, però els rajos es reflecteixen en miralls i també alenteixen l''usuari.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdaf-70e9-9adb-479b3b349dd7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Doa Doa no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939', 'DEVIL_FRUIT', 'Doa Doa no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939', 'es-ES', 'Permite al usuario crear puertas en casi cualquier cosa que toque, incluido el propio aire.', ARRAY['Puertas a través de sólidos: El usuario abre puertas en paredes, suelos o cualquier superficie sólida para atravesarla libremente, por gruesa o resistente que sea (como hace Blueno contra Luffy).','Puertas de inmovilización: Se pueden crear puertas bajo los pies del rival o sobre su propio cuerpo, atrapándole los pies o inmovilizándolo.','Puertas de aire: Las puertas creadas en el aire llevan a una especie de dimensión de bolsillo dentro del propio aire, ideal para esconderse, huir o moverse sin ser visto.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939', 'en-GB', 'Allows the user to create doors on almost anything they touch, including the air itself.', ARRAY['Doors Through Solids: The user opens doors in walls, floors or any solid surface to pass through it freely, however thick or strong it is (as Blueno does against Luffy).','Restraining Doors: Doors can be created on the ground beneath an opponent or on their own body, trapping their feet or immobilising them.','Air Doors: Doors made in the air lead to a pocket dimension inside the air itself, ideal for hiding, escaping or stealthy movement.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939', 'ca-ES', 'Permet a l''usuari crear portes en gairebé qualsevol cosa que toqui, inclòs l''aire mateix.', ARRAY['Portes a través de sòlids: L''usuari obre portes en parets, terres o qualsevol superfície sòlida per travessar-la lliurement, per gruixuda o resistent que sigui (com fa Blueno contra Luffy).','Portes d''immobilització: Es poden crear portes sota els peus del rival o sobre el seu propi cos, atrapant-li els peus o immobilitzant-lo.','Portes d''aire: Les portes creades a l''aire porten a una mena de dimensió de butxaca dins de l''aire mateix, ideal per amagar-se, fugir o moure''s sense ser vist.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb0-75f0-8a1d-9bdb7e7a7939')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Beri Beri no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5', 'DEVIL_FRUIT', 'Beri Beri no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5', 'es-ES', 'Permite al usuario dividir su cuerpo en multitud de bolas con forma de baya que sirven a la vez de armadura y de proyectiles.', ARRAY['Cuerpo de bayas: El usuario se divide en bolas redondas similares a balas de cañón y se vuelve inmune a los golpes contundentes, como el Strong Right de Franky (como hace Very Good).','Ataque desde varios ángulos: Las bolas separadas cargan contra el rival desde varias direcciones a la vez y, en el anime, también pueden flotar en el aire.','Defensa frágil: La ventaja defensiva cuesta fuerza y velocidad, y si capturan una pieza, como la cabeza, el usuario queda indefenso.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5', 'en-GB', 'Allows the user to split their body into a multitude of berry-like balls that act as both armour and projectiles.', ARRAY['Berry Body: The user divides into round, cannonball-like balls, becoming immune to blunt attacks such as Franky''s Strong Right (as Very Good does).','Multi-Angle Assault: The separated balls charge at the opponent from several directions at once, and in the anime they can also float in mid-air.','Fragile Defence: The defensive gain costs strength and speed, and a captured piece, such as the head, leaves the user helpless.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5', 'ca-ES', 'Permet a l''usuari dividir el seu cos en una multitud de boles en forma de baia que serveixen alhora d''armadura i de projectils.', ARRAY['Cos de baies: L''usuari es divideix en boles rodones semblants a bales de canó i esdevé immune als cops contundents, com el Strong Right de Franky (com fa Very Good).','Atac des de diversos angles: Les boles separades carreguen contra el rival des de diverses direccions alhora i, a l''anime, també poden flotar a l''aire.','Defensa fràgil: L''avantatge defensiu costa força i velocitat, i si capturen una peça, com el cap, l''usuari queda indefens.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb1-7eca-8bfb-170d2f7bd7f5')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Sabi Sabi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb2-7134-8d5e-43fa4218a86b', 'DEVIL_FRUIT', 'Sabi Sabi no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb2-7134-8d5e-43fa4218a86b', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb2-7134-8d5e-43fa4218a86b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb2-7134-8d5e-43fa4218a86b', 'es-ES', 'Permite al usuario oxidar y desmenuzar el metal con solo tocarlo.', ARRAY['Toque oxidante: Al tocar un objeto metálico, el usuario lo oxida y lo hace pedazos casi al instante (como hace Shu con la espada de Zoro).','Defensa antimetal: Las espadas y otras armas metálicas que tocan el cuerpo del usuario se corroen antes de poder cortar, lo que da una clara ventaja frente a rivales con armas de filo.','Carne oxidada: En el anime, el usuario también puede oxidar el cuerpo de un objetivo vivo, aunque el efecto se desvanece si este se libera antes de quedar totalmente corroído.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb2-7134-8d5e-43fa4218a86b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb2-7134-8d5e-43fa4218a86b', 'en-GB', 'Allows the user to make metal rust and crumble simply by touching it.', ARRAY['Rusting Touch: By touching a metal object, the user causes it to rust and break apart almost instantly (as Shu does to Zoro''s sword).','Anti-Metal Defence: Swords and other metal weapons that touch the user''s body corrode before they can cut, giving a clear edge against bladed opponents.','Rusting Flesh: In the anime, the user can also rust a living target''s body, though the effect fades if the target breaks free before being fully corroded.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb2-7134-8d5e-43fa4218a86b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb2-7134-8d5e-43fa4218a86b', 'ca-ES', 'Permet a l''usuari oxidar i esmicolar el metall només tocant-lo.', ARRAY['Toc oxidant: En tocar un objecte metàl·lic, l''usuari l''oxida i el fa miques gairebé a l''instant (com fa Shu amb l''espasa de Zoro).','Defensa antimetall: Les espases i altres armes metàl·liques que toquen el cos de l''usuari es corroeixen abans de poder tallar, cosa que dona un clar avantatge davant de rivals amb armes de tall.','Carn oxidada: A l''anime, l''usuari també pot oxidar el cos d''un objectiu viu, tot i que l''efecte s''esvaeix si aquest s''allibera abans de quedar totalment corroït.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb2-7134-8d5e-43fa4218a86b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Shari Shari no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb3-7649-a3c6-400b221efb32', 'DEVIL_FRUIT', 'Shari Shari no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb3-7649-a3c6-400b221efb32', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb3-7649-a3c6-400b221efb32')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb3-7649-a3c6-400b221efb32', 'es-ES', 'Permite al usuario hacer girar partes de su cuerpo como si fueran ruedas.', ARRAY['Brazos giratorios: El usuario rota los brazos a gran velocidad y los convierte en armas giratorias formidables (como hace Sharinguru contra Franky).','Piernas giratorias: La misma rotación puede aplicarse a los pies, añadiendo otra forma de golpear al rival en combate.','Poder limitado: Los cuerpos metálicos resistentes, como los brazos de Franky, soportan los ataques giratorios, lo que demuestra que su fuerza tiene límites.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb3-7649-a3c6-400b221efb32')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb3-7649-a3c6-400b221efb32', 'en-GB', 'Allows the user to spin their body parts like wheels.', ARRAY['Spinning Arms: The user rotates their arms at high speed, turning them into formidable spinning weapons (as Sharinguru does against Franky).','Spinning Legs: The same rotation can be applied to the feet, adding another way to strike an opponent in combat.','Limited Power: Hard metallic bodies, such as Franky''s arms, resist the spinning attacks, showing that their force has limits.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb3-7649-a3c6-400b221efb32')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb3-7649-a3c6-400b221efb32', 'ca-ES', 'Permet a l''usuari fer girar parts del seu cos com si fossin rodes.', ARRAY['Braços giratoris: L''usuari rota els braços a gran velocitat i els converteix en armes giratòries formidables (com fa Sharinguru contra Franky).','Cames giratòries: La mateixa rotació es pot aplicar als peus, afegint una altra manera de colpejar el rival en combat.','Poder limitat: Els cossos metàl·lics resistents, com els braços de Franky, aguanten els atacs giratoris, cosa que demostra que la seva força té límits.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb3-7649-a3c6-400b221efb32')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Suke Suke no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb4-7910-ab95-9387f95694d5', 'DEVIL_FRUIT', 'Suke Suke no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb4-7910-ab95-9387f95694d5', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb4-7910-ab95-9387f95694d5')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb4-7910-ab95-9387f95694d5', 'es-ES', 'Permite al usuario volverse invisible, junto con todo lo que toque.', ARRAY['Invisibilidad: El usuario se vuelve totalmente invisible, lo que le permite espiar conversaciones y moverse sin ser detectado.','Objetos invisibles: Todo lo que el usuario toca también desaparece de la vista, como armas ocultas o incluso un barco pequeño usado para una huida discreta, hasta que deja de estar en contacto con él.','Puntos débiles reveladores: Quedar cubierto de agua, sal o sangre, o recibir un golpe potente, revela brevemente al usuario, y los rivales aún pueden rastrearlo por el oído o el olfato.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb4-7910-ab95-9387f95694d5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb4-7910-ab95-9387f95694d5', 'en-GB', 'Allows the user to turn themselves, and anything they touch, invisible.', ARRAY['Invisibility: The user becomes completely invisible, allowing them to spy on conversations and move around without being detected.','Invisible Objects: Whatever the user touches also vanishes from sight, such as hidden weapons or even a small ship used for a concealed getaway, until it leaves their contact.','Revealing Weak Points: Being covered in water, salt or blood, or taking a powerful blow, briefly reveals the user, and opponents can still track them by sound or smell.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb4-7910-ab95-9387f95694d5')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb4-7910-ab95-9387f95694d5', 'ca-ES', 'Permet a l''usuari tornar-se invisible, juntament amb tot el que toqui.', ARRAY['Invisibilitat: L''usuari esdevé totalment invisible, cosa que li permet espiar converses i moure''s sense ser detectat.','Objectes invisibles: Tot el que l''usuari toca també desapareix de la vista, com armes amagades o fins i tot un vaixell petit usat per a una fugida discreta, fins que deixa d''estar en contacte amb ell.','Punts febles reveladors: Quedar cobert d''aigua, sal o sang, o rebre un cop potent, revela breument l''usuari, i els rivals encara el poden seguir pel so o l''olfacte.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb4-7910-ab95-9387f95694d5')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Toshi Toshi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb5-74bb-ad28-b5799db99455', 'DEVIL_FRUIT', 'Toshi Toshi no mi', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb5-74bb-ad28-b5799db99455', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb5-74bb-ad28-b5799db99455')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb5-74bb-ad28-b5799db99455', 'es-ES', 'Permite al usuario manipular la edad de personas y objetos con solo tocarlos.', ARRAY['Regresión de edad: Tocar a un objetivo lo devuelve a la forma de un niño pequeño y débil, dejando indefensos incluso a combatientes fuertes (como hace Bonney con un vicealmirante de la Marina).','Envejecimiento rápido: El usuario puede envejecer a un objetivo hasta la vejez, o brevemente hasta un esqueleto para provocarle un susto de muerte, mientras los objetos se corroen y se convierten en polvo.','Disfraz personal: El usuario también puede cambiar su propia edad para disfrazarse, y los años arrebatados a otros aparecen como joyas brillantes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb5-74bb-ad28-b5799db99455')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb5-74bb-ad28-b5799db99455', 'en-GB', 'Allows the user to manipulate the age of people and objects by touch.', ARRAY['Age Regression: Touching a target turns them back into a small, weak child, leaving even strong fighters helpless (as Bonney does to a Marine Vice Admiral).','Rapid Ageing: The user can age a target into old age, or briefly into a skeleton for a terrifying near-death scare, while objects corrode and crumble to dust.','Self Disguise: The user can also change their own age to disguise themselves, and the years taken from others appear as sparkling jewellery.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb5-74bb-ad28-b5799db99455')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb5-74bb-ad28-b5799db99455', 'ca-ES', 'Permet a l''usuari manipular l''edat de persones i objectes només tocant-los.', ARRAY['Regressió d''edat: Tocar un objectiu el torna a la forma d''un nen petit i feble, deixant indefensos fins i tot combatents forts (com fa Bonney amb un vicealmirall de la Marina).','Envelliment ràpid: L''usuari pot envellir un objectiu fins a la vellesa, o breument fins a un esquelet per provocar-li un ensurt de mort, mentre els objectes es corroeixen i es converteixen en pols.','Disfressa personal: L''usuari també pot canviar la seva pròpia edat per disfressar-se, i els anys presos als altres apareixen com a joies brillants.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb5-74bb-ad28-b5799db99455')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Shiro Shiro no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb6-70ac-a683-be591e249201', 'DEVIL_FRUIT', 'Shiro Shiro no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb6-70ac-a683-be591e249201', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb6-70ac-a683-be591e249201')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb6-70ac-a683-be591e249201', 'es-ES', 'Permite al usuario convertirse en una fortaleza viviente, con defensas de castillo y salas en el interior de su propio cuerpo.', ARRAY['Fortaleza viviente: El cuerpo del usuario alberga troneras de cañón, rastrillos y puentes levadizos, con salas interiores que van desde estancias de combate de piedra hasta salones de reuniones amueblados (como el castillo de Capone Bege).','Aura reductora: Las personas y objetos que entran en el aura que rodea al usuario son atraídos al interior y encogidos, y recuperan su tamaño al salir del castillo.','Defensas a tamaño completo: A tamaño completo, la fortaleza puede desplegar sus defensas en círculo para cubrir 360 grados de ataque o lanzar emboscadas sorpresa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb6-70ac-a683-be591e249201')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb6-70ac-a683-be591e249201', 'en-GB', 'Allows the user to become a living fortress, with castle defences and rooms inside their own body.', ARRAY['Living Fortress: The user''s body holds cannon ports, portcullises and drawbridges, with rooms inside ranging from stone combat halls to furnished meeting rooms (as Capone Bege''s castle).','Shrinking Aura: People and objects that enter the aura around the user are pulled inside and shrunk, returning to normal size when they leave the castle.','Full-Size Defences: At full size the fortress can deploy its defences in a circle for 360 degrees of attack coverage, or launch surprise ambushes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb6-70ac-a683-be591e249201')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb6-70ac-a683-be591e249201', 'ca-ES', 'Permet a l''usuari convertir-se en una fortalesa vivent, amb defenses de castell i sales a l''interior del seu propi cos.', ARRAY['Fortalesa vivent: El cos de l''usuari allotja canoneres, rastells i ponts llevadissos, amb sales interiors que van des d''estances de combat de pedra fins a salons de reunions moblats (com el castell de Capone Bege).','Aura reductora: Les persones i objectes que entren a l''aura que envolta l''usuari són atrets a l''interior i encongits, i recuperen la mida en sortir del castell.','Defenses a mida completa: A mida completa, la fortalesa pot desplegar les defenses en cercle per cobrir 360 graus d''atac o llançar emboscades sorpresa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb6-70ac-a683-be591e249201')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Wara Wara no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb7-7285-8a60-21a765413082', 'DEVIL_FRUIT', 'Wara Wara no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb7-7285-8a60-21a765413082', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb7-7285-8a60-21a765413082')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb7-7285-8a60-21a765413082', 'es-ES', 'Permite al usuario crear y manipular paja, incluidos muñecos de paja que redirigen el daño hacia otras personas.', ARRAY['Manipulación de paja: El usuario crea paja para atar enemigos, formar armas afiladas, cubrir su propio cuerpo en combate o construir un enorme monstruo de paja (como hace Hawkins).','Transferencia de daño: Cada muñeco de paja guardado en el cuerpo del usuario está vinculado a una persona, y cualquier herida del usuario se redirige a la persona que el muñeco representa.','Estrategia de rehenes: Al vincular muñecos a los aliados del rival, el usuario puede hacer imposible que el enemigo le haga daño sin herir antes a sus propios compañeros.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb7-7285-8a60-21a765413082')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb7-7285-8a60-21a765413082', 'en-GB', 'Allows the user to create and manipulate straw, including straw dolls that redirect damage onto other people.', ARRAY['Straw Manipulation: The user creates straw to bind enemies, form sharp weapons, cover their own body for combat, or build a huge straw monster (as Hawkins does).','Damage Transfer: Straw dolls stored inside the user''s body are each linked to a person, and any injury to the user is redirected to whoever the doll represents.','Hostage Strategy: By linking dolls to an opponent''s allies, the user can make it impossible for the enemy to hurt them without first harming their own comrades.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb7-7285-8a60-21a765413082')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb7-7285-8a60-21a765413082', 'ca-ES', 'Permet a l''usuari crear i manipular palla, inclosos ninots de palla que redirigeixen el dany cap a altres persones.', ARRAY['Manipulació de palla: L''usuari crea palla per lligar enemics, formar armes afilades, cobrir el seu propi cos en combat o construir un enorme monstre de palla (com fa Hawkins).','Transferència de dany: Cada ninot de palla guardat al cos de l''usuari està vinculat a una persona, i qualsevol ferida de l''usuari es redirigeix cap a la persona que el ninot representa.','Estratègia d''ostatges: En vincular ninots als aliats del rival, l''usuari pot fer impossible que l''enemic li faci mal sense ferir abans els seus propis companys.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb7-7285-8a60-21a765413082')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Oto Oto no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb8-7fd4-ace0-6c643345bb0d', 'DEVIL_FRUIT', 'Oto Oto no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d', 'es-ES', 'Permite al usuario convertir partes de su cuerpo en instrumentos musicales y tocar música con efectos destructivos.', ARRAY['Cuerpo instrumento: El usuario convierte partes de su cuerpo en instrumentos, como dientes en teclas de piano o un brazo en trompeta, e interpreta melodías concretas (como hace Apoo).','Ataques musicales: Cada melodía provoca un efecto distinto, desde puñetazos y cortes que desmiembran hasta explosiones elementales, transmitidos por ondas sonoras difíciles de prever.','Límites del sonido: Algunos efectos solo funcionan con objetivos que oyen la música, y el usuario no puede atacar lo que no ve, por lo que taparse los oídos puede mitigarlos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d', 'en-GB', 'Allows the user to turn parts of their body into musical instruments and play music with destructive effects.', ARRAY['Instrument Body: The user turns body parts into instruments, such as teeth into piano keys or an arm into a trumpet, and plays specific tunes (as Apoo does).','Musical Attacks: Each tune causes a different effect, from punches and dismembering slashes to elemental explosions, delivered by sound waves that are hard to see coming.','Sound Limits: Some effects only work on targets that can hear the music, and the user cannot attack what they cannot see, so covering the ears can blunt them.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d', 'ca-ES', 'Permet a l''usuari convertir parts del seu cos en instruments musicals i tocar música amb efectes destructius.', ARRAY['Cos instrument: L''usuari converteix parts del seu cos en instruments, com dents en tecles de piano o un braç en trompeta, i interpreta melodies concretes (com fa Apoo).','Atacs musicals: Cada melodia provoca un efecte diferent, des de cops de puny i talls que desmembren fins a explosions elementals, transmesos per ones sonores difícils de preveure.','Límits del so: Alguns efectes només funcionen amb objectius que senten la música, i l''usuari no pot atacar el que no veu, de manera que tapar-se les orelles els pot mitigar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb8-7fd4-ace0-6c643345bb0d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Doku Doku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdb9-7ebc-96c5-860faca3deaa', 'DEVIL_FRUIT', 'Doku Doku no mi', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdb9-7ebc-96c5-860faca3deaa', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb9-7ebc-96c5-860faca3deaa')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb9-7ebc-96c5-860faca3deaa', 'es-ES', 'Permite al usuario producir y controlar todo tipo de veneno, siendo además totalmente inmune a él.', ARRAY['Venenos versátiles: El usuario crea venenos en forma líquida o gaseosa, desde irritantes leves hasta agentes paralizantes y toxinas corrosivas mortales (como hace Magellan).','Criaturas de veneno: Grandes cantidades de veneno líquido pueden moldearse en formas monstruosas gigantescas, como una hidra, para atacar a los enemigos.','Inmunidad total: El usuario es inmune a todas las toxinas, ya las produzca la fruta o provengan de fuentes externas, aunque su cuerpo no se transforma en veneno.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb9-7ebc-96c5-860faca3deaa')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb9-7ebc-96c5-860faca3deaa', 'en-GB', 'Allows the user to produce and control all kinds of poison, while being completely immune to it.', ARRAY['Versatile Poisons: The user creates poisons in liquid or gas form, from mild irritants to paralysing agents and deadly corrosive toxins (as Magellan does).','Poison Constructs: Large amounts of liquid poison can be shaped into gigantic monstrous forms, such as a hydra, to attack enemies.','Total Immunity: The user is immune to all toxins, whether made by the fruit or by outside sources, though their body does not turn into poison.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb9-7ebc-96c5-860faca3deaa')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdb9-7ebc-96c5-860faca3deaa', 'ca-ES', 'Permet a l''usuari produir i controlar tota mena de verí, i alhora és totalment immune a ell.', ARRAY['Verins versàtils: L''usuari crea verins en forma líquida o gasosa, des d''irritants lleus fins a agents paralitzants i toxines corrosives mortals (com fa Magellan).','Criatures de verí: Grans quantitats de verí líquid es poden modelar en formes monstruoses gegantines, com una hidra, per atacar els enemics.','Immunitat total: L''usuari és immune a totes les toxines, tant si les produeix la fruita com si provenen de fonts externes, tot i que el seu cos no es transforma en verí.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdb9-7ebc-96c5-860faca3deaa')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Horu Horu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdba-7abf-a17c-d551208db5d7', 'DEVIL_FRUIT', 'Horu Horu no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdba-7abf-a17c-d551208db5d7', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdba-7abf-a17c-d551208db5d7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdba-7abf-a17c-d551208db5d7', 'es-ES', 'Permite al usuario crear hormonas especiales capaces de alterar el cuerpo de cualquiera a quien se las inyecte.', ARRAY['Inyección de hormonas: El usuario convierte sus dedos en agujas para inyectar hormonas que remodelan un cuerpo, como el género, el crecimiento, la temperatura o la pigmentación de la piel (como hace Ivankov).','Mejoras para el combate: Las hormonas pueden dar una descarga de adrenalina, agrandar partes del cuerpo para reforzar un ataque o cambiar el género de una persona para obtener los rasgos del sexo opuesto.','Riesgos curativos: Las hormonas pueden reforzar el sistema inmunitario frente a enfermedades o venenos, pero solo aumentan ligeramente la supervivencia y pueden tener efectos secundarios mortales si se abusa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdba-7abf-a17c-d551208db5d7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdba-7abf-a17c-d551208db5d7', 'en-GB', 'Allows the user to create special hormones that can alter the body of anyone they are injected into.', ARRAY['Hormone Injection: The user turns their fingers into needles to inject hormones that reshape a body, such as gender, growth, temperature or skin pigmentation (as Ivankov does).','Combat Boosts: Hormones can give a burst of adrenaline, enlarge body parts to strengthen an attack, or change a person''s gender to gain the traits of the opposite sex.','Healing Risks: Hormones can boost the immune system against illness or poison, but they only slightly raise survival odds and can have deadly side effects if overused.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdba-7abf-a17c-d551208db5d7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdba-7abf-a17c-d551208db5d7', 'ca-ES', 'Permet a l''usuari crear hormones especials capaces d''alterar el cos de qualsevol a qui les injecti.', ARRAY['Injecció d''hormones: L''usuari converteix els dits en agulles per injectar hormones que remodelen un cos, com el gènere, el creixement, la temperatura o la pigmentació de la pell (com fa Ivankov).','Millores per al combat: Les hormones poden donar una descàrrega d''adrenalina, engrandir parts del cos per reforçar un atac o canviar el gènere d''una persona per obtenir els trets del sexe oposat.','Riscos curatius: Les hormones poden reforçar el sistema immunitari davant de malalties o verins, però només augmenten lleugerament la supervivència i poden tenir efectes secundaris mortals si se n''abusa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdba-7abf-a17c-d551208db5d7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Choki Choki no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdbb-7dea-84c8-a8227543661d', 'DEVIL_FRUIT', 'Choki Choki no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdbb-7dea-84c8-a8227543661d', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbb-7dea-84c8-a8227543661d')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbb-7dea-84c8-a8227543661d', 'es-ES', 'Permite al usuario convertir partes de su cuerpo en tijeras capaces de cortar objetos sólidos como si fueran papel.', ARRAY['Cuerpo de tijera: Cualquier parte del cuerpo, desde dos dedos hasta las manos enteras, puede volverse unas tijeras de cualquier tamaño, de modo que siempre hay una hoja a mano.','Corte de piedra: Las tijeras atraviesan materia sólida como la piedra, y lo cortado se maneja después con la ligereza y flexibilidad del papel.','Moldeado del terreno: Las tiras de suelo cortado pueden lanzarse contra los enemigos, alzarse como muro protector o tenderse como puente entre dos puntos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbb-7dea-84c8-a8227543661d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbb-7dea-84c8-a8227543661d', 'en-GB', 'Allows the user to turn parts of their body into scissors that can cut through solid objects as if they were paper.', ARRAY['Scissor Body: Any part of the body, from two fingers to whole hands, can become scissors of any size, so the user always has a blade to hand.','Stone Cutting: The scissors slice through solid matter such as stone, and cut material can then be handled as light and flexible as paper.','Terrain Shaping: Strips of cut ground can be thrown at enemies, raised as a protective wall or laid down as a bridge between two points.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbb-7dea-84c8-a8227543661d')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbb-7dea-84c8-a8227543661d', 'ca-ES', 'Permet a l''usuari convertir parts del seu cos en tisores capaces de tallar objectes sòlids com si fossin paper.', ARRAY['Cos de tisora: Qualsevol part del cos, des de dos dits fins a les mans senceres, pot esdevenir unes tisores de qualsevol mida, de manera que sempre hi ha una fulla a mà.','Tall de pedra: Les tisores travessen matèria sòlida com la pedra, i el que s''ha tallat es pot manejar després amb la lleugeresa i la flexibilitat del paper.','Modelatge del terreny: Les tires de sòl tallat es poden llançar contra els enemics, alçar com a mur protector o estendre com a pont entre dos punts.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbb-7dea-84c8-a8227543661d')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kira Kira no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdbc-7a0b-9008-65f3801bcb63', 'DEVIL_FRUIT', 'Kira Kira no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdbc-7a0b-9008-65f3801bcb63', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbc-7a0b-9008-65f3801bcb63')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbc-7a0b-9008-65f3801bcb63', 'es-ES', 'Permite al usuario convertir su cuerpo en diamante, lo que le otorga una dureza excepcional y golpes físicos muy potentes.', ARRAY['Cuerpo de diamante: Partes del cuerpo del usuario se endurecen hasta convertirse en diamante, una de las sustancias más duras del mundo, y resulta muy difícil herirle.','Barrera viviente: Al interponerse en la trayectoria de un ataque a distancia, el usuario absorbe el impacto con su piel de diamante y protege a quienes tiene detrás.','Golpes densos: El peso y la dureza del diamante multiplican la fuerza de los ataques físicos del usuario, como una embestida con el antebrazo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbc-7a0b-9008-65f3801bcb63')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbc-7a0b-9008-65f3801bcb63', 'en-GB', 'Allows the user to turn their body into diamond, granting exceptional toughness and powerful physical strikes.', ARRAY['Diamond Body: Parts of the user''s body harden into diamond, one of the hardest substances in the world, making them extremely hard to wound.','Living Barrier: By stepping into the path of a long-range attack, the user can absorb the impact with their diamond skin and shield those behind them.','Dense Strikes: The weight and hardness of the diamond multiply the force of the user''s physical attacks, such as a charging forearm slam.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbc-7a0b-9008-65f3801bcb63')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbc-7a0b-9008-65f3801bcb63', 'ca-ES', 'Permet a l''usuari convertir el seu cos en diamant, cosa que li atorga una duresa excepcional i cops físics molt potents.', ARRAY['Cos de diamant: Parts del cos de l''usuari s''endureixen fins a esdevenir diamant, una de les substàncies més dures del món, i resulta molt difícil ferir-lo.','Barrera vivent: En interposar-se en la trajectòria d''un atac a distància, l''usuari absorbeix l''impacte amb la seva pell de diamant i protegeix qui té al darrere.','Cops densos: El pes i la duresa del diamant multipliquen la força dels atacs físics de l''usuari, com una envestida amb l''avantbraç.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbc-7a0b-9008-65f3801bcb63')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Poke Poke no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdbd-7b52-abdb-8f372e10f655', 'DEVIL_FRUIT', 'Poke Poke no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdbd-7b52-abdb-8f372e10f655', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbd-7b52-abdb-8f372e10f655')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbd-7b52-abdb-8f372e10f655', 'es-ES', 'Permite al usuario crear bolsillos en cualquier parte de su cuerpo, capaces de guardar objetos mucho más grandes que su abertura.', ARRAY['Bolsillos en cualquier parte: El usuario puede abrir bolsillos en cualquier zona de su cuerpo, y se dice que puede tener un número ilimitado de ellos.','Almacén elástico: Los objetos guardados cambian de tamaño para encajar, así que incluso piezas mucho mayores que la abertura caben dentro.','Arsenal oculto: Armas como un martillo gigante pueden guardarse en un bolsillo y sacarse cuando hagan falta, incluso desde la barbilla.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbd-7b52-abdb-8f372e10f655')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbd-7b52-abdb-8f372e10f655', 'en-GB', 'Allows the user to create pockets anywhere on their body, which can hold objects far larger than the openings themselves.', ARRAY['Pockets Anywhere: The user can open pockets on any part of their body, and is said to be able to have an unlimited number of them.','Stretching Storage: Objects stored inside change size to fit, so items much larger than the pocket opening can still be tucked away.','Hidden Arsenal: Weapons such as a giant hammer can be kept inside a pocket and drawn out whenever needed, even from the chin.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbd-7b52-abdb-8f372e10f655')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbd-7b52-abdb-8f372e10f655', 'ca-ES', 'Permet a l''usuari crear butxaques a qualsevol part del seu cos, capaces de guardar objectes molt més grans que la seva obertura.', ARRAY['Butxaques a qualsevol lloc: L''usuari pot obrir butxaques en qualsevol zona del seu cos, i es diu que pot tenir-ne un nombre il·limitat.','Magatzem elàstic: Els objectes guardats canvien de mida per encaixar, així que fins i tot peces molt més grans que l''obertura hi caben.','Arsenal ocult: Armes com un martell gegant es poden guardar en una butxaca i treure quan calgui, fins i tot des de la barbeta.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbd-7b52-abdb-8f372e10f655')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Woshu Woshu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdbe-7b62-80bb-ebe7af70c586', 'DEVIL_FRUIT', 'Woshu Woshu no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdbe-7b62-80bb-ebe7af70c586', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbe-7b62-80bb-ebe7af70c586')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbe-7b62-80bb-ebe7af70c586', 'es-ES', 'Permite al usuario lavar y tender a personas y objetos como si fueran ropa.', ARRAY['Colada humana: El usuario puede lavar y tender a personas y objetos como si fueran ropa, dejando a sus objetivos incapaces de luchar o moverse.','Limpieza del corazón: Los villanos tendidos también ven su corazón malvado un poco más limpio, lo que atenúa la maldad de los piratas peligrosos.','Derribo silencioso: El poder inutiliza a los rivales sin un combate convencional, lo que lo vuelve eficaz para reducir a piratas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbe-7b62-80bb-ebe7af70c586')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbe-7b62-80bb-ebe7af70c586', 'en-GB', 'Allows the user to wash and hang out to dry people and objects as if they were clothes.', ARRAY['Human Laundry: The user can wash and hang up people and objects like laundry, leaving their targets unable to fight back or move.','Cleansing of the Heart: Dried villains also have their evil hearts made a little cleaner, softening the malice of dangerous pirates.','Quiet Takedown: The power disables opponents without a conventional fight, making it an effective way to subdue pirates.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbe-7b62-80bb-ebe7af70c586')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbe-7b62-80bb-ebe7af70c586', 'ca-ES', 'Permet a l''usuari rentar i estendre persones i objectes com si fossin roba.', ARRAY['Bugada humana: L''usuari pot rentar i estendre persones i objectes com si fossin roba, deixant els seus objectius incapaços de lluitar o moure''s.','Neteja del cor: Els malvats estesos també veuen el seu cor malvat una mica més net, cosa que atenua la maldat dels pirates perillosos.','Derrota silenciosa: El poder neutralitza els rivals sense un combat convencional, i resulta eficaç per reduir pirates.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbe-7b62-80bb-ebe7af70c586')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Fuwa Fuwa no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdbf-738d-b101-231ef61e065e', 'DEVIL_FRUIT', 'Fuwa Fuwa no mi', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdbf-738d-b101-231ef61e065e', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbf-738d-b101-231ef61e065e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbf-738d-b101-231ef61e065e', 'es-ES', 'Permite al usuario levitar a sí mismo y a cualquier objeto inanimado que haya tocado.', ARRAY['Autolevitación: El usuario puede anular la gravedad sobre su propio cuerpo para flotar y volar, esquivando muchos ataques con facilidad.','Objetos flotantes: Cualquier cosa sin vida que el usuario haya tocado, por pesada que sea, puede hacerse flotar y se mantiene en el aire hasta que él decida lo contrario.','Objetos como armas: Los objetos levitados, como rocas enormes o buques de guerra, pueden dejarse caer sobre los enemigos, usarse de escudo o remodelarse en figuras que embisten.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbf-738d-b101-231ef61e065e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbf-738d-b101-231ef61e065e', 'en-GB', 'Allows the user to levitate themselves and any inanimate objects they have touched.', ARRAY['Self Levitation: The user can nullify gravity on their own body to float and fly, avoiding many attacks with ease.','Floating Objects: Anything non-living the user has touched, however heavy, can be made to float and stays aloft until the user decides otherwise.','Objects as Weapons: Levitated objects, such as huge rocks or warships, can be dropped on enemies, used as shields or reshaped into charging figures.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbf-738d-b101-231ef61e065e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdbf-738d-b101-231ef61e065e', 'ca-ES', 'Permet a l''usuari fer levitar-se a si mateix i qualsevol objecte inanimat que hagi tocat.', ARRAY['Autolevitació: L''usuari pot anul·lar la gravetat sobre el seu propi cos per flotar i volar, esquivant molts atacs amb facilitat.','Objectes flotants: Qualsevol cosa sense vida que l''usuari hagi tocat, per pesada que sigui, es pot fer flotar i es manté a l''aire fins que ell decideixi el contrari.','Objectes com a armes: Els objectes levitats, com roques enormes o vaixells de guerra, es poden deixar caure sobre els enemics, usar d''escut o remodelar en figures que embesteixen.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdbf-738d-b101-231ef61e065e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Deka Deka no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc0-75d6-bfb8-3b65303e9269', 'DEVIL_FRUIT', 'Deka Deka no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc0-75d6-bfb8-3b65303e9269', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc0-75d6-bfb8-3b65303e9269')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc0-75d6-bfb8-3b65303e9269', 'es-ES', 'Permite al usuario agrandar drásticamente su cuerpo hasta convertirse en un gigante colosal.', ARRAY['Tamaño colosal: El cuerpo del usuario crece hasta una altura enorme, muy superior a la de un gigante normal, y pasa a ser uno de los seres más grandes del mundo.','Fuerza descomunal: Con un tamaño tan inmenso, el usuario puede levantar y recolocar edificios enteros con facilidad.','Vadeo marino: La altura le permite mantenerse en pie en ciertas zonas del océano, aunque hacerlo resulta agotador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc0-75d6-bfb8-3b65303e9269')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc0-75d6-bfb8-3b65303e9269', 'en-GB', 'Allows the user to drastically enlarge their body, becoming a colossal giant.', ARRAY['Colossal Size: The user''s body grows to an enormous height, far beyond that of an ordinary giant, making them one of the largest beings in the world.','Towering Strength: With such immense size, the user can lift and rearrange entire buildings with ease.','Sea Wading: The sheer height lets the user stand in certain stretches of the ocean, although doing so is exhausting.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc0-75d6-bfb8-3b65303e9269')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc0-75d6-bfb8-3b65303e9269', 'ca-ES', 'Permet a l''usuari engrandir dràsticament el seu cos fins a esdevenir un gegant colossal.', ARRAY['Mida colossal: El cos de l''usuari creix fins a una alçada enorme, molt superior a la d''un gegant normal, i passa a ser un dels éssers més grans del món.','Força descomunal: Amb una mida tan immensa, l''usuari pot aixecar i recol·locar edificis sencers amb facilitat.','Travessia marina: L''alçada li permet mantenir-se dret en certes zones de l''oceà, tot i que fer-ho resulta esgotador.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc0-75d6-bfb8-3b65303e9269')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Mato Mato no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc1-7026-8676-1e03844499e8', 'DEVIL_FRUIT', 'Mato Mato no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc1-7026-8676-1e03844499e8', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc1-7026-8676-1e03844499e8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc1-7026-8676-1e03844499e8', 'es-ES', 'Permite al usuario marcar objetivos con el tacto y apuntarles con cualquier cosa desde cualquier distancia, con el proyectil persiguiendo su marca.', ARRAY['Marcado de objetivos: Al tocar a alguien con una mano queda grabado como objetivo, y el usuario puede memorizar tantos objetivos como manos tenga.','Lanzamientos teledirigidos: Cualquier cosa enviada hacia un objetivo marcado, grande o pequeña, lo sigue desde cualquier lugar y cambia de rumbo si este se mueve.','Proyectiles vivientes: El usuario no necesita lanzar el objeto, pues al tocarlo con una mano marcada sale volando, incluso un barco enorme o un grupo de personas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc1-7026-8676-1e03844499e8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc1-7026-8676-1e03844499e8', 'en-GB', 'Allows the user to mark targets by touch and then aim anything at them from any distance, with the projectile homing in on its mark.', ARRAY['Target Marking: Touching someone with a hand imprints them as a target, and the user can memorise as many targets as they have hands.','Homing Throws: Anything sent toward a marked target, large or small, tracks them from any location and changes course to follow them if they move.','Living Projectiles: The user need not throw the object, as touching it with a marked hand sends it flying, even a huge ship or a group of people.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc1-7026-8676-1e03844499e8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc1-7026-8676-1e03844499e8', 'ca-ES', 'Permet a l''usuari marcar objectius amb el tacte i apuntar-los amb qualsevol cosa des de qualsevol distància, amb el projectil perseguint la seva marca.', ARRAY['Marcatge d''objectius: En tocar algú amb una mà queda gravat com a objectiu, i l''usuari pot memoritzar tants objectius com mans tingui.','Llançaments teledirigits: Qualsevol cosa enviada cap a un objectiu marcat, gran o petita, el segueix des de qualsevol lloc i canvia de rumb si es mou.','Projectils vivents: L''usuari no necessita llançar l''objecte, ja que en tocar-lo amb una mà marcada surt volant, fins i tot un vaixell enorme o un grup de persones.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc1-7026-8676-1e03844499e8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Fuku Fuku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc2-7e2e-bfa8-194b1667d58a', 'DEVIL_FRUIT', 'Fuku Fuku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a', 'es-ES', 'Permite al usuario crear ropa y otras prendas a partir de hojas o piedras con solo visualizarlas.', ARRAY['Atuendos instantáneos: Al colocar una hoja o una piedra en la cabeza de alguien se convierte en la vestimenta que el usuario imagina, y dura mientras se lleve puesta.','Disfraces: La ropa creada sirve como disfraz, lo que permite al usuario y a sus aliados pasar desapercibidos en labores de espionaje.','Prendas útiles: Además de disfraces, el usuario puede crear ropa de abrigo contra el frío e incluso armaduras y armas funcionales, como espadas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a', 'en-GB', 'Allows the user to create clothing and other apparel out of leaves or stones simply by visualising it.', ARRAY['Instant Outfits: Placing a leaf or stone on someone''s head turns it into the garb the user pictures, which lasts for as long as it is worn.','Disguises: The clothes created work as disguises, letting the user and their allies blend in during espionage work.','Useful Apparel: Beyond disguises, the user can make warm clothing against the cold, and even functioning armour and weapons such as swords.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a', 'ca-ES', 'Permet a l''usuari crear roba i altres peces a partir de fulles o pedres només visualitzant-les.', ARRAY['Vestits instantanis: En col·locar una fulla o una pedra al cap d''algú es converteix en la indumentària que l''usuari s''imagina, i dura mentre es porti posada.','Disfresses: La roba creada serveix de disfressa, cosa que permet a l''usuari i als seus aliats passar desapercebuts en tasques d''espionatge.','Peces útils: A més de disfresses, l''usuari pot crear roba d''abric contra el fred i fins i tot armadures i armes funcionals, com espases.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc2-7e2e-bfa8-194b1667d58a')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Buki Buki no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc3-7435-90d5-1030020443d7', 'DEVIL_FRUIT', 'Buki Buki no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc3-7435-90d5-1030020443d7', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc3-7435-90d5-1030020443d7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc3-7435-90d5-1030020443d7', 'es-ES', 'Permite al usuario transformar partes de su cuerpo en cualquier tipo de arma.', ARRAY['Cuerpo arma: Los dedos, los brazos o incluso todo el cuerpo pueden convertirse en espadas, armas de fuego y otros armamentos, así que no se necesitan armas externas.','Formas explosivas: El usuario puede transformarse en misiles o bombas y no sufre daño por la explosión de su propio cuerpo transformado.','Cuerpo regenerable: Si el cuerpo convertido en arma se rompe en pedazos, estos se recomponen en un cuerpo entero, aunque las armas dirigidas contra el usuario pueden herirle.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc3-7435-90d5-1030020443d7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc3-7435-90d5-1030020443d7', 'en-GB', 'Allows the user to transform their body parts into any kind of weapon.', ARRAY['Weapon Body: Fingers, arms or even the whole body can become blades, firearms and other weapons, so the user needs no external arms.','Explosive Forms: The user can turn into missiles or bombs, and takes no damage from the blast of their own transformed body.','Reforming Body: If the user''s weaponised body is broken apart, the pieces reform into a whole body, though weapons aimed at the user can still hurt.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc3-7435-90d5-1030020443d7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc3-7435-90d5-1030020443d7', 'ca-ES', 'Permet a l''usuari transformar parts del seu cos en qualsevol tipus d''arma.', ARRAY['Cos arma: Els dits, els braços o fins i tot tot el cos poden esdevenir espases, armes de foc i altres armaments, així que no calen armes externes.','Formes explosives: L''usuari es pot transformar en míssils o bombes i no pateix cap dany per l''explosió del seu propi cos transformat.','Cos regenerable: Si el cos convertit en arma es trenca a trossos, aquests es tornen a unir en un cos sencer, tot i que les armes dirigides contra l''usuari el poden ferir.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc3-7435-90d5-1030020443d7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Guru Guru no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc4-7384-9eef-d6cf130c4129', 'DEVIL_FRUIT', 'Guru Guru no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc4-7384-9eef-d6cf130c4129', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc4-7384-9eef-d6cf130c4129')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc4-7384-9eef-d6cf130c4129', 'es-ES', 'Permite al usuario convertir cualquier parte de su cuerpo o de la ropa que lleva en hélices que giran a gran velocidad.', ARRAY['Piezas de hélice: Cualquier parte del cuerpo o prenda puede volverse una hélice, y el usuario puede incluso hacer girar las piernas como patines sobre el suelo.','Vuelo propulsado: El giro es lo bastante fuerte como para llevar al usuario a distancias muy largas, permitiéndole volar con todo el cuerpo.','Fuerza de vendaval: La rotación rápida crea fuertes tormentas de viento y puede lanzar objetos u otras personas a gran velocidad.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc4-7384-9eef-d6cf130c4129')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc4-7384-9eef-d6cf130c4129', 'en-GB', 'Allows the user to turn any part of their body or worn clothing into propellers that spin at high speed.', ARRAY['Propeller Parts: Any body part or piece of clothing can become a propeller, and the user can even spin their legs like roller skates on the ground.','Propelled Flight: The spinning is strong enough to carry the user over very long distances, letting them fly with their whole body.','Gale Force: Rapid rotation creates powerful windstorms and can launch objects or other people at high speed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc4-7384-9eef-d6cf130c4129')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc4-7384-9eef-d6cf130c4129', 'ca-ES', 'Permet a l''usuari convertir qualsevol part del seu cos o de la roba que du en hèlixs que giren a gran velocitat.', ARRAY['Peces d''hèlix: Qualsevol part del cos o peça de roba pot esdevenir una hèlix, i l''usuari pot fins i tot fer girar les cames com patins sobre el terra.','Vol propulsat: El gir és prou fort per portar l''usuari a distàncies molt llargues, permetent-li volar amb tot el cos.','Força de ventada: La rotació ràpida crea fortes tempestes de vent i pot llançar objectes o altres persones a gran velocitat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc4-7384-9eef-d6cf130c4129')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Beta Beta no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc5-7566-a9a5-b4d07ceec577', 'DEVIL_FRUIT', 'Beta Beta no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc5-7566-a9a5-b4d07ceec577', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc5-7566-a9a5-b4d07ceec577')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc5-7566-a9a5-b4d07ceec577', 'es-ES', 'Permite al usuario crear y controlar cantidades ilimitadas de mucosidad.', ARRAY['Mucosidad infinita: El usuario puede producir mucosidad sin límite, lo bastante fuerte para sujetar un barco, inmovilizar a un rival o adherirse a paredes y techos.','Proyección de mucosidad: La mucosidad puede extenderse para atrapar objetivos lejanos, dispararse con gran fuerza y precisión o usarse como mayal para blandir objetos pesados.','Capa inflamable: Envuelto en una gruesa capa de mucosidad, el usuario puede engañar a los enemigos para que golpeen el vacío, y puede prenderla para causar daño explosivo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc5-7566-a9a5-b4d07ceec577')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc5-7566-a9a5-b4d07ceec577', 'en-GB', 'Allows the user to create and control unlimited amounts of mucus.', ARRAY['Endless Mucus: The user can produce limitless mucus, strong enough to hold a ship, to restrain movement or to stick to walls and ceilings.','Mucus Projection: Mucus can be extended to seize distant targets, flung with great force and precision, or used as a flail to swing heavy objects.','Flammable Coat: Wrapped in a thick coat of mucus, the user can fool enemies into hitting empty space, and can ignite it to cause explosive fire damage.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc5-7566-a9a5-b4d07ceec577')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc5-7566-a9a5-b4d07ceec577', 'ca-ES', 'Permet a l''usuari crear i controlar quantitats il·limitades de mucositat.', ARRAY['Mucositat infinita: L''usuari pot produir mucositat sense límit, prou forta per sostenir un vaixell, immobilitzar un rival o adherir-se a parets i sostres.','Projecció de mucositat: La mucositat es pot estendre per atrapar objectius llunyans, disparar amb gran força i precisió o usar com a maça per brandar objectes pesants.','Capa inflamable: Embolcallat en una gruixuda capa de mucositat, l''usuari pot enganyar els enemics perquè colpegin el buit, i pot encendre-la per causar dany explosiu.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc5-7566-a9a5-b4d07ceec577')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Zushi Zushi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc6-7ac1-99f3-8326680030e9', 'DEVIL_FRUIT', 'Zushi Zushi no mi', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc6-7ac1-99f3-8326680030e9', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc6-7ac1-99f3-8326680030e9')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc6-7ac1-99f3-8326680030e9', 'es-ES', 'Permite al usuario crear y controlar fuerzas gravitatorias de enorme potencia.', ARRAY['Campos de gravedad: El usuario puede crear gravedad de intensidad y dirección variables, inmovilizando enemigos o atrayendo cosas hacia un punto, incluso meteoritos del cielo.','Control telecinético: La materia puede moverse con precisión, desde lanzar escombros contra los rivales hasta desplazar un enorme buque de guerra.','Vuelo simulado: Al anular la gravedad y empujar las cosas hacia arriba, el usuario puede hacer levitar una roca o un barco para viajar por el aire, aunque no a sí mismo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc6-7ac1-99f3-8326680030e9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc6-7ac1-99f3-8326680030e9', 'en-GB', 'Allows the user to create and control gravitational forces of extreme potency.', ARRAY['Gravity Fields: The user can create gravity of varying intensity and direction, pinning enemies down or pulling things toward a point, even meteors from the sky.','Telekinetic Control: Matter can be moved with precision, from hurling debris at foes to shifting a massive battleship.','Simulated Flight: By negating gravity and pushing things upward, the user can levitate a rock or ship to travel through the air, though not themselves.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc6-7ac1-99f3-8326680030e9')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc6-7ac1-99f3-8326680030e9', 'ca-ES', 'Permet a l''usuari crear i controlar forces gravitatòries d''enorme potència.', ARRAY['Camps de gravetat: L''usuari pot crear gravetat d''intensitat i direcció variables, immobilitzant enemics o atraient coses cap a un punt, fins i tot meteorits del cel.','Control telecinètic: La matèria es pot moure amb precisió, des de llançar enderrocs contra els rivals fins a desplaçar un enorme vaixell de guerra.','Vol simulat: En anul·lar la gravetat i empènyer les coses cap amunt, l''usuari pot fer levitar una roca o un vaixell per viatjar per l''aire, tot i que no a si mateix.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc6-7ac1-99f3-8326680030e9')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Bari Bari no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc7-76b3-9fbd-f149af9629ff', 'DEVIL_FRUIT', 'Bari Bari no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc7-76b3-9fbd-f149af9629ff', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc7-76b3-9fbd-f149af9629ff')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc7-76b3-9fbd-f149af9629ff', 'es-ES', 'Permite al usuario conjurar barreras transparentes de la nada.', ARRAY['Creación de barreras: El usuario conjura paneles y cúpulas invisibles o similares al cristal, normalmente cruzando los dedos, para protegerse de los ataques.','Defensa inquebrantable: Las barreras pueden bloquear por completo golpes y tajos extremadamente potentes, aunque solo se puede mantener una a la vez.','Barreras moldeables: Las barreras pueden lanzarse hacia delante como una fuerza, usarse para devolver los ataques o formar estructuras sólidas, como escaleras para subir.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc7-76b3-9fbd-f149af9629ff')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc7-76b3-9fbd-f149af9629ff', 'en-GB', 'Allows the user to conjure transparent barriers out of thin air.', ARRAY['Barrier Creation: The user conjures invisible or glass-like panes and domes, usually by crossing their fingers, to shield themselves from attacks.','Unbreakable Defence: The barriers can completely block extremely powerful blows and slashes, though only one can be sustained at a time.','Shaped Barriers: Barriers can be launched forward as a force, used to deflect attacks back, or formed into solid shapes such as stairs to stand on.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc7-76b3-9fbd-f149af9629ff')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc7-76b3-9fbd-f149af9629ff', 'ca-ES', 'Permet a l''usuari conjurar barreres transparents del no-res.', ARRAY['Creació de barreres: L''usuari conjura panells i cúpules invisibles o semblants al vidre, normalment creuant els dits, per protegir-se dels atacs.','Defensa inquebrantable: Les barreres poden blocar per complet cops i talls extremadament potents, tot i que només se''n pot mantenir una alhora.','Barreres modelables: Les barreres es poden llançar endavant com una força, usar per retornar els atacs o formar estructures sòlides, com escales per pujar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc7-76b3-9fbd-f149af9629ff')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Giro Giro no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc8-75a2-be62-6f63e680be50', 'DEVIL_FRUIT', 'Giro Giro no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc8-75a2-be62-6f63e680be50', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc8-75a2-be62-6f63e680be50')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc8-75a2-be62-6f63e680be50', 'es-ES', 'Permite al usuario ver a través de cualquier cosa, incluidas la ropa, la mente y los recuerdos de las personas, tanto de cerca como a enormes distancias.', ARRAY['Visión de rayos X: Al formar círculos con los dedos frente a los ojos, el usuario ve a través de la ropa, la piel y los obstáculos como si fueran transparentes.','Lectura de mente: El usuario puede asomarse a los pensamientos y recuerdos de una persona, descubriendo sus mentiras, y también transmitir sus propios recuerdos a la mente de otra.','Vista de largo alcance: El usuario percibe todo lo que ocurre en un radio de miles de kilómetros, sin necesidad de formar los círculos ni de tener los ojos abiertos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc8-75a2-be62-6f63e680be50')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc8-75a2-be62-6f63e680be50', 'en-GB', 'Allows the user to see through anything, including people''s clothes, minds and memories, from close range or across vast distances.', ARRAY['X-Ray Vision: Forming circles with the fingers in front of the eyes lets the user see through clothing, skin and obstacles as if they were transparent.','Mind Reading: The user can look into a person''s thoughts and memories, exposing lies, and can also transfer their own memories into another person''s mind.','Long-Range Sight: The user can perceive everything within a radius of thousands of kilometres, without needing to form the circles or even keep their eyes open.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc8-75a2-be62-6f63e680be50')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc8-75a2-be62-6f63e680be50', 'ca-ES', 'Permet a l''usuari veure a través de qualsevol cosa, incloses la roba, la ment i els records de les persones, tant de prop com a enormes distàncies.', ARRAY['Visió de raigs X: En formar cercles amb els dits davant dels ulls, l''usuari veu a través de la roba, la pell i els obstacles com si fossin transparents.','Lectura de ment: L''usuari pot veure els pensaments i els records d''una persona, descobrint-ne les mentides, i també transmetre els seus propis records a la ment d''una altra.','Vista de llarg abast: L''usuari percep tot el que passa en un radi de milers de quilòmetres, sense haver de formar els cercles ni tenir els ulls oberts.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc8-75a2-be62-6f63e680be50')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Ato Ato no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdc9-70c4-975c-96e3228e9119', 'DEVIL_FRUIT', 'Ato Ato no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdc9-70c4-975c-96e3228e9119', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc9-70c4-975c-96e3228e9119')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc9-70c4-975c-96e3228e9119', 'es-ES', 'Permite al usuario transformar cualquier objetivo en arte abstracto, deformando su aspecto y volviéndolo inútil.', ARRAY['Distorsión artística: El usuario lanza una nube contra un objetivo y retuerce su forma y su función, de modo que armas, vehículos e incluso armas naturales como pezuñas o cuernas dejan de funcionar.','Efecto reversible: El usuario puede devolver un objeto distorsionado a su estado normal del mismo modo, y todas las distorsiones se deshacen si el usuario pierde el conocimiento.','Lienzo viviente: Las personas y su equipo pueden convertirse en obras abstractas, lo que permite al usuario inutilizar enemigos y sus vías de escape, como los barcos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc9-70c4-975c-96e3228e9119')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc9-70c4-975c-96e3228e9119', 'en-GB', 'Allows the user to warp any target into abstract art, distorting its shape and rendering it useless.', ARRAY['Art Distortion: The user throws a cloud at a target, twisting its appearance and function so that weapons, vehicles and even natural weapons like hooves or antlers stop working.','Reversible Effect: The user can return a distorted object to its normal state in the same way, and all distortions fade if the user is knocked unconscious.','Living Canvas: People and their equipment can be reshaped into abstract artworks, which lets the user disable enemies and their escape routes, such as ships.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc9-70c4-975c-96e3228e9119')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdc9-70c4-975c-96e3228e9119', 'ca-ES', 'Permet a l''usuari transformar qualsevol objectiu en art abstracte, deformant-ne l''aspecte i deixant-lo inservible.', ARRAY['Distorsió artística: L''usuari llança un núvol contra un objectiu i en retorça la forma i la funció, de manera que armes, vehicles i fins i tot armes naturals com peülles o banyes deixen de funcionar.','Efecte reversible: L''usuari pot tornar un objecte distorsionat al seu estat normal de la mateixa manera, i totes les distorsions es desfan si l''usuari perd el coneixement.','Llenç viu: Les persones i el seu equipament es poden convertir en obres abstractes, cosa que permet a l''usuari inutilitzar enemics i les seves vies de fuga, com ara els vaixells.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdc9-70c4-975c-96e3228e9119')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Jake Jake no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693', 'DEVIL_FRUIT', 'Jake Jake no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693', 'es-ES', 'Permite al usuario transformarse en una chaqueta de cuerpo entero que otra persona puede vestir, tomando el control de su cuerpo.', ARRAY['Chaqueta viviente: El usuario se transforma en una chaqueta que humanos o animales pueden ponerse, sin perder la capacidad de moverse ni de hablar.','Control corporal: Una vez puesta, el usuario controla el cuerpo de quien la lleva y obtiene sus habilidades, lo que puede aumentar mucho la fuerza conjunta.','Ajuste flexible: La chaqueta se adapta al tamaño de quien la viste, incluso si es mucho más grande que el usuario, aunque necesita a otra persona o criatura que se la ponga.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693', 'en-GB', 'Allows the user to turn into a full-body jacket that another person can wear, taking control of the wearer''s body.', ARRAY['Living Jacket: The user transforms into a jacket that humans or animals can wear, while still being able to move and speak.','Body Control: Once worn, the user controls the wearer''s body and gains their abilities, which can greatly increase the combined strength.','Flexible Fit: The jacket adapts to the size of the wearer, even when they are far larger than the user, although it needs another person or creature to wear it.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693', 'ca-ES', 'Permet a l''usuari transformar-se en una jaqueta de cos sencer que una altra persona es pot posar, prenent el control del seu cos.', ARRAY['Jaqueta vivent: L''usuari es transforma en una jaqueta que humans o animals es poden posar, sense perdre la capacitat de moure''s ni de parlar.','Control corporal: Un cop posada, l''usuari controla el cos de qui la porta i n''obté les habilitats, cosa que pot augmentar molt la força conjunta.','Ajust flexible: La jaqueta s''adapta a la mida de qui la vesteix, fins i tot si és molt més gran que l''usuari, encara que necessita una altra persona o criatura que se la posi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdca-73ba-9a3a-d9d5d7ab6693')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Pamu Pamu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdcb-7766-ac25-03d65a30150f', 'DEVIL_FRUIT', 'Pamu Pamu no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdcb-7766-ac25-03d65a30150f', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcb-7766-ac25-03d65a30150f')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcb-7766-ac25-03d65a30150f', 'es-ES', 'Permite al usuario hacer que cualquier objeto inorgánico que toque se hinche y explote violentamente.', ARRAY['Hinchazón explosiva: Todo lo inorgánico que toca el usuario se infla y luego estalla, dañando lo que haya cerca mientras el usuario sale ileso.','Autodetonación: El usuario puede aplicar el poder a su propio cuerpo, como hinchar un brazo o un casco y hacerlo estallar para herir a los enemigos con la metralla.','Superficies minadas: El usuario puede implantar puntos de hinchazón en una superficie que explotan al tocarlos, pero el efecto desaparece si pierde el conocimiento o suelta el objetivo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcb-7766-ac25-03d65a30150f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcb-7766-ac25-03d65a30150f', 'en-GB', 'Allows the user to make any inorganic object they touch swell up and explode violently.', ARRAY['Explosive Swelling: Anything inorganic the user touches inflates and then ruptures, damaging everything nearby while the user remains unharmed.','Self-Detonation: The user can apply the power to their own body, such as swelling an arm or helmet and bursting it to hit enemies with shrapnel.','Landmine Surfaces: The user can implant swelling points in a surface that explode when touched, but the effect fades if the user loses consciousness or lets go of the target.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcb-7766-ac25-03d65a30150f')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcb-7766-ac25-03d65a30150f', 'ca-ES', 'Permet a l''usuari fer que qualsevol objecte inorgànic que toqui s''inflï i exploti violentament.', ARRAY['Inflor explosiva: Tot el que és inorgànic i toca l''usuari s''infla i després rebenta, fent mal a tot el que hi ha a prop mentre l''usuari en surt il·lès.','Autodetonació: L''usuari pot aplicar el poder al seu propi cos, com inflar un braç o un casc i fer-lo rebentar per ferir els enemics amb la metralla.','Superfícies minades: L''usuari pot implantar punts d''inflor en una superfície que exploten en tocar-los, però l''efecte desapareix si perd el coneixement o deixa anar l''objectiu.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcb-7766-ac25-03d65a30150f')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Sui Sui no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc', 'DEVIL_FRUIT', 'Sui Sui no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc', 'es-ES', 'Permite al usuario nadar a través de materia sólida, como el suelo y las paredes, como si fuera agua.', ARRAY['Nado sólido: El usuario puede nadar por el suelo y las paredes, incluso hacia arriba desafiando la gravedad, y la materia vuelve a su sitio sin dejar rastro.','Aproximación sigilosa: Al quedar casi oculto bajo la superficie, el usuario puede infiltrarse en lugares y acercarse sin ser visto para atacar a sus rivales por la espalda.','Golpes con impulso: El usuario puede ganar velocidad mientras nada, incluso arrastrando a un rival, para dar más fuerza a sus derribos y otros ataques.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc', 'en-GB', 'Allows the user to swim through solid matter such as ground and walls as if it were water.', ARRAY['Solid Swimming: The user can swim through the ground and walls, even upwards against gravity, and the matter flows back into place without leaving a trace.','Stealth Approach: Being mostly hidden beneath the surface lets the user infiltrate places and sneak up on opponents to attack them from behind.','Momentum Strikes: The user can pick up speed while swimming, even carrying an opponent along, to add force to throws and other attacks.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc', 'ca-ES', 'Permet a l''usuari nedar a través de matèria sòlida, com el terra i les parets, com si fos aigua.', ARRAY['Natació sòlida: L''usuari pot nedar pel terra i les parets, fins i tot cap amunt desafiant la gravetat, i la matèria torna al seu lloc sense deixar rastre.','Aproximació sigil·losa: En quedar gairebé amagat sota la superfície, l''usuari pot infiltrar-se en llocs i apropar-se sense ser vist per atacar els rivals per l''esquena.','Cops amb impuls: L''usuari pot guanyar velocitat mentre neda, fins i tot arrossegant un rival, per donar més força als seus derrocaments i altres atacs.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcc-74a4-8b76-a2fb7cb62cdc')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Ton Ton no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdcd-7896-9b92-69c1c3d1fa02', 'DEVIL_FRUIT', 'Ton Ton no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02', 'es-ES', 'Permite al usuario aumentar su peso corporal en cantidades enormes, medidas en toneladas.', ARRAY['Peso aplastante: El usuario puede llegar a pesar miles de toneladas, lo bastante para agrietar el suelo y romper los huesos de gigantes al caer sobre un rival.','Peso añadido: El peso del usuario puede aumentar sosteniendo objetos pesados, como un gran escudo, para aplastar a sus rivales con aún más fuerza.','Flotación casi ingrávida: Aunque solo lo hace más pesado, el fruto también permite al usuario flotar en el aire casi como si no pesara nada.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02', 'en-GB', 'Allows the user to increase their body weight by enormous amounts, measured in tonnes.', ARRAY['Crushing Weight: The user can become as heavy as thousands of tonnes, enough to crack the ground and break the bones of giants when they land on an opponent.','Heavy Supplement: The user''s weight can be added to by holding heavy objects, such as a large shield, to crush opponents even harder.','Near-Weightless Floating: Despite only making the user heavier, the fruit also lets them float in the air almost as if they weighed nothing.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02', 'ca-ES', 'Permet a l''usuari augmentar el seu pes corporal en quantitats enormes, mesurades en tones.', ARRAY['Pes aixafant: L''usuari pot arribar a pesar milers de tones, prou per esquerdar el terra i trencar els ossos de gegants en caure sobre un rival.','Pes afegit: El pes de l''usuari pot augmentar sostenint objectes pesats, com un gran escut, per aixafar els rivals amb encara més força.','Flotació quasi sense pes: Encara que només el fa més pesat, el fruit també permet a l''usuari flotar a l''aire gairebé com si no pesés res.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcd-7896-9b92-69c1c3d1fa02')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Hira Hira no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdce-7052-81aa-a07fbbcd75f7', 'DEVIL_FRUIT', 'Hira Hira no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdce-7052-81aa-a07fbbcd75f7', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdce-7052-81aa-a07fbbcd75f7')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdce-7052-81aa-a07fbbcd75f7', 'es-ES', 'Permite al usuario dotar a su cuerpo y a todo lo que toque de las propiedades maleables de la tela.', ARRAY['Propiedades de tela: Los objetos afectados se vuelven planos, flexibles y plegables, de modo que incluso el metal endurecido ondea como una bandera y conserva su dureza y su filo.','Cuerpo bandera: El usuario puede aplanar su propio cuerpo hasta quedar como una bandera, lo que le permite flotar sin riesgo al caer o esquivar golpes.','Armas remodeladas: El usuario puede estirar y remodelar sus armas, como alargar un estoque hasta convertirlo en una hoja flexible como un látigo, o plegar objetos para ocultarlos.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdce-7052-81aa-a07fbbcd75f7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdce-7052-81aa-a07fbbcd75f7', 'en-GB', 'Allows the user to give their body and anything they touch the malleable properties of cloth.', ARRAY['Cloth Properties: Affected objects become flat, flexible and foldable, so even hardened metal can flutter like a flag, while keeping its toughness and sharpness.','Flag Body: The user can flatten their own body into a flag-like state, which lets them float safely when falling or avoid blows.','Reshaped Weapons: The user can stretch and reshape their weapons, such as lengthening a rapier into a whip-like blade, or fold items small to conceal them.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdce-7052-81aa-a07fbbcd75f7')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdce-7052-81aa-a07fbbcd75f7', 'ca-ES', 'Permet a l''usuari dotar el seu cos i tot el que toqui de les propietats mal·leables de la tela.', ARRAY['Propietats de tela: Els objectes afectats esdevenen plans, flexibles i plegables, de manera que fins i tot el metall endurit oneja com una bandera i conserva la duresa i el tall.','Cos bandera: L''usuari pot aplanar el seu propi cos fins a quedar com una bandera, cosa que li permet flotar sense risc en caure o esquivar cops.','Armes remodelades: L''usuari pot estirar i remodelar les seves armes, com allargar un estoc fins a convertir-lo en una fulla flexible com un fuet, o plegar objectes per amagar-los.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdce-7052-81aa-a07fbbcd75f7')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Ishi Ishi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdcf-7fee-98e2-2f4d82651510', 'DEVIL_FRUIT', 'Ishi Ishi no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdcf-7fee-98e2-2f4d82651510', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcf-7fee-98e2-2f4d82651510')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcf-7fee-98e2-2f4d82651510', 'es-ES', 'Permite al usuario fusionarse con cualquier piedra y controlarla como si fuera una extensión de su propio cuerpo.', ARRAY['Asimilación de piedra: El usuario se fusiona con estructuras de piedra, ocultándose en ellas y controlándolas, incluso remodelando los muros y pasillos de todo un edificio de piedra.','Piedra flexible: La piedra controlada se vuelve sorprendentemente flexible y puede estirarse y remodelarse de muchas formas, y cuanta más piedra haya cerca, mayor es el poder del usuario.','Regeneración pétrea: El daño a la piedra controlada no hiere al usuario, que puede regenerar las partes perdidas en segundos asimilando más piedra.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcf-7fee-98e2-2f4d82651510')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcf-7fee-98e2-2f4d82651510', 'en-GB', 'Allows the user to merge with any stone and take control of it as though it were an extension of their own body.', ARRAY['Stone Assimilation: The user merges with stone structures, hiding inside them and controlling them, even reshaping walls and halls of an entire stone building.','Flexible Stone: Controlled stone becomes surprisingly flexible and can be stretched and reshaped in many ways, and the more stone nearby, the greater the user''s power.','Stone Regeneration: Damage to the controlled stone does not hurt the user, who can regenerate lost parts within seconds by assimilating more stone.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcf-7fee-98e2-2f4d82651510')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdcf-7fee-98e2-2f4d82651510', 'ca-ES', 'Permet a l''usuari fusionar-se amb qualsevol pedra i controlar-la com si fos una extensió del seu propi cos.', ARRAY['Assimilació de pedra: L''usuari es fusiona amb estructures de pedra, amagant-s''hi i controlant-les, fins i tot remodelant les parets i els passadissos de tot un edifici de pedra.','Pedra flexible: La pedra controlada esdevé sorprenentment flexible i es pot estirar i remodelar de moltes maneres, i com més pedra hi ha a prop, més gran és el poder de l''usuari.','Regeneració pètria: El dany a la pedra controlada no fereix l''usuari, que pot regenerar les parts perdudes en segons assimilant més pedra.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdcf-7fee-98e2-2f4d82651510')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Nagi Nagi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd0-7014-a416-1c61e6a9bf89', 'DEVIL_FRUIT', 'Nagi Nagi no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd0-7014-a416-1c61e6a9bf89', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd0-7014-a416-1c61e6a9bf89')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd0-7014-a416-1c61e6a9bf89', 'es-ES', 'Permite al usuario crear campos insonorizados y silenciar cualquier ruido que hagan él mismo u otros.', ARRAY['Campo insonorizado: El usuario crea un campo circular del que no puede salir ni entrar ningún sonido, ideal para conversar en privado lejos de oídos indiscretos.','Acciones silenciosas: El usuario puede anular el ruido de sus propias acciones o de las de otros, lo que facilita mucho el sigilo y los ataques por sorpresa.','Se disipa al caer: Los campos desaparecen si el usuario queda incapacitado, y el silencio no protege contra los ataques en sí.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd0-7014-a416-1c61e6a9bf89')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd0-7014-a416-1c61e6a9bf89', 'en-GB', 'Allows the user to create soundproof fields and silence any noise made by themselves or others.', ARRAY['Soundproof Field: The user creates a circular field in which no sound can leave or enter, ideal for private conversations away from eavesdroppers.','Silent Actions: The user can cancel the noise of their own or someone else''s actions, making stealth and surprise attacks much easier.','Dispelled by Incapacitation: The fields vanish if the user is incapacitated, and silence offers no protection against the attacks themselves.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd0-7014-a416-1c61e6a9bf89')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd0-7014-a416-1c61e6a9bf89', 'ca-ES', 'Permet a l''usuari crear camps insonoritzats i silenciar qualsevol soroll que facin ell mateix o altres.', ARRAY['Camp insonoritzat: L''usuari crea un camp circular del qual no pot sortir ni entrar cap so, ideal per conversar en privat lluny d''orelles indiscretes.','Accions silencioses: L''usuari pot anul·lar el soroll de les seves accions o de les d''altres persones, cosa que facilita molt el sigil i els atacs per sorpresa.','Es dissipa en caure: Els camps desapareixen si l''usuari queda incapacitat, i el silenci no protegeix contra els atacs en si.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd0-7014-a416-1c61e6a9bf89')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Chiyu Chiyu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd1-7b7f-b595-46babe44f7bd', 'DEVIL_FRUIT', 'Chiyu Chiyu no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd1-7b7f-b595-46babe44f7bd', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd1-7b7f-b595-46babe44f7bd')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd1-7b7f-b595-46babe44f7bd', 'es-ES', 'Permite al usuario curar a cualquier ser vivo mediante sus lágrimas o agua curativa que brota de sus palmas.', ARRAY['Lágrimas curativas: Las lágrimas que tocan a un ser vivo lo devuelven a la plena salud y pueden volver a unir partes del cuerpo, aunque no pueden crear tejido de la nada.','Agua curativa: El usuario también puede producir agua con las mismas propiedades restauradoras desde las palmas de las manos, y puede curarse a sí mismo.','Restauración de objetos: El usuario puede reparar objetos inanimados dañados, aunque al hacerlo se acorta su propia esperanza de vida.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd1-7b7f-b595-46babe44f7bd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd1-7b7f-b595-46babe44f7bd', 'en-GB', 'Allows the user to heal any living thing through their tears or healing water produced from their palms.', ARRAY['Healing Tears: Tears that touch a living being restore it to full health and can reattach body parts, though they cannot create tissue from nothing.','Healing Water: The user can also produce water with the same restorative properties from their palms, and can heal themselves too.','Object Restoration: The user can repair damaged inanimate objects, though doing so shortens the user''s own lifespan.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd1-7b7f-b595-46babe44f7bd')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd1-7b7f-b595-46babe44f7bd', 'ca-ES', 'Permet a l''usuari curar qualsevol ésser viu mitjançant les seves llàgrimes o aigua curativa que li surt dels palmells.', ARRAY['Llàgrimes curatives: Les llàgrimes que toquen un ésser viu el retornen a la salut plena i poden tornar a unir parts del cos, tot i que no poden crear teixit del no-res.','Aigua curativa: L''usuari també pot produir aigua amb les mateixes propietats restauradores des dels palmells, i es pot curar a si mateix.','Restauració d''objectes: L''usuari pot reparar objectes inanimats danyats, tot i que en fer-ho s''escurça la seva pròpia esperança de vida.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd1-7b7f-b595-46babe44f7bd')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Maki Maki no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd2-794d-b084-79c6c8e9bbf8', 'DEVIL_FRUIT', 'Maki Maki no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8', 'es-ES', 'Permite al usuario invocar y controlar telequinéticamente pergaminos que pueden cambiar de tamaño y guardar objetos en su interior.', ARRAY['Pergaminos gigantes: El usuario invoca pergaminos que se despliegan casi al instante y pueden crecer lo bastante como para empequeñecer a las personas o envolver a un enorme dragón.','Sellado de ataques: Un pergamino puede capturar el ataque de un rival, incluso ataques de energía como un aliento de fuego, y liberarlo cuando se quiera para volverlo contra él.','Dibujos vivientes: El usuario puede hacer que los dibujos de los pergaminos cobren forma como objetos realistas, como clones de sí mismo que confunden a los rivales.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8', 'en-GB', 'Allows the user to summon and telekinetically control scrolls that can change size and store objects inside them.', ARRAY['Giant Scrolls: The user summons scrolls that unroll almost instantly and can grow big enough to dwarf people or wrap around a huge dragon.','Attack Sealing: A scroll can capture an opponent''s attack, even energy attacks like fire breath, and release it on command to turn it against them.','Living Drawings: The user can manifest drawings from scrolls as lifelike objects, such as clones of themselves that confuse opponents.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8', 'ca-ES', 'Permet a l''usuari invocar i controlar telecinèticament pergamins que poden canviar de mida i guardar objectes a l''interior.', ARRAY['Pergamins gegants: L''usuari invoca pergamins que es despleguen gairebé a l''instant i poden créixer prou per empetitir les persones o embolcallar un enorme drac.','Segellat d''atacs: Un pergamí pot capturar l''atac d''un rival, fins i tot atacs d''energia com un alè de foc, i alliberar-lo quan es vulgui per tornar-lo contra ell.','Dibuixos vivents: L''usuari pot fer que els dibuixos dels pergamins prenguin forma com a objectes realistes, com clons d''ell mateix que confonen els rivals.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd2-794d-b084-79c6c8e9bbf8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Soru Soru no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd3-780c-bd70-2222155a8d47', 'DEVIL_FRUIT', 'Soru Soru no mi', 'MYTHICAL')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd3-780c-bd70-2222155a8d47', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd3-780c-bd70-2222155a8d47')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd3-780c-bd70-2222155a8d47', 'es-ES', 'Permite al usuario extraer las almas de quienes le temen y usarlas para robar años de vida o dar vida a objetos.', ARRAY['Extracción de almas: El usuario arranca el alma de quienes le temen y les roba tanta vida como decida, desde unos segundos hasta toda la que les quede.','Homies: Al infundir fragmentos de alma en objetos, plantas o animales, el usuario les da inteligencia, habla y un rostro humano, creando subordinados obedientes.','Infusión de almas: Las almas pueden infundirse incluso en sustancias creadas por otras frutas del diablo, dándoles vida y a veces permitiendo al usuario controlarlas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd3-780c-bd70-2222155a8d47')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd3-780c-bd70-2222155a8d47', 'en-GB', 'Allows the user to extract the souls of people who fear them and use them to steal lifespan or give life to objects.', ARRAY['Soul Extraction: The user pulls the souls out of those who fear them, stealing as much of their lifespan as they choose, from seconds to their entire remaining life.','Homies: By infusing soul fragments into objects, plants or animals, the user gives them intelligence, speech and a human face, creating obedient subordinates.','Soul Infusion: Souls can even be infused into substances made by other Devil Fruits, giving them life and sometimes letting the user control them.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd3-780c-bd70-2222155a8d47')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd3-780c-bd70-2222155a8d47', 'ca-ES', 'Permet a l''usuari extreure les ànimes de qui el temen i fer-les servir per robar anys de vida o donar vida a objectes.', ARRAY['Extracció d''ànimes: L''usuari arrenca l''ànima de qui el tem i els roba tanta vida com decideixi, des d''uns segons fins a tota la que els quedi.','Homies: En infondre fragments d''ànima en objectes, plantes o animals, l''usuari els dona intel·ligència, parla i un rostre humà, creant subordinats obedients.','Infusió d''ànimes: Les ànimes es poden infondre fins i tot en substàncies creades per altres fruits del diable, donant-los vida i de vegades permetent a l''usuari controlar-les.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd3-780c-bd70-2222155a8d47')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Mira Mira no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd4-713b-913e-2c16f1e0970c', 'DEVIL_FRUIT', 'Mira Mira no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd4-713b-913e-2c16f1e0970c', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd4-713b-913e-2c16f1e0970c')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd4-713b-913e-2c16f1e0970c', 'es-ES', 'Permite al usuario crear espejos que reflejan ataques, abren portales a una dimensión espejo y convierten al usuario en el reflejo de otros.', ARRAY['Espejos reflectantes: El usuario crea espejos con las manos que absorben los ataques enemigos y se los devuelven al atacante.','Mundo Espejo: Los espejos funcionan como portales a una dimensión de bolsillo donde se puede atrapar a los rivales, y el usuario puede viajar por ella hasta otros espejos y barcos a voluntad.','Copia especular: El usuario puede convertirse en el reflejo de otra persona, imitar sus acciones y obligarla a repetir los movimientos del propio usuario.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd4-713b-913e-2c16f1e0970c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd4-713b-913e-2c16f1e0970c', 'en-GB', 'Allows the user to create mirrors that reflect attacks, open portals to a mirror dimension, and turn the user into reflections of others.', ARRAY['Reflecting Mirrors: The user creates mirrors with their hands that absorb enemy attacks and reflect them back at the attacker.','Mirro-World: Mirrors serve as portals to a pocket dimension where opponents can be trapped, and the user can travel through it to other mirrors and ships at will.','Mirror Copy: The user can become the reflection of another person, imitate their actions and force them to repeat the user''s own movements.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd4-713b-913e-2c16f1e0970c')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd4-713b-913e-2c16f1e0970c', 'ca-ES', 'Permet a l''usuari crear miralls que reflecteixen atacs, obren portals a una dimensió mirall i converteixen l''usuari en el reflex d''altres.', ARRAY['Miralls reflectors: L''usuari crea miralls amb les mans que absorbeixen els atacs enemics i els tornen a l''atacant.','Món Mirall: Els miralls fan de portals a una dimensió de butxaca on es pot atrapar els rivals, i l''usuari hi pot viatjar fins a altres miralls i vaixells a voluntat.','Còpia especular: L''usuari pot convertir-se en el reflex d''una altra persona, imitar-ne les accions i obligar-la a repetir els moviments de l''usuari.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd4-713b-913e-2c16f1e0970c')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Pero Pero no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd5-77bf-bdcc-01ad7d26383b', 'DEVIL_FRUIT', 'Pero Pero no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b', 'es-ES', 'Permite al usuario crear y controlar cantidades casi ilimitadas de caramelo comestible. El caramelo puede endurecerse hasta resistir el fuego de cañón.', ARRAY['Caramelo Ilimitado: El usuario crea y moldea enormes cantidades de caramelo, incluso con movimiento propio, como escaleras mecánicas o una prótesis para un miembro perdido.','Caramelo Endurecido: Al endurecerse, el caramelo resiste cañonazos y puñetazos potentes y sirve para inmovilizar a rivales o armar trampas y armas.','Conversión en Caramelo: El usuario cubre a un oponente con caramelo que se filtra poco a poco en su cuerpo hasta convertirlo en caramelo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b', 'en-GB', 'Allows the user to create and control seemingly limitless amounts of edible candy. The candy can be hardened until it withstands cannon fire.', ARRAY['Limitless Candy: The user creates and shapes huge amounts of candy, even making it move on its own, such as escalators or a prosthetic to replace a lost limb.','Hardened Candy: When hardened, the candy withstands cannon fire and powerful punches, and is used to pin down enemies or form weapons and traps.','Candy Conversion: The user coats an opponent in candy that slowly seeps into their body until it turns them into candy.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b', 'ca-ES', 'Permet a l''usuari crear i controlar quantitats pràcticament il·limitades de caramel comestible. El caramel es pot endurir fins a resistir les canonades.', ARRAY['Caramel Il·limitat: L''usuari crea i modela enormes quantitats de caramel, fins i tot amb moviment propi, com escales mecàniques o una pròtesi per a un membre perdut.','Caramel Endurit: En endurir-se, el caramel resisteix canonades i cops potents i serveix per immobilitzar rivals o crear armes i trampes.','Conversió en Caramel: L''usuari cobreix un oponent amb caramel que s''infiltra a poc a poc en el seu cos fins a convertir-lo en caramel.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd5-77bf-bdcc-01ad7d26383b')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Bisu Bisu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd6-7b29-b547-90dcb056d3a0', 'DEVIL_FRUIT', 'Bisu Bisu no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd6-7b29-b547-90dcb056d3a0', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd6-7b29-b547-90dcb056d3a0')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd6-7b29-b547-90dcb056d3a0', 'es-ES', 'Permite al usuario conjurar cantidades casi ilimitadas de galleta al dar una palmada. Con ellas modela figuras, armas y armaduras a voluntad.', ARRAY['Galletas Ilimitadas: El usuario conjura galletas al aplaudir y las convierte en migas finas que condensa y moldea con la forma que desee.','Marionetas de Galleta: El usuario forma figuras humanoides móviles de aspecto muy realista, las controla desde dentro y puede añadirles extremidades y armas.','Armadura de Galleta: Las galletas son muy duras, como el bizcocho de barco, y reforzadas con Haki pueden servir de escudo o armadura, aunque se debilitan con agua.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd6-7b29-b547-90dcb056d3a0')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd6-7b29-b547-90dcb056d3a0', 'en-GB', 'Allows the user to conjure seemingly limitless amounts of biscuit by clapping their hands. The biscuit can be shaped into figures, weapons and armour at will.', ARRAY['Limitless Biscuits: The user conjures biscuits by clapping and crushes them into fine crumbs that can be condensed and moulded into any shape they wish.','Biscuit Puppets: The user forms lifelike, moving humanoid figures, controls them from within and can add extra limbs and weapons to them.','Biscuit Armour: The biscuits are as hard as hardtack and, reinforced with Haki, can act as a tough shield or armour, though water weakens them.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd6-7b29-b547-90dcb056d3a0')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd6-7b29-b547-90dcb056d3a0', 'ca-ES', 'Permet a l''usuari conjurar quantitats pràcticament il·limitades de galeta en picar de mans. Amb elles modela figures, armes i armadures a voluntat.', ARRAY['Galetes Il·limitades: L''usuari conjura galetes en aplaudir i les converteix en molles fines que condensa i modela amb la forma que vulgui.','Titelles de Galeta: L''usuari forma figures humanoides mòbils d''aspecte molt realista, les controla des de dins i els pot afegir extremitats i armes.','Armadura de Galeta: Les galetes són molt dures, com el pa de vaixell, i reforçades amb Haki poden servir d''escut o armadura, tot i que l''aigua les debilita.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd6-7b29-b547-90dcb056d3a0')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Bata Bata no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd7-7a73-aebd-117251c7dea6', 'DEVIL_FRUIT', 'Bata Bata no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd7-7a73-aebd-117251c7dea6', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd7-7a73-aebd-117251c7dea6')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd7-7a73-aebd-117251c7dea6', 'es-ES', 'Permite al usuario crear y controlar cantidades casi ilimitadas de mantequilla. Esta es pegajosa y sirve para atrapar a los enemigos.', ARRAY['Mantequilla Ilimitada: El usuario genera enormes cantidades de mantequilla y las mueve a distancia a su antojo.','Mantequilla Adhesiva: En estado semilíquido la mantequilla es pegajosa, no resbaladiza, y se enrolla en manos y cuerpos para inmovilizar a los oponentes.','Apoyo en Combate: Al dejar a los rivales atrapados e indefensos, el usuario permite que sus aliados los dominen con facilidad.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd7-7a73-aebd-117251c7dea6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd7-7a73-aebd-117251c7dea6', 'en-GB', 'Allows the user to create and control seemingly limitless amounts of butter. The butter is sticky and is used to trap enemies.', ARRAY['Limitless Butter: The user generates huge amounts of butter and moves it from a distance at will.','Adhesive Butter: In its semi-liquid state the butter is sticky rather than slippery, and wraps around hands and bodies to immobilise opponents.','Combat Support: By leaving foes trapped and defenceless, the user lets their allies deal with them easily.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd7-7a73-aebd-117251c7dea6')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd7-7a73-aebd-117251c7dea6', 'ca-ES', 'Permet a l''usuari crear i controlar quantitats pràcticament il·limitades de mantega. Aquesta és enganxosa i serveix per atrapar els enemics.', ARRAY['Mantega Il·limitada: L''usuari genera enormes quantitats de mantega i les mou a distància al seu gust.','Mantega Adhesiva: En estat semilíquid la mantega és enganxosa, no relliscosa, i s''enrotlla a mans i cossos per immobilitzar els oponents.','Suport en Combat: En deixar els rivals atrapats i indefensos, l''usuari permet que els seus aliats els dominin amb facilitat.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd7-7a73-aebd-117251c7dea6')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Buku Buku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd8-7789-80b7-847c410dee85', 'DEVIL_FRUIT', 'Buku Buku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd8-7789-80b7-847c410dee85', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd8-7789-80b7-847c410dee85')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd8-7789-80b7-847c410dee85', 'es-ES', 'Permite al usuario controlar libros cercanos y atrapar a seres vivos dentro de su ambientación. Los mundos del interior los define su imaginación.', ARRAY['Control de Libros: El usuario mueve libros de cualquier tamaño de forma telequinética y los usa incluso como puntos de apoyo para desplazarse por el aire.','Ilusión del Libro: Al sostener un libro abierto sobre un objetivo, este cree estar dentro del mundo del libro; la ilusión desaparece al cerrarlo.','Prisión de Papel: El usuario encierra a seres vivos dentro de un libro convirtiéndolos en ilustración, sin envejecer, y también puede conectar varios Den Den Mushi a un libro.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd8-7789-80b7-847c410dee85')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd8-7789-80b7-847c410dee85', 'en-GB', 'Allows the user to control nearby books and trap living beings inside their settings. The worlds inside are shaped by the user''s imagination.', ARRAY['Book Control: The user telekinetically moves books of any size and can even use them as footholds to travel through the air.','Book Illusion: Holding an open book over a target makes them believe they are inside the book''s world, an illusion that ends when the book is closed.','Paper Prison: The user traps living beings inside a book as illustrations, where they do not age, and can also link several Den Den Mushi to a book.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd8-7789-80b7-847c410dee85')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd8-7789-80b7-847c410dee85', 'ca-ES', 'Permet a l''usuari controlar llibres propers i atrapar éssers vius dins de la seva ambientació. Els mons de l''interior els defineix la seva imaginació.', ARRAY['Control de Llibres: L''usuari mou llibres de qualsevol mida de manera telecinètica i fins i tot els fa servir com a punts de suport per desplaçar-se per l''aire.','Il·lusió del Llibre: En sostenir un llibre obert sobre un objectiu, aquest creu que és dins del món del llibre; la il·lusió desapareix en tancar-lo.','Presó de Paper: L''usuari tanca éssers vius dins d''un llibre convertits en il·lustració, sense envellir, i també pot connectar diversos Den Den Mushi a un llibre.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd8-7789-80b7-847c410dee85')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kuri Kuri no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdd9-7aef-9ae8-c56853e3b28e', 'DEVIL_FRUIT', 'Kuri Kuri no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e', 'es-ES', 'Permite al usuario crear y controlar cantidades casi ilimitadas de nata. El azúcar de la nata causa quemaduras dolorosas a quien la toca.', ARRAY['Nata Ilimitada: El usuario genera enormes cantidades de nata y las controla para atacar y atrapar a sus enemigos.','Nata Abrasiva: El azúcar de la nata provoca quemaduras de segundo o tercer grado a cualquiera que entre en contacto con ella.','Cream Monster: El usuario forma con la nata varios tentáculos, con aspecto de kraken, y ataca con ellos quemando a sus rivales.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e', 'en-GB', 'Allows the user to create and control seemingly limitless amounts of cream. The sugar in the cream inflicts painful burns on anything it touches.', ARRAY['Limitless Cream: The user generates huge amounts of cream and controls it to attack and capture their enemies.','Scalding Cream: The sugar in the cream causes second or third degree burns to anyone who comes into contact with it.','Cream Monster: The user forms several tentacles out of the cream, giving them the look of a kraken, and attacks with them to burn their foes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e', 'ca-ES', 'Permet a l''usuari crear i controlar quantitats pràcticament il·limitades de nata. El sucre de la nata causa cremades doloroses a qui la toca.', ARRAY['Nata Il·limitada: L''usuari genera enormes quantitats de nata i les controla per atacar i atrapar els seus enemics.','Nata Abrasiva: El sucre de la nata provoca cremades de segon o tercer grau a qualsevol que hi entri en contacte.','Cream Monster: L''usuari forma amb la nata diversos tentacles, amb aspecte de kraken, i ataca amb ells cremant els seus rivals.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdd9-7aef-9ae8-c56853e3b28e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Shibo Shibo no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdda-7766-8d1c-dc76cb6f0c76', 'DEVIL_FRUIT', 'Shibo Shibo no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76', 'es-ES', 'Permite al usuario extraer el líquido de seres vivos y objetos al estrujarlos con las manos. Los líquidos obtenidos pueden beberse como zumos.', ARRAY['Estrujado: Al retorcer a un objetivo con las manos, el usuario le extrae todo el líquido sin dañarlo directamente, aunque a los seres vivos los deja deshidratados y secos.','Zumos Variados: Los líquidos extraídos se convierten en bebidas de sabores deliciosos según su origen, incluso de cosas peligrosas como una roca volcánica.','Absorción y Descarga: El usuario puede absorber líquido para crecer y pesar más, soltarlo en chorros a presión, expulsar venenos de su cuerpo o deshidratar con su espada.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76', 'en-GB', 'Allows the user to extract the liquid from living beings and objects by wringing them with their hands. The liquids obtained can be drunk as juices.', ARRAY['Wringing: By twisting a target in their hands, the user extracts all its liquid without directly harming it, though living beings are left dehydrated and shrivelled.','Assorted Juices: The extracted liquids become drinks with delicious flavours depending on their source, even from dangerous things such as volcanic rock.','Absorb and Release: The user can absorb liquid to grow larger and heavier, release it in pressurised blasts, wring poison out of their body, or dehydrate enemies with their sword.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76', 'ca-ES', 'Permet a l''usuari extreure el líquid d''éssers vius i objectes en escórrer-los amb les mans. Els líquids obtinguts es poden beure com a sucs.', ARRAY['Escorrer: En retorçar un objectiu amb les mans, l''usuari n''extreu tot el líquid sense fer-li dany directe, tot i que els éssers vius queden deshidratats i eixuts.','Sucs Variats: Els líquids extrets es converteixen en begudes de sabors deliciosos segons l''origen, fins i tot de coses perilloses com una roca volcànica.','Absorció i Descàrrega: L''usuari pot absorbir líquid per créixer i pesar més, alliberar-lo en raigs a pressió, expulsar verins del seu cos o deshidratar amb l''espasa.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdda-7766-8d1c-dc76cb6f0c76')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Memo Memo no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bddb-7c53-9b18-ca7b52698aee', 'DEVIL_FRUIT', 'Memo Memo no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bddb-7c53-9b18-ca7b52698aee', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddb-7c53-9b18-ca7b52698aee')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddb-7c53-9b18-ca7b52698aee', 'es-ES', 'Permite al usuario extraer los recuerdos de una persona en forma de tiras de película y manipularlos. Así puede borrar recuerdos o implantar otros.', ARRAY['Extracción de Recuerdos: El usuario introduce la mano en la cabeza del objetivo y saca sus recuerdos como tiras de película, lo que puede dolerle y hacerle perder el conocimiento.','Edición de Recuerdos: Con unas tijeras y pegamento, el usuario corta fotogramas para borrar un recuerdo y pega otros para alterar lo que el objetivo cree haber vivido.','Archivo de Recuerdos: Los fotogramas extraídos pueden conservarse mucho tiempo y reutilizarse para crear recuerdos falsos en otras personas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddb-7c53-9b18-ca7b52698aee')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddb-7c53-9b18-ca7b52698aee', 'en-GB', 'Allows the user to extract a person''s memories as strips of film and manipulate them. They can erase memories or implant new ones.', ARRAY['Memory Extraction: The user reaches into the target''s head and pulls out their memories as film strips, which can be painful and cause the target to faint.','Memory Editing: With scissors and paste, the user cuts out frames to erase a memory and pastes in others to alter what the target believes happened.','Memory Archive: Extracted frames can be stored for long periods and reused to create false memories in other people.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddb-7c53-9b18-ca7b52698aee')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddb-7c53-9b18-ca7b52698aee', 'ca-ES', 'Permet a l''usuari extreure els records d''una persona en forma de tires de pel·lícula i manipular-los. Així pot esborrar records o implantar-ne d''altres.', ARRAY['Extracció de Records: L''usuari introdueix la mà al cap de l''objectiu i en treu els records com a tires de pel·lícula, cosa que pot fer-li mal i deixar-lo inconscient.','Edició de Records: Amb unes tisores i cola, l''usuari talla fotogrames per esborrar un record i en enganxa d''altres per alterar el que l''objectiu creu haver viscut.','Arxiu de Records: Els fotogrames extrets es poden conservar molt de temps i reutilitzar per crear records falsos en altres persones.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddb-7c53-9b18-ca7b52698aee')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Hoya Hoya no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bddc-70c2-962d-2abb797ee2eb', 'DEVIL_FRUIT', 'Hoya Hoya no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bddc-70c2-962d-2abb797ee2eb', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddc-70c2-962d-2abb797ee2eb')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddc-70c2-962d-2abb797ee2eb', 'es-ES', 'Permite al usuario invocar un genio desde su propio cuerpo frotándolo. El genio lucha por él y es capaz de enfrentarse a combatientes de élite.', ARRAY['Invocación del Genio: Al frotarse el cuerpo, el usuario hace surgir de su abdomen un enorme genio musculoso, con sentido propio y capacidad de hablar.','Genio Combatiente: Armado con un bisento, el genio puede igualar o superar a un maestro de artes marciales y cortar por la mitad varios barcos de un solo golpe.','Tamaño Gigante: El genio puede crecer hasta un tamaño colosal, pero solo se mantiene mientras el usuario siga frotándose y no se aleje demasiado de él.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddc-70c2-962d-2abb797ee2eb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddc-70c2-962d-2abb797ee2eb', 'en-GB', 'Allows the user to summon a genie from their own body by rubbing it. The genie fights for them and can match elite combatants.', ARRAY['Genie Summoning: By rubbing their body, the user makes a huge, muscular genie emerge from their abdomen, one that is sentient and able to speak.','Genie Combatant: Wielding a bisento, the genie can match or overpower a martial arts master and cut several ships in half with a single strike.','Giant Size: The genie can grow to a colossal size, but it lasts only while the user keeps rubbing their body and stays within a limited distance.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddc-70c2-962d-2abb797ee2eb')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddc-70c2-962d-2abb797ee2eb', 'ca-ES', 'Permet a l''usuari invocar un geni des del seu propi cos fregant-lo. El geni lluita per ell i és capaç d''enfrontar-se a combatents d''elit.', ARRAY['Invocació del Geni: En fregar-se el cos, l''usuari fa sorgir del seu abdomen un enorme geni musculós, amb sentit propi i capacitat de parlar.','Geni Combatent: Armat amb un bisento, el geni pot igualar o superar un mestre d''arts marcials i tallar per la meitat diversos vaixells d''un sol cop.','Mida Gegant: El geni pot créixer fins a una mida colossal, però només es manté mentre l''usuari continua fregant-se i no s''allunya massa d''ell.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddc-70c2-962d-2abb797ee2eb')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Netsu Netsu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bddd-74b6-9ae6-4f018dcea529', 'DEVIL_FRUIT', 'Netsu Netsu no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bddd-74b6-9ae6-4f018dcea529', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddd-74b6-9ae6-4f018dcea529')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddd-74b6-9ae6-4f018dcea529', 'es-ES', 'Permite al usuario emanar calor desde su cuerpo. Puede calentar partes de sí mismo, objetos o incluso el agua del mar.', ARRAY['Cuerpo Ardiente: El usuario calienta partes de su cuerpo a temperaturas que casi nadie soporta, lo que aumenta el daño de sus golpes o de un simple contacto.','Transferencia de Calor: El usuario pasa calor a armas y objetos para potenciar las suyas, o para volver las de sus rivales demasiado calientes de sostener, incluso sin tocarlas.','Nekkai Jigoku: El usuario calienta sus manos y las sumerge en el océano, haciendo hervir el agua en un radio enorme y dañando a quienes naveguen o naden en ella.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddd-74b6-9ae6-4f018dcea529')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddd-74b6-9ae6-4f018dcea529', 'en-GB', 'Allows the user to emanate heat from their body. They can heat parts of themselves, objects or even the sea.', ARRAY['Burning Body: The user heats parts of their body to temperatures few can withstand, increasing the damage of their blows or even a simple touch.','Heat Transfer: The user passes heat into weapons and objects to boost their own, or to make an opponent''s too hot to hold, even without touching them.','Nekkai Jigoku: The user heats their hands and plunges them into the ocean, boiling the water across an enormous radius and harming anyone sailing or swimming in it.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddd-74b6-9ae6-4f018dcea529')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddd-74b6-9ae6-4f018dcea529', 'ca-ES', 'Permet a l''usuari emanar calor des del seu cos. Pot escalfar parts de si mateix, objectes o fins i tot l''aigua del mar.', ARRAY['Cos Ardent: L''usuari escalfa parts del seu cos a temperatures que gairebé ningú suporta, cosa que augmenta el dany dels seus cops o d''un simple contacte.','Transferència de Calor: L''usuari passa calor a armes i objectes per potenciar les seves, o per fer que les dels rivals siguin massa calentes per sostenir-les, fins i tot sense tocar-les.','Nekkai Jigoku: L''usuari escalfa les mans i les submergeix a l''oceà, fent bullir l''aigua en un radi enorme i fent mal a qui hi navegui o hi nedi.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddd-74b6-9ae6-4f018dcea529')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kuku Kuku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdde-7e2e-9528-ca00302371b3', 'DEVIL_FRUIT', 'Kuku Kuku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdde-7e2e-9528-ca00302371b3', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdde-7e2e-9528-ca00302371b3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdde-7e2e-9528-ca00302371b3', 'es-ES', 'Permite al usuario convertir objetos en comida. No necesita cocinar: basta con transformar el objeto.', ARRAY['Objetos en Comida: El usuario transforma cualquier objeto en alimento al instante, sin necesidad de cocinarlo, para alimentarse a sí mismo y a otros.','Veränderung Gourmet: El usuario clava su espada en un objeto y lo transforma en comida, incluso uno tan grande como un castillo entero.','Neutralizar Peligros: Convertir un objeto en comida puede hacerlo menos peligroso o más útil, por ejemplo al transformar escombros que caen en un pastel esponjoso.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdde-7e2e-9528-ca00302371b3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdde-7e2e-9528-ca00302371b3', 'en-GB', 'Allows the user to turn objects into food. There is no need to cook: transforming the object is enough.', ARRAY['Objects into Food: The user instantly transforms any object into food, with no cooking required, to feed themselves and others.','Veränderung Gourmet: The user stabs their sword into an object and transforms it into food, even one as large as an entire castle.','Neutralising Hazards: Turning an object into food can make it less dangerous or more useful, such as turning falling debris into a soft sponge cake.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdde-7e2e-9528-ca00302371b3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdde-7e2e-9528-ca00302371b3', 'ca-ES', 'Permet a l''usuari convertir objectes en menjar. No cal cuinar: n''hi ha prou amb transformar l''objecte.', ARRAY['Objectes en Menjar: L''usuari transforma qualsevol objecte en aliment a l''instant, sense necessitat de cuinar-lo, per alimentar-se ell i els altres.','Veränderung Gourmet: L''usuari clava l''espasa en un objecte i el transforma en menjar, fins i tot un de tan gran com un castell sencer.','Neutralitzar Perills: Convertir un objecte en menjar pot fer-lo menys perillós o més útil, per exemple en transformar runes que cauen en un pastís esponjós.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdde-7e2e-9528-ca00302371b3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Gocha Gocha no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bddf-7609-81ba-66dde1d6d419', 'DEVIL_FRUIT', 'Gocha Gocha no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bddf-7609-81ba-66dde1d6d419', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddf-7609-81ba-66dde1d6d419')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddf-7609-81ba-66dde1d6d419', 'es-ES', 'Permite al usuario fusionarse con otras personas. La forma fusionada es mucho más grande y fuerte que cada individuo.', ARRAY['Fusión de Personas: El usuario se mezcla con otras personas en un único ser de tamaño y fuerza muy superiores a los de cada uno por separado.','Fusión con Objetos: Los objetos inorgánicos también entran en la mezcla, de modo que las armas y la ropa de los fusionados se combinan en una versión gigante.','Fuerza Colectiva: La potencia de la fusión depende de la fuerza de cada integrante, por lo que resulta más eficaz con muchos aliados cerca.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddf-7609-81ba-66dde1d6d419')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddf-7609-81ba-66dde1d6d419', 'en-GB', 'Allows the user to merge themselves with other people. The merged form is far larger and stronger than any individual.', ARRAY['People Fusion: The user mixes with other people into a single being of far greater size and strength than any of them alone.','Object Fusion: Inorganic objects can join the fusion too, so the weapons and clothing of those merged combine into a giant version.','Collective Strength: The fusion''s power depends on the strength of each member, so it works best with many allies nearby.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddf-7609-81ba-66dde1d6d419')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bddf-7609-81ba-66dde1d6d419', 'ca-ES', 'Permet a l''usuari fusionar-se amb altres persones. La forma fusionada és molt més gran i forta que cada individu.', ARRAY['Fusió de Persones: L''usuari es barreja amb altres persones en un únic ésser de mida i força molt superiors a les de cadascú per separat.','Fusió amb Objectes: Els objectes inorgànics també entren en la barreja, de manera que les armes i la roba dels fusionats es combinen en una versió gegant.','Força Col·lectiva: La potència de la fusió depèn de la força de cada integrant, per la qual cosa és més eficaç amb molts aliats a prop.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bddf-7609-81ba-66dde1d6d419')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Oshi Oshi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8', 'DEVIL_FRUIT', 'Oshi Oshi no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8', 'es-ES', 'Permite al usuario mover y moldear el suelo como si fuera arcilla con solo empujarlo. Funciona sobre tierra, roca e incluso ladrillos o calles.', ARRAY['Moldeado del Suelo: Al empujar el suelo, el usuario le da la forma que quiera, ya sea tierra blanda, roca dura o piedra trabajada como ladrillos y calles.','Ataque Terrestre: El usuario altera el terreno bajo el enemigo, por ejemplo volcando un bloque entero de calle sobre él, o se agarra al suelo clavando un arma.','Túneles Seguros: El usuario excava túneles empujando la tierra, que se queda en su sitio y evita derrumbes, lo que sirve para infiltrarse o crear cámaras ocultas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8', 'en-GB', 'Allows the user to move and shape the ground as though it were clay, simply by pushing it. It works on earth, rock and even bricks or streets.', ARRAY['Ground Shaping: By pushing the ground, the user shapes it as they please, whether soft earth, hard rock or worked stone such as bricks and streets.','Terrain Attack: The user alters the ground beneath an enemy, such as flipping an entire chunk of street onto them, or anchors themselves by jabbing a weapon into it.','Safe Tunnelling: The user digs tunnels by pushing the earth, which stays in place and prevents cave-ins, useful for infiltration or creating hidden chambers.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8', 'ca-ES', 'Permet a l''usuari moure i modelar el terra com si fos argila només empenyent-lo. Funciona sobre terra, roca i fins i tot maons o carrers.', ARRAY['Modelatge del Terra: En empènyer el terra, l''usuari li dona la forma que vulgui, ja sigui terra tova, roca dura o pedra treballada com maons i carrers.','Atac Terrestre: L''usuari altera el terreny sota l''enemic, per exemple bolcant-li un bloc sencer de carrer al damunt, o s''agafa al terra clavant-hi una arma.','Túnels Segurs: L''usuari excava túnels empenyent la terra, que es queda al seu lloc i evita esfondraments, cosa útil per infiltrar-se o crear cambres amagades.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde0-757a-b8b5-2fd35e2d0ff8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kobu Kobu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde1-7c17-9e01-f4e4663c385e', 'DEVIL_FRUIT', 'Kobu Kobu no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde1-7c17-9e01-f4e4663c385e', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde1-7c17-9e01-f4e4663c385e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde1-7c17-9e01-f4e4663c385e', 'es-ES', 'Permite al usuario despertar la fuerza latente de otras personas animándolas. Sus palabras aumentan su espíritu de lucha y su fuerza física.', ARRAY['Palabras de Ánimo: Al animar a otros con sus palabras, el usuario eleva su espíritu de lucha y su fuerza física.','Refuerzo Colectivo: El usuario puede animar a varias personas a la vez y convertir a un grupo de gente corriente en luchadores capaces de abrumar a una tripulación pirata con simples palos.','Voluntad Ajena: El efecto depende de que los animados quieran luchar, pues el usuario solo puede inspirarlos y son ellos quienes deciden seguir adelante.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde1-7c17-9e01-f4e4663c385e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde1-7c17-9e01-f4e4663c385e', 'en-GB', 'Allows the user to awaken the latent strength of others by encouraging them. Their words raise others'' fighting spirit and physical strength.', ARRAY['Words of Encouragement: By encouraging others with their words, the user raises their fighting spirit and physical strength.','Group Empowerment: The user can encourage several people at once, turning ordinary people into fighters able to overwhelm a pirate crew armed with just sticks.','Willing Allies: The effect depends on those encouraged being willing to fight, as the user can only inspire them and they must choose to carry on.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde1-7c17-9e01-f4e4663c385e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde1-7c17-9e01-f4e4663c385e', 'ca-ES', 'Permet a l''usuari despertar la força latent dels altres animant-los. Les seves paraules augmenten el seu esperit de lluita i la seva força física.', ARRAY['Paraules d''Ànim: En animar els altres amb les seves paraules, l''usuari eleva el seu esperit de lluita i la seva força física.','Reforç Col·lectiu: L''usuari pot animar diverses persones alhora i convertir un grup de gent corrent en lluitadors capaços d''aclaparar una tripulació pirata amb simples bastons.','Voluntat Aliena: L''efecte depèn que els animats vulguin lluitar, ja que l''usuari només pot inspirar-los i són ells qui decideixen continuar.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde1-7c17-9e01-f4e4663c385e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Kibi Kibi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde2-792a-bad5-d393c498ae54', 'DEVIL_FRUIT', 'Kibi Kibi no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde2-792a-bad5-d393c498ae54', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde2-792a-bad5-d393c498ae54')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde2-792a-bad5-d393c498ae54', 'es-ES', 'Permite al usuario crear dangos que vuelven completamente leal a cualquier animal o portador de SMILE que se los coma.', ARRAY['Dango de lealtad: Pellizcándose las mejillas y formando un círculo con los dedos, el usuario crea pequeños dangos que solo él puede producir y que no se le pueden arrebatar.','Obediencia absoluta: Quien come un dango conserva sus recuerdos, pero obedece las órdenes verbales del usuario y lo defiende, anteponiendo su vida a la propia.','Vínculo de un mes: El efecto dura un mes, incluso si el usuario pierde el conocimiento, y funciona con animales y comedores de SMILE, pero no con usuarios Zoan naturales.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde2-792a-bad5-d393c498ae54')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde2-792a-bad5-d393c498ae54', 'en-GB', 'Allows the user to create dango dumplings that make any animal or Gifter who eats them utterly loyal to the user.', ARRAY['Loyalty Dango: By pinching their cheeks and forming a circle with their fingers, the user creates small dumplings that only they can produce and that cannot be taken from them.','Absolute Obedience: Whoever eats a dango keeps their memories but obeys the user''s verbal orders and defends them, placing the user''s life above their own.','Month-Long Bond: The taming lasts one month, even if the user is unconscious, and works on animals and SMILE eaters, but not on natural Zoan users.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde2-792a-bad5-d393c498ae54')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde2-792a-bad5-d393c498ae54', 'ca-ES', 'Permet a l''usuari crear dangos que fan lleial del tot qualsevol animal o portador de SMILE que se''ls mengi.', ARRAY['Dango de lleialtat: Pessigant-se les galtes i formant un cercle amb els dits, l''usuari crea petits dangos que només ell pot produir i que no se li poden prendre.','Obediència absoluta: Qui menja un dango conserva els records, però obeeix les ordres verbals de l''usuari i el defensa, posant la vida d''ell per damunt de la seva.','Vincle d''un mes: L''efecte dura un mes, fins i tot si l''usuari perd el coneixement, i funciona amb animals i menjadors de SMILE, però no amb usuaris Zoan naturals.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde2-792a-bad5-d393c498ae54')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Toki Toki no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde3-7403-b0a6-8a5ab65c4bab', 'DEVIL_FRUIT', 'Toki Toki no mi', 'EPIC')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab', 'es-ES', 'Permite al usuario enviarse a sí mismo o a otros hacia el futuro, hasta veinte años de una vez.', ARRAY['Salto al futuro: El usuario puede desaparecer al instante del presente y reaparecer en un momento elegido del futuro, útil para escapar de situaciones peligrosas.','Lapso a elegir: La cantidad de tiempo saltado se decide en cada viaje, siendo veinte años el salto máximo conocido.','Viaje sin retorno: Es imposible viajar al pasado, así que cada salto es irreversible, y los viajeros permanecen en el mismo lugar, por lo que hay que cuidar dónde se está.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab', 'en-GB', 'Allows the user to send themselves or others forward in time, up to twenty years at once.', ARRAY['Leap Forward: The user can instantly vanish from the present and reappear at a chosen moment in the future, a handy way to escape a dangerous situation.','Chosen Span: The amount of time skipped can be decided for each trip, with twenty years being the longest known jump.','One-Way Trip: Travel into the past is impossible, so every jump is irreversible, and travellers stay in the same place, so the user must mind where they stand.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab', 'ca-ES', 'Permet a l''usuari enviar-se a si mateix o a altres cap al futur, fins a vint anys d''un cop.', ARRAY['Salt al futur: L''usuari pot desaparèixer a l''instant del present i reaparèixer en un moment triat del futur, útil per escapar de situacions perilloses.','Lapse a escollir: La quantitat de temps saltat es decideix en cada viatge, i vint anys és el salt màxim conegut.','Viatge sense retorn: És impossible viatjar al passat, així que cada salt és irreversible, i els viatgers es queden al mateix lloc, de manera que cal vigilar on es trobin.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde3-7403-b0a6-8a5ab65c4bab')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Juku Juku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde4-78e9-836b-83c16069a3b2', 'DEVIL_FRUIT', 'Juku Juku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde4-78e9-836b-83c16069a3b2', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde4-78e9-836b-83c16069a3b2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde4-78e9-836b-83c16069a3b2', 'es-ES', 'Permite al usuario envejecer y descomponer rápidamente todo lo que toque, ya sean objetos, terreno o seres vivos.', ARRAY['Descomposición rápida: Al tocar un objeto, el usuario lo hace envejecer y pudrirse al instante, debilitando muros y barreras hasta que se derrumban o pueden romperse.','Escape por el suelo: Aplicado al suelo, el toque lo convierte en un socavón por el que el usuario puede hundirse para evitar un ataque, aunque el deterioro sigue avanzando.','Envejecimiento forzado: Tocar a un ser vivo lo hace envejecer físicamente al instante, pero su mente sigue igual y el envejecimiento nunca se puede revertir.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde4-78e9-836b-83c16069a3b2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde4-78e9-836b-83c16069a3b2', 'en-GB', 'Allows the user to rapidly age and decay anything they touch, whether objects, terrain or living beings.', ARRAY['Rapid Decay: By touching an object, the user makes it age and rot at once, weakening walls and barriers until they collapse or can be broken through.','Sinkhole Escape: Used on the ground, the touch turns it into a sinkhole, letting the user sink away from an attack, though the decay keeps spreading once started.','Forced Ageing: Touching a living being makes it grow physically older in an instant, but the mind stays as it was and the ageing can never be reversed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde4-78e9-836b-83c16069a3b2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde4-78e9-836b-83c16069a3b2', 'ca-ES', 'Permet a l''usuari fer envellir i descompondre ràpidament tot allò que toqui, ja siguin objectes, terreny o éssers vius.', ARRAY['Descomposició ràpida: En tocar un objecte, l''usuari el fa envellir i podrir-se a l''acte, debilitant murs i barreres fins que s''esfondren o es poden trencar.','Escapada pel terra: Aplicat al terra, el toc el converteix en un esvoranc pel qual l''usuari pot enfonsar-se per evitar un atac, tot i que el deteriorament continua avançant.','Envelliment forçat: Tocar un ésser viu el fa envellir físicament a l''instant, però la ment es manté igual i l''envelliment mai no es pot revertir.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde4-78e9-836b-83c16069a3b2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Shiku Shiku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde5-7767-9cf0-8de84ad1bed1', 'DEVIL_FRUIT', 'Shiku Shiku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde5-7767-9cf0-8de84ad1bed1', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde5-7767-9cf0-8de84ad1bed1')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde5-7767-9cf0-8de84ad1bed1', 'es-ES', 'Permite al usuario crear enfermedades que entorpecen a los rivales y pueden propagarse por todo un grupo.', ARRAY['Enfermedades a medida: El usuario puede crear dolencias de todo tipo con efectos distintos, como una que convierte en mujer a todo hombre afectado.','Contagio masivo: Las enfermedades pueden pasar de un enemigo a otro y alcanzar un rango muy amplio, afectando a una tripulación incluso tras sumergirse muy por debajo del mar.','Detección de contrarrestos: El usuario percibe cuándo su enfermedad es contrarrestada, aunque un Haki intenso puede superarla y generar anticuerpos que curan a los demás.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde5-7767-9cf0-8de84ad1bed1')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde5-7767-9cf0-8de84ad1bed1', 'en-GB', 'Allows the user to create diseases that hinder opponents and can spread through a whole group.', ARRAY['Custom Diseases: The user can create illnesses of many kinds with varying effects, such as one that turns every man affected into a woman.','Contagious Spread: The diseases can pass between enemies and reach a very wide range, affecting a crew even after it dives far beneath the sea.','Counter Awareness: The user can sense when the illness is countered, though intense Haki can break through it and produce antibodies that cure others.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde5-7767-9cf0-8de84ad1bed1')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde5-7767-9cf0-8de84ad1bed1', 'ca-ES', 'Permet a l''usuari crear malalties que destorben els rivals i es poden propagar per tot un grup.', ARRAY['Malalties a mida: L''usuari pot crear malalties de tota mena amb efectes diferents, com una que converteix en dona tot home afectat.','Contagi massiu: Les malalties poden passar d''un enemic a un altre i assolir un abast molt ampli, afectant una tripulació fins i tot després d''enfonsar-se molt per sota del mar.','Detecció de contrarestos: L''usuari percep quan la seva malaltia és contrarestada, tot i que un Haki intens pot superar-la i generar anticossos que curen els altres.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde5-7767-9cf0-8de84ad1bed1')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Wapu Wapu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde6-7403-9a7a-b49b0436c616', 'DEVIL_FRUIT', 'Wapu Wapu no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde6-7403-9a7a-b49b0436c616', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde6-7403-9a7a-b49b0436c616')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde6-7403-9a7a-b49b0436c616', 'es-ES', 'Permite al usuario teletransportarse al instante de un lugar a otro, y teletransportar a otros tras tocarlos.', ARRAY['Salto instantáneo: El usuario se desplaza a otro lugar en un instante, lo que permite a un francotirador reposicionarse y emboscar al enemigo desde ángulos inesperados.','Teletransporte por contacto: Tras tocar a otras personas, el usuario puede llevarlas consigo, trasladando a compañeros o rescatando a uno que cae en pleno aire.','Alcance limitado: Sin dominio de la habilidad la distancia es bastante corta, y el usuario debe reaccionar a tiempo para esquivar, así que los ataques rápidos y sorpresivos pueden alcanzarlo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde6-7403-9a7a-b49b0436c616')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde6-7403-9a7a-b49b0436c616', 'en-GB', 'Allows the user to warp instantly from place to place, and to teleport others after touching them.', ARRAY['Instant Warp: The user teleports to another location in an instant, which lets a sniper reposition and ambush enemies from unexpected angles.','Touch Teleport: After touching other people, the user can warp them along, moving whole crewmates or rescuing one who is falling in mid-air.','Limited Range: Without mastery the warping distance is fairly short, and the user must react in time to dodge, so fast surprise attacks can still land.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde6-7403-9a7a-b49b0436c616')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde6-7403-9a7a-b49b0436c616', 'ca-ES', 'Permet a l''usuari teletransportar-se a l''instant d''un lloc a un altre, i teletransportar els altres després de tocar-los.', ARRAY['Salt instantani: L''usuari es desplaça a un altre lloc en un instant, cosa que permet a un franctirador reposicionar-se i emboscar l''enemic des d''angles inesperats.','Teletransport per contacte: Després de tocar altres persones, l''usuari pot endur-se-les, traslladant companys o rescatant-ne un que cau en ple aire.','Abast limitat: Sense domini de l''habilitat la distància és força curta, i l''usuari ha de reaccionar a temps per esquivar, de manera que els atacs ràpids i sorpresa poden arribar-li.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde6-7403-9a7a-b49b0436c616')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Riki Riki no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde7-719d-a35d-fd7214667d44', 'DEVIL_FRUIT', 'Riki Riki no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde7-719d-a35d-fd7214667d44', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde7-719d-a35d-fd7214667d44')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde7-719d-a35d-fd7214667d44', 'es-ES', 'Otorga al usuario una fuerza física anormal, suficiente para levantar y lanzar objetos enormes con las manos desnudas.', ARRAY['Fuerza monstruosa: El poder físico del usuario, ya notable, se potencia mucho más allá de los límites normales, sin que se conozca un tope exacto.','Levantar montañas: La fuerza basta para arrancar una montaña del suelo con total naturalidad, sostenerla sobre la cabeza y arrojarla lejos.','Poder directo: La fruta ofrece fuerza bruta sin más desventajas conocidas que las debilidades habituales de las Frutas del Diablo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde7-719d-a35d-fd7214667d44')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde7-719d-a35d-fd7214667d44', 'en-GB', 'Grants the user abnormal physical strength, enough to lift and hurl enormous objects with bare hands.', ARRAY['Monstrous Strength: The user''s already great physical power is boosted far beyond normal limits, with no exact ceiling known.','Mountain Lifting: The strength is enough to casually tear a mountain out of the ground, hold it overhead and throw it away.','Straightforward Power: The fruit offers raw power with no known drawbacks beyond the standard Devil Fruit weaknesses.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde7-719d-a35d-fd7214667d44')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde7-719d-a35d-fd7214667d44', 'ca-ES', 'Atorga a l''usuari una força física anormal, suficient per aixecar i llançar objectes enormes amb les mans nues.', ARRAY['Força monstruosa: El poder físic de l''usuari, ja notable, es potencia molt més enllà dels límits normals, sense que se''n conegui un sostre exacte.','Aixecar muntanyes: La força basta per arrencar una muntanya del terra amb tota naturalitat, sostenir-la sobre el cap i llançar-la lluny.','Poder directe: La fruita ofereix força bruta sense més desavantatges coneguts que les debilitats habituals de les Fruites del Dimoni.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde7-719d-a35d-fd7214667d44')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Nomi Nomi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde8-7eb3-a874-14f1f6ae08d3', 'DEVIL_FRUIT', 'Nomi Nomi no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3', 'es-ES', 'Permite al usuario almacenar una cantidad indefinida de conocimiento en su cerebro y dominarlo al instante.', ARRAY['Memoria ilimitada: El usuario puede guardar una cantidad indefinida de información, recordando con todo detalle todo lo que ha aprendido.','Dominio instantáneo: El conocimiento nuevo se comprende por completo al momento, incluso materias avanzadas que normalmente requieren toda una vida de estudio.','Cabeza creciente: El cerebro y la cabeza deben crecer para albergar el saber, lo que resulta incómodo, y un cerebro vivo puede conservar la información tras la muerte del usuario.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3', 'en-GB', 'Allows the user to store an indefinite amount of knowledge in their brain and master it instantly.', ARRAY['Limitless Memory: The user can store an indefinite amount of information, remembering everything they have ever learned in perfect detail.','Instant Mastery: New knowledge is understood fully at once, even advanced subjects that normally take a lifetime of study.','Swelling Head: The brain and head must grow to hold the knowledge, which becomes inconvenient, and a living brain can keep its information after the user dies.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3', 'ca-ES', 'Permet a l''usuari emmagatzemar una quantitat indefinida de coneixement al cervell i dominar-lo a l''instant.', ARRAY['Memòria il·limitada: L''usuari pot guardar una quantitat indefinida d''informació, recordant amb tot detall tot el que ha après.','Domini instantani: El coneixement nou es comprèn del tot al moment, fins i tot matèries avançades que normalment requereixen tota una vida d''estudi.','Cap creixent: El cervell i el cap han de créixer per encabir el saber, cosa que resulta incòmoda, i un cervell viu pot conservar la informació després de la mort de l''usuari.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde8-7eb3-a874-14f1f6ae08d3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Shima Shima no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bde9-7340-94fe-b03cde6293d8', 'DEVIL_FRUIT', 'Shima Shima no mi', 'RARE')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bde9-7340-94fe-b03cde6293d8', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde9-7340-94fe-b03cde6293d8')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde9-7340-94fe-b03cde6293d8', 'es-ES', 'Permite al usuario fusionarse con el entorno de cualquier isla y manipularlo libremente como si fuera su propio cuerpo.', ARRAY['Asimilación de la isla: El usuario se funde con el terreno natural o artificial de una isla que toque, como una ladera o suelos de madera, y lo controla a voluntad.','Paisaje vivo: El terreno fusionado puede hablar, mostrar expresiones o formar extremidades gigantes que interactúan con objetos, y el usuario percibe lo que la gente hace sobre él.','Daño compartido: La isla actúa como el cuerpo del usuario, así que el daño intenso que reciba lo sufre él, y los golpes fuertes o con Haki de armadura pueden herirlo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde9-7340-94fe-b03cde6293d8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde9-7340-94fe-b03cde6293d8', 'en-GB', 'Allows the user to merge with the environment of any island and freely manipulate it as if it were their own body.', ARRAY['Island Assimilation: The user becomes one with natural or artificial island terrain they touch, such as a mountainside or wooden floors, and controls it at will.','Living Landscape: The merged terrain can talk, show expressions or form giant limbs that interact with objects, and the user senses what people do on it.','Shared Damage: The island acts as the user''s body, so heavy damage to it is felt by the user, and strong or Armament Haki-coated blows can hurt them.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde9-7340-94fe-b03cde6293d8')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bde9-7340-94fe-b03cde6293d8', 'ca-ES', 'Permet a l''usuari fusionar-se amb l''entorn de qualsevol illa i manipular-lo lliurement com si fos el seu propi cos.', ARRAY['Assimilació de l''illa: L''usuari es fon amb el terreny natural o artificial d''una illa que toqui, com un vessant o terres de fusta, i el controla a voluntat.','Paisatge viu: El terreny fusionat pot parlar, mostrar expressions o formar extremitats gegants que interactuen amb objectes, i l''usuari percep què fa la gent sobre seu.','Dany compartit: L''illa actua com el cos de l''usuari, així que el dany intens que rebi el pateix ell, i els cops forts o amb Haki d''armadura poden ferir-lo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bde9-7340-94fe-b03cde6293d8')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Gabu Gabu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdea-7118-919c-cbdf6568d3d2', 'DEVIL_FRUIT', 'Gabu Gabu no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdea-7118-919c-cbdf6568d3d2', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdea-7118-919c-cbdf6568d3d2')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdea-7118-919c-cbdf6568d3d2', 'es-ES', 'Convierte al usuario en un gran bebedor cuyo estilo de combate consiste en tragar alcohol y escupirlo en forma de fuego.', ARRAY['Gran bebedor: La fruta convierte al usuario en un especialista en tragar, y sus habilidades exactas no se conocen del todo más allá de las debilidades habituales de las Frutas del Diablo.','Trago de alcohol: En combate, el usuario engulle grandes cantidades de alcohol y lo mantiene listo para escupirlo como arma.','Shugō Roppu: El usuario escupe el alcohol tragado y lo prende, creando una corriente de fuego que envuelve al objetivo en llamas.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdea-7118-919c-cbdf6568d3d2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdea-7118-919c-cbdf6568d3d2', 'en-GB', 'Turns the user into a heavy drinker whose fighting style revolves around gulping down alcohol and spitting it out as fire.', ARRAY['Heavy Drinker: The fruit makes the user a gulping specialist, and its exact abilities are not fully known beyond the standard Devil Fruit weaknesses.','Alcohol Gulp: In combat, the user swallows large amounts of alcohol and keeps it ready to be spat out as a weapon.','Shugo Roppu: The user spits the gulped alcohol and ignites it, producing a stream of fire that engulfs the target in flames.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdea-7118-919c-cbdf6568d3d2')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdea-7118-919c-cbdf6568d3d2', 'ca-ES', 'Converteix l''usuari en un gran bevedor l''estil de lluita del qual consisteix a empassar-se alcohol i escopir-lo en forma de foc.', ARRAY['Gran bevedor: La fruita converteix l''usuari en un especialista a empassar, i les seves habilitats exactes no es coneixen del tot més enllà de les debilitats habituals de les Fruites del Dimoni.','Glop d''alcohol: En combat, l''usuari empassa grans quantitats d''alcohol i el té a punt per escopir-lo com a arma.','Shugō Roppu: L''usuari escup l''alcohol empassat i l''encén, creant un corrent de foc que embolcalla l''objectiu en flames.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdea-7118-919c-cbdf6568d3d2')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Muchi Muchi no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdeb-7303-8a50-966d369f49d3', 'DEVIL_FRUIT', 'Muchi Muchi no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdeb-7303-8a50-966d369f49d3', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdeb-7303-8a50-966d369f49d3')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdeb-7303-8a50-966d369f49d3', 'es-ES', 'Permite al usuario dar órdenes a todo lo que golpee con un látigo, incluidos objetos inanimados como edificios.', ARRAY['Látigo de obediencia: Todo lo que el usuario golpea con su látigo recibe obediencia y cumple sus órdenes verbales.','Mando sobre objetos: Incluso los objetos inmóviles obedecen, de modo que una orden de moverse se cumple aunque normalmente fuera físicamente imposible.','Ventaja urbana: En las ciudades, los muchos objetos resistentes del entorno permiten al usuario controlar lo que lo rodea y atrapar multitudes enemigas casi sin ser visto.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdeb-7303-8a50-966d369f49d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdeb-7303-8a50-966d369f49d3', 'en-GB', 'Allows the user to command anything they lash with a whip, including inanimate objects such as buildings.', ARRAY['Whip of Obedience: Whatever the user strikes with their whip is implanted with obedience and will follow the user''s verbal commands.','Commanding Objects: Even immobile objects obey, so an order to move is carried out even when it would normally be physically impossible.','Urban Advantage: In cities, the many sturdy objects around let the user control the surroundings and trap enemy crowds largely unnoticed.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdeb-7303-8a50-966d369f49d3')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdeb-7303-8a50-966d369f49d3', 'ca-ES', 'Permet a l''usuari donar ordres a tot allò que copegi amb un fuet, inclosos objectes inanimats com edificis.', ARRAY['Fuet d''obediència: Tot allò que l''usuari copeja amb el fuet rep obediència i compleix les seves ordres verbals.','Comandament sobre objectes: Fins i tot els objectes immòbils obeeixen, de manera que una ordre de moure''s es compleix encara que normalment fos físicament impossible.','Avantatge urbà: A les ciutats, els molts objectes resistents de l''entorn permeten a l''usuari controlar el que l''envolta i atrapar multituds enemigues quasi sense ser vist.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdeb-7303-8a50-966d369f49d3')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Nori Nori no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdec-793b-bc90-fc7492de2b20', 'DEVIL_FRUIT', 'Nori Nori no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdec-793b-bc90-fc7492de2b20', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdec-793b-bc90-fc7492de2b20')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdec-793b-bc90-fc7492de2b20', 'es-ES', 'Permite al usuario controlar todo aquello o a todo aquel que monte, sin importar la voluntad del objetivo.', ARRAY['Control al montar: El usuario toma el mando de cualquier máquina, criatura o persona sobre la que se suba, mecánica o no, sin importar lo que el objetivo quiera.','Anulación de órdenes: El poder puede anular las órdenes propias del objetivo, como la jerarquía de mando de un Pacifista, y también alcanza a objetivos dentro del rango del Haki del usuario.','Montura única: El usuario debe subirse físicamente al objetivo, por lo que solo puede controlar uno a la vez y puede quedar expuesto a ataques.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdec-793b-bc90-fc7492de2b20')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdec-793b-bc90-fc7492de2b20', 'en-GB', 'Allows the user to control anything or anyone they ride, regardless of the target''s own will.', ARRAY['Control by Riding: The user takes command of any machine, creature or person they ride, mechanical or otherwise, whatever the target would want.','Overriding Commands: The power can override a target''s own orders, such as a Pacifista''s command hierarchy, and also reaches targets within the user''s Haki range.','Single Mount: The user must physically sit on the target, so only one can be controlled at a time, and the rider may be left open to attack.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdec-793b-bc90-fc7492de2b20')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdec-793b-bc90-fc7492de2b20', 'ca-ES', 'Permet a l''usuari controlar tot allò o tothom que munti, sense importar la voluntat de l''objectiu.', ARRAY['Control en muntar: L''usuari pren el comandament de qualsevol màquina, criatura o persona sobre la qual es pugi, mecànica o no, sense importar què vulgui l''objectiu.','Anul·lació d''ordres: El poder pot anul·lar les ordres pròpies de l''objectiu, com la jerarquia de comandament d''un Pacifista, i també arriba a objectius dins de l''abast del Haki de l''usuari.','Muntura única: L''usuari s''ha de pujar físicament a l''objectiu, de manera que només en pot controlar un alhora i pot quedar exposat als atacs.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdec-793b-bc90-fc7492de2b20')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Tsutsu Tsutsu no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bded-75aa-aca0-f637efd57664', 'DEVIL_FRUIT', 'Tsutsu Tsutsu no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bded-75aa-aca0-f637efd57664', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-75aa-aca0-f637efd57664')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-75aa-aca0-f637efd57664', 'es-ES', 'Permite al usuario transformar su cabeza en un cañón que dispara proyectiles explosivos.', ARRAY['Cañón en la cabeza: El usuario hace surgir cañones de su cabeza, que luego puede apuntar contra sus objetivos.','Disparos explosivos: Los cañones lanzan proyectiles explosivos desde lo alto de la cabeza, usando el sombrero como boca de fuego, aptos tanto para el combate como para ejecuciones.','Mantenimiento necesario: Las piezas del cañón requieren cuidado y limpieza tras su uso, y el usuario sufre además las debilidades habituales de las Frutas del Diablo.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-75aa-aca0-f637efd57664')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-75aa-aca0-f637efd57664', 'en-GB', 'Allows the user to turn their head into a cannon that fires explosive projectiles.', ARRAY['Head Cannon: The user manifests cannons from their head, which can then be aimed at targets.','Explosive Shots: The cannons fire explosive projectiles, from the top of the user''s head through a hat that serves as the barrel, suited to both combat and executions.','Maintenance Needed: Cannon parts need upkeep and cleaning after use, and the user is otherwise subject to the standard Devil Fruit weaknesses.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-75aa-aca0-f637efd57664')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bded-75aa-aca0-f637efd57664', 'ca-ES', 'Permet a l''usuari transformar el cap en un canó que dispara projectils explosius.', ARRAY['Canó al cap: L''usuari fa sorgir canons del cap, que després pot apuntar contra els seus objectius.','Trets explosius: Els canons llancen projectils explosius des del cim del cap, fent servir el barret com a boca de foc, aptes tant per al combat com per a execucions.','Manteniment necessari: Les peces del canó necessiten cura i neteja després de l''ús, i l''usuari pateix a més les debilitats habituals de les Fruites del Dimoni.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bded-75aa-aca0-f637efd57664')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Iku Iku no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdee-7011-81c8-a150d78ac43a', 'DEVIL_FRUIT', 'Iku Iku no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdee-7011-81c8-a150d78ac43a', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-7011-81c8-a150d78ac43a')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-7011-81c8-a150d78ac43a', 'es-ES', 'Permite al usuario hacer crecer los objetos inanimados a su cuidado, como si fueran organismos vivos.', ARRAY['Objetos crecientes: Los objetos que están en el territorio del usuario se agrandan hasta un tamaño gigantesco, como un libro que creció en cuanto entró en su biblioteca.','Pensado para gigantes: El poder beneficia a los gigantes, pues los textos impresos a tamaño humano pasan a ser lo bastante grandes para que puedan leerlos.','Crecimiento reversible: Los objetos agrandados pueden volver a su tamaño original, pero nunca encogerse por debajo de él, y a los humanos puede costarles manejar objetos tan grandes.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-7011-81c8-a150d78ac43a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-7011-81c8-a150d78ac43a', 'en-GB', 'Allows the user to make inanimate objects in their care grow larger, as though they were living organisms.', ARRAY['Growing Objects: Items in the user''s territory enlarge to gigantic size, such as a book that grew the moment it entered the user''s library.','Giant-Friendly: The power serves giants well, since texts printed at human size become large enough for them to read.','Reversible Growth: Enlarged objects can be returned to their original size but never shrunk below it, and humans may find the oversized items awkward to use.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-7011-81c8-a150d78ac43a')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdee-7011-81c8-a150d78ac43a', 'ca-ES', 'Permet a l''usuari fer créixer els objectes inanimats que té al seu càrrec, com si fossin organismes vius.', ARRAY['Objectes creixents: Els objectes que són al territori de l''usuari s''engrandeixen fins a una mida gegantina, com un llibre que va créixer tan bon punt va entrar a la seva biblioteca.','Pensat per a gegants: El poder beneficia els gegants, ja que els textos impresos a mida humana passen a ser prou grans perquè els puguin llegir.','Creixement reversible: Els objectes engrandits poden tornar a la mida original, però mai encongir-se per sota d''aquesta, i als humans els pot costar fer servir objectes tan grans.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdee-7011-81c8-a150d78ac43a')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Aro Aro no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdef-7a1b-8b51-90ee57e1f97e', 'DEVIL_FRUIT', 'Aro Aro no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e', 'es-ES', 'Permite al usuario generar tiras con forma de flecha a partir de su cuerpo y controlarlas telequinéticamente.', ARRAY['Tiras flecha: El usuario separa de su cuerpo tiras con forma de flecha, parecidas a vendas, con un gesto cortante y las dirige con movimientos de manos; son casi indestructibles.','Flechas perforantes: Las tiras pueden dispararse desde las manos como una ráfaga de balas, o clavarse como estacas que inmovilizan a los rivales contra el suelo.','Atar y arrastrar: Enroscadas en una extremidad o el cuello, las tiras aprietan con fuerza y pueden arrastrar al objetivo, por pesado que sea, mediante la telequinesis del usuario.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e', 'en-GB', 'Allows the user to generate arrow-shaped strips from their body and control them telekinetically.', ARRAY['Arrow Strips: The user detaches bandage-like arrow strips from their body with a chopping motion and steers them with hand gestures; they are nearly indestructible.','Piercing Arrows: The strips can be fired like gunfire from the hands, or driven in as stakes that pin opponents to the ground.','Binding and Carrying: Wrapped around a limb or neck, the strips constrict powerfully and can haul the target around, however heavy, using the user''s telekinesis.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e', 'ca-ES', 'Permet a l''usuari generar tires amb forma de fletxa a partir del seu cos i controlar-les telecinèticament.', ARRAY['Tires fletxa: L''usuari separa del cos tires amb forma de fletxa, semblants a benes, amb un gest tallant i les dirigeix amb moviments de mans; són gairebé indestructibles.','Fletxes perforants: Les tires es poden disparar des de les mans com una ràfega de bales, o clavar com estaques que immobilitzen els rivals contra el terra.','Lligar i arrossegar: Enrotllades en una extremitat o al coll, les tires estrenyen amb força i poden arrossegar l''objectiu, per pesat que sigui, mitjançant la telecinèsia de l''usuari.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdef-7a1b-8b51-90ee57e1f97e')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- Iba Iba no mi
INSERT INTO powers (id, kind, name, rarity) VALUES ('01a11ac1-bdf0-7656-96cd-b1ed238ff830', 'DEVIL_FRUIT', 'Iba Iba no mi', 'COMMON')
  ON CONFLICT DO NOTHING;
INSERT INTO devil_fruits (id, fruit_type)
  SELECT '01a11ac1-bdf0-7656-96cd-b1ed238ff830', 'PARAMECIA' WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-7656-96cd-b1ed238ff830')
  ON CONFLICT (id) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-7656-96cd-b1ed238ff830', 'es-ES', 'Permite al usuario crear y controlar espinas afiladas como cuchillas y enredaderas cubiertas de ellas, convirtiéndolo en un Hombre Espina.', ARRAY['Enredaderas espinosas: El usuario hace surgir zarzas y espinas sueltas de su cuerpo o de superficies cercanas, ajustando su longitud, grosor y, hasta cierto punto, su forma.','Zarzas de hierro: Las espinas perforan la carne al más leve roce y las zarzas son tan resistentes que se describen como espinas de hierro, útiles como tentáculos punzantes o para defenderse.','Terreno peligroso: Las zarzas que cubren el campo de batalla limitan el movimiento enemigo, y todo lo cubierto de espinas, desde un arma hasta el propio suelo, resulta dañino al tacto.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-7656-96cd-b1ed238ff830')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-7656-96cd-b1ed238ff830', 'en-GB', 'Allows the user to create and control razor-sharp thorns and thorn-covered vines, making them a Thorn-Man.', ARRAY['Thorny Vines: The user manifests briars and individual thorns from their body or nearby surfaces, adjusting their length, thickness and, to a degree, shape.','Iron Briars: The thorns pierce flesh at the slightest touch and the briars are so resilient they are described as iron thorns, useful as stabbing tendrils or defence.','Hazardous Terrain: Briars grown over the battlefield restrict enemy movement, and anything covered in thorns, from a weapon to the ground itself, becomes harmful to touch.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-7656-96cd-b1ed238ff830')
  ON CONFLICT (power_id, locale) DO NOTHING;
INSERT INTO power_translations (power_id, locale, description, skills)
  SELECT '01a11ac1-bdf0-7656-96cd-b1ed238ff830', 'ca-ES', 'Permet a l''usuari crear i controlar espines esmolades com ganivets i lianes cobertes d''elles, convertint-lo en un Home Espina.', ARRAY['Lianes espinoses: L''usuari fa sorgir esbarzers i espines soltes del seu cos o de superfícies properes, ajustant-ne la longitud, el gruix i, fins a cert punt, la forma.','Esbarzers de ferro: Les espines travessen la carn al més lleu contacte i els esbarzers són tan resistents que es descriuen com espines de ferro, útils com a tentacles punxants o per defensar-se.','Terreny perillós: Els esbarzers que cobreixen el camp de batalla limiten el moviment enemic, i tot allò cobert d''espines, des d''una arma fins al mateix terra, resulta nociu al tacte.']::text[]
  WHERE EXISTS (SELECT 1 FROM powers WHERE id = '01a11ac1-bdf0-7656-96cd-b1ed238ff830')
  ON CONFLICT (power_id, locale) DO NOTHING;

-- +goose Down
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Bane Bane no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Noro Noro no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Doa Doa no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Beri Beri no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Sabi Sabi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Shari Shari no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Suke Suke no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Toshi Toshi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Shiro Shiro no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Wara Wara no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Oto Oto no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Doku Doku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Horu Horu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Choki Choki no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Kira Kira no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Poke Poke no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Woshu Woshu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Fuwa Fuwa no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Deka Deka no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Mato Mato no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Fuku Fuku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Buki Buki no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Guru Guru no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Beta Beta no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Zushi Zushi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Bari Bari no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Giro Giro no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Ato Ato no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Jake Jake no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Pamu Pamu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Sui Sui no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Ton Ton no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Hira Hira no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Ishi Ishi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Nagi Nagi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Chiyu Chiyu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Maki Maki no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Soru Soru no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Mira Mira no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Pero Pero no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Bisu Bisu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Bata Bata no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Buku Buku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Kuri Kuri no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Shibo Shibo no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Memo Memo no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Hoya Hoya no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Netsu Netsu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Kuku Kuku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Gocha Gocha no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Oshi Oshi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Kobu Kobu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Kibi Kibi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Toki Toki no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Juku Juku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Shiku Shiku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Wapu Wapu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Riki Riki no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Nomi Nomi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Shima Shima no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Gabu Gabu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Muchi Muchi no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Nori Nori no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Tsutsu Tsutsu no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Iku Iku no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Aro Aro no mi';
DELETE FROM powers WHERE kind = 'DEVIL_FRUIT' AND name = 'Iba Iba no mi';
