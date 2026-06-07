package ai

import (
	"math/rand"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/google/uuid"
)

type managerArchetype struct {
	Name        string
	Style       string
	Targets     []domain.Position
	Personality domain.Personality
}

var managerArchetypes = []managerArchetype{
	{"Guardiola Jr.", "Possession Football", []domain.Position{domain.PosCM, domain.PosCDM, domain.PosCAM}, domain.PersonalityBalanced},
	{"Mourinho Lite", "Defensive Master", []domain.Position{domain.PosCB, domain.PosRB, domain.PosLB}, domain.PersonalityPatient},
	{"Kloppstein", "Gegenpressing", []domain.Position{domain.PosLW, domain.PosRW, domain.PosST}, domain.PersonalityAggressive},
	{"Ancelotti III", "Balanced Tactics", []domain.Position{domain.PosCM, domain.PosCB, domain.PosST}, domain.PersonalityBalanced},
	{"Simeone Jr.", "Counter Attack", []domain.Position{domain.PosCDM, domain.PosCB, domain.PosST}, domain.PersonalityPatient},
	{"Conte Lite", "Wing Back System", []domain.Position{domain.PosLB, domain.PosRB, domain.PosCM}, domain.PersonalityAggressive},
	{"Arteta Jr.", "Progressive Build-up", []domain.Position{domain.PosCM, domain.PosCAM, domain.PosCB}, domain.PersonalityYouthFocused},
	{"Tuchel Jr.", "Tactical Flexibility", []domain.Position{domain.PosCB, domain.PosCM, domain.PosLW}, domain.PersonalityBalanced},
	{"Wenger Jr.", "Youth Development", []domain.Position{domain.PosCAM, domain.PosLW, domain.PosCM}, domain.PersonalityYouthFocused},
	{"Ferguson IX", "Star Hunter", []domain.Position{domain.PosST, domain.PosRW, domain.PosCAM}, domain.PersonalityStarHunter},
	{"Pochettino II", "High Press", []domain.Position{domain.PosCM, domain.PosST, domain.PosLB}, domain.PersonalityAggressive},
	{"Zidane Jr.", "Galactico Builder", []domain.Position{domain.PosLW, domain.PosST, domain.PosCAM}, domain.PersonalityStarHunter},
}

type BidDecision struct {
	ParticipantID uuid.UUID
	Bid           bool
	Amount        int64
}

type Manager struct {
	Participant domain.Participant
	Difficulty  domain.Difficulty
	RNG         *rand.Rand
}

func GenerateManagers(count int, difficulty domain.Difficulty, budget int64, seed int64) []Manager {
	rng := rand.New(rand.NewSource(seed))
	used := make(map[string]bool)
	managers := make([]Manager, count)

	for i := 0; i < count; i++ {
		var arch managerArchetype
		for {
			arch = managerArchetypes[rng.Intn(len(managerArchetypes))]
			if !used[arch.Name] {
				used[arch.Name] = true
				break
			}
		}
		managers[i] = Manager{
			Participant: domain.Participant{
				ID:              uuid.New(),
				Name:            arch.Name,
				Type:            domain.ParticipantAI,
				Personality:     arch.Personality,
				Profile: domain.ManagerProfile{
					Style:   arch.Style,
					Targets: arch.Targets,
				},
				RemainingBudget: budget,
				InitialBudget:   budget,
				Formation:       domain.Formation433,
				Squad:           []domain.SquadPlayer{},
			},
			Difficulty: difficulty,
			RNG:        rand.New(rand.NewSource(seed + int64(i+1)*1000)),
		}
	}
	return managers
}

func (m *Manager) Evaluate(player domain.GeneratedPlayer, currentBid int64, highestBidder *uuid.UUID) BidDecision {
	if len(m.Participant.Squad) >= 15 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	if m.Participant.HasPassed {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}

	targetBonus := m.targetPositionBonus(player.Position)

	switch m.Difficulty {
	case domain.DifficultyEasy:
		return m.evaluateEasy(player, currentBid)
	case domain.DifficultyMedium:
		return m.evaluateMedium(player, currentBid, targetBonus)
	case domain.DifficultyHard:
		return m.evaluateHard(player, currentBid, targetBonus)
	case domain.DifficultyLegendary:
		return m.evaluateLegendary(player, currentBid, targetBonus)
	default:
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
}

func (m *Manager) targetPositionBonus(pos domain.Position) float64 {
	for _, t := range m.Participant.Profile.Targets {
		if t == pos {
			return 1.4
		}
		if t.Group() == pos.Group() {
			return 1.15
		}
	}
	return 1.0
}

func (m *Manager) evaluateEasy(player domain.GeneratedPlayer, currentBid int64) BidDecision {
	if m.RNG.Float64() < 0.55 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	increments := []int64{5_000_000, 10_000_000, 20_000_000}
	inc := increments[m.RNG.Intn(len(increments))]
	bid := currentBid + inc
	if currentBid == 0 {
		bid = player.MarketValue
	}
	if bid > m.Participant.RemainingBudget {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	return BidDecision{ParticipantID: m.Participant.ID, Bid: true, Amount: bid}
}

func (m *Manager) evaluateMedium(player domain.GeneratedPlayer, currentBid int64, targetBonus float64) BidDecision {
	need := m.positionNeed(player.Position) * targetBonus
	if need < 0.5 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	maxBid := int64(float64(player.MarketValue) * (0.85 + m.RNG.Float64()*0.25) * need / 2.0)
	if currentBid >= maxBid {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	bid := currentBid + 5_000_000
	if currentBid == 0 {
		bid = player.MarketValue
	}
	if bid > m.Participant.RemainingBudget || bid > maxBid {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	return BidDecision{ParticipantID: m.Participant.ID, Bid: true, Amount: bid}
}

func (m *Manager) positionNeed(pos domain.Position) float64 {
	counts := map[domain.PositionGroup]int{}
	for _, sp := range m.Participant.Squad {
		counts[sp.Player.Position.Group()]++
	}
	group := pos.Group()
	switch group {
	case domain.GroupGK:
		if counts[group] == 0 {
			return 3.0
		}
		return 0.3
	case domain.GroupDEF:
		if counts[group] < 2 {
			return 2.5
		}
		if counts[group] < 4 {
			return 1.5
		}
		return 0.3
	case domain.GroupMID:
		if counts[group] < 2 {
			return 2.0
		}
		if counts[group] < 3 {
			return 1.2
		}
		return 0.3
	case domain.GroupATT:
		if counts[group] < 2 {
			return 2.0
		}
		return 0.3
	}
	return 0.5
}

func (m *Manager) evaluateHard(player domain.GeneratedPlayer, currentBid int64, targetBonus float64) BidDecision {
	need := m.positionNeed(player.Position) * targetBonus
	if need < 0.5 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}

	remaining := 15 - len(m.Participant.Squad)
	if remaining <= 0 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	maxPerRemaining := m.Participant.RemainingBudget / int64(remaining)
	maxBid := int64(float64(maxPerRemaining) * 1.4)

	valueRatio := float64(player.Rating) / float64(player.MarketValue/1_000_000)
	if valueRatio < 0.003 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}

	fairValue := int64(float64(player.MarketValue) * (0.9 + need*0.08))
	if fairValue > maxBid {
		fairValue = maxBid
	}

	if currentBid >= fairValue {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}

	bid := currentBid + 5_000_000
	if currentBid == 0 {
		bid = player.MarketValue
	}
	if bid > fairValue {
		bid = fairValue
	}
	if bid > m.Participant.RemainingBudget {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	return BidDecision{ParticipantID: m.Participant.ID, Bid: true, Amount: bid}
}

func (m *Manager) evaluateLegendary(player domain.GeneratedPlayer, currentBid int64, targetBonus float64) BidDecision {
	need := m.positionNeed(player.Position) * targetBonus
	valueRatio := float64(player.Rating) / float64(player.MarketValue/1_000_000)

	remaining := 15 - len(m.Participant.Squad)
	if remaining <= 0 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	reservePerPlayer := m.Participant.RemainingBudget / int64(remaining)

	fairValue := int64(float64(player.MarketValue) * (0.85 + valueRatio*0.5))
	personalityMod := personalityMultiplier(m.Participant.Personality, player)
	fairValue = int64(float64(fairValue) * personalityMod)

	if fairValue > reservePerPlayer*2 {
		fairValue = reservePerPlayer * 2
	}

	chemistryBonus := m.chemistryBonus(player)
	fairValue = int64(float64(fairValue) * (1.0 + chemistryBonus))

	if currentBid >= fairValue || need < 0.3 {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}

	bid := currentBid + 5_000_000
	if currentBid == 0 {
		bid = player.MarketValue
	}
	if bid > fairValue {
		bid = fairValue
	}
	if bid > m.Participant.RemainingBudget {
		return BidDecision{ParticipantID: m.Participant.ID, Bid: false}
	}
	return BidDecision{ParticipantID: m.Participant.ID, Bid: true, Amount: bid}
}

func personalityMultiplier(p domain.Personality, player domain.GeneratedPlayer) float64 {
	switch p {
	case domain.PersonalityAggressive:
		return 1.3
	case domain.PersonalityPatient:
		return 0.85
	case domain.PersonalityStarHunter:
		if player.Rating >= 85 {
			return 1.4
		}
		return 0.6
	case domain.PersonalityYouthFocused:
		if player.Age <= 23 {
			return 1.3
		}
		return 0.8
	case domain.PersonalityBudgetManager:
		return 0.9
	default:
		return 1.0
	}
}

func (m *Manager) chemistryBonus(player domain.GeneratedPlayer) float64 {
	if len(m.Participant.Squad) == 0 {
		return 0
	}
	bonus := 0.0
	for _, sp := range m.Participant.Squad {
		if sp.Player.Nation == player.Nation {
			bonus += 0.02
		}
		if sp.Player.Club == player.Club {
			bonus += 0.03
		}
	}
	if bonus > 0.15 {
		bonus = 0.15
	}
	return bonus
}
