# Arcana Server Scaling Plan for 2000 Player Worlds

Arcana should scale through measured, incremental changes to its authoritative
Go world rather than an early rewrite or a switch away from WebSockets. The
target is up to 2,000 connected players in one logical world, assuming players
are normally distributed across the map. A single location containing all
2,000 players requires explicit crowd limits and lower-priority visibility.

The immediate priorities are reducing per-connection snapshot work, preventing
private state exposure and measuring the real cost of each server tick. Spatial
partitioning and delta updates follow after those foundations are proven.

## Current baseline

- The authoritative simulation advances every 50 ms, or 20 Hz.
- Combat and most timed actions use a separate 600 ms pulse.
- The Godot client predicts local movement and interpolates remote movement.
- Each connection currently requests a full world snapshot on its own timer.
- Players, targets, drops, gathering spots, chests and merchants are included
  in every snapshot.
- Test supply chest contents are represented as infinite world drops.
- PostgreSQL stores durable character state; gameplay remains in memory.

## Completed first slice

- Keep the authoritative simulation at 20 Hz.
- Send production network snapshots every 100 ms, or 10 Hz.
- Omit remote players' inventory, bank, skills, combat calculations,
  ammunition pouch, ring rack and rune pouch from JSON snapshots.
- Preserve complete private state for the owning player.
- Test that private fields are absent from remote-player JSON.

This is intentionally a transport-level first step. The simulation snapshot
still constructs some private state before transport filtering; a later shared
frame should avoid that unnecessary work entirely.

## Performance targets

Establish targets before optimising individual systems:

- No authoritative tick may routinely exceed the 50 ms budget.
- Normal tick work should remain well below 50 ms to leave spike headroom.
- Outbound queues must remain bounded for every connection.
- Snapshot size should depend primarily on nearby activity, not total world
  population.
- Database latency must not block simulation ticks.
- Memory should stabilise during a multi-hour soak test.
- A slow client must not increase memory use without limit.

Exact numerical limits for bandwidth, entity visibility and players per cell
should be chosen from load-test measurements rather than assumptions.

## Slice 2: Add baseline measurements

Measure before making deeper changes:

- Simulation tick duration and ticks exceeding budget.
- Full world snapshot construction duration.
- JSON encoding and WebSocket write duration.
- Encoded snapshot bytes per connection.
- Connected and disconnecting player counts.
- Counts of players, targets, drops, fires and gathering spots.
- PostgreSQL save duration and failures.
- Process memory, allocations and garbage-collection pauses.

Emit a compact periodic summary and expose machine-readable metrics later when
the deployment platform and monitoring service are chosen.

Acceptance criteria:

- Metrics add negligible cost and contain no credentials or private data.
- A two-player session provides tick, snapshot, bandwidth and save baselines.
- Slow ticks and failed saves are easy to identify by timestamp.

## Slice 3: Construct public and private state separately

Replace transport-time removal with explicit simulation snapshots:

```text
PublicPlayerState
  position, movement, visible equipment, health, combat level, target

PrivatePlayerState
  inventory, storage, skills, action progress, detailed combat calculations
```

Only construct private state for the owning connection or persistence request.
Do not clone every player's inventory and bank into a general world frame.

Acceptance criteria:

- Public snapshots cannot contain private fields by construction.
- Persistence still captures the complete character.
- Existing client inventory, skills, equipment and combat interfaces work.
- Benchmarks show reduced allocations as player count grows.

## Slice 4: Remove static data from frequent snapshots

Send immutable definitions during welcome or area entry:

- Bank chest positions and names.
- Merchant positions and definitions.
- Gathering spot definitions.
- Other static interactable metadata.

Frequent updates should contain only changing values, such as resource amount
or merchant stock revision.

Move test supply chest contents from infinite world drops into a chest-specific
catalogue that is sent when the chest opens. This keeps beta tooling from
inflating every gameplay snapshot.

Acceptance criteria:

- Static fixtures do not appear in routine movement snapshots.
- Test chests retain the same interface and infinite-item behaviour.
- Opening and closing a test chest does not affect unrelated clients.

## Slice 5: Send transient events once

Track each connection's last delivered event ID. Send only newer combat,
feedback and world events rather than repeatedly transmitting event history.

Important state must still be recoverable from a periodic authoritative
snapshot. Events provide presentation and feedback, not the sole durable truth.

Acceptance criteria:

- Each combat event is presented once per relevant client.
- Reconnection cannot replay an unbounded event history.
- Missing an event cannot leave authoritative client state permanently wrong.

## Slice 6: Build one immutable frame per network interval

Stop invoking `World.Snapshot()` independently for every connection:

```text
Simulation tick
    -> immutable public frame
    -> connection filtering and private attachment
    -> network encoding outside the world lock
```

The simulation remains the sole writer. Network workers may only read detached
frames and must not mutate authoritative state.

Acceptance criteria:

- Public world cloning happens once per network interval.
- A slow WebSocket writer cannot delay the simulation.
- Frame data is immutable and race tests pass.
- Connection-specific private data remains isolated.

## Slice 7: Add bounded outbound delivery

Give every connection a small bounded outbound queue. Coalesce obsolete state
frames and retain the newest state. Reliable account responses and important
events must be handled separately from replaceable movement frames.

Acceptance criteria:

- Slow clients cannot cause unbounded memory growth.
- Obsolete movement frames are discarded safely.
- Persistently stalled clients disconnect with a useful reason.
- Normal clients continue receiving current authoritative state.

## Slice 8: Add spatial indexing and interest management

Partition the map into cells. Maintain indexes for players, enemies, drops,
fires and interactables. Each connection subscribes to its current cell and a
configured neighbourhood.

Use the same spatial index for:

- Network visibility.
- Enemy target searches.
- Nearby interactions and item pickup.
- Area effects.
- Movement and collision queries where applicable.

Acceptance criteria:

- Distant players and drops are absent from a client's updates.
- Entering and leaving visibility creates and removes client entities cleanly.
- Boundary movement does not flicker or duplicate entities.
- Snapshot cost follows nearby entity count rather than total world count.

## Slice 9: Add state revisions and delta updates

Assign revisions to dynamic subsystems and entities. After an initial area
snapshot, send additions, changes and removals. Continue sending occasional
recovery snapshots.

Useful revisions include:

- Player public state.
- Local inventory and equipment.
- Skills and active boosts.
- Bank and specialist storage.
- Resource quantities.
- Merchant stock.

Acceptance criteria:

- Unchanged state produces no repeated payload.
- Out-of-order or stale changes are rejected using tick or revision numbers.
- A recovery snapshot repairs deliberately dropped test updates.

## Slice 10: Optimise client entity processing

The Godot client should process only received changes:

- Pool projectiles, hit indicators and other frequent effects.
- Avoid rebuilding unchanged item and world views.
- Cull names, bars and animations by distance and importance.
- Limit fully rendered incidental players in dense crowds.
- Prioritise the local player, party members, attackers and targets.
- Track frame time, active nodes and effect counts.

Acceptance criteria:

- Client frame time remains stable as nearby entity count grows.
- Expired effects and entities return to a stable baseline.
- A crowd policy is documented for locations exceeding the normal visibility
  budget.

## Slice 11: Improve persistence throughput

- Mark characters dirty when durable state changes.
- Stagger periodic saves instead of creating synchronized bursts.
- Use a bounded asynchronous persistence queue.
- Keep database work outside the simulation lock.
- Drain pending saves during graceful shutdown.
- Measure pool usage, save latency and retries.
- Test backups and restoration before progress is considered permanent.

Acceptance criteria:

- Idle characters do not generate unnecessary writes.
- Database slowdown does not stall movement or combat ticks.
- Queue saturation is visible and handled without silent data loss.

## Slice 12: Build a headless load harness

Simulated clients should authenticate, connect, move, gather, fight, idle,
disconnect and reconnect. Increase load gradually:

```text
10 -> 50 -> 100 -> 250 -> 500 -> 1000 -> 1500 -> 2000
```

Run three distributions:

1. Players spread across the map.
2. Several busy activity hubs.
3. A deliberately crowded worst-case location.

Acceptance criteria:

- Tests produce reproducible tick, bandwidth, memory and database results.
- Multi-hour soak tests reveal no unbounded growth.
- The supported population cap includes operational headroom rather than using
  the first point at which the server fails.

## Later work only if profiling requires it

Do not introduce these before the preceding measurements show a need:

- Compact binary high-frequency protocol.
- Parallel simulation of independent spatial cells.
- Multiple processes cooperating on one seamless world.
- Dedicated network gateways.
- Active-passive world failover.
- Kubernetes or automatic world-process scheduling.

WebSockets, Go and the modular authoritative world can remain. At 2,000 players,
selective state construction and transmission are more important than changing
the underlying transport.

## Recommended next slice

Implement Slice 2 next: tick, snapshot-size, encoding and connection metrics.
Use the resulting two-player baseline to measure the first slice and guide the
static-data and shared-frame work that follows.
