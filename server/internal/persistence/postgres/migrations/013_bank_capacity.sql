ALTER TABLE characters
    ADD COLUMN bank_capacity INTEGER NOT NULL DEFAULT 100
    CHECK (bank_capacity BETWEEN 100 AND 1000 AND (bank_capacity - 100) % 50 = 0);

ALTER TABLE character_bank_items
    DROP CONSTRAINT character_bank_items_slot_check;

ALTER TABLE character_bank_items
    ADD CONSTRAINT character_bank_items_slot_check CHECK (slot BETWEEN 0 AND 999);
