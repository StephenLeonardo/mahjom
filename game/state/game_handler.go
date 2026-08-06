package state

import (
	"encoding/json"
	"fmt"
	"mahjom/game/dataset"
	"mahjom/game/domain"
	"mahjom/game/domain/constants"
	"strings"
	"time"
)

type GameHandler struct {
	ID     string
	Seed   int
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
		Seed:   int(seed),
		Config: config,
		State: &domain.GameState{
			Round: &domain.Round{
				Number:        0,
				Phase:         domain.PhaseCheckWin,
				CurrentPlayer: domain.SeatIndex(domain.East),
				LastDiscardBy: domain.SeatNone,
			},
			Players: [4]*domain.PlayerState{
				{
					ID:       "0",
					Position: "E",
				},
				{
					ID:       "1",
					Position: "S",
				},
				{
					ID:       "2",
					Position: "W",
				},
				{
					ID:       "3",
					Position: "N",
				},
			},
			Wall: &domain.Wall{DrawPile: wallTiles},
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
	handler.currState = handler.checkWinState
	return handler
}

func Play() {

	for range 1000 {
		seed := uint64(time.Now().Unix())
		g := NewGameHandler("gameID", &domain.GameConfig{}, seed)

		for !g.State.IsGameEnd() {
			// fmt.Printf("After resolving %s\n", g.currState.GetStateName())
			// fmt.Printf("Player %s's turn\n", g.State.GetCurrentPlayerState().ID)
			if err := g.currState.Resolve(); err != nil {
				fmt.Println("Error: ", err)
			}
			// g.State.PrintPlayersHand()
			// fmt.Scanln()

			fmt.Println(ToTrainingSampleString(g))
		}

		// fmt.Println()
		// fmt.Println("==============================")
		// fmt.Println()
		// fmt.Println("GG")
		// playersJson, _ := json.MarshalIndent(g.State.Players, "", "  ")
		// fmt.Println(string(playersJson))
		// fmt.Println("MAHJONG")
		fmt.Println("isWin: ", g.State.IsWin)
		// fmt.Println("Seed: ", seed)
		if g.State.IsWin {
			break
		}
	}
}

func ToTrainingSampleString(game *GameHandler) string {
	sample := ToTrainingSample(game)
	jsonBytes, _ := json.Marshal(sample)
	return strings.ReplaceAll(string(jsonBytes), "\n", "")
}

func ToTrainingSample(game *GameHandler) *dataset.TrainingSample {
	// TODO: implement
	trainingSample := &dataset.TrainingSample{
		Seed:      int(game.Seed),
		GameID:    game.ID,
		StateName: game.currState.GetStateName(),
		Player:    int(game.State.Round.CurrentPlayer),
		Action:    ToAction(game),
		Outcome:   ToGameOutcome(game),
	}

	return trainingSample
}

func ToAction(game *GameHandler) *dataset.Action {
	if game.currState == game.discardState {
		return &dataset.Action{
			Type: dataset.ActionDiscard,
		}
	}
	if game.currState == game.claimState {
		// TODO: implement
	}
	return nil
}

func ToGameOutcome(game *GameHandler) *dataset.GameOutcome {
	// TODO: implement
	return nil
}
