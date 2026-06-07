package scoring

import (
	"fmt"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/pkg/engine/chemistry"
)

type formationSlots struct {
	gk  int
	def int
	mid int
	att int
}

var formationSlotMap = map[domain.Formation]formationSlots{
	domain.Formation433:  {1, 4, 3, 3},
	domain.Formation442:  {1, 4, 4, 2},
	domain.Formation352:  {1, 3, 5, 2},
	domain.Formation4231: {1, 4, 5, 1},
}

func positionWeight(pos domain.Position) float64 {
	weights := map[domain.Position]float64{
		domain.PosST: 1.0, domain.PosCF: 1.0,
		domain.PosLW: 0.85, domain.PosRW: 0.85,
		domain.PosCAM: 0.7,
		domain.PosCM: 1.0, domain.PosCDM: 0.9,
		domain.PosLM: 0.8, domain.PosRM: 0.8,
		domain.PosCB: 1.0, domain.PosLB: 0.85, domain.PosRB: 0.85,
		domain.PosGK: 1.2,
	}
	if w, ok := weights[pos]; ok {
		return w
	}
	return 0.7
}

func formModifier(form int) float64 {
	return 0.8 + (float64(form)/100.0)*0.4
}

func selectStartingXI(squad []domain.SquadPlayer, formation domain.Formation) []domain.GeneratedPlayer {
	slots := formationSlotMap[formation]
	var gks, defs, mids, atts []domain.GeneratedPlayer

	for _, sp := range squad {
		switch sp.Player.Position.Group() {
		case domain.GroupGK:
			gks = append(gks, sp.Player)
		case domain.GroupDEF:
			defs = append(defs, sp.Player)
		case domain.GroupMID:
			mids = append(mids, sp.Player)
		case domain.GroupATT:
			atts = append(atts, sp.Player)
		}
	}

	sortByRatingDesc(gks)
	sortByRatingDesc(defs)
	sortByRatingDesc(mids)
	sortByRatingDesc(atts)

	var xi []domain.GeneratedPlayer
	xi = append(xi, takeN(gks, slots.gk)...)
	xi = append(xi, takeN(defs, slots.def)...)
	xi = append(xi, takeN(mids, slots.mid)...)
	xi = append(xi, takeN(atts, slots.att)...)
	return xi
}

func takeN(players []domain.GeneratedPlayer, n int) []domain.GeneratedPlayer {
	if len(players) < n {
		return players
	}
	return players[:n]
}

func sortByRatingDesc(players []domain.GeneratedPlayer) {
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[j].Rating > players[i].Rating {
				players[i], players[j] = players[j], players[i]
			}
		}
	}
}

func lineScore(players []domain.GeneratedPlayer, group domain.PositionGroup) float64 {
	if len(players) == 0 {
		return 0
	}
	var total float64
	var maxPossible float64
	count := 0
	for _, p := range players {
		if p.Position.Group() == group {
			w := positionWeight(p.Position)
			total += float64(p.Rating) * w * formModifier(p.Form)
			maxPossible += 92.0 * w * 1.2
			count++
		}
	}
	if count == 0 {
		return 0
	}
	if maxPossible == 0 {
		return 0
	}
	return (total / maxPossible) * 100.0
}

func benchScore(squad []domain.SquadPlayer, startingXI []domain.GeneratedPlayer) float64 {
	startSet := make(map[string]bool)
	for _, p := range startingXI {
		startSet[p.ID.String()] = true
	}
	var bench []domain.GeneratedPlayer
	for _, sp := range squad {
		if !startSet[sp.Player.ID.String()] {
			bench = append(bench, sp.Player)
		}
	}
	if len(bench) == 0 {
		return 0
	}
	var sum float64
	for _, p := range bench {
		sum += float64(p.Rating)
	}
	avg := sum / float64(len(bench))
	depthFactors := map[int]float64{0: 0, 1: 0.4, 2: 0.7, 3: 0.85, 4: 1.0}
	factor := depthFactors[len(bench)]
	if len(bench) > 4 {
		factor = 1.0
	}
	return (avg / 92.0) * 100.0 * factor
}

func formScore(startingXI []domain.GeneratedPlayer) float64 {
	if len(startingXI) == 0 {
		return 0
	}
	var sum float64
	for _, p := range startingXI {
		sum += float64(p.Form)
	}
	return sum / float64(len(startingXI))
}

func formationFitScore(squad []domain.SquadPlayer, formation domain.Formation, startingXI []domain.GeneratedPlayer) float64 {
	slots := formationSlotMap[formation]
	totalSlots := slots.gk + slots.def + slots.mid + slots.att
	if totalSlots == 0 {
		return 0
	}

	filled := len(startingXI)
	if filled > totalSlots {
		filled = totalSlots
	}

	var fitQuality float64
	for _, p := range startingXI {
		fitQuality += slotFitQuality(p, formation)
	}
	if len(startingXI) == 0 {
		return 0
	}
	slotRatio := float64(filled) / float64(totalSlots)
	avgFit := fitQuality / float64(len(startingXI))
	return slotRatio * avgFit * 100.0
}

func slotFitQuality(player domain.GeneratedPlayer, formation domain.Formation) float64 {
	slots := formationSlotMap[formation]
	group := player.Position.Group()
	switch group {
	case domain.GroupGK:
		if slots.gk > 0 {
			return 1.0
		}
	case domain.GroupDEF:
		if slots.def > 0 {
			return 1.0
		}
	case domain.GroupMID:
		if slots.mid > 0 {
			return 1.0
		}
	case domain.GroupATT:
		if slots.att > 0 {
			return 1.0
		}
	}
	if player.SecondaryPosition != "" && player.SecondaryPosition.Group() != group {
		return 0.7
	}
	return 0.3
}

func squadDepthScore(squad []domain.SquadPlayer) float64 {
	ideal := map[domain.PositionGroup]int{
		domain.GroupGK: 2, domain.GroupDEF: 5, domain.GroupMID: 5, domain.GroupATT: 4,
	}
	counts := map[domain.PositionGroup]int{}
	for _, sp := range squad {
		counts[sp.Player.Position.Group()]++
	}
	var total float64
	for group, idealCount := range ideal {
		coverage := float64(counts[group]) / float64(idealCount)
		if coverage > 1.0 {
			coverage = 1.0
		}
		total += coverage
	}
	return (total / 4.0) * 100.0
}

func budgetEfficiencyScore(squad []domain.SquadPlayer, initialBudget int64) float64 {
	if initialBudget == 0 || len(squad) == 0 {
		return 50
	}
	var totalSpent int64
	var totalRating float64
	for _, sp := range squad {
		totalSpent += sp.PurchasePrice
		totalRating += float64(sp.Player.Rating)
	}
	if totalSpent == 0 {
		return 50
	}
	avgRating := totalRating / float64(len(squad))
	valuePerMillion := avgRating / float64(totalSpent/1_000_000)
	score := valuePerMillion * 8.0
	if score > 100 {
		score = 100
	}
	return score
}

func balanceBonusScore(attack, midfield, defense float64) float64 {
	minScore := attack
	if midfield < minScore {
		minScore = midfield
	}
	if defense < minScore {
		minScore = defense
	}
	maxScore := attack
	if midfield > maxScore {
		maxScore = midfield
	}
	if defense > maxScore {
		maxScore = defense
	}
	gap := maxScore - minScore
	switch {
	case gap < 12:
		return 100
	case gap < 25:
		return 75
	case gap < 40:
		return 45
	default:
		return 15
	}
}

func BestFormation(squad []domain.SquadPlayer) domain.Formation {
	formations := []domain.Formation{
		domain.Formation433, domain.Formation442,
		domain.Formation352, domain.Formation4231,
	}
	best := domain.Formation433
	bestScore := -1.0
	for _, f := range formations {
		xi := selectStartingXI(squad, f)
		score := formationFitScore(squad, f, xi)
		if score > bestScore {
			bestScore = score
			best = f
		}
	}
	return best
}

func Evaluate(squad []domain.SquadPlayer, formation domain.Formation) domain.TeamAnalysis {
	return EvaluateWithBudget(squad, formation, 0)
}

func EvaluateWithBudget(squad []domain.SquadPlayer, formation domain.Formation, initialBudget int64) domain.TeamAnalysis {
	if formation == "" {
		formation = BestFormation(squad)
	}
	xi := selectStartingXI(squad, formation)

	attack := lineScore(xi, domain.GroupATT)
	midfield := lineScore(xi, domain.GroupMID)
	defense := lineScore(xi, domain.GroupDEF)
	chem := chemistry.Calculate(xi)
	bench := benchScore(squad, xi)
	form := formScore(xi)
	fit := formationFitScore(squad, formation, xi)
	depth := squadDepthScore(squad)
	budgetEff := budgetEfficiencyScore(squad, initialBudget)
	balance := balanceBonusScore(attack, midfield, defense)

	total := attack*0.17 + midfield*0.17 + defense*0.17 +
		chem*0.10 + bench*0.07 + form*0.08 + fit*0.09 + depth*0.05 +
		budgetEff*0.05 + balance*0.05

	breakdown := domain.ScoreBreakdown{
		AttackScore:      attack,
		MidfieldScore:    midfield,
		DefenseScore:     defense,
		ChemistryScore:   chem,
		BenchStrength:    bench,
		CurrentForm:      form,
		FormationFit:     fit,
		SquadDepth:       depth,
		BudgetEfficiency: budgetEff,
		BalanceBonus:     balance,
		TotalScore:       total,
	}

	return domain.TeamAnalysis{
		Breakdown:  breakdown,
		Strengths:  detectStrengths(breakdown, squad),
		Weaknesses: detectWeaknesses(breakdown, squad),
	}
}

func detectStrengths(b domain.ScoreBreakdown, squad []domain.SquadPlayer) []string {
	var s []string
	if b.AttackScore >= 80 {
		s = append(s, "Strong attack")
	}
	if b.MidfieldScore >= 80 {
		s = append(s, "Strong midfield control")
	}
	if b.DefenseScore >= 80 {
		s = append(s, "Solid defense")
	}
	if b.ChemistryScore >= 75 {
		s = append(s, "Excellent team chemistry")
	}
	if b.FormationFit >= 85 {
		s = append(s, "Perfect formation alignment")
	}
	if b.BalanceBonus >= 75 {
		s = append(s, "Well-balanced squad")
	}
	if b.BudgetEfficiency >= 70 {
		s = append(s, "Excellent value for money")
	}
	if len(s) == 0 {
		s = append(s, "Balanced squad composition")
	}
	return s
}

func detectWeaknesses(b domain.ScoreBreakdown, squad []domain.SquadPlayer) []string {
	var w []string
	if b.AttackScore < 40 {
		w = append(w, "Weak attack")
	}
	if b.MidfieldScore < 40 {
		w = append(w, "Weak midfield")
	}
	if b.DefenseScore < 40 {
		w = append(w, "Weak defense")
	}
	if b.ChemistryScore < 30 {
		w = append(w, "Poor team chemistry")
	}
	if b.BenchStrength < 20 {
		w = append(w, "No bench depth")
	}
	if b.BalanceBonus < 30 {
		w = append(w, "Squad imbalance — too many stars in one area")
	}
	if b.DefenseScore < 50 && b.AttackScore > 75 {
		w = append(w, "All-out attack with weak defense")
	}
	hasGK := false
	for _, sp := range squad {
		if sp.Player.Position == domain.PosGK {
			hasGK = true
			break
		}
	}
	if !hasGK {
		w = append(w, "No goalkeeper — critical weakness")
	}
	if len(squad) < 13 {
		w = append(w, fmt.Sprintf("Squad size %d below recommended 13", len(squad)))
	}
	if len(w) == 0 {
		w = append(w, "No major weaknesses detected")
	}
	return w
}
