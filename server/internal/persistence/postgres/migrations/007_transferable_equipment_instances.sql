DROP INDEX IF EXISTS character_inventory_instance_id_key;

CREATE UNIQUE INDEX character_inventory_instance_id_key
    ON character_inventory (player_id, instance_id)
    WHERE instance_id IS NOT NULL;
