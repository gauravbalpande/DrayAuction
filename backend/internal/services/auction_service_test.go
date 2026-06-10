package services

import (
	"context"
	"testing"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/internal/repositories"
	"github.com/drayauction/auctionxi/pkg/engine/ai"
	"github.com/drayauction/auctionxi/pkg/engine/auction"
	"github.com/drayauction/auctionxi/pkg/engine/player"
	"github.com/google/uuid"
)

func TestFinalizeAuctionReal(t *testing.T) {
	userRepo := repositories.NewMemoryUserRepository()
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "PlayerOne", "player@example.com", "hash")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	userID := user.ID

	service := NewAuctionService(userRepo)
	
	difficulty := domain.DifficultyLegendary
	tier := player.ResolveAuctionTier(difficulty)
	seed := int64(12345)

	pool := player.GeneratePool(tier.PlayerPoolSize, seed)
	aiManagers := ai.GenerateManagers(3, difficulty, tier.Budget, seed)

	auctionID := uuid.New()
	
	// Create participants with pre-populated squads!
	human := domain.Participant{
		ID:              uuid.New(),
		Name:            "You",
		Type:            domain.ParticipantHuman,
		RemainingBudget: tier.Budget - 200_000_000,
		InitialBudget:   tier.Budget,
		Formation:       domain.Formation433,
		Squad:           []domain.SquadPlayer{},
	}
	for i := 0; i < 12; i++ {
		human.Squad = append(human.Squad, domain.SquadPlayer{
			Player:        pool[i],
			PurchasePrice: pool[i].MarketValue + 5_000_000,
		})
	}

	// Wrap AI managers with squads
	aiList := make([]ai.Manager, len(aiManagers))
	for idx, m := range aiManagers {
		m.Participant.Squad = []domain.SquadPlayer{}
		for i := 0; i < 12; i++ {
			playerIdx := (idx+1)*12 + i
			m.Participant.Squad = append(m.Participant.Squad, domain.SquadPlayer{
				Player:        pool[playerIdx],
				PurchasePrice: pool[playerIdx].MarketValue + 5_000_000,
			})
		}
		m.Participant.RemainingBudget = tier.Budget - 200_000_000
		aiList[idx] = m
	}

	config := domain.AuctionConfig{
		Budget:          tier.Budget,
		PlayerPoolSize:  tier.PlayerPoolSize,
		AIOpponents:     3,
		Difficulty:      difficulty,
		AuctionType:     tier.Type,
		TimerPerPlayer:  15,
		BidResetTimer:   10,
		MinBidIncrement: 5_000_000,
	}

	engine := auction.NewEngine(auctionID, config, human, pool, aiList, seed)

	// Inject the engine into the service
	service.mu.Lock()
	service.engines[auctionID] = engine
	service.userIDs[auctionID] = userID
	service.mu.Unlock()

	_, err = engine.Start()
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	var completed bool
	for i := 0; i < 60; i++ {
		_, completed, err = engine.ResolveCurrentPlayer()
		if err != nil {
			t.Fatalf("ResolveCurrentPlayer failed at %d: %v", i, err)
		}
		if completed {
			break
		}
	}

	if !completed {
		t.Fatalf("expected engine to be completed")
	}

	// Now status should be domain.AuctionStatusCalculating.
	// Let's call service.finalizeAuction!
	service.finalizeAuction(auctionID, engine)

	// Check results
	results, err := service.GetResults(ctx, auctionID)
	if err != nil {
		t.Fatalf("failed to get results: %v", err)
	}

	if !results.Ready {
		t.Fatalf("expected results to be ready")
	}

	if results.WinnerName == "" {
		t.Fatalf("expected winner name to be set")
	}

	t.Logf("Winner of simulation: %s", results.WinnerName)
	t.Logf("Awards: %+v", results.Awards)
	if results.Rewards != nil {
		t.Logf("Rewards: %+v", results.Rewards)
	}
}
