package state

type CheckWinState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *CheckWinState) GetStateName() string {
	return "check win state"
}
