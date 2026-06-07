package services

import (
	"context"
	"sync"
	"time"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/pkg/engine/ai"
	"github.com/drayauction/auctionxi/pkg/engine/auction"
	"github.com/drayauction/auctionxi/pkg/engine/player"
	"github.com/drayauction/auctionxi/pkg/engine/scoring"
	"github.com/google/uuid"
)

type CreateAuctionInput struct {
	AIOpponents int
	Difficulty  domain.Difficulty
}

type AuctionService struct {
	mu      sync.RWMutex
	engines map[uuid.UUID]*auction.Engine
	events  map[uuid.UUID][]chan domain.ActivityEvent
	results map[uuid.UUID]*AuctionResultResponse
}

type AuctionResultResponse struct {
	AuctionID  string               `json:"auction_id"`
	WinnerID   string               `json:"winner_id"`
	WinnerName string               `json:"winner_name"`
	Teams      []TeamResultResponse `json:"teams"`
	Rewards    *ProgressionReward   `json:"rewards,omitempty"`
	Ready      bool                 `json:"ready"`
}

type TeamResultResponse struct {
	ParticipantID string                `json:"participant_id"`
	Name          string                `json:"name"`
	Type          string                `json:"type"`
	Formation     string                `json:"formation"`
	TotalScore    float64               `json:"total_score"`
	Breakdown     domain.ScoreBreakdown `json:"breakdown"`
	Strengths     []string              `json:"strengths"`
	Weaknesses    []string              `json:"weaknesses"`
	SquadSize     int                   `json:"squad_size"`
}

type ProgressionReward struct {
	Coins      int `json:"coins"`
	XP         int `json:"xp"`
	RankPoints int `json:"rank_points"`
}

func NewAuctionService() *AuctionService {
	return &AuctionService{
		engines: make(map[uuid.UUID]*auction.Engine),
		events:  make(map[uuid.UUID][]chan domain.ActivityEvent),
		results: make(map[uuid.UUID]*AuctionResultResponse),
	}
}

func (s *AuctionService) Create(ctx context.Context, userID uuid.UUID, input CreateAuctionInput) (map[string]interface{}, error) {
	tier := player.ResolveAuctionTier(input.Difficulty)
	seed := time.Now().UnixNano()
	pool := player.GeneratePool(tier.PlayerPoolSize, seed)
	aiManagers := ai.GenerateManagers(input.AIOpponents, input.Difficulty, tier.Budget, seed)

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
		AIOpponents:     input.AIOpponents,
		Difficulty:      input.Difficulty,
		AuctionType:     tier.Type,
		TimerPerPlayer:  auction.InitialTimerSeconds,
		BidResetTimer:   auction.BidResetSeconds,
		MinBidIncrement: auction.MinBidIncrement,
	}

	engine := auction.NewEngine(auctionID, config, human, pool, aiManagers, seed)

	s.mu.Lock()
	s.engines[auctionID] = engine
	s.mu.Unlock()

	aiList := make([]map[string]interface{}, len(aiManagers))
	for i, m := range aiManagers {
		targets := make([]string, len(m.Participant.Profile.Targets))
		for j, t := range m.Participant.Profile.Targets {
			targets[j] = string(t)
		}
		aiList[i] = map[string]interface{}{
			"id":          m.Participant.ID.String(),
			"name":        m.Participant.Name,
			"personality": string(m.Participant.Personality),
			"style":       m.Participant.Profile.Style,
			"targets":     targets,
			"difficulty":  string(input.Difficulty),
		}
	}

	return map[string]interface{}{
		"id":               auctionID.String(),
		"status":           "setup",
		"ai_opponents":     input.AIOpponents,
		"difficulty":       string(input.Difficulty),
		"auction_type":     tier.Type,
		"budget":           tier.Budget,
		"player_pool_size": tier.PlayerPoolSize,
		"ai_managers":      aiList,
		"created_at":       time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *AuctionService) Start(ctx context.Context, auctionID uuid.UUID) (map[string]interface{}, error) {
	s.mu.RLock()
	engine, ok := s.engines[auctionID]
	s.mu.RUnlock()
	if !ok {
		return nil, domain.ErrNotFound
	}

	engine.SetEventHandler(func(event domain.ActivityEvent) {
		s.broadcastEvent(auctionID, event)
	})

	engine.Start()
	go s.runAuctionLoop(auctionID, engine)

	return s.buildStateResponse(engine), nil
}

func (s *AuctionService) runAuctionLoop(auctionID uuid.UUID, engine *auction.Engine) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	aiCounter := 0
	for range ticker.C {
		state := engine.GetState()
		if state.Status == domain.AuctionStatusCompleted {
			s.finalizeAuction(auctionID, engine)
			return
		}
		if state.Status != domain.AuctionStatusLive {
			continue
		}

		aiCounter++
		if aiCounter%2 == 0 {
			engine.RunAICycle()
		}

		expired, _ := engine.TickTimer()
		if expired {
			_, completed := engine.ResolveCurrentPlayer()
			aiCounter = 0
			if completed {
				s.finalizeAuction(auctionID, engine)
				return
			}
		}
	}
}

func (s *AuctionService) finalizeAuction(auctionID uuid.UUID, engine *auction.Engine) {
	s.mu.Lock()
	if _, exists := s.results[auctionID]; exists {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	participants := engine.GetParticipants()
	var teams []TeamResultResponse
	var winner TeamResultResponse
	winnerScore := -1.0

	for _, p := range participants {
		formation := scoring.BestFormation(p.Squad)
		analysis := scoring.EvaluateWithBudget(p.Squad, formation, p.InitialBudget)
		team := TeamResultResponse{
			ParticipantID: p.ID.String(),
			Name:          p.Name,
			Type:          string(p.Type),
			Formation:     string(formation),
			TotalScore:    analysis.Breakdown.TotalScore,
			Breakdown:     analysis.Breakdown,
			Strengths:     analysis.Strengths,
			Weaknesses:    analysis.Weaknesses,
			SquadSize:     len(p.Squad),
		}
		teams = append(teams, team)
		if team.TotalScore > winnerScore {
			winnerScore = team.TotalScore
			winner = team
		}
	}

	result := &AuctionResultResponse{
		AuctionID:  auctionID.String(),
		WinnerID:   winner.ParticipantID,
		WinnerName: winner.Name,
		Teams:      teams,
		Ready:      true,
	}

	s.mu.Lock()
	s.results[auctionID] = result
	s.mu.Unlock()

	engine.MarkResultsReady()
}

func (s *AuctionService) GetState(ctx context.Context, auctionID uuid.UUID) (map[string]interface{}, error) {
	s.mu.RLock()
	engine, ok := s.engines[auctionID]
	s.mu.RUnlock()
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s.buildStateResponse(engine), nil
}

func (s *AuctionService) buildStateResponse(engine *auction.Engine) map[string]interface{} {
	state := engine.GetState()
	var currentPlayer interface{}
	if state.CurrentIndex < len(state.PlayerPool) {
		p := state.PlayerPool[state.CurrentIndex]
		currentPlayer = map[string]interface{}{
			"id": p.ID.String(), "name": p.Name, "position": string(p.Position),
			"club": p.Club, "nation": p.Nation, "league": p.League,
			"rating": p.Rating, "market_value": p.MarketValue,
			"form": p.Form, "form_label": p.FormLabel, "age": p.Age,
			"attack": p.Attack, "passing": p.Passing,
			"defending": p.Defending, "physical": p.Physical,
		}
	}

	participants := make([]map[string]interface{}, len(state.Participants))
	for i, p := range state.Participants {
		entry := map[string]interface{}{
			"id": p.ID.String(), "name": p.Name, "type": string(p.Type),
			"remaining_budget": p.RemainingBudget, "squad_size": len(p.Squad),
			"has_passed": p.HasPassed,
		}
		if p.Type == domain.ParticipantAI {
			targets := make([]string, len(p.Profile.Targets))
			for j, t := range p.Profile.Targets {
				targets[j] = string(t)
			}
			entry["style"] = p.Profile.Style
			entry["targets"] = targets
			entry["personality"] = string(p.Personality)
		}
		participants[i] = entry
	}

	bidHistory := make([]map[string]interface{}, len(state.BidHistory))
	for i, b := range state.BidHistory {
		bidHistory[i] = map[string]interface{}{
			"participant_id":   b.ParticipantID.String(),
			"participant_name": b.ParticipantName,
			"amount":           b.Amount,
			"timestamp":        b.Timestamp.Format(time.RFC3339),
		}
	}

	feed := engine.GetActivityFeed()
	feedJSON := make([]map[string]interface{}, len(feed))
	for i, e := range feed {
		entry := map[string]interface{}{
			"type": e.Type, "message": e.Message, "icon": e.Icon,
			"timestamp": e.Timestamp.Format(time.RFC3339),
		}
		if e.ParticipantName != "" {
			entry["participant_name"] = e.ParticipantName
		}
		if e.Amount != nil {
			entry["amount"] = *e.Amount
		}
		feedJSON[i] = entry
	}

	resp := map[string]interface{}{
		"id": state.ID.String(), "status": string(state.Status),
		"auction_type": state.Config.AuctionType,
		"difficulty": string(state.Config.Difficulty), "budget": state.Config.Budget,
		"player_pool_size": state.Config.PlayerPoolSize,
		"current_player_index": state.CurrentIndex,
		"current_bid": state.CurrentBid,
		"timer_seconds": state.TimerSeconds,
		"timer_max": state.TimerMax,
		"participants": participants,
		"bid_history": bidHistory,
		"activity_feed": feedJSON,
		"results_ready": state.ResultsReady,
	}
	if currentPlayer != nil {
		resp["current_player"] = currentPlayer
	}
	if state.HighestBidder != nil {
		resp["highest_bidder"] = state.HighestBidder.String()
		resp["highest_bidder_name"] = state.HighestBidderName
	}
	return resp
}

func (s *AuctionService) PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount int64) (int64, error) {
	s.mu.RLock()
	engine, ok := s.engines[auctionID]
	s.mu.RUnlock()
	if !ok {
		return 0, domain.ErrNotFound
	}

	participantID := s.findHumanParticipant(engine)
	if err := engine.PlaceBid(participantID, amount); err != nil {
		return 0, domain.New("BID_ERROR", err.Error(), 400)
	}
	return amount, nil
}

func (s *AuctionService) BidWithIncrement(ctx context.Context, auctionID, userID uuid.UUID, incrementM int64) (int64, error) {
	s.mu.RLock()
	engine, ok := s.engines[auctionID]
	s.mu.RUnlock()
	if !ok {
		return 0, domain.ErrNotFound
	}

	state := engine.GetState()
	increment := incrementM * 1_000_000
	var amount int64
	if state.CurrentBid == 0 {
		p := state.PlayerPool[state.CurrentIndex]
		amount = p.MarketValue + increment - 5_000_000
	} else {
		amount = state.CurrentBid + increment
	}
	return s.PlaceBid(ctx, auctionID, userID, amount)
}

func (s *AuctionService) Pass(ctx context.Context, auctionID, userID uuid.UUID) error {
	s.mu.RLock()
	engine, ok := s.engines[auctionID]
	s.mu.RUnlock()
	if !ok {
		return domain.ErrNotFound
	}
	participantID := s.findHumanParticipant(engine)
	return engine.Pass(participantID)
}

func (s *AuctionService) GetResults(ctx context.Context, auctionID uuid.UUID) (*AuctionResultResponse, error) {
	s.mu.RLock()
	result, ok := s.results[auctionID]
	engine, engineOk := s.engines[auctionID]
	s.mu.RUnlock()

	if ok {
		return result, nil
	}

	if engineOk {
		state := engine.GetState()
		if state.Status == domain.AuctionStatusCompleted {
			s.finalizeAuction(auctionID, engine)
			s.mu.RLock()
			result, ok = s.results[auctionID]
			s.mu.RUnlock()
			if ok {
				return result, nil
			}
		}
	}

	return nil, domain.ErrNotFound
}

func (s *AuctionService) SubscribeEvents(auctionID uuid.UUID) <-chan domain.ActivityEvent {
	ch := make(chan domain.ActivityEvent, 64)
	s.mu.Lock()
	s.events[auctionID] = append(s.events[auctionID], ch)
	s.mu.Unlock()
	return ch
}

func (s *AuctionService) broadcastEvent(auctionID uuid.UUID, event domain.ActivityEvent) {
	s.mu.RLock()
	channels := s.events[auctionID]
	s.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *AuctionService) findHumanParticipant(engine *auction.Engine) uuid.UUID {
	state := engine.GetState()
	for _, p := range state.Participants {
		if p.Type == domain.ParticipantHuman {
			return p.ID
		}
	}
	return uuid.Nil
}
