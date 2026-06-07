package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRecord struct {
	user         domain.User
	passwordHash string
}

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]userRecord
	byEmail map[string]uuid.UUID
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:   make(map[uuid.UUID]userRecord),
		byEmail: make(map[string]uuid.UUID),
	}
}

func (r *MemoryUserRepository) Create(ctx context.Context, username, email, passwordHash string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[email]; exists {
		return nil, domain.ErrEmailTaken
	}

	user := domain.User{
		ID:         uuid.New(),
		Username:   username,
		Email:      email,
		Coins:      1000,
		XP:         0,
		RankPoints: 0,
		RankTier:   domain.RankBronze,
		Wins:       0,
		Losses:     0,
		CreatedAt:  time.Now(),
	}
	r.users[user.ID] = userRecord{user: user, passwordHash: passwordHash}
	r.byEmail[email] = user.ID
	return &user, nil
}

func (r *MemoryUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	u := rec.user
	return &u, nil
}

func (r *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	u := r.users[id].user
	return &u, nil
}

func (r *MemoryUserRepository) VerifyPassword(ctx context.Context, email, password string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}
	rec := r.users[id]
	if err := bcrypt.CompareHashAndPassword([]byte(rec.passwordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	u := rec.user
	return &u, nil
}

func (r *MemoryUserRepository) UpdateProgression(ctx context.Context, id uuid.UUID, coins, xp int64, rankPoints, wins, losses int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	rec.user.Coins += coins
	rec.user.XP += xp
	rec.user.RankPoints += rankPoints
	rec.user.Wins += wins
	rec.user.Losses += losses
	rec.user.RankTier = domain.RankTierFromPoints(rec.user.RankPoints)
	r.users[id] = rec
	return nil
}

type tokenRecord struct {
	userID    uuid.UUID
	expiresAt int64
}

type MemoryRefreshTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]tokenRecord
}

func NewMemoryRefreshTokenRepository() *MemoryRefreshTokenRepository {
	return &MemoryRefreshTokenRepository{tokens: make(map[string]tokenRecord)}
}

func (r *MemoryRefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[tokenHash] = tokenRecord{userID: userID, expiresAt: expiresAt}
	return nil
}

func (r *MemoryRefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.tokens[tokenHash]
	if !ok {
		return uuid.Nil, fmt.Errorf("token not found")
	}
	return rec.userID, nil
}

func (r *MemoryRefreshTokenRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, tokenHash)
	return nil
}

func (r *MemoryRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, rec := range r.tokens {
		if rec.userID == userID {
			delete(r.tokens, hash)
		}
	}
	return nil
}
