package state

import "mahjom/game/domain"

const (
	WIN_STATE = "WIN"
)

type WinState struct {
	defaultState
	game *GameHandler
}

func (s *WinState) GetStateName() string {
	return WIN_STATE
}

func (s *WinState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrNotInWin
	}
	s.game.State.IsWin = true
	return nil
}
