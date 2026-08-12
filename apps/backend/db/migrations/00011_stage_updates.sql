-- +goose Up
-- Canon-accurate renames: 00008_stages.sql seeded these two names before the
-- localized descriptions in 00010_stage_content.sql already called them by
-- their correct in-universe names ("Skypiea", "Marineford") - fixing stages.name
-- to match. stage_translations is keyed by stage_id, not name, so no backfill
-- needed there.
-- +goose StatementBegin
UPDATE stages SET name = 'Skypiea' WHERE manga = 'ONE_PIECE' AND name = 'Sky Island';
UPDATE stages SET name = 'Marineford' WHERE manga = 'ONE_PIECE' AND name = 'Summit War';
-- +goose StatementEnd

-- New stages: JoJo's latest ongoing part and One Piece's latest saga.
-- +goose StatementBegin
INSERT INTO stages (manga, position, name) VALUES
    ('JOJO', 8, 'The JOJOLands'),
    ('ONE_PIECE', 11, 'Elbaph');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO stage_translations (stage_id, locale, description)
SELECT s.id, v.locale::locale, v.description
FROM stages s
         JOIN (VALUES
    ('JOJO', 'The JOJOLands', 'en-GB', 'Jodio Joestar gets pulled into a heist across a Hawaii reshaped by Stand-fuelled crime and family secrets.'),
    ('JOJO', 'The JOJOLands', 'es-ES', 'Jodio Joestar se ve envuelto en un golpe en un Hawái marcado por el crimen impulsado por Stands y secretos familiares.'),
    ('JOJO', 'The JOJOLands', 'ca-ES', 'Jodio Joestar es veu embolicat en un cop en un Hawaii marcat pel crim impulsat per Stands i secrets familiars.'),

    ('ONE_PIECE', 'Elbaph', 'en-GB', 'The Straw Hats reach the hidden giants'' homeland of Elbaph, drawn into a war entwined with ancient legend.'),
    ('ONE_PIECE', 'Elbaph', 'es-ES', 'Los Sombrero de Paja llegan a Elbaph, la tierra oculta de los gigantes, y se ven envueltos en una guerra ligada a una antigua leyenda.'),
    ('ONE_PIECE', 'Elbaph', 'ca-ES', 'Els Barret de Palla arriben a Elbaph, la terra amagada dels gegants, i es veuen embolicats en una guerra lligada a una antiga llegenda.')
) AS v (manga, name, locale, description)
              ON v.manga = s.manga::text AND v.name = s.name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM stages WHERE (manga, name) IN (('JOJO', 'The JOJOLands'), ('ONE_PIECE', 'Elbaph'));
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE stages SET name = 'Sky Island' WHERE manga = 'ONE_PIECE' AND name = 'Skypiea';
UPDATE stages SET name = 'Summit War' WHERE manga = 'ONE_PIECE' AND name = 'Marineford';
-- +goose StatementEnd
