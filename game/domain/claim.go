package domain

// ClaimDecl is one player's declaration in an open ClaimWindow.
//
// The Meld is not yet built at declaration time; it is constructed only if
// this declaration wins. Resolved() is set on the winning ClaimDecl after
// ClaimWindow.Resolve runs.
type ClaimDecl struct {
	Claimant SeatIndex
	Type     MeldType // MeldPong, MeldKong, or MeldChow
	Kong     KongType // only meaningful when Type == MeldKong (KongExposed)
	Tile     *Tile
}

// ClaimAction is what a non-active player sends back to the engine during a
// ClaimWindow. The frontend drives seat-by-seat and submits one action per
// seat; the engine advances PendingSeat on Pass and appends to Declarations
// on Declare.
type ClaimAction struct {
	Type ClaimActionType
	Meld MeldType // when Type == ClaimDeclare
	Kong KongType // when Type == ClaimDeclare and Meld == MeldKong
}

type ClaimActionType uint8

const (
	ClaimDeclare ClaimActionType = iota
	ClaimPass
)

// seatDistance returns the number of seats from `from` to `to` going
// clockwise in turn order. The result is in [0,3]; 0 means same seat,
// 1 is the next player, 2 is opposite, 3 is previous.
//
// This is the canonical tiebreak primitive: when two claims of the same
// type compete, the smaller distance wins.
func seatDistance(from, to SeatIndex) int {
	return (int(to) - int(from) + 4) % 4
}

// nextSeatAfter wraps seatDistance to find the seat one step clockwise.
func nextSeatAfter(seat SeatIndex) SeatIndex {
	return SeatIndex((int(seat) + 1) % 4)
}

// Resolve picks the winning claim from a populated ClaimWindow.
//
// Selection rules (per house rules):
//   - Kong > Pong > Chow (by Type)
//   - on type tie, smaller seatDistance(FromSeat, Claimant) wins
//   - on full tie, earlier-declared ClaimDecl wins (insertion order)
//
// Returns nil if there are no declarations.
func (w *ClaimWindow) Resolve() *ClaimDecl {
	if w == nil || len(w.Declarations) == 0 {
		return nil
	}
	best := w.Declarations[0]
	for i := 1; i < len(w.Declarations); i++ {
		curr := w.Declarations[i]
		if betterClaim(curr, best, w.FromSeat) {
			best = curr
		}
	}
	return best
}

// betterClaim reports whether `a` should beat `b` under the house rules:
// higher MeldType wins; on equal type, smaller seat distance from the
// discarder wins; on full tie, the earlier-declared one wins (so a == b
// here, and a does not beat b).
func betterClaim(a, b *ClaimDecl, discarder SeatIndex) bool {
	if a.Type != b.Type {
		return a.Type > b.Type // MeldKong > MeldPong > MeldChow by enum order
	}
	da := seatDistance(discarder, a.Claimant)
	db := seatDistance(discarder, b.Claimant)
	return da < db
}

// IsResolved reports whether the window has been resolved.
func (w *ClaimWindow) IsResolved() bool {
	return w == nil || w.Resolved
}

// AllSeatsActed reports whether every non-discarding seat has either
// declared or passed. The caller drives this by tracking who has acted
// in the window's `Acted` field; the frontend uses PendingSeat for the
// next-to-act seat.
func (w *ClaimWindow) AllSeatsActed() bool {
	if w == nil {
		return false
	}
	for i := range w.Acted {
		if !w.Acted[i] {
			return false
		}
	}
	return true
}
