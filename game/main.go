package main

import (
	"fmt"
	"mahjom/game/domain"
	"mahjom/game/domain/constants"
	"time"
)

func main() {

	wall := domain.Wall{
		DrawPile: constants.Catalog,
	}

	wall.Shuffle(uint64(time.Now().Unix()))

	for _, tile := range wall.DrawPile {
		fmt.Println(tile.Describe())
	}

}
