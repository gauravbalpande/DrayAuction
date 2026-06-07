# AuctionXI — Team Evaluation & Scoring Engine

## Design Philosophy

**Winner is NOT determined by average player rating alone.** The scoring system rewards strategic squad building: formation fit, positional balance, chemistry, depth, and form. A 78-rated squad with perfect chemistry and formation fit can beat an 85-rated squad with poor balance.

## Scoring Pipeline

```
Squad (11-15 players)
    │
    ├──▶ Formation Fit Score
    ├──▶ Attack Score
    ├──▶ Midfield Score
    ├──▶ Defense Score
    ├──▶ Chemistry Score
    ├──▶ Bench Strength Score
    ├──▶ Current Form Score
    └──▶ Squad Depth Score
              │
              ▼
        Weighted Total Score
              │
              ▼
        Analysis (strengths/weaknesses)
```

## Score Components

### 1. Attack Score (Weight: 20%)

Evaluates starting attackers (formation-dependent count):

```
attack_score = Σ (player_rating × position_weight × form_modifier) / max_possible

Position weights in attack:
  ST/CF: 1.0
  LW/RW: 0.85
  CAM: 0.7

form_modifier = 0.8 + (player.form / 100) × 0.4  → range [0.8, 1.2]
```

### 2. Midfield Score (Weight: 18%)

```
midfield_score = Σ (player_rating × position_weight × form_modifier) / max_possible

Position weights:
  CM: 1.0
  CDM: 0.9
  CAM: 0.85
  LM/RM: 0.8
```

### 3. Defense Score (Weight: 18%)

```
defense_score = Σ (player_rating × position_weight × form_modifier) / max_possible

Position weights:
  CB: 1.0
  LB/RB: 0.85
  GK: 1.2 (GK is critical — low GK rating heavily penalizes)
```

**GK Penalty:** If no GK in squad, defense_score = 0.

### 4. Chemistry Score (Weight: 12%)

See Chemistry Engine below. Range: 0–100.

### 5. Bench Strength Score (Weight: 8%)

Evaluates non-starting XI players (positions 12–15):

```
bench_score = avg(bench_player_ratings) × bench_depth_factor

bench_depth_factor:
  0 bench players: 0.0
  1 bench player:  0.4
  2 bench players: 0.7
  3 bench players: 0.85
  4 bench players: 1.0
```

### 6. Current Form Score (Weight: 10%)

Average form across starting XI:

```
form_score = avg(starting_xi.form) × 10  → range [0, 100]
```

Form is generated per player (60–99) during player generation.

### 7. Formation Fit Score (Weight: 10%)

How well the squad fills the chosen formation's slots:

```
formation_fit = (filled_required_slots / total_required_slots) × avg_slot_fit_quality

Slot fit quality per player:
  Primary position match:   1.0
  Secondary position match: 0.7
  Wrong position:           0.3
```

**Formations and Required Slots:**

| Formation | GK | DEF | MID | ATT |
|-----------|----|----|-----|-----|
| 4-3-3 | 1 | 4 (2CB, LB, RB) | 3 (1CDM, 2CM) | 3 (LW, ST, RW) |
| 4-4-2 | 1 | 4 | 4 (2CM, LM, RM) | 2 (2ST) |
| 3-5-2 | 1 | 3 (3CB) | 5 (2CM, LM, RM, CAM) | 2 |
| 4-2-3-1 | 1 | 4 | 5 (2CDM, CAM, LM, RM) | 1 (ST) |

Auto-selects best-fit formation if participant didn't explicitly choose.

### 8. Squad Depth Score (Weight: 4%)

Position coverage across entire squad:

```
For each position group (GK, DEF, MID, ATT):
  coverage = min(count / ideal_count, 1.0)

ideal_counts: GK=2, DEF=5, MID=5, ATT=4

squad_depth = avg(coverage_scores) × 100
```

## Chemistry Engine

Chemistry is calculated from nation and club links between starting XI players:

```
For each pair of starting XI players (i, j):
  if same_nation:  +3 chemistry points
  if same_club:     +5 chemistry points
  if same_nation AND same_club: +7 (not stacked, max of pair)

chemistry_score = min(total_chemistry / max_possible_chemistry, 1.0) × 100

max_possible_chemistry = C(11, 2) × 7 = 385 (theoretical max, unrealistic)
Normalization: chemistry_score = (total / 150) × 100, capped at 100
```

**Chemistry Thresholds:**
| Score | Label |
|-------|-------|
| 80+ | Excellent |
| 60–79 | Good |
| 40–59 | Average |
| < 40 | Poor |

## Total Score Calculation

```
total = (attack × 0.20)
      + (midfield × 0.18)
      + (defense × 0.18)
      + (chemistry × 0.12)
      + (bench × 0.08)
      + (form × 0.10)
      + (formation_fit × 0.10)
      + (squad_depth × 0.04)

Range: 0–100 (normalized)
```

## Match Analysis Generation

After scoring, generate human-readable analysis:

**Strengths** (auto-detected):
- Any sub-score ≥ 80 → "Strong {category}"
- Chemistry ≥ 75 → "Excellent team chemistry"
- Formation fit ≥ 85 → "Perfect formation alignment"

**Weaknesses** (auto-detected):
- Any sub-score < 40 → "Weak {category}"
- No GK → "No goalkeeper — critical weakness"
- Chemistry < 30 → "Poor team chemistry"
- Bench < 20 → "No bench depth"
- Squad size < 13 → "Thin squad — injury risk"

## Progression Rewards

Based on auction outcome and difficulty:

```
Base rewards:
  Win:  +150 coins, +250 XP, +35 rank_points
  Loss: +50 coins,  +100 XP, +10 rank_points

Difficulty multiplier:
  Easy:      ×0.8
  Medium:    ×1.0
  Hard:      ×1.3
  Legendary: ×1.6

Performance bonus (score-based):
  score > 90: +50 bonus coins, +50 bonus XP
  score > 80: +25 bonus coins, +25 bonus XP
```

## Rank System

| Rank | Points Required |
|------|----------------|
| Bronze | 0 |
| Silver | 500 |
| Gold | 1500 |
| Platinum | 3500 |
| Diamond | 7000 |
| Legend | 12000 |

Rank points accumulate across auctions. Rank never decreases.

## Example Breakdown

```
Team: "manager_xi" — Formation: 4-3-3 — Total: 847.5

  Attack Score:      182.3 / 200  (91.2%)  ★ Strong
  Midfield Score:    165.8 / 180  (92.1%)  ★ Strong
  Defense Score:     158.2 / 180  (87.9%)  ★ Strong
  Chemistry Score:    92.1 / 100  (92.1%)  ★ Excellent
  Bench Strength:     45.6 / 80   (57.0%)  ○ Average
  Current Form:       78.4 / 100  (78.4%)  ○ Good
  Formation Fit:      88.0 / 100  (88.0%)  ★ Strong
  Squad Depth:        37.1 / 100  (37.1%)  ✗ Weak

  Strengths: Strong attack, Excellent chemistry, Strong formation fit
  Weaknesses: Weak bench, Thin squad depth
```
