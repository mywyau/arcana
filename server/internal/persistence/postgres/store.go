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
		       prayer_points, active_prayer_set,
		       equipped_weapon, COALESCE(equipped_weapon_instance_id, ''), equipped_weapon_reinforcement,
		       combat_style, inventory_capacity, bank_capacity, experience_locked,
		       rune_pouch_unlocked_slots, ammunition_pouch_unlocked_slots
		FROM characters WHERE player_id = $1`, playerID,
	).Scan(
		&state.Position.X, &state.Position.Y, &state.Position.Z, &state.Health,
		&state.Prayer, &state.ActivePrayerSet,
		&state.EquippedWeapon, &state.EquippedWeaponInstanceID, &state.EquippedWeaponReinforcement,
		&state.CombatStyle, &state.Inventory.Capacity, &state.Bank.Capacity, &state.ExperienceLocked,
		&state.RunePouch.UnlockedSlots, &state.AmmunitionPouch.UnlockedSlots,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return simulation.PersistentPlayerState{}, false, nil
	}
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query character: %w", err)
	}
	state.Bank.NextExpansionCost = simulation.BankExpansionCost(state.Bank.Capacity)

	rows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, quantity, COALESCE(instance_id, ''), reinforcement FROM character_inventory
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item simulation.InventoryItem
		if err := rows.Scan(&item.Slot, &item.ItemID, &item.Quantity, &item.InstanceID, &item.Reinforcement); err != nil {
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
	state.AmmunitionPouch.Capacity = simulation.AmmunitionPouchCapacity
	ammunitionRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, quantity FROM character_ammunition_pouch
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query ammunition pouch: %w", err)
	}
	defer ammunitionRows.Close()
	for ammunitionRows.Next() {
		var item simulation.InventoryItem
		if err := ammunitionRows.Scan(&item.Slot, &item.ItemID, &item.Quantity); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan ammunition pouch: %w", err)
		}
		state.AmmunitionPouch.Items = append(state.AmmunitionPouch.Items, item)
	}
	if err := ammunitionRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read ammunition pouch: %w", err)
	}
	if err := ammunitionRows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close ammunition pouch rows: %w", err)
	}
	state.RingRack.Capacity = simulation.RingRackCapacity
	ringRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, instance_id, reinforcement FROM character_ring_rack
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query ring rack: %w", err)
	}
	defer ringRows.Close()
	for ringRows.Next() {
		var item simulation.InventoryItem
		item.Quantity = 1
		if err := ringRows.Scan(&item.Slot, &item.ItemID, &item.InstanceID, &item.Reinforcement); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan ring rack: %w", err)
		}
		state.RingRack.Items = append(state.RingRack.Items, item)
	}
	if err := ringRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read ring rack: %w", err)
	}
	if err := ringRows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close ring rack rows: %w", err)
	}
	state.RunePouch.Capacity = simulation.RunePouchCapacity
	runeRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, quantity FROM character_rune_pouch
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query rune pouch: %w", err)
	}
	defer runeRows.Close()
	for runeRows.Next() {
		var item simulation.InventoryItem
		if err := runeRows.Scan(&item.Slot, &item.ItemID, &item.Quantity); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan rune pouch: %w", err)
		}
		state.RunePouch.Items = append(state.RunePouch.Items, item)
	}
	if err := runeRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read rune pouch: %w", err)
	}
	if err := runeRows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close rune pouch rows: %w", err)
	}
	bankRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, quantity, COALESCE(instance_id, ''), reinforcement FROM character_bank_items
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query bank: %w", err)
	}
	defer bankRows.Close()
	for bankRows.Next() {
		var item simulation.InventoryItem
		if err := bankRows.Scan(&item.Slot, &item.ItemID, &item.Quantity, &item.InstanceID, &item.Reinforcement); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan bank: %w", err)
		}
		state.Bank.Items = append(state.Bank.Items, item)
	}
	if err := bankRows.Err(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("read bank: %w", err)
	}
	if err := bankRows.Close(); err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("close bank rows: %w", err)
	}
	equipmentRows, err := store.database.QueryContext(ctx, `
		SELECT slot, item_id, COALESCE(instance_id, ''), reinforcement FROM character_equipment
		WHERE player_id = $1 ORDER BY slot`, playerID)
	if err != nil {
		return simulation.PersistentPlayerState{}, false, fmt.Errorf("query equipment: %w", err)
	}
	defer equipmentRows.Close()
	state.EquippedArmour = make(map[simulation.ArmourSlot]string)
	state.EquippedArmourInstances = make(map[simulation.ArmourSlot]string)
	state.EquippedArmourReinforcements = make(map[simulation.ArmourSlot]int)
	for equipmentRows.Next() {
		var slot string
		var itemID string
		var instanceID string
		var reinforcement int
		if err := equipmentRows.Scan(&slot, &itemID, &instanceID, &reinforcement); err != nil {
			return simulation.PersistentPlayerState{}, false, fmt.Errorf("scan equipment: %w", err)
		}
		if slot == "spellbook" {
			state.EquippedSpellbook = itemID
		} else {
			state.EquippedArmour[simulation.ArmourSlot(slot)] = itemID
			state.EquippedArmourInstances[simulation.ArmourSlot(slot)] = instanceID
			state.EquippedArmourReinforcements[simulation.ArmourSlot(slot)] = reinforcement
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
			prayer_points, active_prayer_set,
			equipped_weapon, equipped_weapon_instance_id, equipped_weapon_reinforcement,
			combat_style, inventory_capacity, bank_capacity, experience_locked,
			rune_pouch_unlocked_slots, ammunition_pouch_unlocked_slots, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, now())
		ON CONFLICT (player_id) DO UPDATE SET
			position_x = EXCLUDED.position_x,
			position_y = EXCLUDED.position_y,
			position_z = EXCLUDED.position_z,
			health = EXCLUDED.health,
			prayer_points = EXCLUDED.prayer_points,
			active_prayer_set = EXCLUDED.active_prayer_set,
			equipped_weapon = EXCLUDED.equipped_weapon,
			equipped_weapon_instance_id = EXCLUDED.equipped_weapon_instance_id,
			equipped_weapon_reinforcement = EXCLUDED.equipped_weapon_reinforcement,
			combat_style = EXCLUDED.combat_style,
			inventory_capacity = EXCLUDED.inventory_capacity,
			bank_capacity = EXCLUDED.bank_capacity,
			experience_locked = EXCLUDED.experience_locked,
			rune_pouch_unlocked_slots = EXCLUDED.rune_pouch_unlocked_slots,
			ammunition_pouch_unlocked_slots = EXCLUDED.ammunition_pouch_unlocked_slots,
			updated_at = now()`,
		playerID, state.Position.X, state.Position.Y, state.Position.Z, state.Health,
		state.Prayer, state.ActivePrayerSet,
		state.EquippedWeapon, nullableString(state.EquippedWeaponInstanceID), state.EquippedWeaponReinforcement,
		state.CombatStyle, state.Inventory.Capacity, state.Bank.Capacity, state.ExperienceLocked,
		state.RunePouch.UnlockedSlots, state.AmmunitionPouch.UnlockedSlots,
	); err != nil {
		return fmt.Errorf("upsert character: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_inventory WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace inventory: %w", err)
	}
	for _, item := range state.Inventory.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_inventory (player_id, slot, item_id, quantity, instance_id, reinforcement)
			VALUES ($1, $2, $3, $4, $5, $6)`, playerID, item.Slot, item.ItemID, item.Quantity, nullableString(item.InstanceID), item.Reinforcement); err != nil {
			return fmt.Errorf("insert inventory item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_ammunition_pouch WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace ammunition pouch: %w", err)
	}
	for _, item := range state.AmmunitionPouch.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_ammunition_pouch (player_id, slot, item_id, quantity)
			VALUES ($1, $2, $3, $4)`, playerID, item.Slot, item.ItemID, item.Quantity); err != nil {
			return fmt.Errorf("insert ammunition pouch item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_ring_rack WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace ring rack: %w", err)
	}
	for _, item := range state.RingRack.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_ring_rack (player_id, slot, item_id, instance_id, reinforcement)
			VALUES ($1, $2, $3, $4, $5)`, playerID, item.Slot, item.ItemID, item.InstanceID, item.Reinforcement); err != nil {
			return fmt.Errorf("insert ring rack item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_rune_pouch WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace rune pouch: %w", err)
	}
	for _, item := range state.RunePouch.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_rune_pouch (player_id, slot, item_id, quantity)
			VALUES ($1, $2, $3, $4)`, playerID, item.Slot, item.ItemID, item.Quantity); err != nil {
			return fmt.Errorf("insert rune pouch item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_bank_items WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace bank: %w", err)
	}
	for _, item := range state.Bank.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_bank_items (player_id, slot, item_id, quantity, instance_id, reinforcement)
			VALUES ($1, $2, $3, $4, $5, $6)`, playerID, item.Slot, item.ItemID, item.Quantity, nullableString(item.InstanceID), item.Reinforcement); err != nil {
			return fmt.Errorf("insert bank item: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_equipment WHERE player_id = $1", playerID); err != nil {
		return fmt.Errorf("replace equipment: %w", err)
	}
	for slot, itemID := range state.EquippedArmour {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO character_equipment (player_id, slot, item_id, instance_id, reinforcement)
			VALUES ($1, $2, $3, $4, $5)`, playerID, slot, itemID,
			nullableString(state.EquippedArmourInstances[slot]), state.EquippedArmourReinforcements[slot]); err != nil {
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

func (store *Store) Delete(ctx context.Context, playerID string) error {
	if _, err := store.database.ExecContext(
		ctx,
		"DELETE FROM characters WHERE player_id = $1",
		playerID,
	); err != nil {
		return fmt.Errorf("delete character: %w", err)
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
