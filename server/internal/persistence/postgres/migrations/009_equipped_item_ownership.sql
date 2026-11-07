ALTER TABLE characters
    ADD COLUMN equipped_weapon_reinforcement INTEGER NOT NULL DEFAULT 0 CHECK (equipped_weapon_reinforcement BETWEEN 0 AND 1);

ALTER TABLE character_equipment
    ADD COLUMN reinforcement INTEGER NOT NULL DEFAULT 0 CHECK (reinforcement BETWEEN 0 AND 1);

UPDATE characters AS character
SET equipped_weapon_reinforcement = inventory.reinforcement
FROM character_inventory AS inventory
WHERE inventory.player_id = character.player_id
  AND inventory.instance_id = character.equipped_weapon_instance_id;

UPDATE character_equipment AS equipment
SET reinforcement = inventory.reinforcement
FROM character_inventory AS inventory
WHERE inventory.player_id = equipment.player_id
  AND inventory.instance_id = equipment.instance_id;
