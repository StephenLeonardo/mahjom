package main

import (
	"fmt"
	"mahjom/game/domain"
	"time"
)

func main() {
	game := domain.NewGame("demo", domain.GameConfig{}, uint64(time.Now().Unix()))

	fmt.Println("-------------------------")
	fmt.Println("East Player")
	for _, tile := range game.State.Players[domain.East].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.East].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.East].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("South Player")
	for _, tile := range game.State.Players[domain.South].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.South].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.South].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("West Player")
	for _, tile := range game.State.Players[domain.West].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.West].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.West].GetAnimalCount())

	fmt.Println("-------------------------")
	fmt.Println("North Player")
	for _, tile := range game.State.Players[domain.North].Hand {
		fmt.Println(tile.Describe())
	}
	fmt.Println("Flower Count: ", game.State.Players[domain.North].GetFlowerCount())
	fmt.Println("Animal Count: ", game.State.Players[domain.North].GetAnimalCount())
}
