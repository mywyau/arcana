# Status Effects and Crowd Control Roadmap

Status: Slice 1 and the Slice 2 root/slow/silence vocabulary were implemented
on 2026-11-18; periodic and support effects remain planned.

## Purpose

Arcana needs effects that persist beyond a single hit. The first use case is a
brief stun from a Tyrant's telegraphed Heavy Strike, but the design should also
support poison, burning, slows, shields, regeneration, cleansing and future
support spells without adding unrelated one-off fields to player state.

The server remains authoritative over applying, advancing and removing every
effect. The client displays effects and submits intentions; it never decides
whether a character is stunned or whether a damage-over-time tick occurs.

## Crowd control versus other status effects

Crowd control and damage-over-time effects have different gameplay behavior,
but they should use the same lifecycle infrastructure.

| Category | Examples | Primary responsibility |
| --- | --- | --- |
| Control | Stun, root, silence, disarm | Restrict specific player actions |
| Movement modifier | Slow, haste | Modify movement without fully disabling it |
| Periodic | Poison, burn, bleed, regeneration | Apply an effect on scheduled pulses |
| Attribute modifier | Weaken, fortify, accuracy buff | Modify a calculated stat for a duration |
| Absorption | Shield, barrier | Consume a temporary pool before health |
| Triggered | Next-hit protection, emergency heal | Wait for an event, then resolve once |

They share:

- a stable effect ID and catalogue definition;
- a source character or enemy;
- an application and expiration pulse;
- stacking and refresh rules;
- visible and dispellable tags;
- server-side removal and death cleanup;
- snapshot revisions and one-time feedback events.

They do not share all resolution logic. A stun checks action permissions, while
poison schedules damage ticks. Those behaviors should stay in small typed
handlers rather than becoming a general scripting engine prematurely.

## Proposed state model

Each affected entity owns a bounded collection of active statuses. An active
status should contain only runtime state:

- `effect_id`
- `source_id`
- `applied_at_pulse`
- `expires_at_pulse`
- `stacks`
- `next_tick_pulse`, when relevant
- a revision used for network and persistence gating

The catalogue owns static rules such as the display name, category, maximum
stacks, refresh policy, tick interval, dispel tags and whether bosses or players
may receive the effect.

Suggested initial refresh policies are:

- **Replace duration:** reapplication moves the expiry forward without adding
  strength. Use this for stun and root.
- **Add stack:** reapplication adds a bounded stack and refreshes duration. Use
  this selectively for poison or bleed.
- **Independent source:** each source maintains its own instance. Avoid this
  until group combat needs it because it increases state and UI complexity.
- **Ignore weaker:** retain the stronger version and optionally refresh it. Use
  this for stat modifiers later.

Status collections should have a conservative limit, such as 32 active effects
per entity. Rejected or replaced effects must not grow unbounded history.

## Slice 1: shared foundation and stun

This is the recommended next implementation slice.

### Server

1. Add a small status catalogue with `stun` as its only definition.
2. Add active status state and a revision to players.
3. Advance and expire statuses on the existing combat pulse.
4. Add central action checks rather than scattering raw stun comparisons.
5. When stun begins, stop current movement and clear server-driven approaches.
6. Block movement, attacks, gathering and production while stunned.
7. Continue allowing food, potions and prayer changes so a stunned player still
   has defensive choices.
8. Do not reset attack cooldowns; they continue to elapse while the player is
   stunned.
9. Clear stun on death and respawn.
10. Apply a three-pulse stun when a player is hit by a Tyrant Heavy Strike.

Three pulses equals 1.8 seconds with the current 600 ms action pulse. It matches
the planned Tyrant vulnerability window, so a player who fails the movement
check cannot exploit that opening while teammates who dodged it can.

### Client

1. Show an obvious but compact stun indicator on the character.
2. Show the remaining time near other timed combat effects.
3. Keep accepting input locally, but do not predict movement that the server has
   rejected due to stun.
4. Display a concise message such as `You are stunned.` when an action is
   rejected.
5. Play one application effect rather than repeating it with every snapshot.

### Tests

- stun expires on the expected pulse;
- reapplication refreshes but does not stack duration;
- movement and attacks are rejected during stun;
- food, potions and prayer changes remain available;
- applying stun cancels movement and automatic approach paths;
- death and respawn clear stun;
- snapshot revisions transmit changes and meaningful empty state;
- Heavy Strike applies stun only to players hit by the area;
- players who leave the marked area are neither damaged nor stunned.

## Slice 2: control effect vocabulary

Framework support is implemented. These effects are available to encounter
code but have not yet been assigned to an enemy, spell or item.

Add effects only when content needs them:

- **Root:** blocks movement but allows attacks and casting.
- **Slow:** multiplies movement speed by 0.7 (a 30% reduction) with a defined
  minimum. The catalogue value can be tuned centrally.
- **Silence:** blocks spells but leaves weapon attacks available.
- **Disarm:** blocks weapon attacks but leaves movement and spells available.

Action permissions should be expressed as queries such as
`canMove`, `canAttack`, `canCast` and `canInteract`. Individual systems should
not need to know every status ID.

This slice should also add immunity tags for enemies. Ordinary bosses should
usually resist hard control such as stun while remaining susceptible to selected
slows or encounter-specific control mechanics.

## Slice 3: poison, burn and bleed

Periodic effects require additional rules beyond crowd control:

1. Schedule ticks from authoritative pulses.
2. Attribute damage to the original source for loot and contribution credit.
3. Define whether armour, protection prayers or elemental resistance mitigate
   each effect.
4. Prevent damage ticks after expiration, death or removal.
5. Decide stacking per effect rather than globally.
6. Emit each tick once as a combat event instead of retaining repeated history.

Recommended identities:

- **Poison:** longer duration, modest damage, cleansable.
- **Burn:** shorter duration, higher damage, associated with fire effects.
- **Bleed:** physical damage that may become stronger while the target moves,
  if playtesting shows that movement trade-off is understandable and fair.

Combat statuses should initially disappear on logout rather than being persisted.
Long-duration non-combat effects can be considered separately later.

## Slice 4: support effects

Once ally targeting and party identity exist, the same framework can support:

- regeneration;
- temporary shields;
- cleanses that remove effects by tag;
- resistance to the next control effect;
- movement assistance and regrouping effects;
- marked enemies that reward allied attacks;
- short defensive or offensive team buffs.

Support effects should be capped and non-stacking by effect ID unless a design
explicitly calls for stacking. This prevents a large group from multiplying the
same buff into a mandatory strategy.

## Slice 5: PvP and control resistance

Do not add a complicated diminishing-return system before PvP exists. When it
is needed, evaluate:

- shorter control durations against players;
- a brief immunity window after hard control ends;
- category-based diminishing returns;
- equipment or prayers that improve control resistance;
- safe-zone and friendly-fire rules;
- cleanse cooldowns and counterplay.

Bosses should use explicit immunity or duration multipliers rather than sharing
player PvP rules.

## Networking and scale

Statuses should follow the revision-gated snapshot work already used by private
containers and production progress:

- transmit the active list only when its revision changes;
- expose only publicly relevant effects on remote players;
- keep private timing or source details on the affected player's snapshot;
- deliver application, removal and tick feedback as one-time events;
- avoid a per-status goroutine or timer;
- advance all statuses from the fixed simulation pulse;
- keep collections bounded and catalogue definitions static.

Visible control effects such as stun and root must be public so remote clients
can animate them. Hidden or purely numerical modifiers do not automatically need
to be exposed.

## Balance guardrails

- Telegraph hard control when it comes from an enemy encounter.
- Hard control should be short and uncommon.
- Do not combine unavoidable damage with long control.
- Preserve at least one defensive response during most control effects.
- Do not reset cooldowns merely because control was applied.
- Avoid control chains with no opportunity to act.
- Make application, duration and removal visually understandable.
- Give bosses deliberate immunity rules rather than silently ignoring effects.

## Recommended delivery order

1. Shared active-status state and catalogue.
2. Stun action gates and client feedback.
3. Tyrant Heavy Strike stun integration.
4. Root or slow for the first directional movement encounter.
5. Poison and burn with damage attribution.
6. Cleanse and shield effects for support play.
7. Party-aware buffs and contribution credit.
8. PvP resistance and diminishing returns only when PvP is being built.

## Open decisions

- Should crafting be cancelled or merely paused when control is applied?
- Should drinking a potion be permitted during every stun source?
- Which control effects should work on bosses, and at what duration multiplier?
- Should bleed punish movement, or would that conflict too strongly with
  movement-heavy encounters?
- Which statuses are visible to enemies, party members and nearby strangers?
- Should cleanses choose the oldest effect, the strongest effect or a selected
  effect?
- How should healing, cleansing and shielding contribute to group rewards?
