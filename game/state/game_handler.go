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
					ID: "0",
				},
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
			},
			Wall: domain.Wall{DrawPile: wallTiles},
		},
	}

	handler.drawState = &DrawState{
		game: handler,
	}
	handler.discardState = &DiscardState{
		game: handler,
	}
	handler.claimState = &ClaimState{
		game: handler,
	}
	handler.checkWinState = &CheckWinState{
		game: handler,
	}
	handler.winState = &WinState{
		game: handler,
	}

	handler.State.Wall.Shuffle(seed)
	handler.State.DealInitialHands()
	return handler
}

func Play() {
	g := NewGameHandler("gameID", &domain.GameConfig{}, uint64(time.Now().Unix()))

	roundJson, _ := json.MarshalIndent(g.State.Round, "", "  ")
	fmt.Println(string(roundJson))

	fmt.Println()
	fmt.Println("------------------------------")
	fmt.Println()

	playersJson, _ := json.MarshalIndent(g.State.Players, "", "  ")
	fmt.Println(string(playersJson))

	for !g.State.IsGameEnd() {
		if err := g.currState.Resolve(); err != nil {
			fmt.Println("Error: ", err)
		}
	}

	fmt.Println()
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("GG")
	playersJson, _ = json.MarshalIndent(g.State.Players, "", "  ")
	fmt.Println(string(playersJson))
}
