package scoring

import (
	"fmt"
	"github.com/drayauction/auctionxi/internal/domain"
)

type Awards struct {
	BestSigning    AwardInfo `json:"best_signing"`
	BiggestOverpay AwardInfo `json:"biggest_overpay"`
	BestChemistry  AwardInfo `json:"best_chemistry"`
	MostEfficient  AwardInfo `json:"most_efficient"`
}

type AwardInfo struct {
	PlayerName      string `json:"player_name,omitempty"`
	ParticipantName string `json:"participant_name"`
	Details         string `json:"details"`
}

type AuctionResult struct {
	WinnerID   string              `json:"winner_id"`
	WinnerName string              `json:"winner_name"`
	Awards     Awards              `json:"awards"`
	Teams      []TeamResultSummary `json:"teams"`
}

type TeamResultSummary struct {
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

func CalculateWinnerAndAwards(participants []domain.Participant) *AuctionResult {
	if len(participants) == 0 {
		return &AuctionResult{}
	}

	var summaries []TeamResultSummary
	var winner TeamResultSummary
	winnerScore := -1.0
	winnerBudgetRemaining := int64(-1)

	// Chemistry, Efficiency trackers
	bestChemScore := -1.0
	bestChemManager := ""
	
	bestEffScore := -1.0
	bestEffManager := ""

	// Player-level trackers
	bestSigningRatio := -1.0
	bestSigningPlayer := ""
	bestSigningManager := ""
	bestSigningPrice := int64(0)
	bestSigningRating := 0

	biggestOverpayAmt := int64(-1)
	biggestOverpayPlayer := ""
	biggestOverpayManager := ""
	biggestOverpayPrice := int64(0)
	biggestOverpayVal := int64(0)

	for _, p := range participants {
		formation := BestFormation(p.Squad)
		analysis := EvaluateWithBudget(p.Squad, formation, p.InitialBudget)
		
		summary := TeamResultSummary{
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
		summaries = append(summaries, summary)

		// Tiebreaker: Higher total score. If equal, higher remaining budget.
		isNewWinner := false
		if summary.TotalScore > winnerScore {
			isNewWinner = true
		} else if summary.TotalScore == winnerScore {
			if p.RemainingBudget > winnerBudgetRemaining {
				isNewWinner = true
			}
		}

		if isNewWinner {
			winnerScore = summary.TotalScore
			winnerBudgetRemaining = p.RemainingBudget
			winner = summary
		}

		// Track Best Chemistry
		if analysis.Breakdown.ChemistryScore > bestChemScore {
			bestChemScore = analysis.Breakdown.ChemistryScore
			bestChemManager = p.Name
		}

		// Track Most Efficient
		if analysis.Breakdown.BudgetEfficiency > bestEffScore {
			bestEffScore = analysis.Breakdown.BudgetEfficiency
			bestEffManager = p.Name
		}

		// Track player purchases
		for _, sp := range p.Squad {
			if sp.PurchasePrice > 0 {
				// Best Signing: High rating relative to price
				// Ratio: Rating / (Price in millions)
				ratio := float64(sp.Player.Rating) / (float64(sp.PurchasePrice) / 1_000_000.0)
				if ratio > bestSigningRatio {
					bestSigningRatio = ratio
					bestSigningPlayer = sp.Player.Name
					bestSigningManager = p.Name
					bestSigningPrice = sp.PurchasePrice
					bestSigningRating = sp.Player.Rating
				}

				// Biggest Overpay: Price - MarketValue
				overpay := sp.PurchasePrice - sp.Player.MarketValue
				if overpay > biggestOverpayAmt {
					biggestOverpayAmt = overpay
					biggestOverpayPlayer = sp.Player.Name
					biggestOverpayManager = p.Name
					biggestOverpayPrice = sp.PurchasePrice
					biggestOverpayVal = sp.Player.MarketValue
				}
			}
		}
	}

	awards := Awards{
		BestChemistry: AwardInfo{
			ParticipantName: bestChemManager,
			Details:         fmt.Sprintf("Achieved a team chemistry score of %.1f%%", bestChemScore),
		},
		MostEfficient: AwardInfo{
			ParticipantName: bestEffManager,
			Details:         fmt.Sprintf("Achieved a budget efficiency score of %.1f%%", bestEffScore),
		},
	}

	if bestSigningPlayer != "" {
		awards.BestSigning = AwardInfo{
			PlayerName:      bestSigningPlayer,
			ParticipantName: bestSigningManager,
			Details:         fmt.Sprintf("Signed %s (%d Rating) for just %dM", bestSigningPlayer, bestSigningRating, bestSigningPrice/1_000_000),
		}
	} else {
		awards.BestSigning = AwardInfo{
			ParticipantName: "N/A",
			Details:         "No players were signed",
		}
	}

	if biggestOverpayPlayer != "" && biggestOverpayAmt > 0 {
		awards.BiggestOverpay = AwardInfo{
			PlayerName:      biggestOverpayPlayer,
			ParticipantName: biggestOverpayManager,
			Details:         fmt.Sprintf("Signed %s for %dM (Market Value: %dM, Overpay: +%dM)", biggestOverpayPlayer, biggestOverpayPrice/1_000_000, biggestOverpayVal/1_000_000, biggestOverpayAmt/1_000_000),
		}
	} else {
		awards.BiggestOverpay = AwardInfo{
			ParticipantName: "N/A",
			Details:         "No overpays detected",
		}
	}

	return &AuctionResult{
		WinnerID:   winner.ParticipantID,
		WinnerName: winner.Name,
		Awards:     awards,
		Teams:      summaries,
	}
}
