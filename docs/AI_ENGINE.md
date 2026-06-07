# AuctionXI — AI Manager Engine Design

## Philosophy

AI managers use **deterministic algorithmic strategies** — no LLM APIs. Each manager has a personality that influences bidding behavior within its difficulty tier. All decisions are reproducible given the same seed and game state.

## Architecture

```
AIManagerEngine
├── ManagerProfile (personality, strategy weights)
├── DifficultyStrategy (interface)
│   ├── EasyStrategy
│   ├── MediumStrategy
│   ├── HardStrategy
│   └── LegendaryStrategy
├── BidEvaluator
│   ├── ValueAssessment
│   ├── SquadNeedAnalysis
│   └── BudgetGuard
└── DecisionEmitter → AuctionEvent
```

## Manager Personalities

Each AI manager is assigned a personality at auction creation:

| Personality | Behavior Bias |
|-------------|---------------|
| **Aggressive** | Overpays for star players, early bidding |
| **Patient** | Waits for timer pressure, snipes late |
| **Balanced** | Even spending across positions |
| **Youth Focused** | Prefers high-potential lower-rated players |
| **Star Hunter** | Targets 85+ rated players exclusively |
| **Budget Manager** | Maximizes value per coin spent |

Personality modifies base strategy weights, not the difficulty tier itself.

## Difficulty Tiers

### Easy — Random Bidding

```
Strategy: RandomBidStrategy
- 40% chance to bid random increment (+5M, +10M, +20M)
- 60% chance to pass
- No awareness of position, budget, or squad needs
- May bid above market value randomly
- May exhaust budget early on mediocre players
```

**Purpose:** Tutorial opponent. Teaches auction mechanics without strategic pressure.

### Medium — Position Aware

```
Strategy: PositionAwareStrategy
- Evaluates current squad composition
- Identifies position gaps (missing GK, weak defense, etc.)
- Bids on players filling needed positions
- Bid amount: market_value * random(0.8, 1.2)
- Passes on positions already well-staffed
- No budget optimization — may overspend on needed positions
```

**Position Priority Weights:**
| Squad Gap | Priority Multiplier |
|-----------|-------------------|
| Missing GK | 3.0x |
| < 2 Defenders | 2.5x |
| < 2 Midfielders | 2.0x |
| < 2 Attackers | 2.0x |
| Position full | 0.3x |

### Hard — Budget & Squad Aware

```
Strategy: BudgetSquadStrategy
- All Medium capabilities plus:
- Tracks remaining budget vs. remaining players in pool
- Calculates max_bid_per_remaining = budget / (15 - squad_size)
- Won't bid more than 1.5x max_bid_per_remaining
- Evaluates player value: rating / market_value ratio
- Prefers high value-per-coin players
- Adjusts aggression based on remaining pool size
- Late-auction: increases bid frequency as pool shrinks
```

**Budget Guard Rules:**
```
max_single_bid = min(
    remaining_budget - (min_remaining_players * avg_market_value * 0.5),
    market_value * 1.3
)
```

### Legendary — Full Optimization

```
Strategy: LegendaryStrategy
- All Hard capabilities plus:
- Formation planning: selects target formation early, bids for fit
- Squad optimization: evaluates marginal team score improvement per bid
- Market value evaluation: precise fair value calculation
- Opponent modeling: estimates user bidding patterns from history
- Endgame planning: reserves budget for final high-value targets
- Chemistry awareness: prefers players matching existing squad nations/clubs
```

**Legendary Decision Pipeline:**
```
1. Calculate marginal_score_if_acquired(player)
2. Calculate fair_value = f(rating, position, age, form)
3. Calculate max_willing_to_pay = fair_value * personality_multiplier
4. If current_bid < max_willing_to_pay AND marginal_score > threshold:
     bid = min(current_bid + optimal_increment, max_willing_to_pay)
5. Else: pass
```

**Formation Planning:**
- At auction start, select target formation based on personality
- Weight player bids by formation slot fit score
- Example: 4-3-3 target → prioritize wingers and a CAM

## AI Manager Generation

At auction creation, N AI managers are generated:

```go
type AIManager struct {
    ID          uuid.UUID
    Name        string          // Generated from name pool
    Personality Personality
    Difficulty  Difficulty
    Strategy    BidStrategy     // Assigned by difficulty
    Budget      int64
    Squad       []Player
    Formation   Formation
}
```

**Name Pool:** 50+ fictional manager names (no real person likeness).

## Parallel Evaluation

When the auction timer ticks, all AI managers evaluate in parallel:

```
goroutine per AI manager:
  1. Receive current auction state snapshot
  2. Run strategy.Evaluate(state, manager)
  3. Return BidDecision or PassDecision
  4. Auction engine processes decisions in timestamp order
```

AI decisions for the same tick are processed with a small random delay (50–200ms) to simulate human-like timing.

## Testing Strategy

- **Unit tests:** Each strategy tested in isolation with fixture game states
- **Simulation tests:** Run 1000 auctions per difficulty, verify win rates:
  - Easy: Human wins ~85%
  - Medium: Human wins ~65%
  - Hard: Human wins ~45%
  - Legendary: Human wins ~25%
- **Determinism tests:** Same seed → same AI decisions

## Configuration

```go
type AIConfig struct {
    EasyBidProbability     float64 // 0.40
    MediumPositionWeight   float64 // 2.0
    HardBudgetMultiplier   float64 // 1.5
    LegendaryScoreThreshold float64 // 5.0
    DecisionDelayMinMs     int     // 50
    DecisionDelayMaxMs     int     // 200
}
```
