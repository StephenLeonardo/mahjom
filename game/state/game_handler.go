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

	// Event log
	EventLog *domain.EventLog
}

func NewGameHandler(id string, config *domain.GameConfig, seed uint64) *GameHandler {
	wallTiles := append([]*domain.Tile(nil), constants.Catalog...)

	players := [4]*domain.PlayerState{
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
	}

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
			Players: players,
			Wall:    &domain.Wall{DrawPile: wallTiles},
		},

		EventLog: &domain.EventLog{
			Players: players,
			Seed:    int(seed),
			GameID:  id,
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
		// seed := uint64(1786119936)
		g := NewGameHandler("gameID", &domain.GameConfig{}, seed)
		eventLog := g.EventLog

		for !g.State.IsGameEnd() {

			g.State.Round.Claim = nil
			g.State.Round.ClaimV2 = nil
			g.State.Round.ClaimV3 = nil

			eventLog.CurrPosition = g.State.GetCurrentPlayerState().Position
			state := g.currState.GetStateName()
			eventLog.Action = state
			eventLog.RoundNumber = g.State.Round.Number

			if err := g.currState.Resolve(); err != nil {
				fmt.Println("Error: ", err)
			}

			if state == CLAIM_STATE &&
				g.State.Round.ClaimV3 != nil {
				winnerPosition := g.State.Round.ClaimV3.Claimant
				discarderPosition := g.State.Round.ClaimV3.Discarder
				eventLog.CurrPosition = g.State.Players[winnerPosition].Position
				eventLog.Meld = &domain.Meld{
					Type:         g.State.Round.ClaimV3.Type,
					Kong:         g.State.Round.ClaimV3.Kong,
					Tiles:        g.State.Round.ClaimV3.Tiles,
					FromSeat:     discarderPosition,
					FromPosition: g.State.Players[discarderPosition].Position,
				}
			}

			// g.State.PrintPlayersHand()
			// fmt.Scanln()

			// fmt.Println(ToTrainingSampleString(g))
			if state != CHECK_WIN_STATE &&
				state != WIN_STATE &&
				(state != CLAIM_STATE || g.State.Round.ClaimV3 != nil) {
				jsonBytes, _ := json.MarshalIndent(eventLog, "", "  ")
				fmt.Println(string(jsonBytes))
			}

			eventLog = resetEventLog(eventLog)

			fmt.Println()
			fmt.Println()
			fmt.Println("==============================")
			fmt.Println()
			fmt.Println()

		}

		// fmt.Println()
		// fmt.Println("==============================")
		// fmt.Println()
		// fmt.Println("GG")
		// playersJson, _ := json.MarshalIndent(g.State.Players, "", "  ")
		// fmt.Println(string(playersJson))
		// fmt.Println("MAHJONG")

		outcomeType := "Draw"
		var winner *string = nil
		var winningTiles []*domain.Tile
		if g.State.IsWin {
			outcomeType = "Win"
			winner = &g.State.GetCurrentPlayerState().Position
			winningTiles = g.State.GetWinningTiles()
		}

		eventLog.Outcome = &domain.GameOutcome{
			OutcomeType:  outcomeType,
			Winner:       winner,
			WinningTiles: winningTiles,
			Tai:          g.State.GetCurrentPlayerState().Score,
		}

		jsonBytes, _ := json.MarshalIndent(eventLog, "", "  ")
		fmt.Println(string(jsonBytes))

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

func resetEventLog(eventLog *domain.EventLog) *domain.EventLog {
	return &domain.EventLog{
		Players: eventLog.Players,
		Seed:    eventLog.Seed,
		GameID:  eventLog.GameID,
	}
}
