package domain

import (
	"encoding/json"
	"fmt"
)

// Canonical Singapore set: 148 physical tiles.
//
//	108 suited (Characters, Bamboo, Dots 1–9, ×4 each)
//	 16 winds (East/South/West/North, ×4)
//	 12 dragons (Zhong/Fa/Bai, ×4 each)
//	  8 flowers (four kinds × two copies; ranks 0–3 identify kind, two TileIDs per kind)
//	  4 animals (ranks 0–3)
//
// TileID is the instance index in the full set (0..147), not rank within a suit.
const TileSetSize = 148

type TileID uint8

type Suit uint8

const (
	SuitCharacter Suit = iota
	SuitBamboo
	SuitDot
	SuitWind
	SuitDragon
	SuitFlower
	SuitAnimal
)

// Dragon honor ranks (SuitDragon).
const (
	Zhong uint8 = iota
	Fa
	Bai
)

type Tile struct {
	ID   TileID
	Suit Suit
	Rank uint8 // suited 1–9; wind East..North as 0–3; dragon Zhong/Fa/Bai; flower/animal per catalog
}

var (
	windNames   = []string{"East Wind", "South Wind", "West Wind", "North Wind"}
	dragonNames = []string{"Hong Zhong", "Fa Cai", "Bai Ban"}
	animalNames = []string{"Cat", "Mouse", "Rooster", "Centipede"}
	flowerNames = []string{"Plum", "Orchid", "Chrysanthemum", "Bamboo"}
)

func (t *Tile) Describe() string {
	switch t.Suit {
	case SuitCharacter:
		return fmt.Sprintf("#%d %d Wan", t.ID, t.Rank)
	case SuitBamboo:
		return fmt.Sprintf("#%d %d Bamboo", t.ID, t.Rank)
	case SuitDot:
		return fmt.Sprintf("#%d %d Dot", t.ID, t.Rank)
	case SuitWind:
		return fmt.Sprintf("#%d %s", t.ID, windNames[t.Rank])
	case SuitDragon:
		return fmt.Sprintf("#%d %s", t.ID, dragonNames[t.Rank])
	case SuitAnimal:
		return fmt.Sprintf("#%d %s", t.ID, animalNames[t.Rank])
	case SuitFlower:
		return fmt.Sprintf("#%d %s", t.ID, flowerNames[t.Rank])
	}
	return fmt.Sprintf("Unknown tile (suit %d, rank %d)", t.Suit, t.Rank)
}

func (t *Tile) IsFlower() bool {
	return t.Suit == SuitFlower
}

func (t *Tile) IsAnimal() bool {
	return t.Suit == SuitAnimal
}

func (t *Tile) IsBonus() bool {
	return t.IsFlower() || t.IsAnimal()
}

func (t *Tile) IsSuitedForChow() bool {
	return t.Suit == SuitCharacter || t.Suit == SuitBamboo || t.Suit == SuitDot
}

func (t *Tile) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return json.Marshal(t.Describe())
}
