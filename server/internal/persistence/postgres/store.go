package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/simulation"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	database *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return &Store{database: database}, nil
}

func (store *Store) Close() error {
	return store.database.Close()
}

func (store *Store) Ping(ctx context.Context) error {
	if err := store.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return nil
}

func (store *Store) Migrate(ctx context.Context) error {
	if _, err := store.database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		if err := store.applyMigration(ctx, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (store *Store) applyMigration(ctx context.Context, version string) error {
	contents, err := migrations.ReadFile("migrations/" + version)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer tx.Rollback()
	// Serializes migrations when two server processes start at the same time.
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(8675309)"); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	var applied bool
	if err := tx.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version,
	).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %s: %w", version, err)
	}
	if applied {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}

func (store *Store) FindPlayerID(
	ctx context.Context,
	tokenHash [sha256.Size]byte,
) (string, bool, error) {
	var playerID string
	err := store.database.QueryRowContext(ctx,
		"SELECT player_id FROM guest_identities WHERE token_hash = $1", tokenHash[:],
	).Scan(&playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query guest identity: %w", err)
	}
	return playerID, true, nil
}

func (store *Store) Create(
	ctx context.Context,
	playerID string,
	tokenHash [sha256.Size]byte,
) error {
	result, err := store.database.ExecContext(ctx, `
		INSERT INTO guest_identities (player_id, token_hash)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, playerID, tokenHash[:])
	if err != nil {
		return fmt.Errorf("insert guest identity: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect guest identity insert: %w", err)
	}
	if rows == 0 {
		return identity.ErrIdentityExists
	}
	return nil
}

func (store *Store) Load(
	ctx context.Context,
	playerID string,
) (simulation.PersistentPlayerState, bool, error) {
	var state simulation.PersistentPlayerState
	err := store.database.QueryRowContext(ctx, `
		SELECT position_x, position_y, position_z, health,
		       equipped_weapon, combat_style, inventory_capacity
		FROM characters WHERE player_id = $1`, playerID,
	).Scan(
		&state.Position.X, &state.Position.Y, &state.Position.Z, &state.Health,
		&state.EquippedWeapon, &state.CombatStyle, &state.Inventory.Capacity,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return simulation.PersistentPlayerState{}, false, nil
	}
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query character: %w", err)
	}

	rows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, quantity FROM character_inventory
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item simulation.InventoryItem
		if err := rows.Scan(&item.Slot, &item.ItemID, &item.Quantity); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan inventory: %w", err)
		}
		state.Inventory.Items = append(state.Inventory.Items, item)
	}
	if err := rows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read inventory: %w", err)
	}
	if err := rows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close inventory rows: %w", err)
	}
	equipmentRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id FROM character_equipment
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query equipment: %w", err)
	}
	defer equipmentRows.Close()
	state.EquippedArmour = make(map[simulation.ArmourSlot]string)
	for equipmentRows.Next() {
		var slot string
		var itemID string
		if err := equipmentRows.Scan(&slot, &itemID); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan equipment: %w", err)
		}
		if slot == "spellbook" {
			state.EquippedSpellbook = itemID
		} else {
			state.EquippedArmour[simulation.ArmourSlot(slot)] = itemID
		}
	}
	if err := equipmentRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read equipment: %w", err)
	}
	if err := equipmentRows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close equipment rows: %w", err)
	}
	skillRows, err := store.database.QueryContext(ctx, `
		SELECT skill_id, experience FROM character_skills
		WHERE player_id = $1 ORDER BY skill_id`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query skills: %w", err)
	}
	defer skillRows.Close()
	for skillRows.Next() {
		var skill simulation.PersistentSkillState
		if err := skillRows.Scan(&skill.ID, &skill.Experience); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan skill: %w", err)
		}
		state.Skills = append(state.Skills, skill)
	}
	if err := skillRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read skills: %w", err)
	}
	return state, true, nil
}

func (store *Store) Save(
	ctx context.Context,
	playerID string,
	state simulation.PersistentPlayerState,
) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin character save: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO characters (
			player_id, position_x, position_y, position_z, health,
			equipped_weapon, combat_style, inventory_capacity, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (player_id) DO UPDATE SET
			position_x = EXCLUDED.position_x,
			position_y = EXCLUDED.position_y,
			position_z = EXCLUDED.position_z,
			health = EXCLUDED.health,
			equipped_weapon = EXCLUDED.equipped_weapon,
			combat_style = EXCLUDED.combat_style,
			inventory_capacity = EXCLUDED.inventory_capacity,
			updated_at = now()`,
		playerID, state.Position.X, state.Position.Y, state.Position.Z, state.Health,
		state.EquippedWeapon, state.CombatStyle, state.Inventory.Capacity,
	); err != nil {
		return fmt.Errorf("upsert character: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_inventory WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace inventory: %w", err)
	}
	for _, item := range state.Inventory.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_inventory (player_id, slot, item_id, quantity)
			VALUES ($1, $2, $3, $4)`, playerID, item.Slot, item.ItemID, item.Quantity); err != nil {
			return fmt.Errorf("insert inventory item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_equipment WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace equipment: %w", err)
	}
	for slot, itemID := range state.EquippedArmour {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_equipment (player_id, slot, item_id)
			VALUES ($1, $2, $3)`, playerID, slot, itemID); err != nil {
			return fmt.Errorf("insert equipment: %w", err)
		}
	}
	if state.EquippedSpellbook != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_equipment (player_id, slot, item_id)
			VALUES ($1, 'spellbook', $2)`, playerID, state.EquippedSpellbook); err != nil {
			return fmt.Errorf("insert equipped spellbook: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_skills WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace skills: %w", err)
	}
	for _, skill := range state.Skills {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_skills (player_id, skill_id, experience)
			VALUES ($1, $2, $3)`, playerID, skill.ID, skill.Experience); err != nil {
			return fmt.Errorf("insert skill: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit character save: %w", err)
	}
	return nil
}
