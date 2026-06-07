# AuctionXI Auction Room Redesign

## 1. Updated UI Wireframe

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AUCTION ROOM · Competitive · 800M Budget          Player 12/50    ⏱ 10s   │
│  ████████████░░░░░░░░  (soft timer — resets on bid)                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                    ┌─────────────────────────────┐                          │
│         AI MGR 1   │      JUDE BELLINGHAM        │   AI MGR 2              │
│    ┌──────────┐    │  CM · Real Madrid · England │    ┌──────────┐          │
│    │ Guardiola│    │  Rating 89 · Value 120M     │    │Mourinho  │          │
│    │ 420M · 5 │    │  ── Scouting Report ──      │    │ 380M · 6 │          │
│    └──────────┘    │  ATK 82 PAS 88 DEF 72 PHY 85│    └──────────┘          │
│         ▲ glow     │  Form: Excellent            │         ▲               │
│                    └─────────────────────────────┘                          │
│                              YOU (center)                                   │
│                         ┌──────────────┐                                    │
│                         │  🔥 You · 520M│                                   │
│                         │  8 players    │                                   │
│                         └──────────────┘                                    │
│                              AI MGR 3                                       │
│                         ┌──────────┐                                        │
│                         │ Kloppstein│                                       │
│                         │ 410M · 4  │                                       │
│                         └──────────┘                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│  CURRENT BID          HIGHEST BIDDER           BID HISTORY                  │
│  148M                 🔥 You                  You 148M · Guardiola 140M   │
├─────────────────────────────────────────────────────────────────────────────┤
│  [ +5M ]  [ +10M ]  [ +20M ]  [ PASS ]                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  LIVE FEED (newest first, animated)                                         │
│  ⚡ Guardiola Jr bid 140M                                                     │
│  🔥 You bid 148M                                                             │
│  🎯 Rodri presented — market value 95M                                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 2. Component Architecture

```
frontend/src/
├── app/auction/
│   ├── setup/page.tsx              # AI count + difficulty only
│   └── [id]/
│       ├── rulebook/page.tsx
│       ├── live/page.tsx           # Orchestrates room layout
│       └── results/page.tsx        # Fixed polling + transition
├── components/auction/
│   ├── AuctionRoom.tsx             # Virtual room layout
│   ├── ManagerCard.tsx             # Personality, style, targets, glow
│   ├── PlayerSpotlight.tsx         # Center player + scouting report
│   ├── BidPanel.tsx                # Current bid, highest bidder, history
│   ├── SoftTimer.tsx               # Animated soft timer bar
│   ├── BidControls.tsx             # +5M / +10M / +20M / Pass
│   └── LiveFeed.tsx                # Emoji feed with Framer Motion
├── hooks/
│   └── useAuctionRoom.ts           # Polling, events, bid state
└── stores/index.ts                 # Extended auction store
```

## 3. Updated Auction State Machine

```
LIVE (player N)
  │
  ├─ timer starts at 15s
  │
  ├─ on BID (any manager)
  │    ├─ validate bid
  │    ├─ update current_bid, highest_bidder
  │    ├─ append bid_history
  │    ├─ RESET timer → 10s          ← soft timer
  │    └─ emit bid_placed event
  │
  ├─ every 1s: timer_tick
  ├─ every 2s: AI evaluation cycle (during live, not after expiry)
  │
  └─ timer == 0
       ├─ RESOLVING → assign player
       ├─ emit player_sold / player_unsold
       └─ next player → timer = 15s OR auction COMPLETED
            └─ finalize results (sync, before marking complete)
```

## 4. Required Backend Changes

| Area | Change |
|------|--------|
| Timer | Soft reset: 15s initial, 10s on bid |
| Auction loop | Continuous 1s ticker; AI during live phase |
| Results bug | Sync finalize; GetResults lazy-compute fallback |
| Players | Real footballer database (Top 5 leagues) |
| Config | Auto budget/pool from difficulty |
| Managers | Style, targets, personality profiles |
| Feed | Rich emoji messages on all events |
| Scoring | Budget efficiency + star-hoard penalty |
| API | bid_history, manager profiles, scouting attrs |

## 5. Required Frontend Changes

| Area | Change |
|------|--------|
| Setup | Remove budget/pool; show auto auction type preview |
| Live page | Full auction room layout |
| Polling | 1s poll; sync feed from server events |
| Results | Poll state until completed + results ready |
| Animations | Manager glow, bid pulse, feed slide-in |

## 6. Implementation Roadmap (Priority)

| P | Task | Impact |
|---|------|--------|
| P0 | Soft timer + auction loop rewrite | Core auction feel |
| P0 | Results bug fix (sync finalize) | Unblocks completion |
| P0 | Real player database | Credibility |
| P1 | Auto budget from difficulty | Game balance |
| P1 | Auction room UI + manager cards | AAA feel |
| P1 | Bid history + highest bidder | Competition clarity |
| P2 | Live feed emojis + animations | Excitement |
| P2 | Scouting report UI | Strategic depth |
| P2 | Scoring: budget efficiency | Balanced teams win |
| P3 | SSE hook (replace polling) | Real-time polish |
