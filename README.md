# AuctionXI

**AuctionXI** is a production-grade football auction strategy game where human players compete against algorithmic AI managers to build the strongest squad under budget constraints.

> This is NOT a CRUD app. This is NOT fantasy football. This is NOT betting.
> It's a strategy game that rewards squad balance, formation planning, and player valuation.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, TypeScript, TailwindCSS, Shadcn UI, Zustand, TanStack Query, Framer Motion |
| Backend | Go 1.24+, Gin, Clean Architecture, JWT, PostgreSQL, Redis, SQLC, Goose, Zap |
| Game Engines | Auction, AI Manager, Player Generator, Scoring, Chemistry |
| DevOps | Docker, Docker Compose, Makefile |

## Quick Start

```bash
# Clone and start all services
make dev

# Or run individually:
make backend-dev    # Go API on :8080
make frontend-dev   # Next.js on :3000
make backend-test   # Run Go unit tests
make migrate        # Run database migrations
```

## Project Structure

```
DrayAuction/
├── docs/                    # Architecture, API, engine designs, roadmap
├── backend/
│   ├── cmd/api/             # Application entrypoint
│   ├── internal/
│   │   ├── api/             # Handlers, middleware, DTOs, router
│   │   ├── domain/          # Entities, interfaces, errors
│   │   ├── services/        # Application services
│   │   └── repositories/    # Data access (memory + SQLC)
│   ├── pkg/engine/          # Pure game engines (zero HTTP/DB deps)
│   │   ├── auction/         # State machine, timer, bid resolution
│   │   ├── ai/              # 4-tier algorithmic AI managers
│   │   ├── player/          # Procedural player generation
│   │   ├── scoring/         # Multi-factor team evaluation
│   │   └── chemistry/       # Nation/club synergy
│   └── migrations/          # Goose SQL migrations
├── frontend/
│   └── src/
│       ├── app/             # Next.js App Router pages
│       ├── components/      # UI components
│       ├── lib/             # API client, utilities
│       └── stores/          # Zustand state management
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Core User Flow

1. **Landing Page** → Register / Login
2. **Dashboard** → Profile, stats, START AUCTION
3. **Auction Setup** → AI opponents, difficulty, budget, pool size
4. **Rulebook** → Rules, squad requirements, scoring system
5. **Live Auction** → Real-time bidding with 15s timer, AI opponents, activity feed
6. **Results** → Score breakdown, strengths/weaknesses, rewards

## Game Engines

### Player Generation
Every auction generates a unique random player pool using seeded PRNG. No two auctions feel identical.

### AI Managers (4 Difficulty Tiers)
- **Easy**: Random bidding (tutorial opponent)
- **Medium**: Position-aware squad building
- **Hard**: Budget + squad aware with value assessment
- **Legendary**: Full optimization — formation planning, chemistry, market evaluation

### Team Scoring (NOT rating-only)
8 weighted factors: Attack, Midfield, Defense, Chemistry, Bench Strength, Form, Formation Fit, Squad Depth.

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | System design, data flow, scalability path |
| [API Spec](docs/API.md) | REST API endpoints and schemas |
| [AI Engine](docs/AI_ENGINE.md) | AI manager design and difficulty tiers |
| [Auction Engine](docs/AUCTION_ENGINE.md) | State machine, timer, bid validation |
| [Scoring Engine](docs/SCORING_ENGINE.md) | Team evaluation and chemistry |
| [Roadmap](docs/ROADMAP.md) | MVP phases and multiplayer roadmap |

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
DATABASE_URL=postgres://auctionxi:auctionxi@localhost:5432/auctionxi?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_ACCESS_SECRET=your-secret-here
JWT_REFRESH_SECRET=your-secret-here
CORS_ORIGIN=http://localhost:3000
```

## License

Private — Portfolio project.
