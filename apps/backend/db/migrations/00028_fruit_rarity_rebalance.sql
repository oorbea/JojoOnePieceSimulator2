-- +goose Up
-- Devil Fruit rarity rebalance. Until now MYTHICAL held a single fruit and 31 of
-- the 75 fruits were LEGENDARY, so the tier no longer said anything. The target
-- mirrors the Stand catalogue's proportions (~7% MYTHICAL, ~11% LEGENDARY, ~19%
-- EPIC): MYTHICAL is for the world-level fruits, Ancient Zoans drop to EPIC. See
-- ObsidianVault/catalog-seed-p7-p8-fruits-2026-10-08.md.
--
-- Rarity does not affect the draw (uniform) - only display and the Versus bots'
-- loadout score. Each update only applies while the row still holds the old
-- value, so a rarity an admin changed since wins.

-- LEGENDARY -> MYTHICAL
UPDATE powers SET rarity = 'MYTHICAL' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'LEGENDARY'
  AND name IN ('Hito Hito no mi: Model Nika', 'Gura Gura no mi', 'Yami Yami no mi', 'Ope Ope no Mi',
               'Uo Uo no mi: Model Seiryu', 'Magu Magu no mi', 'Pika Pika no mi', 'Hie Hie no mi');

-- MYTHICAL -> LEGENDARY (Magu Magu takes its slot)
UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'MYTHICAL'
  AND name = 'Hebi Hebi no mi: Model Yamata no Orochi';

-- LEGENDARY -> EPIC (Ancient Zoans and Mori Mori)
UPDATE powers SET rarity = 'EPIC' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'LEGENDARY'
  AND name IN ('Ryu Ryu no mi: Model Allosaurus', 'Ryu Ryu no mi: Model Brachiosaurus',
               'Ryu Ryu no mi: Model Pachycephalosaurus', 'Ryu Ryu no mi: Model Pteranodon',
               'Ryu Ryu no mi: Model Spinosaurus', 'Ryu Ryu no mi: Model Triceratops',
               'Zou Zou no mi: Model Mammoth', 'Neko Neko no mi: Model Saber-Toothed Tiger',
               'Kumo Kumo no mi: Model Rosamygale Grauvogeli', 'Mori Mori no mi');

-- LEGENDARY -> RARE
UPDATE powers SET rarity = 'RARE' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'LEGENDARY'
  AND name = 'Tama Tama no mi';

-- +goose Down
UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'RARE'
  AND name = 'Tama Tama no mi';

UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'EPIC'
  AND name IN ('Ryu Ryu no mi: Model Allosaurus', 'Ryu Ryu no mi: Model Brachiosaurus',
               'Ryu Ryu no mi: Model Pachycephalosaurus', 'Ryu Ryu no mi: Model Pteranodon',
               'Ryu Ryu no mi: Model Spinosaurus', 'Ryu Ryu no mi: Model Triceratops',
               'Zou Zou no mi: Model Mammoth', 'Neko Neko no mi: Model Saber-Toothed Tiger',
               'Kumo Kumo no mi: Model Rosamygale Grauvogeli', 'Mori Mori no mi');

UPDATE powers SET rarity = 'MYTHICAL' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'LEGENDARY'
  AND name = 'Hebi Hebi no mi: Model Yamata no Orochi';

UPDATE powers SET rarity = 'LEGENDARY' WHERE kind = 'DEVIL_FRUIT' AND rarity = 'MYTHICAL'
  AND name IN ('Hito Hito no mi: Model Nika', 'Gura Gura no mi', 'Yami Yami no mi', 'Ope Ope no Mi',
               'Uo Uo no mi: Model Seiryu', 'Magu Magu no mi', 'Pika Pika no mi', 'Hie Hie no mi');
