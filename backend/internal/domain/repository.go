package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateProgression(ctx context.Context, id uuid.UUID, coins, xp int64, rankPoints, wins, losses int) error
}

type AuctionRepository interface {
	Create(ctx context.Context, userID uuid.UUID, config AuctionConfig, seed int64) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuctionRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status AuctionStatus) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]AuctionSummary, int, error)
}

type AuctionRecord struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Status             AuctionStatus
	Config             AuctionConfig
	Seed               int64
	CurrentPlayerIndex int
	CurrentBid         int64
	HighestBidderID    *uuid.UUID
	TimerSeconds       int
	Version            int
	Participants       []Participant
	Players            []GeneratedPlayer
}

type AuctionSummary struct {
	ID             uuid.UUID
	Status         AuctionStatus
	Difficulty     Difficulty
	Budget         int64
	PlayerPoolSize int
	AIOpponents    int
	Result         string
	Score          float64
	CreatedAt      string
}

type BidRepository interface {
	Create(ctx context.Context, auctionID, playerID, participantID uuid.UUID, amount int64) error
}

type ResultRepository interface {
	Create(ctx context.Context, auctionID uuid.UUID, winnerID, winnerUserID *uuid.UUID, isUserWin bool, coins, xp, rankPoints int) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt int64) error
	GetByHash(ctx context.Context, tokenHash string) (uuid.UUID, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type LeaderboardRepository interface {
	GetTop(ctx context.Context, period string, limit int) ([]LeaderboardEntry, error)
	Upsert(ctx context.Context, userID uuid.UUID, period string, score int64, wins, losses int) error
}

type LeaderboardEntry struct {
	Rank     int       `json:"rank"`
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	RankTier RankTier  `json:"rank_tier"`
	Score    int64     `json:"total_score"`
	Wins     int       `json:"wins"`
	WinRate  float64   `json:"win_rate"`
}
