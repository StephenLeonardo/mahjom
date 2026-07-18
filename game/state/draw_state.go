package state

type DrawState struct {
	defaultState
	gameHandler *GameHandler
}

func (s *DrawState) Draw() error {

	return nil
}
