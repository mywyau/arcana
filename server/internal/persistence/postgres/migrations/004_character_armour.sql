CREATE TABLE character_equipment (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    slot TEXT NOT NULL,
    item_id TEXT NOT NULL,
    PRIMARY KEY (player_id, slot)
);

INSERT INTO character_inventory (player_id, slot, item_id, quantity)
SELECT characters.player_id, free_slot.slot, 'training_helm', 1
FROM characters
CROSS JOIN LATERAL (
    SELECT candidate AS slot
    FROM generate_series(0, characters.inventory_capacity - 1) AS candidate
    WHERE NOT EXISTS (
        SELECT 1 FROM character_inventory
        WHERE player_id = characters.player_id AND slot = candidate
    )
    ORDER BY candidate
    LIMIT 1
) AS free_slot
WHERE NOT EXISTS (
    SELECT 1 FROM character_inventory
    WHERE player_id = characters.player_id AND item_id = 'training_helm'
);

INSERT INTO character_inventory (player_id, slot, item_id, quantity)
SELECT characters.player_id, free_slot.slot, 'training_plate', 1
FROM characters
CROSS JOIN LATERAL (
    SELECT candidate AS slot
    FROM generate_series(0, characters.inventory_capacity - 1) AS candidate
    WHERE NOT EXISTS (
        SELECT 1 FROM character_inventory
        WHERE player_id = characters.player_id AND slot = candidate
    )
    ORDER BY candidate
    LIMIT 1
) AS free_slot
WHERE NOT EXISTS (
    SELECT 1 FROM character_inventory
    WHERE player_id = characters.player_id AND item_id = 'training_plate'
);
