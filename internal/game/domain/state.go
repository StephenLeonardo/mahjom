package domain

type GameState struct {
	Round Round
	Wall  []Tile

	Players [4]PlayerState
}
