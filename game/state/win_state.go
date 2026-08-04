package state

import "mahjom/game/domain"

type WinState struct {
	defaultState
	game *GameHandler
}

func (s *WinState) GetStateName() string {
	return "win state"
}

func (s *WinState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrNotInWin
	}
	s.game.State.IsWin = true
	return nil
}
