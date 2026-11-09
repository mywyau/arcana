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

Run the dependency-free Godot client regression tests with:

```sh
./run-client-tests.sh
```

The suite uses temporary identity files and does not modify normal client
profiles.

### Client/server compatibility

WebSocket authentication includes a numeric protocol version. The server
rejects missing or incompatible versions before loading a character, and the
client displays an update-required screen without deleting its saved session.
Increment `protocol.CurrentVersion` and `NetworkClient.PROTOCOL_VERSION`
together when a network change is intentionally incompatible. The Godot
`application/config/version` value is a human-readable build label and does
not control compatibility.

Only one live connection may control a character. A newer authenticated
connection saves and replaces the older one; the displaced client stops
automatic reconnects and offers an explicit **Reconnect here** action.

### Hosted server

The repository includes a production server image in `server/Dockerfile` and a
Render Blueprint in `render.yaml`. The Blueprint creates one Frankfurt web
service and one managed PostgreSQL database. Keep the service at one instance:
the active world is owned by one server process and is not horizontally shared.

Create a Render Blueprint from this repository, review the paid service and
database plans, and deploy it. Render supplies `PORT` and `DATABASE_URL`; the
server applies its embedded migrations and exposes `/ready` for the health
check. Production mode refuses to start without PostgreSQL and refuses to
enable development content.

The relevant environment variables are:

- `ARCANA_ENV`: `development`, `staging`, or `production`.
- `DATABASE_URL`: required outside development.
- `PORT`: hosting platform HTTP port; defaults to `8080` locally.
- `ARCANA_ENABLE_DEBUG_CONTENT`: defaults on only in development and cannot be
  enabled in production.
- `ARCANA_ALLOW_GUEST_AUTH`: defaults on in development and off elsewhere; it
  cannot be enabled in production.
- `ARCANA_ALLOWED_ORIGINS`: optional comma-separated WebSocket origin hosts for
  future browser clients. Native Godot clients do not send a browser origin.

Build and exercise the same container locally with:

```sh
docker build -t arcana-server ./server
docker run --rm -p 8080:8080 \
  -e ARCANA_ENV=development \
  arcana-server
```

The Godot client reads `network/server_url` from `client/project.godot`. It can
also be overridden for staging tests without editing the project:

```sh
ARCANA_SERVER_URL=wss://arcana-beta-server.onrender.com/ws ./run-client.sh beta
```

The client opens an account gate before entering the world. Registration and
login use the server's `/api/account/*` HTTP endpoints, then the returned
session token authenticates the WebSocket connection. The account API address
is derived from the WebSocket address (`wss` becomes `https` and `/ws` is
removed), or it can be set explicitly with `ARCANA_ACCOUNT_API_URL`. Guest play
is shown only for a local development server and must be selected explicitly.

For a persistent local setting, copy `.env.client.example` to `.env.client`
and set `ARCANA_SERVER_URL` there. `run-client.sh` loads that file automatically,
and `.env.client` is ignored by Git:

```sh
cp .env.client.example .env.client
./run-client.sh player-one
```

To connect to a server running locally without changing `.env.client`, use the
local wrapper instead:

```sh
./run-server.sh
./run-client-local.sh player-one
```

The health endpoint uses HTTPS (`/ready`), while the game client connects to
the WebSocket endpoint using WSS (`/ws`).

Guest identities are stored per profile and per server URL. The same profile
name can therefore connect to localhost and the hosted beta without presenting
one server's guest token to the other. Existing profile tokens are migrated as
localhost identities the first time that profile connects locally.

If a connection drops, the client reconnects and authenticates automatically.
Ordinary failures retry after 1, 2, 4, and progressively longer delays capped
at 30 seconds. A stale guest token is cleared and retried immediately. Once the
server welcomes the client, authoritative snapshots restore the visible world.

An exported build can use the project setting or a user argument such as
`-- --server-url=wss://arcana-beta-server.onrender.com/ws`. Public deployments
must use `wss://`, not `ws://`.

### Build the first macOS beta client

Install the export templates matching the installed Godot version from
**Editor > Manage Export Templates**. Then build the tested Universal 2 macOS
archive with:

```sh
./build-client-macos.sh
```

The script runs the headless client tests and writes a versioned application
archive plus SHA-256 checksum under `dist/`. The `macOS Beta` export preset adds
the `arcana_beta` feature, which selects the hosted WSS endpoint baked into
`client/project.godot`; ordinary editor and local-script runs continue to use
localhost. Send the archive, checksum and `BETA_TESTER_README.md` to a tester.

This initial build is ad-hoc signed but not Apple-notarised. Testers may need to
right-click the extracted application and choose **Open**. Proper Developer ID
signing and notarisation remain later distribution work.

Build the 64-bit Intel/AMD Windows beta archive from the same macOS development
machine with:

```sh
./build-client-windows.sh
```

This writes `Arcana.exe`, its separate `Arcana.pck` data file, and the tester
instructions into one versioned ZIP under `dist/`. Keeping the PCK separate
reduces avoidable antivirus false positives and leaves a cleaner path to future
Windows code signing. This initial private build is not code-signed, so Windows
SmartScreen may require the tester to choose **More info → Run anyway**.

Windows testers should extract the complete ZIP, keep `Arcana.exe` and
`Arcana.pck` in the same folder, and then double-click `Arcana.exe`. They should
not launch the executable from inside the ZIP or receive the executable without
its PCK file. The packaged `BETA_TESTER_README.md` contains the complete launch,
SmartScreen and server-availability instructions.

### Account foundation

The server database separates private account credentials from character state.
Email addresses are normalized for case-insensitive login, accounts currently
have one character slot, and account creation can either provision a new
`player-*` character or link an existing guest. Passwords use salted Argon2id
hashes with their cost parameters stored in the encoded hash. Successful login
issues a random 30-day session token; PostgreSQL stores only its SHA-256 hash,
while the client stores the raw token per profile and server URL.

The reusable account API exposes `POST /api/account/register`,
`POST /api/account/login`, and `POST /api/account/logout`. Hosted clients and a
future website use these endpoints over HTTPS; logout supplies the session as a
Bearer token. Registration and password login are rejected over WebSocket so
they cannot bypass the HTTP account controls. WebSocket authentication accepts
an issued account session token, plus explicitly enabled local-development
guest identities.

Registration and login attempts are protected by bounded in-memory rate
limits. A limited request receives HTTP 429 with a `Retry-After` header, which
the client presents as a readable wait time. Successful login clears the failed
attempt history for that client/email pair; ordinary session-token reconnects
are not limited. The limiter is process-local, which is appropriate while the
beta runs one server instance. Move this state to a shared store before running
multiple interchangeable account API instances.

`ARCANA_TRUST_PROXY_HEADERS` controls whether the HTTP account API uses the
first valid address in `X-Forwarded-For`. It defaults to `false` locally and
must only be enabled behind a trusted proxy that replaces or sanitises that
header. The Render Blueprint enables it for the hosted service.

Settings > Manage account also supports logging out on the current device. The
server deletes that session, the client removes its local token, and the game
reconnects as a new guest. Other device sessions remain active.

Character persistence is independent from guest credentials. Production rejects
guest WebSocket authentication, while local development retains it for rapid
testing. The hosted service should remain suspended until the client startup
authentication screen uses the account API.

Email verification, password reset, and "log out everywhere" are not
implemented yet. These remain deferred for the small trusted private beta.

## WebSocket protocol

The client authenticates the WebSocket with an account session in production.
Development servers can also accept a saved guest token (or an empty token to
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

Pass a profile name to keep a separate guest identity, HUD layout, and client
settings for local multiplayer testing. Each client window includes its profile
name in the title:

```sh
./run-client.sh player-one
./run-client.sh player-two
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

Each player has an authoritative 20-slot base inventory within a maximum of 50
slots. Capacity and item placement are sent in snapshots, while the HUD renders
the unlocked slots with item icons, stack quantities and drag-and-drop ordering.
Fresh characters begin with an empty inventory and collect their equipment and
tools from the world.

Defeated training targets drop stackable Coins. Drops are authoritative world
entities with unique IDs. Click a drop from within 3 world units to collect it; the server
validates distance, available inventory space, and that another player has not
already collected it. The first coin drop uses a free slot and later coins stack.
Automatic walking to loot, ground-drop persistence, trading, and item use are
not included in this slice.

The Skills tab shows authoritative Attack, Strength, Defence, Range, and Magic
progression. Each successful hit grants XP equal to the target's actual health
lost, routed to the skill associated with the selected combat technique. Sword
and dagger Stab, mace Crush, and battle-axe
Measured Slash train Attack; Slash and Heavy Crush techniques train Strength;
and the melee Defensive technique trains Defence. Bow techniques train Range,
staff techniques train Magic, and Unfocused attacks grant no combat XP. Dodged
attacks also grant no XP. The HUD displays `COMBAT XP LOCKED` while Unfocused is
selected.

Level 2 begins at 100 XP, with prototype level thresholds following
`100 × (level - 1)²` up to level 100. XP is persisted with the character, while
levels are derived from total XP so the curve can be adjusted later. Skills do
not modify combat stats yet, and non-combat skills remain presentation-only
placeholders.

The Equipment tab provides an icon-based overview of the authoritative
main-hand weapon plus placeholder head, body, hands, legs, and trinket slots.
Only the main-hand slot has gameplay state in this slice.

Players and the dummy each have a separate 10% dodge rate. Every ready attack
rolls dodge before damage: a dodge consumes the normal attack cadence, deals no
damage, and produces an authoritative `dodged` event shown as blue `DODGE`
feedback. Dodge is deliberately independent from future percentage defence and
flat armour calculations.

The persistent bottom-left player HUD shows authoritative health, Mana, Prayer,
and dodge. Players begin with 100/100 Mana and 100/100 Prayer; spending and
regeneration are not implemented yet. Enemy health is deliberately separate: a
top-centre Target HUD appears only for the selected enemy or an enemy that
currently has aggro. Styles is the default bottom-right tab, and the Prayer tab
contains disabled placeholder prayers ready for a later activation slice.

Press `F10` to edit the HUD. Drag a highlighted `MOVE` header to reposition a
panel and its corner handle to scale it between 75% and 150%. The layout is
saved locally in `user://hud_layout.cfg`; `Shift+F10` restores the defaults.
The former combat panel and its attack-pulse bar are hidden from players. Press
`F9` to show or hide that panel while debugging combat timing.
