package domain

type Game struct {
	ID     string
	Config GameConfig
	State  GameState
}
