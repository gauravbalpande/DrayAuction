package player

import (
	"testing"

	"github.com/drayauction/auctionxi/internal/domain"
)

func TestGeneratePoolSize(t *testing.T) {
	pool := GeneratePool(50, 12345)
	if len(pool) != 50 {
		t.Fatalf("expected 50 players, got %d", len(pool))
	}
}

func TestGeneratePoolRealNames(t *testing.T) {
	pool := GeneratePool(40, 42)
	found := false
	for _, p := range pool {
		if p.Name == "Jude Bellingham" || p.Name == "Rodri" {
			found = true
		}
		if p.League == "" {
			t.Fatalf("player %s missing league", p.Name)
		}
		if p.FormLabel == "" {
			t.Fatalf("player %s missing form label", p.Name)
		}
	}
	if !found {
		t.Fatal("expected real footballers in pool")
	}
}

func TestResolveAuctionTier(t *testing.T) {
	easy := ResolveAuctionTier(domain.DifficultyEasy)
	if easy.Budget != 500_000_000 {
		t.Fatalf("easy budget wrong: %d", easy.Budget)
	}
	legend := ResolveAuctionTier(domain.DifficultyLegendary)
	if legend.Budget != 1_200_000_000 {
		t.Fatalf("legendary budget wrong: %d", legend.Budget)
	}
}

func TestGeneratePoolDeterministic(t *testing.T) {
	a := GeneratePool(40, 999)
	b := GeneratePool(40, 999)
	for i := range a {
		if a[i].Name != b[i].Name {
			t.Fatalf("same seed should produce identical pools at index %d", i)
		}
	}
}
