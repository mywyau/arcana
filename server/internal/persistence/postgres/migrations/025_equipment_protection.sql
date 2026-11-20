CREATE TABLE character_equipment_protection (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    equipment_key TEXT NOT NULL,
    PRIMARY KEY (player_id, equipment_key)
);
