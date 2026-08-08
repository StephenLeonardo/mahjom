package state

import (
	"encoding/json"
	"fmt"
	"mahjom/game/domain"
	"mahjom/game/domain/constants"
	"os"
	"time"
)

const (
	discardDatasetFilePath = "./dataset/discard.jsonl"
	claimDatasetFilePath   = "./dataset/claim.jsonl"
	outcomeDatasetFilePath = "./dataset/outcome.jsonl"
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

	for range 100 {
		seed := uint64(time.Now().Nanosecond())
		// seed := uint64(1786119936)
		g := NewGameHandler("gameID", &domain.GameConfig{}, seed)
		eventLog := g.EventLog

		for !g.State.IsGameEnd() {
			eventLog.CurrPosition = g.State.GetCurrentPlayerState().Position
			state := g.currState.GetStateName()
			eventLog.Action = state

			if err := g.currState.Resolve(); err != nil {
				fmt.Println("Error: ", err)
			}

			eventLog.RemainingDrawTiles = int(g.State.Wall.Remaining())

			switch state {
			case DISCARD_STATE:
				eventLog.Tile = g.State.Round.LastDiscard
				eventLog.TileDiscardScores = g.State.Round.TileDiscardScores
				eventLog.DiscardTurnNumber += 1
				storeDataset(eventLog, discardDatasetFilePath)
			case DRAW_STATE:
				eventLog.Tile = g.State.Round.NewlyDrawnTile
			case CLAIM_STATE:
				if g.State.Round.ClaimV3 != nil {
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
					eventLog.Tile = g.State.Round.LastDiscard
					storeDataset(eventLog, claimDatasetFilePath)
				}
			}

			eventLog = resetEventLog(eventLog)
			resetRound(g.State.Round)
		}

		if g.currState == g.checkWinState {
			g.currState.Resolve()
		}

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

		storeDataset(eventLog, outcomeDatasetFilePath)
	}
}

func resetEventLog(eventLog *domain.EventLog) *domain.EventLog {
	return &domain.EventLog{
		Players:           eventLog.Players,
		Seed:              eventLog.Seed,
		GameID:            eventLog.GameID,
		DiscardTurnNumber: eventLog.DiscardTurnNumber,
	}
}

func storeDataset(eventLog *domain.EventLog, filepath string) error {
	jsonBytes, err := json.Marshal(eventLog)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(jsonBytes, '\n')); err != nil {
		return err
	}

	return nil
}

func resetRound(round *domain.Round) {
	// round.LastDiscard = nil
	// round.LastDiscardBy = domain.SeatNone
	round.Claim = nil
	round.ClaimV2 = nil
	round.ClaimV3 = nil
	round.NewlyDrawnTile = nil
	round.TileDiscardScores = nil
}
