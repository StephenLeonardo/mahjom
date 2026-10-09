package state

import "mahjom/game/domain"

const (
	CHECK_WIN_STATE = "CHECK_WIN"
)

type CheckWinState struct {
	defaultState
	game *GameHandler
}

func (s *CheckWinState) GetStateName() string {
	return CHECK_WIN_STATE
}

func (s *CheckWinState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return ErrActionNotAllowed
	}

	if domain.IsSimpleWinningHand(s.game.State.GetCurrentPlayerState()) {
		// only win if player's Tai is enough
		playerTai := s.game.State.GetCurrentPlayerState().GetCurrentTai()
		if playerTai >= int64(s.game.Config.MinTai) {
			s.game.currState = s.game.winState
			return nil
		}
	}

	s.game.currState = s.game.discardState
	return nil
}
