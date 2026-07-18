package state

import "mahjom/game/domain"

type GameHandler struct {
	ID     string
	Config *domain.GameConfig
	State  *domain.GameState

	drawState *DrawState
}
