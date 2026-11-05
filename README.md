# Arcana

Arcana is a low-poly multiplayer RPG inspired by the gameplay principles of
Old School RuneScape. It is being built as a Godot 4 client backed by an
authoritative Go server.

## Repository layout

- `server/` — authoritative world simulation and, later, network transport
- `client/` — Godot 4 client

## Server

The server contains a 20 Hz fixed-timestep simulation and a small continuous
click-to-move model. Player positions and speeds are authoritative on the
server. A JSON WebSocket endpoint is available at `ws://localhost:8080/ws`.
Positions use full `X/Y/Z` coordinates in preparation for layered navigation.
The prototype includes layered A* navigation and static obstacle avoidance;
general map loading and dynamic collision are not implemented yet.

With Go installed, run it from the repository root:

```sh
cd server
go run ./cmd/server
```

Run its tests with:

```sh
cd server
go test ./...
```

### Local persistence

By default, the server keeps guest identities and characters in memory. To
retain them across restarts, start the local PostgreSQL container. The
repository-root `run-server.sh` automatically loads `DATABASE_URL` from the
ignored local `.env` file when the variable is not already set:

```sh
docker compose up -d postgres
./run-server.sh
```

Copy `.env.example` to `.env` if the local file is ever removed. An explicitly
exported `DATABASE_URL` takes precedence, so hosted environments do not need an
`.env` file.

The server applies embedded schema migrations automatically. Guest token
hashes, position, health, inventory, equipped weapon, and combat style are
saved every 30 seconds, when a client disconnects, and during a graceful server
shutdown. The raw guest token stays only in the Godot client's
`user://guest_identity.cfg` file; losing that file currently means losing
access to that guest character.

Stop PostgreSQL with `docker compose down`. Its named volume is retained. Only
run `docker compose down -v` when you deliberately want to erase local
identities and characters.

The server exposes `GET /health` for process liveness and `GET /ready` for
database readiness. With the local server running, check them with:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

The normal Go suite skips the database integration test. Run that test against
the local PostgreSQL container with:

```sh
cd server
set -a; source ../.env; set +a
ARCANA_TEST_DATABASE_URL="${DATABASE_URL}" go test ./internal/persistence/postgres -run RoundTrip -v
```

The test creates a random guest identity, verifies identity and character data
through a second database connection, links an email account, and removes its
records afterward.

### Account foundation

The server database separates private account credentials from character state.
Email addresses are normalized for case-insensitive login, each account can own
one character, and account creation links an existing guest character in one
transaction. Passwords use salted Argon2id hashes with their cost parameters
stored in the encoded hash. Registration, login sessions, and client account UI
are intentionally deferred to the next slices; guest authentication remains the
only active login flow for now.

## WebSocket protocol

The client authenticates first with its saved guest token (or an empty token to
create an identity). After the `welcome` message, movement intentions look like:

```json
{
  "type": "move_to",
  "sequence": 1,
  "destination": { "x": 10, "y": 0, "z": 5 }
}
```

The server sends a `welcome` message followed by authoritative `snapshot`
messages. Sequence numbers must increase for the duration of the connection.

## Client

The Godot client provides a placeholder 3D world for testing continuous
click-to-move. It connects to the local server, sends movement intentions, and
predicts local movement immediately while reconciling with authoritative
snapshots. Remote players are smoothly interpolated.

1. Start the Go server with `cd server && go run ./cmd/server`.
2. Import `client/project.godot` in Godot 4.3 or newer.
3. Run the project and left-click the ground to move.

For local development, the repository-root helper scripts can start both sides
from any working directory. Run each command in its own terminal:

```sh
./run-server.sh
./run-client.sh
```

The green capsule is the local player. Other connected players are blue.
Drag with the middle mouse button to orbit the guided camera and use the mouse
wheel to zoom. The camera follows the local player and pulls closer when world
geometry obstructs its view.
The diagnostic line shows the latest server tick, sent and acknowledged command
sequences, and the current authoritative correction distance.
It also shows the authoritative navigation location as `layer:ground`,
`connector:ground_upper_ramp`, or `layer:upper`.

The prototype world has overlapping ground and upper-floor navigation layers
connected by one authored ramp. Clicking either floor makes the server produce
a continuous waypoint route through the ramp. Each layer uses a 0.5-unit A*
navigation grid with player-radius obstacle clearance, diagonal corner checks,
and simplified waypoints. The layout is still hardcoded and does not yet have a
map-import pipeline.

### Floor-aware interaction

World colliders are tagged and masked by navigation layer. Normal clicks target
the local player's authoritative floor while the ramp remains selectable from
either adjacent layer. The upper floor and its obstacle fade while the player
is below and partially fade during a ramp transition.

### Combat prototype

Click the red target dummy to engage it. The authoritative server approaches to
sword range and automatically attacks every four 600 ms action pulses. Clicking
a target produces a brief violet-blue expanding ring, distinct from the gold
movement marker, so engagement clicks have immediate local feedback. Clicking
a walkable surface or pressing `Escape` clears the player's target and stops
auto-approach, allowing manual movement to continue without being pulled back.
The dummy owns its aggro separately: it can continue attacking while the player
is in range and releases them after defeat, departure, or exceeding its leash.
Click the dummy again to resume attacking. A successful sword hit deals 10
damage, and the combat diagnostic shows the target, cadence, dummy health, and
last result. Restart the server to reset the dummy. Each server-confirmed hit drives a temporary sword
swing, target flash and recoil, and floating damage number on the client. These
effects consume uniquely numbered combat events from a rolling server history
in each snapshot, so repeated snapshots do not replay an outcome and the shown
damage comes from the authoritative simulation.

The dummy locks onto the first player who damages it and retaliates for 15
damage every four action pulses while that player remains in range. Players and
the dummy have 100 health. At zero health, movement and target locks are cleared
for a five-pulse defeated state; players then respawn at their original position
and the dummy resets in place, both with full health.

Press `1` to equip the sword (10 damage, four-pulse cadence), `2` to equip
the dagger (7 damage, three-pulse cadence), or `3` to equip the bow (4 damage,
five-pulse cadence, 12-unit range). Targeting immediately stops leftover manual
movement; when the target is within weapon range, the player remains planted.
Attack recovery persists across disengaging, re-engaging, and changing targets,
so target clicks cannot reset a weapon's cadence.
Bow attacks launch a lightweight projectile
presentation along a shallow arc, and hit feedback waits for the arrow to reach
the target while the underlying result remains server authoritative. Equipment changes are validated by
the server and applied on the next action pulse without resetting an attack
already on cooldown. Each weapon also supplies its own attack and approach
range. Server-driven combat paths replace local movement prediction while
approaching, preventing long-distance engagements from accumulating correction
snaps.

The world contains independent training targets for vertical ranged testing.
The upper dummy stands on the platform reached by the ramp and can be targeted
from below. An overlook dummy stands just beyond the platform edge at
`(2.75, 0, 3.5)` and can be
shot from the upper platform. Each dummy has separate health, aggro, defeat,
and respawn state.

Authoritative targets are stored by unique ID and sent as a deterministic
`targets` collection in each snapshot. The Godot client creates target scenes
from that collection, so adding another stationary training target no longer
requires new protocol fields or hardcoded nodes in the main scene.

### Inventory prototype

Each player has an authoritative eight-slot base inventory, with capacity sent
in snapshots so future upgrades can expand it per player. The starter inventory
contains the prototype sword, dagger, and bow. Use the Inventory tab in the
bottom-right HUD to view its 4×2 grid. Slots are display-only in this first
slice for non-equipment actions: weapon slots show their icons and can be
clicked to send the same server-authoritative equip request as the `1`, `2`,
and `3` shortcuts. The selected slot distinguishes equipped and queued state.
Inventory dragging, manual dropping, and item use are not implemented.

Defeated training targets drop one stackable Training Shard. Drops are
authoritative world entities with unique IDs and appear as glowing blue-violet
pickups. Click a drop from within 3 world units to collect it; the server
validates distance, available inventory space, and that another player has not
already collected it. The first shard uses a free slot and later shards stack.
Automatic walking to loot, ground-drop persistence, trading, and item use are
not included in this slice.

The Skills tab shows authoritative Attack, Range, and Magic progression. A
successful sword, dagger, mace, or battle-axe hit grants 10 Attack XP; a bow hit
grants 10 Range XP; and a staff hit grants 10 Magic XP. Dodged attacks grant no
XP. Level 2 begins at 100 XP, with prototype level thresholds following
`100 × (level - 1)²` up to level 100. XP is persisted with the character, while
levels are derived from total XP so the curve can be adjusted later. These
skills do not modify combat stats yet. Strength, Defence, and the non-combat
skills remain presentation-only placeholders.

The Equipment tab provides an icon-based overview of the authoritative
main-hand weapon plus placeholder head, body, hands, legs, and trinket slots.
Only the main-hand slot has gameplay state in this slice.

Players and the dummy each have a separate 10% dodge rate. Every ready attack
rolls dodge before damage: a dodge consumes the normal attack cadence, deals no
damage, and produces an authoritative `dodged` event shown as blue `DODGE`
feedback. Dodge is deliberately independent from future percentage defence and
flat armour calculations.

The bottom-right combat HUD shows player and target health, dodge, enemy aggro,
attack readiness, current combat state, and respawn countdowns. Its sword,
dagger, bow, and disengage buttons send the same authoritative intentions as
the `1`, `2`, `3`, and `Escape` shortcuts.
