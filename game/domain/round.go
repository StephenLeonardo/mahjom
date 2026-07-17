package domain

type RoundPhase uint8

const (
	PhaseDealing RoundPhase = iota
	PhasePlay
	PhaseClaim
	PhaseRoundEnd
)

type Round struct {
	Number        int
	Phase         RoundPhase
	CurrentPlayer SeatIndex

	NewlyDrawnTile *Tile
	LastDiscard    *Tile
	LastDiscardBy  SeatIndex

	Claim *ClaimWindow
}

type ClaimWindow struct {
	Discard     Tile
	FromSeat    SeatIndex
	PendingSeat SeatIndex
}
