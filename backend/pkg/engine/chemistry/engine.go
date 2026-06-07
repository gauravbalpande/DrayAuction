package chemistry

import (
	"github.com/drayauction/auctionxi/internal/domain"
)

func Calculate(startingXI []domain.GeneratedPlayer) float64 {
	if len(startingXI) < 2 {
		return 0
	}

	total := 0
	for i := 0; i < len(startingXI); i++ {
		for j := i + 1; j < len(startingXI); j++ {
			pair := 0
			if startingXI[i].Nation == startingXI[j].Nation {
				pair = 3
			}
			if startingXI[i].Club == startingXI[j].Club {
				if pair < 5 {
					pair = 5
				}
			}
			if startingXI[i].Nation == startingXI[j].Nation && startingXI[i].Club == startingXI[j].Club {
				pair = 7
			}
			total += pair
		}
	}

	score := (float64(total) / 150.0) * 100.0
	if score > 100 {
		score = 100
	}
	return score
}
