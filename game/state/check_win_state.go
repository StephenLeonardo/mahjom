package state

import "mahjom/game/domain"

type CheckWinState struct {
	defaultState
	game *GameHandler
}

func (s *CheckWinState) GetStateName() string {
	return "check win state"
}

func (s *CheckWinState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return ErrActionNotAllowed
	}

	if domain.IsSimpleWinningHand(s.game.State.GetCurrentPlayerState()) {
		s.game.currState = s.game.winState
		return nil
	}

	s.game.currState = s.game.discardState
	return nil
}
