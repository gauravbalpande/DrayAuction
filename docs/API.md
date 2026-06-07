# AuctionXI — API Specification

Base URL: `/api/v1`

All authenticated endpoints require `Authorization: Bearer <access_token>`.

## Authentication

### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "username": "manager_xi",
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `201 Created`
```json
{
  "user": {
    "id": "uuid",
    "username": "manager_xi",
    "email": "user@example.com",
    "coins": 1000,
    "xp": 0,
    "rank": "bronze",
    "wins": 0,
    "losses": 0
  },
  "access_token": "jwt...",
  "refresh_token": "jwt..."
}
```

### POST /auth/login
**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK` — same shape as register.

### POST /auth/refresh
**Request:**
```json
{
  "refresh_token": "jwt..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "jwt...",
  "refresh_token": "jwt..."
}
```

### POST /auth/logout
**Headers:** Authorization required.

**Response:** `204 No Content`

---

## User Profile

### GET /users/me
**Response:** `200 OK`
```json
{
  "id": "uuid",
  "username": "manager_xi",
  "email": "user@example.com",
  "coins": 1250,
  "xp": 3400,
  "rank": "silver",
  "rank_points": 450,
  "wins": 12,
  "losses": 8,
  "created_at": "2026-01-15T10:00:00Z"
}
```

### GET /users/me/auctions
Recent auctions for the authenticated user.

**Query:** `?limit=10&offset=0`

**Response:** `200 OK`
```json
{
  "auctions": [
    {
      "id": "uuid",
      "status": "completed",
      "difficulty": "hard",
      "budget": 1000000000,
      "player_pool_size": 50,
      "ai_opponents": 3,
      "result": "win",
      "score": 847.5,
      "created_at": "2026-06-01T14:30:00Z"
    }
  ],
  "total": 42
}
```

---

## Auctions

### POST /auctions
Create a new auction (setup phase).

**Request:**
```json
{
  "ai_opponents": 3,
  "difficulty": "hard",
  "budget": 1000000000,
  "player_pool_size": 50
}
```

| Field | Values |
|-------|--------|
| ai_opponents | 1–5 |
| difficulty | easy, medium, hard, legendary |
| budget | 500000000, 750000000, 1000000000 |
| player_pool_size | 40, 50, 60 |

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "status": "setup",
  "ai_opponents": 3,
  "difficulty": "hard",
  "budget": 1000000000,
  "player_pool_size": 50,
  "ai_managers": [
    {
      "id": "uuid",
      "name": "Guardiola Jr.",
      "personality": "possession",
      "difficulty": "hard"
    }
  ],
  "created_at": "2026-06-06T10:00:00Z"
}
```

### GET /auctions/:id
Get auction details including current state.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "status": "live",
  "difficulty": "hard",
  "budget": 1000000000,
  "player_pool_size": 50,
  "current_player_index": 12,
  "current_player": {
    "id": "uuid",
    "name": "Marcus Thorne",
    "position": "CM",
    "club": "Northbridge FC",
    "nation": "England",
    "rating": 84,
    "market_value": 45000000
  },
  "current_bid": 65000000,
  "highest_bidder": "ai-manager-uuid",
  "highest_bidder_name": "Guardiola Jr.",
  "timer_seconds": 12,
  "participants": [
    {
      "id": "uuid",
      "name": "You",
      "type": "human",
      "remaining_budget": 720000000,
      "squad_size": 8
    }
  ]
}
```

### POST /auctions/:id/start
Begin the live auction after rulebook acknowledgment.

**Response:** `200 OK` — returns full live auction state.

### POST /auctions/:id/bids
Place a bid on the current player.

**Request:**
```json
{
  "amount": 70000000
}
```

Or use increment shortcuts via query: `?increment=5|10|20` (millions).

**Response:** `200 OK`
```json
{
  "accepted": true,
  "current_bid": 70000000,
  "highest_bidder": "user-uuid",
  "timer_seconds": 15,
  "remaining_budget": 650000000
}
```

**Errors:**
- `400` — bid too low, insufficient budget, auction not live
- `409` — bid conflict (another bid placed simultaneously)

### POST /auctions/:id/pass
Pass on the current player.

**Response:** `200 OK`
```json
{
  "passed": true
}
```

### GET /auctions/:id/events
Server-Sent Events stream for live auction updates.

**Events:**
```
event: player_presented
data: {"player": {...}, "timer_seconds": 15}

event: bid_placed
data: {"bidder_id": "...", "bidder_name": "...", "amount": 65000000}

event: player_passed
data: {"participant_id": "...", "participant_name": "..."}

event: player_sold
data: {"player": {...}, "buyer_id": "...", "amount": 70000000}

event: auction_completed
data: {"auction_id": "..."}

event: timer_tick
data: {"seconds_remaining": 8}
```

### GET /auctions/:id/results
Get auction results and team analysis.

**Response:** `200 OK`
```json
{
  "auction_id": "uuid",
  "winner_id": "user-uuid",
  "winner_name": "manager_xi",
  "teams": [
    {
      "participant_id": "uuid",
      "name": "manager_xi",
      "type": "human",
      "formation": "4-3-3",
      "total_score": 847.5,
      "breakdown": {
        "attack_score": 182.3,
        "midfield_score": 165.8,
        "defense_score": 158.2,
        "chemistry_score": 92.1,
        "bench_strength": 45.6,
        "current_form": 78.4,
        "formation_fit": 88.0,
        "squad_depth": 37.1
      },
      "strengths": ["Strong midfield control", "Excellent chemistry"],
      "weaknesses": ["Weak bench", "No backup GK"],
      "players": [...]
    }
  ],
  "rewards": {
    "coins": 150,
    "xp": 250,
    "rank_points": 35
  }
}
```

---

## Leaderboards

### GET /leaderboards
**Query:** `?period=weekly|alltime&limit=50`

**Response:** `200 OK`
```json
{
  "entries": [
    {
      "rank": 1,
      "user_id": "uuid",
      "username": "top_manager",
      "rank_tier": "diamond",
      "total_score": 9850,
      "wins": 45,
      "win_rate": 0.72
    }
  ]
}
```

---

## Error Format

All errors follow a consistent structure:

```json
{
  "error": {
    "code": "INSUFFICIENT_BUDGET",
    "message": "Bid amount exceeds remaining budget",
    "details": {}
  }
}
```

### Standard HTTP Status Codes

| Code | Usage |
|------|-------|
| 400 | Validation errors, business rule violations |
| 401 | Missing or expired token |
| 403 | Forbidden action |
| 404 | Resource not found |
| 409 | Conflict (concurrent bid) |
| 429 | Rate limit exceeded |
| 500 | Internal server error |
