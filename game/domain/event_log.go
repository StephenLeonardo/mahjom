package domain

type EventLog struct {
	Seed              int             `json:"id"`
	GameID            string          `json:"-"`
	RoundNumber       int             `json:"-"`
	Action            string          `json:"action,omitempty"`       // Draw | Discard | Claim
	CurrPosition      string          `json:"currPosition,omitempty"` // E | S | W | N
	Tile              *Tile           `json:"tile,omitempty"`
	TileDiscardScores map[string]int  `json:"tileDiscardScores,omitempty"`
	Meld              *Meld           `json:"meld,omitempty"`
	Players           [4]*PlayerState `json:"players"`
	// PlayerDiscards   [4][]*Tile
	// PlayerBonusTiles [4][]*Tile
	// PlayerMelds      [4][]*Meld

	// Outcome
	Outcome *GameOutcome `json:"outcome,omitempty"`
}

type GameOutcome struct {
	OutcomeType string `json:"outcomeType"` // Win | Draw

	Winner *string `json:"winner"` // E | S | W | N

	WinningTiles []*Tile `json:"winningTiles"`

	Tai int `json:"tai"`
}
