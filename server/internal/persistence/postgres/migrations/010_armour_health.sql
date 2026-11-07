ALTER TABLE characters
    DROP CONSTRAINT IF EXISTS characters_health_check;

ALTER TABLE characters
    ADD CONSTRAINT characters_health_check CHECK (health BETWEEN 1 AND 10000);
