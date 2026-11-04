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
Pathfinding and collision detection are not implemented yet.

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
  "destination": { "x": 10, "z": 5 }
}
```

The server sends a `welcome` message followed by authoritative `snapshot`
messages. Sequence numbers must increase for the duration of the connection.

## Client

The Godot client provides a placeholder 3D world for testing continuous
click-to-move. It connects to the local server, sends movement intentions, and
smoothly renders authoritative snapshots.

1. Start the Go server with `cd server && go run ./cmd/server`.
2. Import `client/project.godot` in Godot 4.3 or newer.
3. Run the project and left-click the ground to move.

The green capsule is the local player. Other connected players are blue.
