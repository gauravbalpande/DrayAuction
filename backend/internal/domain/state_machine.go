package domain

import (
	"fmt"
)

// ValidateTransition checks if transition from -> to is valid.
func ValidateTransition(from, to AuctionStatus) error {
	if from == to {
		return nil
	}

	// Any non-final state can transition to cancelled
	if to == AuctionStatusCancelled {
		if from != AuctionStatusCompleted && from != AuctionStatusFailed {
			return nil
		}
		return fmt.Errorf("cannot cancel auction from final state %s", from)
	}

	valid := false
	switch from {
	case AuctionStatusSetup:
		valid = to == AuctionStatusRulebook || to == AuctionStatusLive
	case AuctionStatusRulebook:
		valid = to == AuctionStatusLive
	case AuctionStatusLive:
		valid = to == AuctionStatusResolving
	case AuctionStatusResolving:
		valid = to == AuctionStatusLive || to == AuctionStatusCalculating
	case AuctionStatusCalculating:
		valid = to == AuctionStatusCompleted || to == AuctionStatusFailed
	case AuctionStatusCompleted, AuctionStatusFailed, AuctionStatusCancelled:
		// Terminal states cannot transition to any other state
		valid = false
	}

	if !valid {
		return fmt.Errorf("invalid auction state transition from %s to %s", from, to)
	}
	return nil
}
