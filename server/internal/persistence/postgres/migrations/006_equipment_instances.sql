ALTER TABLE character_inventory
    ADD COLUMN IF NOT EXISTS instance_id TEXT,
    ADD COLUMN IF NOT EXISTS reinforcement INTEGER NOT NULL DEFAULT 0 CHECK (reinforcement BETWEEN 0 AND 1);

CREATE UNIQUE INDEX IF NOT EXISTS character_inventory_instance_id_key
    ON character_inventory (instance_id)
    WHERE instance_id IS NOT NULL;

ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS equipped_weapon_instance_id TEXT;

ALTER TABLE character_equipment
    ADD COLUMN IF NOT EXISTS instance_id TEXT;
