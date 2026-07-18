package main

import (
	"encoding/json"
	"fmt"
	"mahjom/game/domain"
	"sort"
	"sync"
	"time"
)

func main() {
	game := domain.NewGame("demo", &domain.GameConfig{}, uint64(time.Now().Unix()))

	fmt.Println("-------------------------")
	fmt.Println("East Player")
	// game.State.Players[domain.East].SortHand()
	for _, tile := range game.State.Players[domain.East].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.East].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.East].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("South Player")
	// game.State.Players[domain.South].SortHand()
	for _, tile := range game.State.Players[domain.South].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.South].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.South].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("West Player")
	// game.State.Players[domain.West].SortHand()
	for _, tile := range game.State.Players[domain.West].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.West].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.West].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("North Player")
	// game.State.Players[domain.North].SortHand()
	for _, tile := range game.State.Players[domain.North].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.North].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.North].GetAnimalCount())

	isWin := false
	i := 0
	for game.State.Wall.Remaining() > 0 {
		fmt.Println()
		fmt.Println("Len wall: ", len(game.State.Wall.DrawPile))
		fmt.Println("Round ", i)

		err := game.DrawForCurrentPlayer()
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}

		err = game.ResolveWin()
		if err != nil {
			fmt.Println("Error Win: ", err)
			break
		}

		currPlayer := game.State.Players[game.State.Round.CurrentPlayer]
		fmt.Println("#################################")
		fmt.Println("Discarding this pile:")
		discardedJson, _ := json.MarshalIndent(currPlayer.Hand[0], "", "	")
		fmt.Println(string(discardedJson))
		fmt.Println("#################################")

		err = game.Discard(domain.TileID(currPlayer.Hand[0].ID))
		if err != nil {
			fmt.Println("Error 2: ", err)
			break
		}

		roundJson, _ := json.MarshalIndent(game.State.Round, "", "  ")
		fmt.Println(string(roundJson))

		for game.State.Round.Phase == domain.PhaseClaim {
			claimDecl := promptClaims(game.State.Round.ClaimV2)
			game.State.Round.ClaimV2.Winner = claimDecl
			claimDeclJson, _ := json.MarshalIndent(claimDecl, "", "	")
			fmt.Println()
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println("Claim Declare:")
			fmt.Println(string(claimDeclJson))
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			err = game.ResolveClaim()
			if err != nil {
				fmt.Println("Error 3: ", err)
				break
			}

			if claimDecl != nil {
				err = game.ResolveWin()
				if err != nil {
					fmt.Println("Error Win: ", err)
					break
				}

				currPlayer := game.State.Players[game.State.Round.CurrentPlayer]
				err = game.Discard(domain.TileID(currPlayer.Hand[0].ID))
				if err != nil {
					fmt.Println("Error 4: ", err)
					break
				}
			}
		}

		roundJson, _ = json.MarshalIndent(game.State.Round, "", "  ")
		fmt.Println(string(roundJson))

		if game.State.Round.Phase == domain.PhaseWin {
			fmt.Println("-------------------------------")
			fmt.Println("-------------------------------")
			fmt.Println("-------------------------------")
			fmt.Println("-------------------------------")
			fmt.Println("WIN")
			isWin = true
			break
		}
		i += 1
	}

	fmt.Println("-------------------------------")
	fmt.Println("-------------------------------")
	fmt.Println("-------------------------------")
	fmt.Println("-------------------------------")
	stateJson, _ := json.MarshalIndent(game.State.Players, "", "	")
	fmt.Println(string(stateJson))
	fmt.Println("isWin: ", isWin)

}

func promptClaims(window *domain.ClaimWindowV2) *domain.ClaimDecl {
	ch := make(chan *domain.ClaimDecl, 3)
	var wg sync.WaitGroup
	for _, possibleClaims := range window.AllPossibleDeclarations {
		// TODO: implement prompting to clients
		if len(possibleClaims) == 0 {
			continue
		}
		seatIndex := possibleClaims[0].Claimant

		wg.Add(1)
		go func(seatIndex domain.SeatIndex, possibleClaims []*domain.ClaimDecl) {
			defer wg.Done()
			// TODO: prompt user and wait for 5 seconds
			// TODO: implement pass action
			ch <- possibleClaims[0]
		}(seatIndex, possibleClaims)
	}

	go func() { wg.Wait(); close(ch) }()

	var responses []*domain.ClaimDecl
	for res := range ch {
		responses = append(responses, res)
	}

	return resolve(responses)
}

func resolve(claims []*domain.ClaimDecl) *domain.ClaimDecl {
	if len(claims) == 0 {
		return nil
	}
	discarder := int(claims[0].Discarder)
	n := 4
	sort.Slice(claims, func(i, j int) bool {
		if claims[i].Type == claims[j].Type {
			rankI := (int(claims[i].Claimant) - discarder + n) % n
			rankJ := (int(claims[j].Claimant) - discarder + n) % n
			return rankI < rankJ
		}

		return claims[i].Type > claims[j].Type
	})

	return claims[0]
}
