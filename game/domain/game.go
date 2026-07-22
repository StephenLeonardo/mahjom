package domain

import (
	"errors"
)

var (
	ErrRoundNotInPlay    = errors.New("round is not in play")
	ErrRoundNotInDiscard = errors.New("round is not in discard")
	ErrNotInWin          = errors.New("round is not in win")
	ErrMustDraw          = errors.New("player must draw before discarding")
	ErrMustDiscard       = errors.New("player must discard before drawing")
	ErrTileNotInHand     = errors.New("tile is not in the player's hand")
	ErrTileNotClaim      = errors.New("tile is not claimable for the requested meld")
	ErrNotClaimSeat      = errors.New("it is not this seat's turn to declare or pass")
	ErrAlreadyActed      = errors.New("seat has already declared or passed in this claim window")
	ErrClaimClosed       = errors.New("claim window is not open")
	ErrWallEmpty         = errors.New("wall is empty")
	ErrCheckWinClosed    = errors.New("not the time to check win")
)

type Game struct {
	ID     string
	Config *GameConfig
	State  *GameState
}
