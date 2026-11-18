# Prayer System Design

Status: living design document.

This document records the current prayer rules and the intended direction for
support-oriented prayers. Implemented behavior is separated from proposed
content so ideas do not accidentally become promises or hidden requirements.

## Design goals

Prayer should reward awareness, timing, positioning, and cooperation. It
should not become another equipment screen made entirely from passive
percentage bonuses.

- Keep prayers useful for solo players without making them mandatory.
- Let support players improve a group through decisions rather than raw DPS.
- Make powerful effects visible and understandable to allies and opponents.
- Use prayer points, trigger costs, cooldowns, positioning, and exclusivity as
  different balancing tools.
- Avoid effects whose power scales without a cap as more players gather.
- Keep combat authoritative on the server.

## Implemented prayers

### Basic boosts

These are level-1 introductory prayers. Only one offensive boost can be active
at once. Evasion, Defence, and Regeneration are utility prayers and may remain
active alongside an offensive boost.

| Prayer | Effect | Drain |
| --- | --- | --- |
| Accuracy | +5% melee hit chance | 1 point every 6 seconds |
| Strength | +5% melee damage | 1 point every 6 seconds |
| Ranged Strength | +5% ranged damage | 1 point every 6 seconds |
| Magic Strength | +5% magic damage | 1 point every 6 seconds |
| Defence | 5% general incoming-damage reduction | 1 point every 6 seconds |
| Evasion | +5% dodge chance | 1 point every 6 seconds |
| Regeneration | Restores 5% maximum health on its drain interval | 1 point every 6 seconds |

### Protection prayers

Only one protection prayer can be active. Correct protection is applied
multiplicatively with Defence and armour, rather than adding all reductions
together.

| Prayer | Effect | Drain |
| --- | --- | --- |
| Protect from Melee | 30% reduction against stab, slash, and crush | 1 point every 3.6 seconds |
| Protect from Ranged | 30% reduction against ranged damage | 1 point every 3.6 seconds |
| Protect from Magic | 30% reduction against magic damage | 1 point every 3.6 seconds |

Incorrect protection provides no reduction. Protection prayers do not change
accuracy or dodge chance.

### Mastery prayers

Masteries unlock at Prayer level 40. They consolidate modest bonuses rather
than introducing a large statistical tier jump. They share the offensive-boost
exclusive group with the introductory boosts.

| Prayer | Effect | Drain |
| --- | --- | --- |
| Melee Mastery | +5% melee accuracy, +5% melee damage, and 5% defence | 1 point every 3.6 seconds |
| Ranged Mastery | +5% ranged accuracy, +5% ranged damage, and 5% defence | 1 point every 3.6 seconds |
| Magic Mastery | +5% magic accuracy, +5% magic damage, and 5% defence | 1 point every 3.6 seconds |

This leaves room for later tiers without forcing current content to be
rebalanced around large bonuses.

### Triggered prayers

Mercy and Wrath share the emergency exclusive group, so only one can be armed.

#### Mercy

- Unlocks at Prayer level 30.
- Remains armed until a damaging hit leaves the player below 20% maximum
  health.
- Restores 20% maximum health.
- Consumes 25 prayer points and switches itself off.
- Does not trigger with fewer than 25 prayer points.
- Does not revive a player from lethal damage.

#### Wrath

- Unlocks at Prayer level 45.
- Drains 1 prayer point every 3.6 seconds while armed.
- On death, creates a visible golden burst with a 4-unit radius.
- Deals damage equal to 20% of the fallen player's maximum health.
- Can defeat nearby monsters and other players.
- Chained Wrath deaths are supported without recursively triggering the same
  player twice.

Wrath currently damages every nearby player. Party membership, safe zones,
duelling rules, and friendly-fire policy do not exist yet and must be decided
before public PvP or shared social hubs.

## Exclusivity model

Exclusivity keeps the interface flexible without allowing every combat benefit
to be stacked simultaneously.

| Group | Current members | Rule |
| --- | --- | --- |
| Offensive boost | Accuracy, Strength, Ranged Strength, Magic Strength, and all Masteries | One active |
| Protection | Protect from Melee, Ranged, or Magic | One active |
| Emergency | Mercy or Wrath | One active |
| Utility | Defence, Evasion, and Regeneration | May coexist for now |

Activating a prayer replaces the existing prayer in the same group and clears
the replaced prayer's drain progress. Future balance testing may introduce a
utility or aura group if unrestricted utility stacking becomes too efficient.

## Support-oriented direction

The next unique prayers should create cooperation rather than simply adding
more accuracy or damage.

### Recommended first support prayer: Radiant Mark

The support player marks one enemy through an attack or explicit target action.
When an allied player damages that enemy, the ally receives a small heal or
prayer restoration. A shared per-target cooldown prevents fast weapons and
large groups from multiplying the reward uncontrollably.

Why it is a strong first choice:

- Encourages the group to coordinate on one target.
- Gives the support player an active offensive decision.
- Works in ordinary combat and boss encounters.
- Does not require the support player to stop participating in combat.
- Can be capped and measured more easily than a permanent aura.

Suggested first-pass rules:

- Level 35 Prayer.
- One marked enemy per caster.
- Ten-second duration with a short reapplication cooldown.
- Restores a small fixed amount, not a percentage of the ally's damage.
- Triggers at most once per second across the whole group.
- One Radiant Mark reward per target; additional casters refresh duration but
  do not stack rewards.

### Other candidate support prayers

#### Intercession

The first nearby party member who would be defeated survives at 1 health. The
effect consumes a large amount of the support player's prayer, deactivates, and
enters a long cooldown. It must never trigger for an already-defeated player or
chain across several support players in one hit.

#### Consecrated Ground

After the caster remains stationary for a short setup time, a visible area is
created around them. Allies inside receive improved healing or recovery. Moving
collapses the area, creating a positioning trade-off.

#### Shared Burden

A capped portion of damage dealt to one linked ally is redirected to the
support player. Redirected damage must not recursively trigger another Shared
Burden link. The prayer ends if either player leaves range.

#### Purifying Chorus

Periodically removes one poison, burn, curse, or future debuff stack from one
nearby party member. It should choose a deterministic target and have a group
cooldown rather than cleansing every ally simultaneously.

#### Guiding Light

Allies moving toward the support player receive a brief movement bonus. This
helps groups regroup after area attacks without increasing general travel or
escape speed indefinitely.

#### Oathbound

Links the caster to one ally. A capped portion of healing received by either
player is shared with the other while they remain within range. Multiple links
to the same target should not stack.

#### Beacon

Improves exploration support by highlighting nearby resources, exits, or
interactable objectives for the party. This offers a prayer role outside
combat and could later interact with raids or dangerous gathering areas.

## Balance guardrails

- Prefer fixed or capped support values over uncapped percentages.
- Give group-wide effects a shared cooldown, not one cooldown per recipient.
- Do not let the same named support effect stack from multiple players.
- Use multiplicative damage reduction when several defensive layers coexist.
- A support prayer should have visible counterplay or an understandable cost.
- Avoid making one support prayer mandatory for every boss.
- Preserve a useful solo behavior where practical, but allow the group version
  to be the prayer's main strength.
- Do not solve every dangerous mechanic with immunity; mitigation, recovery,
  repositioning, and limited rescues create more design space.

## Required supporting systems

Several proposed prayers should wait until these foundations exist:

1. Party membership and an authoritative `is_ally` relationship.
2. Friendly-fire and safe-zone rules.
3. Server-authoritative status effects with duration and source ownership.
4. Per-player and shared effect cooldowns exposed to the client.
5. Spatial ally queries using the existing world grid rather than global
   player scans.
6. Clear world indicators for links, marks, auras, triggers, and cooldowns.
7. Combat-log messages explaining why a prayer triggered or failed.

Unique prayer behavior should remain in small, typed server handlers at first.
A general-purpose effect scripting engine is unnecessary until several effects
share proven lifecycle hooks such as activation, incoming damage, outgoing
damage, defeat, pulse, and ally movement.

## Suggested implementation order

1. Playtest protection, Mastery, Mercy, and Wrath values.
2. Define party membership, ally rules, and Wrath friendly-fire behavior.
3. Add a small status-effect and cooldown representation.
4. Implement Radiant Mark as the first party-support prayer.
5. Add one positioning prayer, preferably Consecrated Ground.
6. Add one rescue prayer, preferably Intercession.
7. Reassess utility-prayer stacking and create an aura-exclusive group only if
   testing shows it is necessary.

## Open decisions

- Maximum party size.
- Whether support prayers affect only formal party members or any friendly
  nearby player.
- Whether Wrath damages party members, only hostile players, or everyone
  outside safe zones.
- Whether prayer equipment modifies drain rate, effect strength, or maximum
  prayer points.
- Whether switching prayers has no cooldown, a global cooldown, or only an
  effect-specific lockout.
- Whether support contributions grant Prayer experience, encounter credit, or
  a separate assist reward.
