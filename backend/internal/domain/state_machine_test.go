package domain

import (
	"testing"
)

func TestValidateTransition_ValidPaths(t *testing.T) {
	tests := []struct {
		name string
		from AuctionStatus
		to   AuctionStatus
	}{
		{"setup to rulebook", AuctionStatusSetup, AuctionStatusRulebook},
		{"setup to live", AuctionStatusSetup, AuctionStatusLive},
		{"rulebook to live", AuctionStatusRulebook, AuctionStatusLive},
		{"live to resolving", AuctionStatusLive, AuctionStatusResolving},
		{"resolving to live", AuctionStatusResolving, AuctionStatusLive},
		{"resolving to calculating", AuctionStatusResolving, AuctionStatusCalculating},
		{"calculating to completed", AuctionStatusCalculating, AuctionStatusCompleted},
		{"calculating to failed", AuctionStatusCalculating, AuctionStatusFailed},
		{"setup to cancelled", AuctionStatusSetup, AuctionStatusCancelled},
		{"live to cancelled", AuctionStatusLive, AuctionStatusCancelled},
		{"calculating to cancelled", AuctionStatusCalculating, AuctionStatusCancelled},
		{"same state", AuctionStatusLive, AuctionStatusLive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransition(tt.from, tt.to); err != nil {
				t.Errorf("expected valid transition from %s to %s, got error: %v", tt.from, tt.to, err)
			}
		})
	}
}

func TestValidateTransition_InvalidPaths(t *testing.T) {
	tests := []struct {
		name string
		from AuctionStatus
		to   AuctionStatus
	}{
		{"setup to completed", AuctionStatusSetup, AuctionStatusCompleted},
		{"setup to calculating", AuctionStatusSetup, AuctionStatusCalculating},
		{"live to completed", AuctionStatusLive, AuctionStatusCompleted},
		{"live to calculating", AuctionStatusLive, AuctionStatusCalculating},
		{"completed to live", AuctionStatusCompleted, AuctionStatusLive},
		{"completed to cancelled", AuctionStatusCompleted, AuctionStatusCancelled},
		{"failed to live", AuctionStatusFailed, AuctionStatusLive},
		{"failed to cancelled", AuctionStatusFailed, AuctionStatusCancelled},
		{"resolving to completed", AuctionStatusResolving, AuctionStatusCompleted},
		{"cancelled to live", AuctionStatusCancelled, AuctionStatusLive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransition(tt.from, tt.to); err == nil {
				t.Errorf("expected invalid transition from %s to %s, got nil error", tt.from, tt.to)
			}
		})
	}
}
