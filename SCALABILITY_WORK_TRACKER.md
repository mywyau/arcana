# Arcana Scalability Work Tracker

Target: up to 2,000 connected players in one logical world.

This is the living execution tracker. Keep architectural discussion in the
longer scaling documents, and update this file after each measured slice.

## Working rules

- Keep the authoritative simulation at 20 Hz and snapshots at 10 Hz unless
  measurements justify changing them.
- Prefer small, measured slices over a server rewrite.
- Do not replace Go, WebSockets, PostgreSQL, or the modular monolith without a
  demonstrated bottleneck.
- Run spread, hotspot, and soak scenarios before claiming a capacity.
- Check with the project owner before gameplay-facing decisions such as crowd
  limits, instancing, or reduced update fidelity.

## Completed foundations

- [x] Public/private player state separation.
- [x] Shared immutable public frame per network interval.
- [x] Spatial area-of-interest grid and 40-unit visibility radius.
- [x] Public player and target deltas with periodic reconciliation.
- [x] Revision-gated inventory, equipment, bank, specialist containers,
  skills, boosts, and production progress.
- [x] Bounded replaceable snapshot queue and stalled-client disconnection.
- [x] Slower environment snapshot cadence.
- [x] Performance summaries and headless WebSocket load testing.

## Completed slice: crowded-area visibility budget

Status: implemented; awaiting hotspot measurement.

First pass:

- [x] Add a configurable maximum number of visible players per connection.
- [x] Always preserve the local player, their player target, and players
  attacking them, even if this exceeds the incidental-player budget.
- [x] Select remaining incidental players by distance with deterministic ID
  tie-breaking.
- [x] Add tests for caps and protected combat participants.
- [x] Document and expose the configuration in local and hosted environments.

Follow-up decision:

- [ ] Choose the beta default after a hotspot test. Initial engineering default:
  100 total visible player entries per connection.
- [ ] Add selection hysteresis if hotspot movement causes visible-player churn.
- [ ] Decide future party/clan priority rules when those systems exist.

Acceptance:

- A dense crowd cannot make per-client player payload grow without bound.
- Combat-relevant players remain visible outside the ordinary crowd selection.
- Spread-world behavior remains unchanged below the configured budget.

## Active slice: static and dynamic environment separation

Status: payload separation implemented; measurement and revisions remain.

- [x] Send gathering, production-station, altar, and bank definitions on first
  visibility and whenever the visible landmark membership changes.
- [x] Send compact gathering resource quantities in periodic environment
  updates instead of repeating complete definitions.
- [x] Keep drops and fires in dynamic environment updates.
- [x] Preserve definition payloads when the replaceable snapshot queue drops
  an older frame for a newer realtime frame.
- [x] Separate merchant identity/position from changing stock.
- [ ] Add content/environment revision identifiers for reconnect and future
  area streaming.
- [ ] Measure environment payload reduction with the spread load scenario.

## Queued slices

### 3. One-time transient event delivery

- Track the last event delivered to each connection.
- Send hit, miss, sound, XP, and action feedback events once.
- Bound event retention while keeping durable state recoverable from snapshots.

### 4. Dirty asynchronous persistence

- Mark characters dirty only after durable changes.
- Coalesce saves per character through a bounded queue outside the simulation
  and connection snapshot paths.
- Prevent stale writes, expose queue metrics, retry safely, and drain on
  graceful shutdown.

### 5. Client crowd-cost controls

- Match server visibility priorities in Godot.
- Reduce distant animation and overhead UI cost.
- Pool projectiles, hit splats, XP drops, and other frequent effects.
- Measure frame time, node count, visible players, and active effects.

### 6. Simulation spatial queries

- Extend spatial indexing to enemy acquisition, AoE, pickup, and nearby action
  queries once profiles show global scans are material.

### 7. Capacity validation and operations

- Run spread and hotspot tests at 100, 250, 500, 1,000, then 2,000 clients.
- Add multi-hour soak tests and p95/p99 measurements.
- Track hosted CPU, resident memory, bandwidth, GC pauses, outbound queue depth,
  save queue depth, and database pool pressure.

## Measurement log

Record date, commit, scenario, client count, duration, server size, tick p99,
payload rate, heap/GC behavior, disconnects, and conclusions here after each
capacity run.

- No post-crowd-budget measurements recorded yet.

