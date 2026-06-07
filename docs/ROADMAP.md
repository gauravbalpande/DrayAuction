# AuctionXI — Development Roadmap

## MVP Implementation Order

### Phase 1: Foundation (Week 1–2)
**Goal:** Runnable project skeleton with auth and database.

| # | Task | Priority |
|---|------|----------|
| 1.1 | Docker Compose (Postgres, Redis, API, Frontend) | P0 |
| 1.2 | Database migrations (all tables) | P0 |
| 1.3 | Go project scaffold (clean architecture) | P0 |
| 1.4 | Auth: register, login, JWT, refresh tokens | P0 |
| 1.5 | Next.js scaffold with dark theme + Shadcn | P0 |
| 1.6 | Landing page (hero, features, CTA) | P0 |
| 1.7 | Auth pages (register, login) | P0 |

**Exit criteria:** User can register, login, see dashboard shell.

### Phase 2: Game Engines (Week 2–3)
**Goal:** Core game logic fully tested, no UI yet.

| # | Task | Priority |
|---|------|----------|
| 2.1 | Player generation engine | P0 |
| 2.2 | Auction state machine | P0 |
| 2.3 | AI manager engine (all 4 difficulties) | P0 |
| 2.4 | Chemistry engine | P0 |
| 2.5 | Team evaluation / scoring engine | P0 |
| 2.6 | Unit tests for all engines (>80% coverage) | P0 |
| 2.7 | Simulation tests (AI win rate validation) | P1 |

**Exit criteria:** All engines pass unit tests. Simulation shows expected win rates.

### Phase 3: Auction Backend (Week 3–4)
**Goal:** Full auction lifecycle via API.

| # | Task | Priority |
|---|------|----------|
| 3.1 | Auction CRUD endpoints | P0 |
| 3.2 | Auction start + player pool generation | P0 |
| 3.3 | Bid/pass endpoints with validation | P0 |
| 3.4 | Timer goroutine + AI decision loop | P0 |
| 3.5 | SSE event stream | P0 |
| 3.6 | Player resolution + squad management | P0 |
| 3.7 | Results endpoint + scoring integration | P0 |
| 3.8 | Redis hot state management | P0 |
| 3.9 | Integration tests for full auction flow | P0 |

**Exit criteria:** Complete auction playable via API calls / curl.

### Phase 4: Auction Frontend (Week 4–5)
**Goal:** Beautiful live auction experience.

| # | Task | Priority |
|---|------|----------|
| 4.1 | Dashboard (profile, stats, recent auctions) | P0 |
| 4.2 | Auction setup page | P0 |
| 4.3 | Rulebook screen | P0 |
| 4.4 | Live auction screen (the hero feature) | P0 |
| 4.5 | SSE hook for real-time updates | P0 |
| 4.6 | Bid controls (+5M, +10M, +20M, Pass) | P0 |
| 4.7 | Activity feed component | P0 |
| 4.8 | Timer countdown animation | P0 |
| 4.9 | Results screen with score breakdown | P0 |
| 4.10 | Framer Motion animations throughout | P1 |

**Exit criteria:** Full user flow from landing → register → auction → results.

### Phase 5: Progression & Polish (Week 5–6)
**Goal:** Game loop complete with rewards and rankings.

| # | Task | Priority |
|---|------|----------|
| 5.1 | Progression system (coins, XP, rank points) | P0 |
| 5.2 | Rank tier display and progression | P0 |
| 5.3 | Leaderboard page | P1 |
| 5.4 | Landing page rankings preview | P1 |
| 5.5 | Rate limiting middleware | P0 |
| 5.6 | Error handling polish | P0 |
| 5.7 | Loading states and empty states | P1 |
| 5.8 | Mobile responsive design pass | P1 |
| 5.9 | E2E test for complete user journey | P1 |

**Exit criteria:** MVP launch-ready. Full game loop with progression.

---

## Future Multiplayer Roadmap

### V2.0 — Async Multiplayer (Month 2–3)

| Feature | Description |
|---------|-------------|
| **Friend Challenges** | Invite a friend to the same auction config, play asynchronously |
| **Auction Replays** | Watch replay of any completed auction bid-by-bid |
| **Custom Rules** | User-defined timer, budget, pool size |
| **Manager Profiles** | Public profile pages with stats and history |

**Architecture changes:**
- Auction room abstraction with participant slots
- Turn-based bidding with notification system
- Auction history stored as event log (event sourcing lite)

### V2.5 — Real-Time Multiplayer (Month 3–5)

| Feature | Description |
|---------|-------------|
| **Live Multiplayer Auctions** | 2–6 human players in same auction room |
| **WebSocket Rooms** | Replace SSE with bidirectional WebSocket |
| **Matchmaking Queue** | Join queue by skill rating, auto-matched auctions |
| **Spectator Mode** | Watch live auctions without participating |
| **Chat System** | In-auction emoji reactions (not full chat — avoid toxicity) |

**Architecture changes:**
```
                    ┌─────────────────┐
                    │  Matchmaking    │
                    │  Service        │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Auction Room   │
                    │  Manager        │
                    │  (Redis-backed) │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ WS Conn 1│  │ WS Conn 2│  │ WS Conn N│
        └──────────┘  └──────────┘  └──────────┘
```

- Extract auction service into standalone process
- Redis pub/sub for cross-instance event fanout
- Distributed timer with Redis keyspace notifications
- Conflict resolution: last-write-wins with version vectors
- Presence system: track connected participants

### V3.0 — Competitive Platform (Month 5–8)

| Feature | Description |
|---------|-------------|
| **Ranked Seasons** | Monthly ranked seasons with reset and rewards |
| **Tournaments** | Bracket-style elimination tournaments |
| **Season Pass** | Premium progression track with cosmetic rewards |
| **Team Badges & Flair** | Cosmetic customization for profiles |
| **Advanced Analytics** | Bid efficiency, value extraction metrics |
| **AI Coach** | Post-auction AI analysis suggesting improvements |

**Architecture changes:**
- Extract scoring service for async post-auction analysis
- Tournament bracket engine
- Season management service
- Analytics pipeline (PostgreSQL → aggregation tables)
- CDN for static assets and replays

### V3.5 — Platform Scale (Month 8–12)

| Feature | Description |
|---------|-------------|
| **Mobile App** | React Native companion app |
| **API for Third Parties** | Public API for stats and replays |
| **Admin Dashboard** | Moderation, analytics, config management |
| **Regional Leaderboards** | Geo-based rankings |
| **Anti-Cheat** | Bid timing analysis, pattern detection |

**Architecture changes:**
- Kubernetes deployment with horizontal pod autoscaling
- Read replicas for PostgreSQL
- Redis Cluster for auction state
- Event bus (NATS/Kafka) for cross-service communication
- Observability stack (Prometheus, Grafana, Jaeger)

---

## Technical Debt Schedule

| Sprint | Debt Item |
|--------|-----------|
| Post-MVP | Add WebSocket alongside SSE (abstract transport layer) |
| Post-MVP | Event sourcing for auction history |
| V2.0 | Database connection pooling optimization |
| V2.5 | Load testing auction concurrency (target: 100 simultaneous) |
| V3.0 | Extract game engines into shared library |

## Success Metrics

| Metric | MVP Target | V2 Target |
|--------|-----------|-----------|
| Auction completion rate | > 80% | > 90% |
| Avg auction duration | 8–12 min | 8–12 min |
| DAU | 50 | 500 |
| User retention (D7) | 20% | 35% |
| API p99 latency | < 200ms | < 100ms |
| Auction event delivery | < 500ms | < 100ms |
