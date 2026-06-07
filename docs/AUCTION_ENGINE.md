# AuctionXI — Auction Engine Design

## Overview

The Auction Engine is a **finite state machine** that orchestrates the live auction experience. It manages player presentation order, bid validation, timer countdown, AI decision cycles, and player resolution.

## State Machine

```
                    ┌─────────┐
                    │  SETUP  │
                    └────┬────┘
                         │ start()
                         ▼
                    ┌─────────┐
              ┌────│ RULEBOOK │────┐
              │    └────┬────┘    │
              │         │ begin() │
              │         ▼         │
              │    ┌─────────┐    │
              │    │  LIVE   │◄───┘ (next player)
              │    └────┬────┘
              │         │
              │    ┌────┴────┐
              │    │         │
              │  bid/pass  timer_expired
              │    │         │
              │    ▼         ▼
              │ ┌──────────────┐
              │ │  RESOLVING   │
              │ └──────┬───────┘
              │        │
              │   more players?
              │    │         │
              │   yes        no
              │    │         │
              │    ▼         ▼
              │  LIVE    ┌───────────┐
              │          │ COMPLETED │
              │          └───────────┘
              │
              └── cancel() → CANCELLED
```

## Core Types

```go
type AuctionState struct {
    ID               uuid.UUID
    Status           AuctionStatus
    Config           AuctionConfig
    PlayerPool       []GeneratedPlayer
    CurrentIndex     int
    CurrentBid       int64
    HighestBidder    *uuid.UUID
    TimerSeconds     int
    Participants     []Participant
    ActivityFeed     []ActivityEvent
    Seed             int64  // For deterministic player gen + AI
}

type AuctionConfig struct {
    Budget          int64
    PlayerPoolSize  int
    AIOpponents     int
    Difficulty      Difficulty
    TimerPerPlayer  int  // 15 seconds
    MinBidIncrement int64 // 5_000_000
}

type Participant struct {
    ID              uuid.UUID
    Name            string
    Type            ParticipantType // human | ai
    RemainingBudget int64
    Squad           []SquadPlayer
    HasPassed       bool // on current player
}
```

## Timer System

```
Server-side timer (authoritative):
- Started when player is presented
- Default: 15 seconds
- Tick interval: 1 second
- On each tick: emit timer_tick event via SSE
- On expiry: transition to RESOLVING

Timer does NOT reset on bids (V1 simplicity).
Future: optional 5-second extension on new bids.
```

## Bid Validation Rules

| Rule | Validation |
|------|-----------|
| Auction must be LIVE | Status check |
| Participant hasn't passed | HasPassed == false |
| Bid > current bid | Strict greater-than |
| Bid ≤ remaining budget | Budget check |
| Minimum increment | bid ≥ current_bid + 5M (or 0 if no current bid) |
| Max squad size | squad_size < 15 (can't bid if full) |
| Concurrent safety | Optimistic lock on auction version |

## Bid Increments

| Action | Calculation |
|--------|------------|
| +5M | current_bid + 5,000,000 (or market_value if first bid) |
| +10M | current_bid + 10,000,000 |
| +20M | current_bid + 20,000,000 |
| Custom | Any valid amount ≥ current + 5M |

First bid on a player: minimum is the player's `market_value`.

## Player Resolution

When timer expires:

```
1. If highest_bidder != nil:
     - Transfer player to winner's squad
     - Deduct bid amount from winner's budget
     - Emit player_sold event
   Else:
     - Player goes unsold (removed from pool)
     - Emit player_unsold event

2. Reset: current_bid = 0, highest_bidder = nil, all has_passed = false

3. If current_index + 1 < pool_size AND any participant can still bid:
     - Advance to next player
     - Start new timer
   Else:
     - Transition to COMPLETED
     - Trigger scoring engine
```

## Early Auction End Conditions

Auction ends when ANY of:
- All players in pool have been presented
- All participants have 15 players (max squad)
- All participants are budget-exhausted (can't afford minimum bid on any remaining player)

## Activity Feed

Every action generates an activity event:

```go
type ActivityEvent struct {
    Timestamp   time.Time
    Type        ActivityType
    Message     string      // Human-readable
    Participant string
    Amount      *int64      // For bids
    PlayerName  string
}

// Examples:
// {type: "bid", message: "You bid 70M", participant: "You", amount: 70000000}
// {type: "pass", message: "Guardiola Jr. passed", participant: "Guardiola Jr."}
// {type: "sold", message: "Marcus Thorne sold to You for 70M", ...}
```

Feed is stored in Redis (last 100 events) and streamed via SSE.

## Player Pool Generation

Each auction generates a unique pool:

```
1. Seed = hash(auction_id + timestamp)
2. PRNG seeded with seed
3. Generate player_pool_size players:
   - Position distribution: 4 GK, 12 DEF, 14 MID, 10 ATT (for pool of 40)
   - Scale proportionally for 50/60
   - Each player: random name, club, nation, rating (65-92), market_value
4. Shuffle presentation order
5. Store in auction_players table
```

## Redis Hot State

During live auction, state is cached in Redis:

```
Key: auction:{id}:state     → JSON serialized AuctionState
Key: auction:{id}:timer     → TTL key expiring in timer_seconds
Key: auction:{id}:events    → List of recent activity events
Channel: auction:{id}:live  → Pub/sub for SSE fanout
```

On auction completion, state is persisted to PostgreSQL and Redis keys are cleaned up.

## Concurrency Model

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  HTTP Bid   │────▶│  Bid Mutex   │────▶│ Update State│
│  Request    │     │  (per auction)│     │ + Emit Event│
└─────────────┘     └──────────────┘     └─────────────┘

┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Timer Tick │────▶│  AI Evaluate │────▶│ Process Bids│
│  (1s loop)  │     │  (parallel)  │     │ + Emit Event│
└─────────────┘     └──────────────┘     └─────────────┘
```

- One goroutine per live auction handles timer + AI
- Bid requests acquire a mutex on the auction ID
- Version counter prevents stale bid overwrites

## Error Recovery

- If timer goroutine crashes: Redis TTL key triggers resolution on expiry
- If server restarts mid-auction: Load state from Redis, resume timer
- Orphaned auctions (> 30 min inactive): Background job marks as cancelled, refunds nothing (players already sold are kept)
