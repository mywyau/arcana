CREATE TABLE character_ring_rack (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 0 AND 4),
    item_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    reinforcement INTEGER NOT NULL DEFAULT 0 CHECK (reinforcement BETWEEN 0 AND 5),
    PRIMARY KEY (player_id, slot),
    UNIQUE (player_id, instance_id)
);
