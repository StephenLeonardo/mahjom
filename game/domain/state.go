package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IGameState interface {
	DealInitialHands()
}

type GameState struct {
	Round   *Round
	Wall    *Wall
	Players [4]*PlayerState
	IsWin   bool
}

func (g *GameState) IsGameEnd() bool {
	return g.Wall.Remaining() == 0 || g.IsWin
}

func (g *GameState) GetCurrentPlayerState() *PlayerState {
	currSeat := g.Round.CurrentPlayer
	return g.Players[currSeat]
}

func (g *GameState) DealInitialHands() {
	for playerIndex := range g.Players {
		for range 13 {
			_, ok := g.drawHandTile(g.Players[playerIndex])
			if !ok {
				return
			}
		}
	}

	dealerTile, ok := g.drawHandTile(g.Players[East])
	if !ok {
		return
	}
	g.Round.CurrentPlayer = SeatIndex(East)
	g.Round.NewlyDrawnTile = dealerTile
	g.Round.Phase = PhasePlay
}

// drawHandTile reveals flowers and animals and draws replacements until it
// finds a tile that belongs in the player's concealed hand.
func (g *GameState) drawHandTile(player *PlayerState) (*Tile, bool) {
	for {
		tile, ok := g.Wall.Draw()
		if !ok {
			return nil, false
		}

		switch {
		case tile.IsFlower():
			player.AddTile(tile)
		case tile.IsAnimal():
			player.AddTile(tile)
		default:
			player.AddTile(tile)
			return tile, true
		}
	}
}

func (g *GameState) PrintPlayersHand() {
	fmt.Println()
	fmt.Println("---------------------------------------------")
	for _, player := range g.Players {
		playerCopy := &PlayerState{}
		deepCopyJSON(player, playerCopy)
		playerCopy.SortHand()
		// playerJson, _ := json.MarshalIndent(playerCopy, "", "  ")
		// fmt.Println(string(playerJson))
		tilesStr := []string{}
		for _, tile := range playerCopy.Hand {
			tilesStr = append(tilesStr, tile.Describe())
		}
		meldsStr := []string{}
		for _, melds := range playerCopy.Melds {
			for _, tile := range melds.Tiles {
				meldsStr = append(meldsStr, tile.Describe())
			}
		}
		animalsStr := []string{}
		for _, tile := range playerCopy.Animals {
			animalsStr = append(animalsStr, tile.Describe())
		}
		flowersStr := []string{}
		for _, tile := range playerCopy.Flowers {
			flowersStr = append(flowersStr, tile.Describe())
		}
		fmt.Println()
		fmt.Printf("Player %s\n", playerCopy.ID)
		fmt.Printf("Hand: %s\n", strings.Join(tilesStr, ", "))
		fmt.Printf("Melds: %s\n", strings.Join(meldsStr, ", "))
		fmt.Printf("Animals: %s\n", strings.Join(animalsStr, ", "))
		fmt.Printf("Flowers: %s\n", strings.Join(flowersStr, ", "))
	}
	fmt.Println("---------------------------------------------")
	fmt.Println()
}

func deepCopyJSON(src interface{}, dst interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dst)
}

func (g *GameState) GetWinningTiles() []*Tile {

	winner := g.GetCurrentPlayerState()

	res := append([]*Tile{}, winner.Hand...)

	for _, meld := range winner.Melds {
		res = append(res, meld.Tiles...)
	}

	res = append(res, winner.Animals...)
	res = append(res, winner.Flowers...)

	return res
}
