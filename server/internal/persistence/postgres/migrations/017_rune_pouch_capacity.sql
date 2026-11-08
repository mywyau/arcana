ALTER TABLE character_rune_pouch
    DROP CONSTRAINT IF EXISTS character_rune_pouch_slot_check;

ALTER TABLE character_rune_pouch
    ADD CONSTRAINT character_rune_pouch_slot_check CHECK (slot BETWEEN 0 AND 8);

ALTER TABLE characters
    ADD COLUMN rune_pouch_unlocked_slots INTEGER NOT NULL DEFAULT 2;

-- Preserve rune stacks from the former six-slot pouch by moving anything
-- beyond the first two slots into the character's bank. Abort rather than
-- lose data if a character has no free bank slot.
DO $$
DECLARE
    rune_stack RECORD;
    destination_slot INTEGER;
    bank_limit INTEGER;
BEGIN
    FOR rune_stack IN
        SELECT player_id, slot, item_id, quantity
        FROM character_rune_pouch
        WHERE slot >= 2
        ORDER BY player_id, slot
    LOOP
        destination_slot := NULL;
        SELECT slot INTO destination_slot
        FROM character_bank_items
        WHERE player_id = rune_stack.player_id
          AND item_id = rune_stack.item_id
          AND instance_id IS NULL
        ORDER BY slot
        LIMIT 1;

        IF destination_slot IS NOT NULL THEN
            UPDATE character_bank_items
            SET quantity = quantity + rune_stack.quantity
            WHERE player_id = rune_stack.player_id
              AND slot = destination_slot;
        ELSE
            SELECT bank_capacity INTO bank_limit
            FROM characters
            WHERE player_id = rune_stack.player_id;

            SELECT candidate INTO destination_slot
            FROM generate_series(0, bank_limit - 1) AS candidate
            WHERE NOT EXISTS (
                SELECT 1
                FROM character_bank_items
                WHERE player_id = rune_stack.player_id
                  AND slot = candidate
            )
            ORDER BY candidate
            LIMIT 1;

            IF destination_slot IS NULL THEN
                RAISE EXCEPTION 'cannot migrate rune pouch for player %: bank is full', rune_stack.player_id;
            END IF;

            INSERT INTO character_bank_items (
                player_id, slot, item_id, quantity, instance_id, reinforcement
            ) VALUES (
                rune_stack.player_id, destination_slot, rune_stack.item_id,
                rune_stack.quantity, NULL, 0
            );
        END IF;
    END LOOP;

    DELETE FROM character_rune_pouch WHERE slot >= 2;
END $$;

ALTER TABLE characters
    ADD CONSTRAINT characters_rune_pouch_unlocked_slots_check
    CHECK (rune_pouch_unlocked_slots BETWEEN 2 AND 9);
