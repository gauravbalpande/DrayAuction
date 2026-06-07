package player

import (
	"math/rand"

	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/google/uuid"
)

type playerTemplate struct {
	Name      string
	Position  domain.Position
	Secondary domain.Position
	Club      string
	Nation    string
	League    string
	Rating    int
	Value     int64
	Form      int
	Age       int
	Attack    int
	Passing   int
	Defending int
	Physical  int
}

// Curated database of real footballers from Top 5 European leagues.
var realPlayers = []playerTemplate{
	// Premier League
	{"Erling Haaland", domain.PosST, domain.PosCF, "Manchester City", "Norway", "Premier League", 91, 180_000_000, 88, 24, 95, 65, 45, 88},
	{"Mohamed Salah", domain.PosRW, domain.PosRM, "Liverpool", "Egypt", "Premier League", 89, 100_000_000, 86, 32, 88, 82, 45, 78},
	{"Bukayo Saka", domain.PosRW, domain.PosRM, "Arsenal", "England", "Premier League", 87, 120_000_000, 84, 23, 82, 85, 52, 74},
	{"William Saliba", domain.PosCB, domain.PosRB, "Arsenal", "France", "Premier League", 86, 80_000_000, 82, 23, 45, 68, 88, 82},
	{"Declan Rice", domain.PosCDM, domain.PosCM, "Arsenal", "England", "Premier League", 87, 110_000_000, 85, 25, 62, 82, 85, 84},
	{"Cole Palmer", domain.PosCAM, domain.PosRW, "Chelsea", "England", "Premier League", 84, 90_000_000, 87, 22, 85, 86, 42, 72},
	{"Bruno Fernandes", domain.PosCAM, domain.PosCM, "Manchester United", "Portugal", "Premier League", 86, 70_000_000, 80, 30, 78, 90, 55, 72},
	{"Trent Alexander-Arnold", domain.PosRB, domain.PosRM, "Liverpool", "England", "Premier League", 86, 75_000_000, 78, 26, 72, 92, 72, 74},
	{"Virgil van Dijk", domain.PosCB, domain.PosLB, "Liverpool", "Netherlands", "Premier League", 89, 45_000_000, 84, 33, 52, 78, 92, 86},
	{"Alisson", domain.PosGK, domain.PosGK, "Liverpool", "Brazil", "Premier League", 89, 50_000_000, 83, 32, 15, 62, 18, 78},
	{"Phil Foden", domain.PosCAM, domain.PosLW, "Manchester City", "England", "Premier League", 88, 110_000_000, 86, 24, 84, 88, 48, 72},
	{"Rodri", domain.PosCDM, domain.PosCM, "Manchester City", "Spain", "Premier League", 90, 130_000_000, 88, 28, 72, 88, 88, 82},
	{"Martin Ødegaard", domain.PosCAM, domain.PosCM, "Arsenal", "Norway", "Premier League", 87, 90_000_000, 84, 25, 78, 90, 52, 68},
	{"Alexander Isak", domain.PosST, domain.PosCF, "Newcastle", "Sweden", "Premier League", 85, 80_000_000, 82, 25, 88, 72, 42, 78},
	{"Moises Caicedo", domain.PosCDM, domain.PosCM, "Chelsea", "Ecuador", "Premier League", 82, 85_000_000, 80, 23, 58, 78, 82, 84},

	// La Liga
	{"Jude Bellingham", domain.PosCAM, domain.PosCM, "Real Madrid", "England", "La Liga", 90, 180_000_000, 90, 21, 86, 88, 72, 82},
	{"Vinicius Jr", domain.PosLW, domain.PosLM, "Real Madrid", "Brazil", "La Liga", 90, 170_000_000, 87, 24, 90, 78, 38, 80},
	{"Lamine Yamal", domain.PosRW, domain.PosRM, "Barcelona", "Spain", "La Liga", 84, 150_000_000, 89, 17, 82, 84, 35, 68},
	{"Pedri", domain.PosCM, domain.PosCAM, "Barcelona", "Spain", "La Liga", 87, 100_000_000, 82, 22, 72, 90, 62, 68},
	{"Federico Valverde", domain.PosCM, domain.PosRM, "Real Madrid", "Uruguay", "La Liga", 88, 120_000_000, 86, 26, 78, 84, 78, 86},
	{"Antoine Griezmann", domain.PosCF, domain.PosCAM, "Atletico Madrid", "France", "La Liga", 86, 35_000_000, 84, 33, 86, 84, 58, 72},
	{"Raphinha", domain.PosRW, domain.PosRM, "Barcelona", "Brazil", "La Liga", 86, 70_000_000, 85, 28, 82, 80, 48, 76},
	{"Eduardo Camavinga", domain.PosCM, domain.PosCDM, "Real Madrid", "France", "La Liga", 85, 90_000_000, 83, 22, 68, 82, 78, 84},
	{"Aurélien Tchouaméni", domain.PosCDM, domain.PosCM, "Real Madrid", "France", "La Liga", 85, 85_000_000, 81, 25, 62, 80, 82, 86},
	{"Thibaut Courtois", domain.PosGK, domain.PosGK, "Real Madrid", "Belgium", "La Liga", 89, 40_000_000, 80, 32, 12, 58, 15, 82},
	{"Robert Lewandowski", domain.PosST, domain.PosCF, "Barcelona", "Poland", "La Liga", 88, 25_000_000, 82, 36, 92, 78, 42, 80},

	// Serie A
	{"Victor Osimhen", domain.PosST, domain.PosCF, "Napoli", "Nigeria", "Serie A", 87, 100_000_000, 84, 26, 90, 68, 42, 86},
	{"Rafael Leão", domain.PosLW, domain.PosLM, "AC Milan", "Portugal", "Serie A", 86, 90_000_000, 83, 25, 86, 76, 38, 84},
	{"Lautaro Martínez", domain.PosST, domain.PosCF, "Inter Milan", "Argentina", "Serie A", 88, 110_000_000, 87, 27, 90, 74, 48, 82},
	{"Nicolò Barella", domain.PosCM, domain.PosCAM, "Inter Milan", "Italy", "Serie A", 87, 85_000_000, 86, 27, 72, 86, 72, 78},
	{"Alessandro Bastoni", domain.PosCB, domain.PosLB, "Inter Milan", "Italy", "Serie A", 86, 80_000_000, 84, 25, 48, 78, 88, 78},
	{"Khvicha Kvaratskhelia", domain.PosLW, domain.PosLM, "Napoli", "Georgia", "Serie A", 86, 85_000_000, 85, 23, 86, 82, 42, 76},
	{"Federico Chiesa", domain.PosLW, domain.PosRW, "Juventus", "Italy", "Serie A", 84, 50_000_000, 78, 27, 82, 78, 45, 76},
	{"Theo Hernández", domain.PosLB, domain.PosLM, "AC Milan", "France", "Serie A", 86, 60_000_000, 82, 27, 78, 72, 78, 84},

	// Bundesliga
	{"Jamal Musiala", domain.PosCAM, domain.PosCM, "Bayern Munich", "Germany", "Bundesliga", 88, 130_000_000, 88, 21, 86, 88, 52, 74},
	{"Florian Wirtz", domain.PosCAM, domain.PosCM, "Bayer Leverkusen", "Germany", "Bundesliga", 88, 130_000_000, 89, 21, 84, 90, 48, 70},
	{"Harry Kane", domain.PosST, domain.PosCF, "Bayern Munich", "England", "Bundesliga", 90, 100_000_000, 87, 31, 92, 82, 42, 78},
	{"Joshua Kimmich", domain.PosCDM, domain.PosRB, "Bayern Munich", "Germany", "Bundesliga", 87, 70_000_000, 84, 29, 65, 88, 82, 76},
	{"Alphonso Davies", domain.PosLB, domain.PosLM, "Bayern Munich", "Canada", "Bundesliga", 84, 60_000_000, 80, 24, 72, 74, 72, 88},
	{"Xavi Simons", domain.PosCAM, domain.PosCM, "RB Leipzig", "Netherlands", "Bundesliga", 84, 80_000_000, 86, 21, 80, 86, 48, 72},
	{"Serhou Guirassy", domain.PosST, domain.PosCF, "Borussia Dortmund", "Guinea", "Bundesliga", 84, 45_000_000, 86, 28, 88, 68, 42, 84},
	{"Manuel Neuer", domain.PosGK, domain.PosGK, "Bayern Munich", "Germany", "Bundesliga", 87, 8_000_000, 78, 38, 12, 55, 15, 72},

	// Ligue 1
	{"Kylian Mbappé", domain.PosLW, domain.PosST, "Real Madrid", "France", "La Liga", 92, 200_000_000, 88, 26, 95, 82, 38, 88},
	{"Achraf Hakimi", domain.PosRB, domain.PosRM, "PSG", "Morocco", "Ligue 1", 86, 70_000_000, 84, 26, 78, 78, 72, 86},
	{"Marquinhos", domain.PosCB, domain.PosCDM, "PSG", "Brazil", "Ligue 1", 87, 60_000_000, 82, 30, 52, 72, 90, 82},
	{"Vitinha", domain.PosCM, domain.PosCDM, "PSG", "Portugal", "Ligue 1", 84, 55_000_000, 83, 24, 68, 84, 72, 74},
	{"Warren Zaïre-Emery", domain.PosCM, domain.PosCDM, "PSG", "France", "Ligue 1", 82, 60_000_000, 84, 18, 65, 82, 72, 78},
	{"Ousmane Dembélé", domain.PosRW, domain.PosRM, "PSG", "France", "Ligue 1", 85, 50_000_000, 82, 27, 86, 80, 38, 76},
	{"Gonçalo Ramos", domain.PosST, domain.PosCF, "PSG", "Portugal", "Ligue 1", 82, 45_000_000, 80, 23, 82, 72, 42, 78},

	// Additional depth across leagues
	{"Gabriel Magalhães", domain.PosCB, domain.PosLB, "Arsenal", "Brazil", "Premier League", 85, 65_000_000, 83, 27, 48, 68, 86, 82},
	{"Enzo Fernández", domain.PosCM, domain.PosCDM, "Chelsea", "Argentina", "Premier League", 84, 85_000_000, 80, 24, 68, 86, 72, 78},
	{"Son Heung-min", domain.PosLW, domain.PosST, "Tottenham", "South Korea", "Premier League", 86, 50_000_000, 84, 32, 88, 80, 42, 76},
	{"Bernardo Silva", domain.PosCM, domain.PosCAM, "Manchester City", "Portugal", "Premier League", 87, 70_000_000, 84, 30, 72, 88, 62, 72},
	{"Gavi", domain.PosCM, domain.PosCAM, "Barcelona", "Spain", "La Liga", 83, 80_000_000, 76, 20, 68, 84, 68, 74},
	{"Frenkie de Jong", domain.PosCM, domain.PosCDM, "Barcelona", "Netherlands", "La Liga", 86, 70_000_000, 82, 27, 68, 88, 72, 78},
	{"Marcus Thuram", domain.PosST, domain.PosLW, "Inter Milan", "France", "Serie A", 84, 55_000_000, 83, 27, 84, 74, 42, 86},
	{"Hakan Çalhanoğlu", domain.PosCDM, domain.PosCM, "Inter Milan", "Turkey", "Serie A", 85, 40_000_000, 84, 30, 72, 88, 72, 74},
	{"Jonathan Tah", domain.PosCB, domain.PosRB, "Bayer Leverkusen", "Germany", "Bundesliga", 84, 35_000_000, 82, 28, 45, 72, 86, 84},
	{"Dayot Upamecano", domain.PosCB, domain.PosRB, "Bayern Munich", "France", "Bundesliga", 84, 50_000_000, 80, 26, 48, 68, 86, 86},
	{"Desire Doué", domain.PosRW, domain.PosCAM, "PSG", "France", "Ligue 1", 80, 50_000_000, 82, 19, 78, 80, 42, 72},
	{"Ederson", domain.PosGK, domain.PosGK, "Manchester City", "Brazil", "Premier League", 88, 40_000_000, 82, 31, 12, 78, 15, 78},
	{"Toni Kroos", domain.PosCM, domain.PosCDM, "Real Madrid", "Germany", "La Liga", 88, 20_000_000, 85, 34, 72, 94, 72, 68},
	{"Dani Carvajal", domain.PosRB, domain.PosRM, "Real Madrid", "Spain", "La Liga", 86, 30_000_000, 82, 32, 68, 78, 78, 76},
	{"Kai Havertz", domain.PosST, domain.PosCAM, "Arsenal", "Germany", "Premier League", 84, 65_000_000, 82, 25, 82, 80, 58, 80},
}

type AuctionTier struct {
	Type           string
	Budget         int64
	PlayerPoolSize int
}

func ResolveAuctionTier(difficulty domain.Difficulty) AuctionTier {
	switch difficulty {
	case domain.DifficultyEasy:
		return AuctionTier{"Casual Auction", 500_000_000, 40}
	case domain.DifficultyMedium:
		return AuctionTier{"Competitive Auction", 800_000_000, 50}
	case domain.DifficultyHard:
		return AuctionTier{"Competitive Auction", 800_000_000, 50}
	case domain.DifficultyLegendary:
		return AuctionTier{"Elite Auction", 1_200_000_000, 60}
	default:
		return AuctionTier{"Competitive Auction", 800_000_000, 50}
	}
}

func GeneratePool(poolSize int, seed int64) []domain.GeneratedPlayer {
	rng := rand.New(rand.NewSource(seed))
	if poolSize > len(realPlayers) {
		poolSize = len(realPlayers)
	}

	indices := rng.Perm(len(realPlayers))[:poolSize]
	players := make([]domain.GeneratedPlayer, poolSize)
	for i, idx := range indices {
		t := realPlayers[idx]
		players[i] = templateToPlayer(t)
	}
	rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	return players
}

func templateToPlayer(t playerTemplate) domain.GeneratedPlayer {
	sec := t.Secondary
	if sec == "" {
		sec = t.Position
	}
	return domain.GeneratedPlayer{
		ID:                uuid.New(),
		Name:              t.Name,
		Position:          t.Position,
		SecondaryPosition: sec,
		Club:              t.Club,
		Nation:            t.Nation,
		League:            t.League,
		Rating:            t.Rating,
		MarketValue:       t.Value,
		Form:              t.Form,
		FormLabel:         domain.FormLabelFromScore(t.Form),
		Age:               t.Age,
		Attack:            t.Attack,
		Passing:           t.Passing,
		Defending:         t.Defending,
		Physical:          t.Physical,
	}
}
