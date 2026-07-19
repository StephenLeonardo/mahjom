package state

type WinState struct {
	defaultState
	game *GameHandler
}

func (s *WinState) GetStateName() string {
	return "win state"
}

func (s *WinState) Resolve() error {
	return nil
}
