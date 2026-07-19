package state

type DrawState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *DrawState) Draw() error {

	return nil
}

func (s *DrawState) GetStateName() string {
	return "draw state"
}
