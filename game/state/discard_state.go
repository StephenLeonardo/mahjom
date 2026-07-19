package state

type DiscardState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *DiscardState) GetStateName() string {
	return "discard state"
}
