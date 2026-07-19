package state

import (
	"encoding/json"
	"fmt"
	"mahjom/game/domain"
	"mahjom/game/domain/constants"
	"time"
)

type GameHandler struct {
	ID     string
	Config *domain.GameConfig
	State  *domain.GameState
	// Players [4]*domain.PlayerState

	currState IStateActionables

	drawState     IStateActionables
	discardState  IStateActionables
	claimState    IStateActionables
	checkWinState IStateActionables
	winState      IStateActionables
}

func NewGameHandler(id string, config *domain.GameConfig, seed uint64) *GameHandler {
	wallTiles := append([]*domain.Tile(nil), constants.Catalog...)
	handler := &GameHandler{
		ID:     id,
		Config: config,
		State: &domain.GameState{
			Round: domain.Round{
				Number:        0,
				Phase:         domain.PhaseCheckWin,
				CurrentPlayer: domain.SeatIndex(domain.East),
				LastDiscardBy: domain.SeatNone,
			},
			Players: [4]*domain.PlayerState{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
				{
					ID: "4",
				},
			},
			Wall: domain.Wall{DrawPile: wallTiles},
		},
	}

	handler.drawState = &DrawState{
		gameHandler: handler,
	}
	handler.discardState = &DiscardState{
		gameHandler: handler,
	}
	handler.claimState = &ClaimState{
		gameHandler: handler,
	}
	handler.checkWinState = &CheckWinState{
		gameHandler: handler,
	}
	handler.winState = &WinState{
		gameHandler: handler,
	}

	handler.State.Wall.Shuffle(seed)
	handler.State.DealInitialHands()
	return handler
}

func Play() {
	g := NewGameHandler("gameID", &domain.GameConfig{}, uint64(time.Now().Unix()))

	roundJson, err := json.MarshalIndent(g.State.Round, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling Round to JSON:", err)
	} else {
		fmt.Println(string(roundJson))
	}

}
