CREATE TABLE character_bank_items (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 0 AND 99),
    item_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    instance_id TEXT,
    reinforcement INTEGER NOT NULL DEFAULT 0 CHECK (reinforcement BETWEEN 0 AND 1),
    PRIMARY KEY (player_id, slot)
);

CREATE UNIQUE INDEX character_bank_instance_id_key
    ON character_bank_items (player_id, instance_id)
    WHERE instance_id IS NOT NULL;
