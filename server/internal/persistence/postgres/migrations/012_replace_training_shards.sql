UPDATE character_inventory
SET item_id = 'coins'
WHERE item_id = 'training_shard';

UPDATE character_bank_items AS coins
SET quantity = coins.quantity + shards.quantity
FROM character_bank_items AS shards
WHERE coins.player_id = shards.player_id
  AND coins.item_id = 'coins'
  AND shards.item_id = 'training_shard';

DELETE FROM character_bank_items AS shards
WHERE shards.item_id = 'training_shard'
  AND EXISTS (
      SELECT 1
      FROM character_bank_items AS coins
      WHERE coins.player_id = shards.player_id
        AND coins.item_id = 'coins'
  );

UPDATE character_bank_items
SET item_id = 'coins'
WHERE item_id = 'training_shard';
