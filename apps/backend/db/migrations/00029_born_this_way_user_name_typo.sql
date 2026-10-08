-- +goose Up
-- Born This Way's user is Kei Nijimura, but the wiki text we translated says
-- "Kyo Nijimura" in the "Absolute Autonomy and Disappearance" skill (the header
-- says Kei). 00025 kept the source typo on purpose; this fixes it in every locale.
-- Skills live only in power_translations.skills. Rows without the typo are left
-- untouched, so an admin edit made since then wins. See
-- ObsidianVault/catalog-seed-p7-p8-fruits-2026-10-08.md.
UPDATE power_translations
   SET skills = ARRAY(
         SELECT replace(s.skill, 'Kyo Nijimura', 'Kei Nijimura')
           FROM unnest(skills) WITH ORDINALITY AS s (skill, ord)
          ORDER BY s.ord)
 WHERE power_id = '01a0ef77-194a-715e-8829-56fb3b472e04'
   AND EXISTS (SELECT 1 FROM unnest(skills) AS k WHERE k LIKE '%Kyo Nijimura%');

-- +goose Down
-- Deliberately a no-op: restoring a known typo has no value.
SELECT 1;
