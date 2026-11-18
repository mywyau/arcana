ALTER TABLE characters
    ADD COLUMN character_name TEXT;

WITH existing_accounts AS (
    SELECT player_id, row_number() OVER (ORDER BY player_id) AS legacy_number
    FROM characters
    WHERE account_id IS NOT NULL
)
UPDATE characters
SET character_name = 'Player' || existing_accounts.legacy_number
FROM existing_accounts
WHERE characters.player_id = existing_accounts.player_id;

ALTER TABLE characters
    ADD CONSTRAINT characters_character_name_length
    CHECK (
        character_name IS NULL
        OR (
            char_length(character_name) BETWEEN 3 AND 20
            AND character_name = btrim(character_name)
        )
    );

CREATE UNIQUE INDEX characters_character_name_lower_unique
    ON characters (lower(character_name))
    WHERE character_name IS NOT NULL;
