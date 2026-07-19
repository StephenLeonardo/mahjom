package state

type ClaimState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *ClaimState) GetStateName() string {
	return "claim state"
}
