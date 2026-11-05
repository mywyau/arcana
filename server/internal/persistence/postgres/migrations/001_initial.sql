CREATE TABLE guest_identities (
    player_id TEXT PRIMARY KEY,
    token_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE characters (
    player_id TEXT PRIMARY KEY REFERENCES guest_identities(player_id) ON DELETE CASCADE,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_z DOUBLE PRECISION NOT NULL DEFAULT 0,
    health INTEGER NOT NULL CHECK (health BETWEEN 1 AND 100),
    equipped_weapon TEXT NOT NULL,
    combat_style TEXT NOT NULL,
    inventory_capacity INTEGER NOT NULL CHECK (inventory_capacity BETWEEN 1 AND 100),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE character_inventory (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    slot INTEGER NOT NULL CHECK (slot >= 0),
    item_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (player_id, slot),
    UNIQUE (player_id, item_id)
);
