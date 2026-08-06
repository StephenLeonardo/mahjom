package domain

type EventLog struct {
	Seed             int
	GameID           string
	RoundNumber      int
	Action           string // Draw | Discard | Claim
	Player           int    // E | S | W | N
	Tile             *Tile
	Meld             *Meld
	Players          [4]*PlayerState
	PlayerDiscards   [4][]*Tile
	PlayerBonusTiles [4][]*Tile
	PlayerMelds      [4][]*Meld

	// Outcome
	Outcome *GameOutcome
}

type GameOutcome struct {
	OutcomeType string // Win | Draw

	Winner *string // E | S | W | N

	WinningTiles []*Tile

	Tai int
}
