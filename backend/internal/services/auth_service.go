package services

import (
	"context"
	"fmt"
	"time"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/internal/repositories"
	"github.com/drayauction/auctionxi/pkg/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users        domain.UserRepository
	refreshTokens domain.RefreshTokenRepository
	jwt          *auth.JWTManager
}

func NewAuthService(users domain.UserRepository, tokens domain.RefreshTokenRepository, jwt *auth.JWTManager) *AuthService {
	return &AuthService{users: users, refreshTokens: tokens, jwt: jwt}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*domain.User, *auth.TokenPair, error) {
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil, nil, domain.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, username, email, string(hash))
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	_ = s.refreshTokens.Create(ctx, user.ID, auth.HashToken(tokens.RefreshToken), expiresAt)

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, *auth.TokenPair, error) {
	userRepo, ok := s.users.(*repositories.MemoryUserRepository)
	if !ok {
		return nil, nil, domain.ErrInvalidCredentials
	}

	user, err := userRepo.VerifyPassword(ctx, email, password)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	_ = s.refreshTokens.Create(ctx, user.ID, auth.HashToken(tokens.RefreshToken), expiresAt)

	return user, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	hash := auth.HashToken(refreshToken)
	userID, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if userID != claims.UserID {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	_ = s.refreshTokens.DeleteByHash(ctx, hash)

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	_ = s.refreshTokens.Create(ctx, user.ID, auth.HashToken(tokens.RefreshToken), expiresAt)

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokens.DeleteByUserID(ctx, userID)
}
