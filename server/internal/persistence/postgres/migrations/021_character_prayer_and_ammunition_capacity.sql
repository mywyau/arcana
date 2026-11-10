ALTER TABLE characters
    ADD COLUMN prayer_points INTEGER NOT NULL DEFAULT 100,
    ADD COLUMN active_prayer_set TEXT NOT NULL DEFAULT 'common_litany',
    ADD COLUMN ammunition_pouch_unlocked_slots INTEGER NOT NULL DEFAULT 2;

ALTER TABLE characters
    ADD CONSTRAINT characters_prayer_points_check
        CHECK (prayer_points BETWEEN 0 AND 100),
    ADD CONSTRAINT characters_active_prayer_set_check
        CHECK (btrim(active_prayer_set) <> ''),
    ADD CONSTRAINT characters_ammunition_pouch_unlocked_slots_check
        CHECK (ammunition_pouch_unlocked_slots BETWEEN 1 AND 10);
