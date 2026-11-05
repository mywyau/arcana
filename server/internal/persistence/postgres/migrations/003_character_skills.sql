CREATE TABLE character_skills (
    player_id TEXT NOT NULL REFERENCES characters(player_id) ON DELETE CASCADE,
    skill_id TEXT NOT NULL,
    experience BIGINT NOT NULL CHECK (experience >= 0),
    PRIMARY KEY (player_id, skill_id)
);
