# Arcana

Arcana is a low-poly multiplayer RPG inspired by the gameplay principles of
Old School RuneScape. It is being built as a Godot 4 client backed by an
authoritative Go server.

## Repository layout

- `server/` — authoritative world simulation and, later, network transport
- `client/` — Godot 4 client (not yet created)

## Server

The server currently contains a 20 Hz fixed-timestep simulation and a small
continuous click-to-move model. Player positions and speeds are authoritative
on the server. It does not yet expose a network endpoint or perform pathfinding
and collision detection.

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
