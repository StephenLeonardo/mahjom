package domain

import (
	"slices"
	"sort"
)

type IPlayerState interface {
	AddTile(tile *Tile)
	SortHand()
	FindInHand(id TileID) (*Tile, bool)
	RemoveFromHand(id TileID) (*Tile, bool)
}

type PlayerState struct {
	ID string

	Hand     []*Tile
	Melds    []*Meld
	Flowers  []*Tile
	Animals  []*Tile
	Discards []*Tile

	Score int
}

func (p *PlayerState) GetFlowerCount() int {
	return len(p.Flowers)
}

func (p *PlayerState) GetAnimalCount() int {
	return len(p.Animals)
}

// AddTile places a tile in the appropriate player collection.
func (p *PlayerState) AddTile(tile *Tile) {
	if tile == nil {
		return
	}

	switch {
	case tile.IsFlower():
		p.Flowers = append(p.Flowers, tile)
	case tile.IsAnimal():
		p.Animals = append(p.Animals, tile)
	default:
		p.Hand = append(p.Hand, tile)
	}
}

// SortHand orders a hand by suit, rank, then physical TileID.
func (p *PlayerState) SortHand() {
	sort.SliceStable(p.Hand, func(i, j int) bool {
		left, right := p.Hand[i], p.Hand[j]
		if left.Suit != right.Suit {
			return left.Suit < right.Suit
		}
		if left.Rank != right.Rank {
			return left.Rank < right.Rank
		}
		return left.ID < right.ID
	})
}

// FindInHand returns the physical tile with id, if it is in the player's hand.
func (p *PlayerState) FindInHand(id TileID) (*Tile, bool) {
	for _, tile := range p.Hand {
		if tile.ID == id {
			return tile, true
		}
	}
	return nil, false
}

// RemoveFromHand removes and returns the physical tile with id.
func (p *PlayerState) RemoveFromHand(id TileID) (*Tile, bool) {
	for i, tile := range p.Hand {
		if tile.ID != id {
			continue
		}

		p.Hand = slices.Delete(p.Hand, i, i+1)
		return tile, true
	}
	return nil, false
}
