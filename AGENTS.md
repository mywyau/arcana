# AGENTS.md — Fantasy MMO

## Project overview

We are building a low-poly 3D multiplayer RPG inspired by the gameplay principles of Old School RuneScape.

The game will use a Godot 4 client and a custom Go server.

## Technology

- Client: Godot 4 and GDScript
- Server: Go
- Networking: WebSockets initially
- Database: PostgreSQL (introduced when persistence is required)
- Infrastructure: Docker
- Version control: Git

## Architecture principles

1. The game server is authoritative over world state.
2. Clients send player intentions, not authoritative state changes.
3. Networking must remain separate from game simulation.
4. The server uses a fixed-timestep simulation loop.
5. Only the world simulation owns and mutates active gameplay state.
6. Prefer a modular monolith over premature microservices.
7. Keep transport-specific logic isolated to make future changes possible.
8. Do not introduce dependencies without explaining why they are needed.

## Development practices

- Implement one clearly defined feature at a time.
- Keep changes small and reviewable.
- Write Go unit tests for game logic.
- Validate incoming network messages.
- Avoid unnecessary abstractions.
- Do not implement features outside the task's scope.
- Run relevant tests and checks after making changes.
- Report tests that could not be run.
- Never commit secrets or credentials.

## Godot conventions

- Use typed GDScript where practical.
- Separate scenes from gameplay and networking logic.
- Prefer reusable scenes for game entities.
- Use placeholder geometry until mechanics are working.
- Follow Godot 4 APIs and conventions.

## Completion requirements

For each task, provide:

1. Summary of changes.
2. Files created or modified.
3. Tests and checks performed.
4. Any known limitations.
5. Manual steps required to verify the feature.
