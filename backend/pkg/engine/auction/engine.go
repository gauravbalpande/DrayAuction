package auction

import (
	"fmt"
	"sync"
	"time"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/pkg/engine/ai"
	"github.com/google/uuid"
)

const (
	InitialTimerSeconds = 15
	BidResetSeconds     = 10
	MinBidIncrement     = 5_000_000
	MinSquadSize        = 11
	MaxSquadSize        = 15
)

type Engine struct {
	mu       sync.Mutex
	state    *State
	aiMgrs   []ai.Manager
	onEvent  func(domain.ActivityEvent)
}

type State struct {
	ID                uuid.UUID
	Status            domain.AuctionStatus
	Config            domain.AuctionConfig
	PlayerPool        []domain.GeneratedPlayer
	CurrentIndex      int
	CurrentBid        int64
	HighestBidder     *uuid.UUID
	HighestBidderName string
	TimerSeconds      int
	TimerMax          int
	BidHistory        []domain.BidRecord
	HumanParticipant  *domain.Participant
	Participants      []domain.Participant
	ActivityFeed      []domain.ActivityEvent
	Seed              int64
	Version           int
	ResultsReady      bool
}

func NewEngine(id uuid.UUID, config domain.AuctionConfig, human domain.Participant, playerPool []domain.GeneratedPlayer, aiManagers []ai.Manager, seed int64) *Engine {
	participants := make([]domain.Participant, 0, len(aiManagers)+1)
	participants = append(participants, human)
	for _, m := range aiManagers {
		participants = append(participants, m.Participant)
	}

	if config.BidResetTimer == 0 {
		config.BidResetTimer = BidResetSeconds
	}
	if config.TimerPerPlayer == 0 {
		config.TimerPerPlayer = InitialTimerSeconds
	}

	return &Engine{
		state: &State{
			ID:               id,
			Status:           domain.AuctionStatusRulebook,
			Config:           config,
			PlayerPool:       playerPool,
			Participants:     participants,
			HumanParticipant: &human,
			Seed:             seed,
			TimerSeconds:     InitialTimerSeconds,
			TimerMax:         InitialTimerSeconds,
			BidHistory:       []domain.BidRecord{},
		},
		aiMgrs: aiManagers,
	}
}

func (e *Engine) SetEventHandler(fn func(domain.ActivityEvent)) {
	e.onEvent = fn
}

func (e *Engine) Start() (domain.ActivityEvent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := domain.ValidateTransition(e.state.Status, domain.AuctionStatusLive); err != nil {
		return domain.ActivityEvent{}, err
	}

	e.state.Status = domain.AuctionStatusLive
	e.state.CurrentIndex = 0
	e.state.TimerSeconds = InitialTimerSeconds
	e.state.TimerMax = InitialTimerSeconds
	e.resetBidding()

	player := e.state.PlayerPool[0]
	event := e.presentPlayerEvent(player)
	e.emit(event)
	return event, nil
}

func (e *Engine) PlaceBid(participantID uuid.UUID, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.Status != domain.AuctionStatusLive {
		return fmt.Errorf("auction is not live")
	}

	participant := e.findParticipant(participantID)
	if participant == nil {
		return fmt.Errorf("participant not found")
	}
	if participant.HasPassed {
		return fmt.Errorf("participant has already passed")
	}
	if len(participant.Squad) >= MaxSquadSize {
		return fmt.Errorf("squad is full")
	}
	if amount > participant.RemainingBudget {
		return fmt.Errorf("insufficient budget")
	}

	minBid := e.state.CurrentBid + MinBidIncrement
	if e.state.CurrentBid == 0 {
		player := e.currentPlayer()
		minBid = player.MarketValue
	}
	if amount < minBid {
		return fmt.Errorf("bid must be at least %d", minBid)
	}

	e.state.CurrentBid = amount
	e.state.HighestBidder = &participantID
	e.state.HighestBidderName = participant.Name
	e.state.TimerSeconds = e.state.Config.BidResetTimer
	e.state.TimerMax = e.state.Config.BidResetTimer
	e.state.Version++

	record := domain.BidRecord{
		ParticipantID:   participantID,
		ParticipantName: participant.Name,
		Amount:          amount,
		Timestamp:       time.Now(),
	}
	e.state.BidHistory = append(e.state.BidHistory, record)

	icon := "⚡"
	if participant.Type == domain.ParticipantHuman {
		icon = "🔥"
	}
	event := domain.ActivityEvent{
		Type:            "bid_placed",
		Icon:            icon,
		Message:         fmt.Sprintf("%s %s bid %dM", icon, participant.Name, amount/1_000_000),
		ParticipantID:   participantID.String(),
		ParticipantName: participant.Name,
		Amount:          &amount,
		PlayerName:      e.currentPlayer().Name,
		Timestamp:       time.Now(),
	}
	e.emit(event)
	return nil
}

func (e *Engine) Pass(participantID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.Status != domain.AuctionStatusLive {
		return fmt.Errorf("auction is not live")
	}

	participant := e.findParticipant(participantID)
	if participant == nil {
		return fmt.Errorf("participant not found")
	}
	participant.HasPassed = true

	event := domain.ActivityEvent{
		Type:            "player_passed",
		Icon:            "💀",
		Message:         fmt.Sprintf("💀 %s passed", participant.Name),
		ParticipantID:   participantID.String(),
		ParticipantName: participant.Name,
		PlayerName:      e.currentPlayer().Name,
		Timestamp:       time.Now(),
	}
	e.emit(event)
	return nil
}

func (e *Engine) TickTimer() (expired bool, event *domain.ActivityEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.Status != domain.AuctionStatusLive {
		return false, nil
	}

	e.state.TimerSeconds--
	if e.state.TimerSeconds > 0 {
		return false, nil
	}
	return true, nil
}

func (e *Engine) RunAICycle() {
	e.mu.Lock()
	if e.state.Status != domain.AuctionStatusLive {
		e.mu.Unlock()
		return
	}
	player := e.currentPlayer()
	currentBid := e.state.CurrentBid
	highestBidder := e.state.HighestBidder
	e.mu.Unlock()

	for i := range e.aiMgrs {
		e.mu.Lock()
		if p := e.findParticipant(e.aiMgrs[i].Participant.ID); p != nil {
			e.aiMgrs[i].Participant = *p
		}
		if e.aiMgrs[i].Participant.HasPassed {
			e.mu.Unlock()
			continue
		}
		participant := e.aiMgrs[i].Participant
		e.mu.Unlock()

		decision := e.aiMgrs[i].Evaluate(player, currentBid, highestBidder)
		if decision.Bid {
			if err := e.PlaceBid(decision.ParticipantID, decision.Amount); err == nil {
				e.mu.Lock()
				currentBid = e.state.CurrentBid
				highestBidder = e.state.HighestBidder
				e.mu.Unlock()
			}
		} else if e.aiMgrs[i].RNG.Float64() < 0.35 {
			_ = e.Pass(participant.ID)
		}
	}
}

func (e *Engine) ResolveCurrentPlayer() (*domain.ActivityEvent, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := domain.ValidateTransition(e.state.Status, domain.AuctionStatusResolving); err != nil {
		return nil, false, err
	}

	e.state.Status = domain.AuctionStatusResolving
	player := e.currentPlayer()
	var event domain.ActivityEvent

	if e.state.HighestBidder != nil {
		winner := e.findParticipant(*e.state.HighestBidder)
		winner.RemainingBudget -= e.state.CurrentBid
		winner.Squad = append(winner.Squad, domain.SquadPlayer{
			Player:        player,
			PurchasePrice: e.state.CurrentBid,
		})
		amount := e.state.CurrentBid
		icon := "🏆"
		event = domain.ActivityEvent{
			Type:            "player_sold",
			Icon:            icon,
			Message:         fmt.Sprintf("🏆 %s sold to %s for %dM", player.Name, winner.Name, amount/1_000_000),
			ParticipantID:   winner.ID.String(),
			ParticipantName: winner.Name,
			Amount:          &amount,
			PlayerName:      player.Name,
			Timestamp:       time.Now(),
		}
	} else {
		event = domain.ActivityEvent{
			Type:       "player_unsold",
			Icon:       "❌",
			Message:    fmt.Sprintf("❌ %s went unsold", player.Name),
			PlayerName: player.Name,
			Timestamp:  time.Now(),
		}
	}
	e.emit(event)

	e.state.CurrentIndex++
	if e.shouldEnd() {
		if err := domain.ValidateTransition(domain.AuctionStatusResolving, domain.AuctionStatusCalculating); err != nil {
			return &event, true, err
		}
		e.state.Status = domain.AuctionStatusCalculating
		completeEvent := domain.ActivityEvent{
			Type:      "auction_completed",
			Icon:      "🎉",
			Message:   "🎉 Auction completed! Calculating results...",
			Timestamp: time.Now(),
		}
		e.emit(completeEvent)
		return &event, true, nil
	}

	if err := domain.ValidateTransition(domain.AuctionStatusResolving, domain.AuctionStatusLive); err != nil {
		return &event, false, err
	}
	e.state.Status = domain.AuctionStatusLive
	e.state.TimerSeconds = InitialTimerSeconds
	e.state.TimerMax = InitialTimerSeconds
	e.resetBidding()

	nextPlayer := e.currentPlayer()
	presentEvent := e.presentPlayerEvent(nextPlayer)
	e.emit(presentEvent)
	return &event, false, nil
}

func (e *Engine) Complete(success bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	targetStatus := domain.AuctionStatusCompleted
	if !success {
		targetStatus = domain.AuctionStatusFailed
	}

	if err := domain.ValidateTransition(e.state.Status, targetStatus); err != nil {
		return err
	}

	e.state.Status = targetStatus
	return nil
}

func (e *Engine) presentPlayerEvent(player domain.GeneratedPlayer) domain.ActivityEvent {
	return domain.ActivityEvent{
		Type:       "player_presented",
		Icon:       "🎯",
		Message:    fmt.Sprintf("🎯 %s is up — %s · %s · Value %dM", player.Name, player.Position, player.Club, player.MarketValue/1_000_000),
		PlayerName: player.Name,
		Timestamp:  time.Now(),
	}
}

func (e *Engine) MarkResultsReady() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.ResultsReady = true
}

func (e *Engine) GetState() State {
	e.mu.Lock()
	defer e.mu.Unlock()
	return *e.state
}

func (e *Engine) GetActivityFeed() []domain.ActivityEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	feed := make([]domain.ActivityEvent, len(e.state.ActivityFeed))
	copy(feed, e.state.ActivityFeed)
	return feed
}

func (e *Engine) GetParticipants() []domain.Participant {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make([]domain.Participant, len(e.state.Participants))
	copy(result, e.state.Participants)
	return result
}

func (e *Engine) shouldEnd() bool {
	if e.state.CurrentIndex >= len(e.state.PlayerPool) {
		return true
	}
	allFull := true
	for _, p := range e.state.Participants {
		if len(p.Squad) < MaxSquadSize {
			allFull = false
			break
		}
	}
	return allFull
}

func (e *Engine) currentPlayer() domain.GeneratedPlayer {
	return e.state.PlayerPool[e.state.CurrentIndex]
}

func (e *Engine) resetBidding() {
	e.state.CurrentBid = 0
	e.state.HighestBidder = nil
	e.state.HighestBidderName = ""
	e.state.BidHistory = []domain.BidRecord{}
	for i := range e.state.Participants {
		e.state.Participants[i].HasPassed = false
	}
}

func (e *Engine) findParticipant(id uuid.UUID) *domain.Participant {
	for i := range e.state.Participants {
		if e.state.Participants[i].ID == id {
			return &e.state.Participants[i]
		}
	}
	return nil
}

func (e *Engine) emit(event domain.ActivityEvent) {
	e.state.ActivityFeed = append([]domain.ActivityEvent{event}, e.state.ActivityFeed...)
	if len(e.state.ActivityFeed) > 100 {
		e.state.ActivityFeed = e.state.ActivityFeed[:100]
	}
	if e.onEvent != nil {
		e.onEvent(event)
	}
}
