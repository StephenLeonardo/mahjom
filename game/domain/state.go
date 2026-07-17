package domain

type GameState struct {
	Round   Round
	Wall    Wall
	Players [4]PlayerState
}
