package domain

import (
	"time"

	"github.com/google/uuid"
)

type RankTier string

const (
	RankBronze   RankTier = "bronze"
	RankSilver   RankTier = "silver"
	RankGold     RankTier = "gold"
	RankPlatinum RankTier = "platinum"
	RankDiamond  RankTier = "diamond"
	RankLegend   RankTier = "legend"
)

type Difficulty string

const (
	DifficultyEasy      Difficulty = "easy"
	DifficultyMedium    Difficulty = "medium"
	DifficultyHard      Difficulty = "hard"
	DifficultyLegendary Difficulty = "legendary"
)

type AuctionStatus string

const (
	AuctionStatusSetup     AuctionStatus = "setup"
	AuctionStatusRulebook  AuctionStatus = "rulebook"
	AuctionStatusLive      AuctionStatus = "live"
	AuctionStatusResolving AuctionStatus = "resolving"
	AuctionStatusCompleted AuctionStatus = "completed"
	AuctionStatusCancelled AuctionStatus = "cancelled"
)

type ParticipantType string

const (
	ParticipantHuman ParticipantType = "human"
	ParticipantAI    ParticipantType = "ai"
)

type Personality string

const (
	PersonalityAggressive    Personality = "aggressive"
	PersonalityPatient       Personality = "patient"
	PersonalityBalanced      Personality = "balanced"
	PersonalityYouthFocused  Personality = "youth_focused"
	PersonalityStarHunter    Personality = "star_hunter"
	PersonalityBudgetManager Personality = "budget_manager"
)

type Position string

const (
	PosGK  Position = "GK"
	PosCB  Position = "CB"
	PosLB  Position = "LB"
	PosRB  Position = "RB"
	PosCDM Position = "CDM"
	PosCM  Position = "CM"
	PosCAM Position = "CAM"
	PosLM  Position = "LM"
	PosRM  Position = "RM"
	PosLW  Position = "LW"
	PosRW  Position = "RW"
	PosST  Position = "ST"
	PosCF  Position = "CF"
)

type PositionGroup string

const (
	GroupGK  PositionGroup = "GK"
	GroupDEF PositionGroup = "DEF"
	GroupMID PositionGroup = "MID"
	GroupATT PositionGroup = "ATT"
)

type Formation string

const (
	Formation433  Formation = "4-3-3"
	Formation442  Formation = "4-4-2"
	Formation352  Formation = "3-5-2"
	Formation4231 Formation = "4-2-3-1"
)

func (p Position) Group() PositionGroup {
	switch p {
	case PosGK:
		return GroupGK
	case PosCB, PosLB, PosRB:
		return GroupDEF
	case PosCDM, PosCM, PosCAM, PosLM, PosRM:
		return GroupMID
	case PosLW, PosRW, PosST, PosCF:
		return GroupATT
	default:
		return GroupMID
	}
}

type User struct {
	ID         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Coins      int64     `json:"coins"`
	XP         int64     `json:"xp"`
	RankPoints int       `json:"rank_points"`
	RankTier   RankTier  `json:"rank_tier"`
	Wins       int       `json:"wins"`
	Losses     int       `json:"losses"`
	CreatedAt  time.Time `json:"created_at"`
}

type GeneratedPlayer struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Position          Position  `json:"position"`
	SecondaryPosition Position  `json:"secondary_position,omitempty"`
	Club              string    `json:"club"`
	Nation            string    `json:"nation"`
	League            string    `json:"league"`
	Rating            int       `json:"rating"`
	MarketValue       int64     `json:"market_value"`
	Form              int       `json:"form"`
	FormLabel         string    `json:"form_label"`
	Age               int       `json:"age"`
	Attack            int       `json:"attack"`
	Passing           int       `json:"passing"`
	Defending         int       `json:"defending"`
	Physical          int       `json:"physical"`
}

type ManagerProfile struct {
	Style   string     `json:"style"`
	Targets []Position `json:"targets"`
}

type BidRecord struct {
	ParticipantID   uuid.UUID `json:"participant_id"`
	ParticipantName string    `json:"participant_name"`
	Amount          int64     `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}

type Participant struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Type            ParticipantType `json:"type"`
	Personality     Personality     `json:"personality,omitempty"`
	Profile         ManagerProfile  `json:"profile,omitempty"`
	RemainingBudget int64           `json:"remaining_budget"`
	InitialBudget   int64           `json:"initial_budget"`
	Squad           []SquadPlayer   `json:"squad"`
	Formation       Formation       `json:"formation"`
	HasPassed       bool            `json:"has_passed"`
}

type SquadPlayer struct {
	Player        GeneratedPlayer `json:"player"`
	PurchasePrice int64           `json:"purchase_price"`
	SlotType      string          `json:"slot_type"`
}

type AuctionConfig struct {
	Budget          int64      `json:"budget"`
	PlayerPoolSize  int        `json:"player_pool_size"`
	AIOpponents     int        `json:"ai_opponents"`
	Difficulty      Difficulty `json:"difficulty"`
	AuctionType     string     `json:"auction_type"`
	TimerPerPlayer  int        `json:"timer_per_player"`
	BidResetTimer   int        `json:"bid_reset_timer"`
	MinBidIncrement int64      `json:"min_bid_increment"`
}

type ScoreBreakdown struct {
	AttackScore      float64 `json:"attack_score"`
	MidfieldScore    float64 `json:"midfield_score"`
	DefenseScore     float64 `json:"defense_score"`
	ChemistryScore   float64 `json:"chemistry_score"`
	BenchStrength    float64 `json:"bench_strength"`
	CurrentForm      float64 `json:"current_form"`
	FormationFit     float64 `json:"formation_fit"`
	SquadDepth       float64 `json:"squad_depth"`
	BudgetEfficiency float64 `json:"budget_efficiency"`
	BalanceBonus     float64 `json:"balance_bonus"`
	TotalScore       float64 `json:"total_score"`
}

type TeamAnalysis struct {
	Breakdown  ScoreBreakdown `json:"breakdown"`
	Strengths  []string       `json:"strengths"`
	Weaknesses []string       `json:"weaknesses"`
}

type ActivityEvent struct {
	Type            string    `json:"type"`
	Message         string    `json:"message"`
	Icon            string    `json:"icon,omitempty"`
	ParticipantID   string    `json:"participant_id,omitempty"`
	ParticipantName string    `json:"participant_name,omitempty"`
	Amount          *int64    `json:"amount,omitempty"`
	PlayerName      string    `json:"player_name,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

func FormLabelFromScore(form int) string {
	switch {
	case form >= 85:
		return "Excellent"
	case form >= 75:
		return "Good"
	case form >= 65:
		return "Average"
	default:
		return "Poor"
	}
}

func RankTierFromPoints(points int) RankTier {
	switch {
	case points >= 12000:
		return RankLegend
	case points >= 7000:
		return RankDiamond
	case points >= 3500:
		return RankPlatinum
	case points >= 1500:
		return RankGold
	case points >= 500:
		return RankSilver
	default:
		return RankBronze
	}
}
