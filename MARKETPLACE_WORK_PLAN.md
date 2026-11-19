# Player Marketplace Work Plan

Status: proposed design and implementation roadmap.

This document describes a player marketplace inspired by exchange-style MMO
markets without committing the game to an exact Grand Exchange clone. The
first version should prove safe item custody and trading before adding
automatic matching, price charts, or a separate market service.

## Goals

- Let players trade without both players being online together.
- Keep item and coin movement authoritative on the server.
- Prevent duplicated items, duplicated coins, overselling, and double claims.
- Keep the first implementation small enough to test and review safely.
- Support future taxes, automatic offers, multiple worlds, and price history.
- Use PostgreSQL and the existing Go server until scale demonstrates a need for
  a separate service.

## Recommended first version

Begin with a direct-listing marketplace rather than an automatically matched
order book.

A seller deposits a stackable item, chooses a unit price and quantity, and
creates a listing. Another player chooses that specific listing and buys some
or all of its remaining quantity. Items and sale proceeds are collected from
the marketplace rather than being inserted into an online inventory during the
transaction.

The first version should support:

- Stackable commodities only.
- Player-selected unit prices.
- Partial purchases from a listing.
- A small active-listing limit per account or character.
- Listing cancellation by the seller.
- Separate collection of purchased items and earned coins.
- Fixed listing expiry.
- Server-side validation and PostgreSQL transactions.

The first version should not support:

- Reinforced or individually modified equipment.
- Automatic buy offers or price matching.
- Price-history charts.
- Cross-world or cross-region markets.
- Offline notifications.
- A standalone marketplace microservice.

## Core custody model

The market must hold assets in escrow. A listing must never remain backed only
by an item in the seller's normal inventory.

```text
Seller inventory -> item escrow -> buyer collection
Buyer coins      -> completed trade -> seller collection
```

Creating a listing removes the listed quantity from the seller's inventory in
the same authoritative operation that creates the escrow record. Buying locks
the listing, verifies its remaining quantity, transfers coins, reduces the
escrow quantity, and creates collection entries in one database transaction.

Cancellation moves unsold escrow into the seller's collection. It must not try
to place the item directly into an inventory that may be full or currently
owned by another server process.

## Relationship to item values

The marketplace price is chosen by players. The proposed canonical item
`BaseValue` used by death retention and NPC shops should not set the market
price.

It can still be used for:

- A suggested initial listing price.
- NPC shop buy and sell prices.
- Listing-fee or tax calculations, if desired.
- Detecting obviously invalid values during development.

Market prices should remain independent because supply and demand will differ
from authored item values.

## Proposed data model

Names may change during implementation, but the responsibilities should remain
separate.

### `market_listings`

- Listing ID.
- Seller character ID and account ID.
- Item ID.
- Original quantity and remaining quantity.
- Unit price in coins.
- Status: active, sold, cancelled, or expired.
- Creation and expiry timestamps.
- Optional version number for optimistic checks and auditability.

### `market_collections`

- Collection ID.
- Owning character or account ID.
- Item ID or coin amount.
- Quantity remaining to collect.
- Source listing or transaction ID.
- Creation timestamp.

Collection entries allow completed trades and cancellations to remain safe
when the player's inventory is full or the player is offline.

### `market_transactions`

- Transaction ID.
- Listing ID.
- Buyer and seller IDs.
- Item ID and quantity.
- Unit price and total price.
- Timestamp.

This immutable trade ledger is useful for support investigations, balancing,
future price history, and duplication audits. It should not be treated as the
source of current escrow quantities.

## Transaction rules

All state-changing operations must be executed by the server and wrapped in
database transactions.

### Create listing

1. Validate item, quantity, price, listing limit, and tradeability.
2. Lock or otherwise safely update the character inventory state.
3. Remove exactly the requested stack quantity.
4. Insert the listing and its escrowed quantity.
5. Commit both changes together.

### Purchase listing

1. Lock the active listing row with `SELECT ... FOR UPDATE` or an equivalent
   conditional update.
2. Verify the requested quantity is still available.
3. Verify the buyer is not the seller and has enough coins.
4. Remove the buyer's coins.
5. Reduce the listing's remaining quantity and complete it when empty.
6. Credit the purchased item to the buyer's collection.
7. Credit the sale proceeds to the seller's collection.
8. Append the immutable transaction record.
9. Commit everything together.

### Cancel or expire listing

1. Lock the listing and confirm it is still active.
2. Mark it cancelled or expired.
3. Move its remaining escrow quantity into the seller's collection.
4. Commit both changes together.

### Collect assets

1. Lock the collection entry and character inventory state.
2. Move as much as can safely fit, or reject the action with a clear message.
3. Reduce or delete the collection entry only by the amount actually moved.
4. Commit the inventory and collection updates together.

Every command should be safe to retry or reject after a duplicated client
request. The client must never decide that payment or delivery succeeded.

## Server structure

Keep the first version inside the existing Go modular monolith, but isolate it
behind clear packages and interfaces.

Suggested responsibilities:

- Marketplace domain logic: validation, statuses, totals, and state changes.
- PostgreSQL repository: transactions, row locking, queries, and migrations.
- Simulation integration: movement to the market interaction point and opening
  the interface.
- Network handlers: validated intentions such as list, buy, cancel, inspect,
  and collect.
- Client marketplace controller: browsing, listing creation, confirmation, and
  collection UI.

The simulation should not scan or match marketplace listings during world
ticks. Market operations are request-driven database work and should remain
outside the fixed-timestep hot path.

## Client experience for the spike

Use a marketplace clerk, board, or chest as the world interaction point. Its
interface can have four concise views:

1. Browse listings.
2. Sell an item.
3. My listings.
4. Collection.

The client should show item icons, item names, unit price, available quantity,
and total purchase price. It should require confirmation before a purchase or
listing is submitted. Search and sorting can initially be limited to item name,
price, and newest listing.

The server must return a fresh result if a listing changes before confirmation;
the client should never assume cached quantity or price is still current.

## Security and integrity requirements

- Use integer coin amounts; never floating-point currency.
- Reject zero, negative, overflowing, or unreasonably large prices and
  quantities.
- Validate tradeability from the authoritative item catalogue.
- Do not trust item names, prices, ownership, or totals sent by the client.
- Prevent a character from buying its own listing.
- Rate-limit search and mutation requests separately.
- Record listing, cancellation, trade, and collection identifiers in logs.
- Keep an immutable transaction ledger for investigation.
- Test concurrent purchases against the final quantity.
- Do not silently destroy assets when an inventory is full.

## Implementation slices

### Slice 0: design decisions

- Decide whether listings and collections belong to an account or character.
- Choose active-listing limit, expiry duration, and price limits.
- Decide whether the spike has no tax, a listing fee, or a sale tax.
- Define the initial set of tradeable stackable items.

Recommended defaults are character ownership, six active listings, seven-day
expiry, no tax during the private spike, and an explicit tradeability flag in
the item catalogue.

### Slice 1: database-only marketplace foundation

- Add migrations for listings, collections, and transactions.
- Add typed domain records and PostgreSQL repository methods.
- Implement listing creation, direct purchase, cancellation, and collection.
- Add transaction and concurrency tests using the PostgreSQL test database.
- Do not add client or world interaction yet.

Exit condition: two concurrent buyers cannot purchase more than the remaining
quantity, and every committed coin or item movement can be accounted for.

### Slice 2: server protocol and authorization

- Add browse, inspect, list, buy, cancel, and collect messages.
- Validate authenticated character ownership and request limits.
- Keep market database work outside the simulation tick.
- Add handler tests for malformed and unauthorized requests.

### Slice 3: basic market interface

- Add a market interaction point.
- Add browse, sell, personal listings, and collection views.
- Add confirmation prompts and useful stale-listing errors.
- Reuse the established item icon catalogue and contextual menus.

### Slice 4: beta hardening

- Add expiry processing in bounded batches.
- Add operational counters for listings, purchases, failures, and latency.
- Add reconciliation queries for escrow and collection balances.
- Exercise disconnects, full inventories, retries, server restarts, and database
  failures.
- Run concurrency and load tests before enabling the feature remotely.

### Later: automatic exchange offers

Only after direct listings are proven should the system add buy offers and
automatic matching. That requires price-time priority, coin escrow for buy
orders, partial fills across several orders, deterministic matching, and more
detailed cancellation and refund rules.

Automatic matching should be implemented as a transactional command or a
bounded worker, not inside the world simulation tick.

## Decisions deliberately deferred

- Exact taxes and long-term coin sinks.
- Buy offers and price-time matching.
- Whether the market is shared by all worlds.
- Price-history aggregation and charts.
- Trading reinforced, charged, or otherwise instanced equipment.
- Market access requirements, memberships, and geographic restrictions.
- Public API or website access.
- Extracting the market into a separate service.

## When a separate service becomes justified

The market can remain in the current Go server while there is one world or a
small beta population. Consider extraction only when several world processes
need the same market, marketplace traffic materially affects gameplay request
latency, independent deployment becomes useful, or a dedicated matching worker
is required.

The database schema and domain boundary should make extraction possible, but
the first implementation should not pay the operational cost of a
microservice.

## Suggested first implementation task

Before implementing UI, build Slice 1 as a PostgreSQL-only spike for a single
stackable item. Prove escrow, partial purchase, cancellation, collection, and a
race between two buyers. Once those invariants hold, generalising the item set
and adding the client interface becomes substantially safer.
