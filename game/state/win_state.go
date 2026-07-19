package state

type WinState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *WinState) GetStateName() string {
	return "win state"
}
