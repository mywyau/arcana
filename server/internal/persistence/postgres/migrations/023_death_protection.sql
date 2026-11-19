CREATE TABLE character_death_protection (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    inventory_slot INTEGER NOT NULL CHECK (inventory_slot >= 0),
    PRIMARY KEY (player_id, inventory_slot)
);
