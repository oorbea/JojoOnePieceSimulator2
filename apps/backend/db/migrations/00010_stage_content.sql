-- +goose Up
-- picture/picture_thumb/picture_status: same pipeline as powers/users - see
-- 00003_picture_renditions.sql, whose picture_status type is reused as-is.
-- +goose StatementBegin
ALTER TABLE stages
    ADD COLUMN picture        text NOT NULL DEFAULT '',
    ADD COLUMN picture_thumb  text NOT NULL DEFAULT '',
    ADD COLUMN picture_status picture_status NOT NULL DEFAULT 'NONE';
-- +goose StatementEnd

-- stage_translations is power_translations' shape without `skills` (a Stage
-- has none) - see 00006_locales.sql. Unlike powers, all three locales are
-- mandatory for every Stage (enforced in the application layer, same
-- non-declarative reasoning 00006 gives for en-GB), so there is no partial-
-- translation state to worry about in the fallback chain in practice.
-- +goose StatementBegin
CREATE TABLE stage_translations (
    stage_id    uuid   NOT NULL REFERENCES stages (id) ON DELETE CASCADE,
    locale      locale NOT NULL,
    description text   NOT NULL CHECK (description <> ''),
    PRIMARY KEY (stage_id, locale)
);
-- +goose StatementEnd

-- Backfill: one description per locale for every stage seeded in
-- 00008_stages.sql, matched by (manga, name) since their ids were
-- gen_random_uuid() at seed time.
-- +goose StatementBegin
INSERT INTO stage_translations (stage_id, locale, description)
SELECT s.id, v.locale::locale, v.description
FROM stages s
         JOIN (VALUES
    ('JOJO', 'Phantom Blood', 'en-GB', 'Jonathan Joestar''s rivalry with his adoptive brother Dio Brando escalates into a battle against a resurrected stone mask vampire.'),
    ('JOJO', 'Phantom Blood', 'es-ES', 'La rivalidad de Jonathan Joestar con su hermano adoptivo Dio Brando se convierte en una batalla contra un vampiro resucitado por una máscara de piedra.'),
    ('JOJO', 'Phantom Blood', 'ca-ES', 'La rivalitat de Jonathan Joestar amb el seu germà adoptiu Dio Brando esdevé una batalla contra un vampir ressuscitat per una màscara de pedra.'),

    ('JOJO', 'Battle Tendency', 'en-GB', 'Joseph Joestar faces the ancient Pillar Men and masters Hamon breathing technique to stop them from conquering the world.'),
    ('JOJO', 'Battle Tendency', 'es-ES', 'Joseph Joestar se enfrenta a los antiguos Hombres Pilar y domina la técnica de respiración Hamon para impedir que conquisten el mundo.'),
    ('JOJO', 'Battle Tendency', 'ca-ES', 'Joseph Joestar s''enfronta als antics Homes Pilar i domina la tècnica de respiració Hamon per evitar que conquereixin el món.'),

    ('JOJO', 'Stardust Crusaders', 'en-GB', 'Jotaro Kujo leads a crew of Stand users across the world to defeat Dio Brando and save his mother''s life.'),
    ('JOJO', 'Stardust Crusaders', 'es-ES', 'Jotaro Kujo lidera a un grupo de usuarios de Stand a través del mundo para derrotar a Dio Brando y salvar la vida de su madre.'),
    ('JOJO', 'Stardust Crusaders', 'ca-ES', 'Jotaro Kujo lidera un grup d''usuaris d''Stand arreu del món per derrotar Dio Brando i salvar la vida de la seva mare.'),

    ('JOJO', 'Diamond is Unbreakable', 'en-GB', 'Josuke Higashikata investigates a string of Stand-related crimes in the quiet town of Morioh.'),
    ('JOJO', 'Diamond is Unbreakable', 'es-ES', 'Josuke Higashikata investiga una serie de crímenes relacionados con Stands en la tranquila ciudad de Morioh.'),
    ('JOJO', 'Diamond is Unbreakable', 'ca-ES', 'Josuke Higashikata investiga una sèrie de crims relacionats amb Stands a la tranquil·la ciutat de Morioh.'),

    ('JOJO', 'Golden Wind', 'en-GB', 'Giorno Giovanna, Dio''s secret son, rises through Passione''s gang ranks to take down the mafia boss from within.'),
    ('JOJO', 'Golden Wind', 'es-ES', 'Giorno Giovanna, el hijo secreto de Dio, asciende en las filas de la banda Passione para derrocar al jefe de la mafia desde dentro.'),
    ('JOJO', 'Golden Wind', 'ca-ES', 'Giorno Giovanna, el fill secret de Dio, ascendeix en les files de la banda Passione per enderrocar el cap de la màfia des de dins.'),

    ('JOJO', 'Stone Ocean', 'en-GB', 'Jolyne Cujoh, wrongly imprisoned, must survive a Stand-infested penitentiary while racing to save her father Jotaro.'),
    ('JOJO', 'Stone Ocean', 'es-ES', 'Jolyne Cujoh, encarcelada injustamente, debe sobrevivir en una prisión infestada de Stands mientras corre para salvar a su padre Jotaro.'),
    ('JOJO', 'Stone Ocean', 'ca-ES', 'Jolyne Cujoh, empresonada injustament, ha de sobreviure en una presó infestada d''Stands mentre corre per salvar el seu pare Jotaro.'),

    ('JOJO', 'Steel Ball Run', 'en-GB', 'Johnny Joestar and Gyro Zeppeli race across America in a brutal horse-racing competition hiding a deeper conspiracy.'),
    ('JOJO', 'Steel Ball Run', 'es-ES', 'Johnny Joestar y Gyro Zeppeli compiten a través de América en una brutal carrera de caballos que oculta una conspiración mayor.'),
    ('JOJO', 'Steel Ball Run', 'ca-ES', 'Johnny Joestar i Gyro Zeppeli competeixen a través d''Amèrica en una brutal cursa de cavalls que amaga una conspiració més gran.'),

    ('JOJO', 'JoJolion', 'en-GB', 'An amnesiac young man washes ashore in a Morioh reshaped by an earthquake, hiding a mystery about his own fused identity.'),
    ('JOJO', 'JoJolion', 'es-ES', 'Un joven amnésico llega a las costas de un Morioh transformado por un terremoto, que esconde un misterio sobre su propia identidad fusionada.'),
    ('JOJO', 'JoJolion', 'ca-ES', 'Un jove amnèsic arriba a les costes d''un Morioh transformat per un terratrèmol, que amaga un misteri sobre la seva pròpia identitat fusionada.'),

    ('ONE_PIECE', 'East Blue', 'en-GB', 'Monkey D. Luffy sets sail from his home sea, gathering the first members of the Straw Hat crew on the way to Grand Line.'),
    ('ONE_PIECE', 'East Blue', 'es-ES', 'Monkey D. Luffy zarpa desde su mar natal, reuniendo a los primeros miembros de la tripulación de los Sombrero de Paja en su camino al Grand Line.'),
    ('ONE_PIECE', 'East Blue', 'ca-ES', 'Monkey D. Luffy surt del seu mar natal, reunint els primers membres de la tripulació dels Barret de Palla en el camí cap al Grand Line.'),

    ('ONE_PIECE', 'Alabasta', 'en-GB', 'The Straw Hats help Princess Vivi stop Crocodile''s plot to plunge her desert kingdom into civil war.'),
    ('ONE_PIECE', 'Alabasta', 'es-ES', 'Los Sombrero de Paja ayudan a la princesa Vivi a detener el plan de Crocodile para sumir su reino desértico en una guerra civil.'),
    ('ONE_PIECE', 'Alabasta', 'ca-ES', 'Els Barret de Palla ajuden la princesa Vivi a aturar el pla de Crocodile per enfonsar el seu regne desèrtic en una guerra civil.'),

    ('ONE_PIECE', 'Sky Island', 'en-GB', 'The crew rides a mysterious updraft to Skypiea, a floating island ruled by a self-proclaimed god obsessed with gold.'),
    ('ONE_PIECE', 'Sky Island', 'es-ES', 'La tripulación asciende por una misteriosa corriente hasta Skypiea, una isla flotante gobernada por un autoproclamado dios obsesionado con el oro.'),
    ('ONE_PIECE', 'Sky Island', 'ca-ES', 'La tripulació ascendeix per un misteriós corrent fins a Skypiea, una illa flotant governada per un déu autoproclamat obsessionat amb l''or.'),

    ('ONE_PIECE', 'Water Seven', 'en-GB', 'The Straw Hats confront betrayal and a corrupt sea-god agency while getting the Going Merry rebuilt in a city of shipwrights.'),
    ('ONE_PIECE', 'Water Seven', 'es-ES', 'Los Sombrero de Paja se enfrentan a una traición y a una agencia marítima corrupta mientras intentan reconstruir el Going Merry en una ciudad de carpinteros navales.'),
    ('ONE_PIECE', 'Water Seven', 'ca-ES', 'Els Barret de Palla s''enfronten a una traïció i a una agència marítima corrupta mentre intenten reconstruir el Going Merry en una ciutat de mestres d''aixa.'),

    ('ONE_PIECE', 'Thriller Bark', 'en-GB', 'The crew battles the zombie-raising Gecko Moria on a ghost ship island to reclaim their stolen shadows.'),
    ('ONE_PIECE', 'Thriller Bark', 'es-ES', 'La tripulación se enfrenta a Gecko Moria, capaz de crear zombis, en una isla de barco fantasma para recuperar sus sombras robadas.'),
    ('ONE_PIECE', 'Thriller Bark', 'ca-ES', 'La tripulació s''enfronta a Gecko Moria, capaç de crear zombis, en una illa de vaixell fantasma per recuperar les seves ombres robades.'),

    ('ONE_PIECE', 'Summit War', 'en-GB', 'The Marines and the Whitebeard Pirates clash at Marineford in a war to save Portgas D. Ace from execution.'),
    ('ONE_PIECE', 'Summit War', 'es-ES', 'La Marina y los Piratas de Barbablanca chocan en Marineford en una guerra por salvar a Portgas D. Ace de la ejecución.'),
    ('ONE_PIECE', 'Summit War', 'ca-ES', 'La Marina i els Pirates de Barbablanca xoquen a Marineford en una guerra per salvar Portgas D. Ace de l''execució.'),

    ('ONE_PIECE', 'Fish-Man Island', 'en-GB', 'The Straw Hats descend into the undersea kingdom of Fish-Man Island to confront centuries of resentment toward humans.'),
    ('ONE_PIECE', 'Fish-Man Island', 'es-ES', 'Los Sombrero de Paja descienden al reino submarino de la Isla Gyojin para enfrentarse a siglos de resentimiento hacia los humanos.'),
    ('ONE_PIECE', 'Fish-Man Island', 'ca-ES', 'Els Barret de Palla baixen al regne submarí de l''Illa Gyojin per enfrontar-se a segles de rancúnia envers els humans.'),

    ('ONE_PIECE', 'Dressrosa', 'en-GB', 'Luffy''s crew joins a rebellion against Donquixote Doflamingo, the shichibukai secretly ruling a toy-cursed kingdom.'),
    ('ONE_PIECE', 'Dressrosa', 'es-ES', 'La tripulación de Luffy se une a una rebelión contra Donquixote Doflamingo, el shichibukai que gobierna en secreto un reino maldito con juguetes.'),
    ('ONE_PIECE', 'Dressrosa', 'ca-ES', 'La tripulació de Luffy s''uneix a una rebel·lió contra Donquixote Doflamingo, el shichibukai que governa en secret un regne maleït amb joguines.'),

    ('ONE_PIECE', 'Whole Cake Island', 'en-GB', 'Sanji is forced into an arranged marriage by the Charlotte family, drawing the crew into Big Mom''s cake-shaped territory.'),
    ('ONE_PIECE', 'Whole Cake Island', 'es-ES', 'Sanji es obligado a un matrimonio arreglado por la familia Charlotte, arrastrando a la tripulación al territorio de Big Mom con forma de tarta.'),
    ('ONE_PIECE', 'Whole Cake Island', 'ca-ES', 'Sanji és obligat a un matrimoni concertat per la família Charlotte, arrossegant la tripulació al territori de Big Mom amb forma de pastís.'),

    ('ONE_PIECE', 'Wano Country', 'en-GB', 'The Straw Hats join an alliance of samurai to overthrow the shogun Kaido and free the isolated country of Wano.'),
    ('ONE_PIECE', 'Wano Country', 'es-ES', 'Los Sombrero de Paja se unen a una alianza de samuráis para derrocar al shogun Kaido y liberar el aislado país de Wano.'),
    ('ONE_PIECE', 'Wano Country', 'ca-ES', 'Els Barret de Palla s''uneixen a una aliança de samurais per enderrocar el shogun Kaido i alliberar l''aïllat país de Wano.'),

    ('ONE_PIECE', 'Egghead', 'en-GB', 'The crew reaches Dr. Vegapunk''s island laboratory, uncovering secrets of the Void Century amid a World Government assault.'),
    ('ONE_PIECE', 'Egghead', 'es-ES', 'La tripulación llega al laboratorio insular del Dr. Vegapunk, descubriendo secretos del Siglo Vacío en medio de un asalto del Gobierno Mundial.'),
    ('ONE_PIECE', 'Egghead', 'ca-ES', 'La tripulació arriba al laboratori insular del Dr. Vegapunk, descobrint secrets del Segle Buit enmig d''un assalt del Govern Mundial.')
) AS v (manga, name, locale, description)
              ON v.manga = s.manga::text AND v.name = s.name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stage_translations;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE stages
    DROP COLUMN picture,
    DROP COLUMN picture_thumb,
    DROP COLUMN picture_status;
-- +goose StatementEnd
