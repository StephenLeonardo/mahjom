package domain

type RoundPhase uint8

const (
	PhaseDealing RoundPhase = iota
	PhaseDraw
	PhasePlay
	PhaseClaim
	PhaseDiscard
	PhaseCheckWin
	PhaseWin
)

type Round struct {
	Number        int
	Phase         RoundPhase
	CurrentPlayer SeatIndex

	NewlyDrawnTile *Tile
	LastDiscard    *Tile
	LastDiscardBy  SeatIndex

	Claim   *ClaimWindow
	ClaimV2 *ClaimWindowV2
	ClaimV3 *ClaimDecl
}

type ClaimWindow struct {
	Discard      *Tile
	FromSeat     SeatIndex
	PendingSeat  SeatIndex
	Acted        [4]bool // true once that seat has declared or passed
	Declarations []*ClaimDecl
	Resolved     bool
	Winner       *ClaimDecl
}

// prompt everyone at once, wait for responses, then resolve
type ClaimWindowV2 struct {
	Discard  *Tile
	FromSeat SeatIndex

	AllPossibleDeclarations [][]*ClaimDecl
	Responded               [4]bool
	Declarations            map[SeatIndex]*ClaimDecl
	DeadlineUnix            int64

	Resolved bool
	Winner   *ClaimDecl
}

func (r *Round) GetNextSeatAfter(seat SeatIndex) SeatIndex {
	return SeatIndex((int(seat) + 1) % 4)
}
