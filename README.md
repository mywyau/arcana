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

## WebSocket protocol

Each connection receives a temporary player identity. To move, send:

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
