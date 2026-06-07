# AuctionXI — System Architecture

## Overview

AuctionXI is a football auction strategy game where human players compete against algorithmic AI managers to build the strongest squad under budget constraints. The system is designed as a **modular monolith** with clear domain boundaries, enabling future extraction into microservices for real-time multiplayer.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           CLIENT LAYER                                   │
│  Next.js 15 · TypeScript · Zustand · TanStack Query · Framer Motion     │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │ HTTPS / REST (+ SSE for live auction)
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           API GATEWAY LAYER                              │
│  Gin Router · JWT Auth · Rate Limiting · Security Headers · CORS        │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          ▼                       ▼                       ▼
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────────────┐
│  Auth Service   │   │ Auction Service │   │  Progression Service    │
│  Register/Login │   │ Setup/Live/Bids │   │  XP · Coins · Ranks     │
│  JWT + Refresh  │   │ SSE Events      │   │  Leaderboards           │
└────────┬────────┘   └────────┬────────┘   └────────────┬────────────┘
         │                       │                          │
         └───────────────────────┼──────────────────────────┘
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         GAME ENGINE LAYER (Pure Go)                      │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌───────────────┐  │
│  │ Auction      │ │ AI Manager   │ │ Team Eval    │ │ Chemistry     │  │
│  │ Engine       │ │ Engine       │ │ Engine       │ │ Engine        │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └───────────────┘  │
│  ┌──────────────┐ ┌──────────────┐                                      │
│  │ Player Gen   │ │ Scoring      │                                      │
│  │ Engine       │ │ Engine       │                                      │
│  └──────────────┘ └──────────────┘                                      │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        REPOSITORY LAYER (SQLC)                           │
│  Users · Auctions · Bids · Teams · Results · Leaderboards               │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  ▼
┌──────────────────────┐   ┌──────────────────────┐
│     PostgreSQL       │   │        Redis          │
│  Persistent State    │   │  Sessions · Auction   │
│  Migrations (Goose)  │   │  State · Rate Limits  │
└──────────────────────┘   └──────────────────────┘
```

## Design Principles

| Principle | Implementation |
|-----------|----------------|
| **Domain-Driven Design** | Game engines are pure Go packages with zero HTTP/DB dependencies |
| **Clean Architecture** | Dependencies point inward: handlers → services → repositories → domain |
| **Event-Driven Auctions** | Auction state machine emits events consumed by SSE/WebSocket (future) |
| **Deterministic AI** | Seeded PRNG for reproducible AI behavior and testability |
| **Idempotent Operations** | Bid placement uses optimistic locking + version counters |
| **Multiplayer-Ready** | Auction room abstraction with participant slots reserved for humans |

## Domain Boundaries

### Auth Domain
- User registration, login, token refresh, logout
- Password hashing (bcrypt), JWT access + refresh token rotation

### Auction Domain
- Auction lifecycle: `setup → rulebook → live → resolving → completed`
- Player pool generation per auction (unique seed)
- Bid validation, timer management, auto-sell on expiry
- Activity feed event stream

### Team Domain
- Squad composition rules (11–15 players, position minimums)
- Formation assignment (4-3-3, 4-4-2, 3-5-2, 4-2-3-1)
- Budget tracking per participant

### Game Engine Domain (Pure Logic)
- **Player Generator**: Procedural player creation with position-weighted attributes
- **AI Manager Engine**: Difficulty-tiered bidding strategies
- **Auction Engine**: State machine, timer, bid resolution
- **Team Evaluation Engine**: Multi-factor scoring (not rating-only)
- **Chemistry Engine**: Nation/club synergy calculations
- **Scoring Engine**: Aggregates all sub-scores into final team score

### Progression Domain
- XP, coins, rank points on auction completion
- Rank tiers: Bronze → Legend
- Leaderboard aggregation

## Data Flow: Live Auction

```
1. User clicks "Begin Auction"
   → POST /api/v1/auctions/{id}/start
   → AuctionService loads config, generates player pool
   → AuctionEngine initializes state machine
   → First player presented, 15s timer starts

2. Timer tick (server-side, 1s interval)
   → AuctionEngine checks expiry
   → AI Engine evaluates all AI managers in parallel
   → AI bids/passes emitted as events
   → Events pushed to Redis pub/sub → SSE to client

3. User bids
   → POST /api/v1/auctions/{id}/bids
   → Validate: budget, increment rules, auction active
   → Update current bid, reset timer extension (optional)
   → Emit bid event

4. Timer expires
   → Resolve: assign player to highest bidder
   → Deduct budget, add to squad
   → Advance to next player or end auction

5. Auction complete
   → Scoring Engine evaluates all squads
   → Results persisted, progression updated
   → GET /api/v1/auctions/{id}/results
```

## Security Model

- **JWT Access Token**: 15-minute expiry, contains user_id + role
- **Refresh Token**: 7-day expiry, stored hashed in DB, rotation on use
- **Rate Limiting**: Redis sliding window — 100 req/min general, 30 bids/min per auction
- **Input Validation**: go-playground/validator on all request DTOs
- **Security Headers**: HSTS, X-Content-Type-Options, X-Frame-Options via middleware
- **CORS**: Configurable allowed origins

## Scalability Path (V1 → Multiplayer)

| V1 (Current) | V2 (Multiplayer) |
|--------------|------------------|
| Single human + N AI | N humans + optional AI fill |
| SSE for live updates | WebSocket rooms via Redis pub/sub |
| In-process auction timer | Distributed timer via Redis + leader election |
| Monolithic Go service | Extract auction service, add matchmaking |
| Polling/SSE client | Real-time bid sync with conflict resolution |

## Technology Decisions

| Choice | Rationale |
|--------|-----------|
| Go + Gin | High concurrency for auction timers + AI parallel evaluation |
| SQLC | Type-safe SQL, no ORM magic, performance |
| PostgreSQL | ACID for bids/budget, JSONB for squad snapshots |
| Redis | Auction hot state, pub/sub for SSE, rate limiting |
| Next.js 15 App Router | SSR landing, client-side auction UX |
| Zustand | Lightweight auction room state without Redux boilerplate |
| TanStack Query | Server state caching, optimistic bid updates |

## Folder Structure

```
DrayAuction/
├── docs/                          # Architecture & design documents
├── backend/
│   ├── cmd/api/                   # Application entrypoint
│   ├── internal/
│   │   ├── api/                   # Router, handlers, middlewares, DTOs
│   │   ├── domain/                # Entities, value objects, interfaces
│   │   ├── services/              # Application services (orchestration)
│   │   ├── repositories/          # SQLC-backed repository implementations
│   │   └── database/              # SQLC queries, generated code
│   ├── pkg/
│   │   ├── engine/                # Pure game engines
│   │   │   ├── auction/
│   │   │   ├── ai/
│   │   │   ├── player/
│   │   │   ├── scoring/
│   │   │   └── chemistry/
│   │   ├── auth/                  # JWT utilities
│   │   ├── config/                # Environment config
│   │   └── logger/                # Zap wrapper
│   ├── migrations/                # Goose SQL migrations
│   ├── sqlc.yaml
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app/                   # Next.js App Router pages
│   │   ├── components/            # UI components
│   │   ├── lib/                   # API client, utils
│   │   ├── stores/                # Zustand stores
│   │   └── hooks/                 # Custom hooks (SSE, auction)
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```
