package auction

import (
	"math/rand"
	"testing"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/pkg/engine/ai"
	"github.com/drayauction/auctionxi/pkg/engine/player"
	"github.com/google/uuid"
)

func TestSimulateFullLegendaryAuction(t *testing.T) {
	difficulty := domain.DifficultyLegendary
	tier := player.ResolveAuctionTier(difficulty)
	seed := int64(12345) // use fixed seed for reproducibility

	pool := player.GeneratePool(tier.PlayerPoolSize, seed)
	aiManagers := ai.GenerateManagers(3, difficulty, tier.Budget, seed)

	auctionID := uuid.New()
	human := domain.Participant{
		ID:              uuid.New(),
		Name:            "You",
		Type:            domain.ParticipantHuman,
		RemainingBudget: tier.Budget,
		InitialBudget:   tier.Budget,
		Formation:       domain.Formation433,
		Squad:           []domain.SquadPlayer{},
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

	engine := NewEngine(auctionID, config, human, pool, aiManagers, seed)

	_, err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// We run a mock loop simulating the game flow.
	// Since we are testing completion, we will simulate a series of ticks.
	// In a real game, the loop ticks every second. Here we will programmatically tick the timer and let AI bid.
	// The human will bid occasionally if budget allows, otherwise pass.
	rng := rand.New(rand.NewSource(seed))
	
	maxTicks := 10000 // prevent infinite loops in tests
	ticks := 0
	
	for ticks < maxTicks {
		ticks++
		state := engine.GetState()
		
		if state.Status == domain.AuctionStatusCalculating || state.Status == domain.AuctionStatusCompleted {
			t.Logf("Auction successfully reached terminal state %s at tick %d!", state.Status, ticks)
			return
		}
		
		if state.Status != domain.AuctionStatusLive {
			t.Fatalf("Unexpected state: %s at tick %d", state.Status, ticks)
		}

		// Run AI cycle
		engine.RunAICycle()

		// Human bidding behavior: 20% chance to bid if they are not highest bidder
		hPart := engine.findParticipant(human.ID)
		if hPart != nil && !hPart.HasPassed && len(hPart.Squad) < 15 {
			isHighest := false
			if state.HighestBidder != nil && *state.HighestBidder == human.ID {
				isHighest = true
			}
			if !isHighest && rng.Float64() < 0.20 {
				var bidAmt int64
				if state.CurrentBid == 0 {
					bidAmt = state.PlayerPool[state.CurrentIndex].MarketValue
				} else {
					bidAmt = state.CurrentBid + 5_000_000
				}
				if bidAmt <= hPart.RemainingBudget {
					_ = engine.PlaceBid(human.ID, bidAmt)
				}
			} else if rng.Float64() < 0.05 {
				_ = engine.Pass(human.ID)
			}
		}

		// Tick timer
		expired, _ := engine.TickTimer()
		if expired {
			_, completed, err := engine.ResolveCurrentPlayer()
			if err != nil {
				t.Fatalf("ResolveCurrentPlayer failed: %v", err)
			}
			if completed {
				// Transition to completed
				err = engine.Complete(true)
				if err != nil {
					t.Fatalf("Failed to complete engine: %v", err)
				}
				t.Logf("Completed successfully at tick %d", ticks)
				return
			}
		}
	}

	t.Fatalf("Test timed out: reached max ticks %d without completing", maxTicks)
}
