-- +goose Up
-- Catalogue fix-ups on rows admins created by hand in prod (they exist in no
-- earlier migration): canonical names, plus a few Stand rarity corrections.
-- See ObsidianVault/catalog-seed-p7-p8-fruits-2026-10-08.md.
--
-- Every statement is guarded so an admin edit always wins and the migration is
-- safe to re-run or to run on a database that lacks these rows: renames match
-- on the exact old name (a row an admin already renamed simply doesn't match),
-- and rarity changes only apply while the row still holds the old value.
--
-- Renaming a Stand is not cosmetic here: game/power_effects.go keys its Spin
-- rules by exact name, so the Tusk rename ships together with the code change
-- that matches 'tusk: act N'.

UPDATE powers SET name = 'Tusk: Act 1' WHERE kind = 'STAND' AND name = 'Tusk: Acto 1';
UPDATE powers SET name = 'Tusk: Act 2' WHERE kind = 'STAND' AND name = 'Tusk: Acto 2';
UPDATE powers SET name = 'Tusk: Act 3' WHERE kind = 'STAND' AND name = 'Tusk: Acto 3';
UPDATE powers SET name = 'Tusk: Act 4' WHERE kind = 'STAND' AND name = 'Tusk: Acto 4';

-- Part 3's The World and Steel Ball Run's are different Stands; the latter was
-- distinguished only by capitals.
UPDATE powers SET name = 'The World (Steel Ball Run)' WHERE kind = 'STAND' AND name = 'THE WORLD';

UPDATE powers SET name = 'Sugar Mountain' WHERE kind = 'STAND' AND name = 'Sugar Mountain''s Spring';
UPDATE powers SET name = 'Milagroman' WHERE kind = 'STAND' AND name = 'Milagro Man';
UPDATE powers SET name = 'Ozon Baby' WHERE kind = 'STAND' AND name = 'Ozone Baby';
UPDATE powers SET name = 'TATOO YOU!' WHERE kind = 'STAND' AND name = 'Tatoo You!';
UPDATE powers SET name = 'Hermit Purple' WHERE kind = 'STAND' AND name = 'Hermit purple';
UPDATE powers SET name = 'D4C: Love Train' WHERE kind = 'STAND' AND name = 'Dirty Deeds Done Dirt Cheap: Love Train';

-- Rarity: both Worlds MYTHICAL (like Star Platinum: The World); the base D4C
-- drops to LEGENDARY so Love Train is its higher tier; High Priestess is one of
-- the weakest Egyptian Stands.
UPDATE powers SET rarity = 'MYTHICAL' WHERE kind = 'STAND' AND name = 'The World' AND rarity = 'LEGENDARY';
UPDATE powers SET rarity = 'MYTHICAL' WHERE kind = 'STAND' AND name = 'The World (Steel Ball Run)' AND rarity = 'LEGENDARY';
UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'STAND' AND name = 'Dirty Deeds Done Dirt Cheap' AND rarity = 'MYTHICAL';
UPDATE powers SET rarity = 'RARE' WHERE kind = 'STAND' AND name = 'High Priestess' AND rarity = 'EPIC';

-- +goose Down
UPDATE powers SET rarity = 'EPIC' WHERE kind = 'STAND' AND name = 'High Priestess' AND rarity = 'RARE';
UPDATE powers SET rarity = 'MYTHICAL' WHERE kind = 'STAND' AND name = 'Dirty Deeds Done Dirt Cheap' AND rarity = 'LEGENDARY';
UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'STAND' AND name = 'The World (Steel Ball Run)' AND rarity = 'MYTHICAL';
UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'STAND' AND name = 'The World' AND rarity = 'MYTHICAL';

UPDATE powers SET name = 'Dirty Deeds Done Dirt Cheap: Love Train' WHERE kind = 'STAND' AND name = 'D4C: Love Train';
UPDATE powers SET name = 'Hermit purple' WHERE kind = 'STAND' AND name = 'Hermit Purple';
UPDATE powers SET name = 'Tatoo You!' WHERE kind = 'STAND' AND name = 'TATOO YOU!';
UPDATE powers SET name = 'Ozone Baby' WHERE kind = 'STAND' AND name = 'Ozon Baby';
UPDATE powers SET name = 'Milagro Man' WHERE kind = 'STAND' AND name = 'Milagroman';
UPDATE powers SET name = 'Sugar Mountain''s Spring' WHERE kind = 'STAND' AND name = 'Sugar Mountain';
UPDATE powers SET name = 'THE WORLD' WHERE kind = 'STAND' AND name = 'The World (Steel Ball Run)';
UPDATE powers SET name = 'Tusk: Acto 4' WHERE kind = 'STAND' AND name = 'Tusk: Act 4';
UPDATE powers SET name = 'Tusk: Acto 3' WHERE kind = 'STAND' AND name = 'Tusk: Act 3';
UPDATE powers SET name = 'Tusk: Acto 2' WHERE kind = 'STAND' AND name = 'Tusk: Act 2';
UPDATE powers SET name = 'Tusk: Acto 1' WHERE kind = 'STAND' AND name = 'Tusk: Act 1';
