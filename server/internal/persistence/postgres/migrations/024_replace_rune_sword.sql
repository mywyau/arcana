UPDATE characters
SET equipped_weapon = 'adamantite_sword'
WHERE equipped_weapon = 'rune_sword';

UPDATE character_inventory
SET item_id = 'adamantite_sword'
WHERE item_id = 'rune_sword';

UPDATE character_bank_items
SET item_id = 'adamantite_sword'
WHERE item_id = 'rune_sword';
