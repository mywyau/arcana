CREATE TABLE character_ammunition_pouch (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 0 AND 9),
    item_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (player_id, slot),
    UNIQUE (player_id, item_id)
);
