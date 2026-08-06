package domain

import (
	"math/rand/v2"
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
	ID       string
	Position string // E | S | W | N

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

func (p *PlayerState) AddToMelds(claimDecl *ClaimDecl) {
	p.Melds = append(p.Melds, &Meld{
		Type:     claimDecl.Type,
		Kong:     claimDecl.Kong,
		Tiles:    claimDecl.Tiles,
		FromSeat: claimDecl.Discarder,
	})
}

// DiscardRandomTileFromHand selects a random tile from the player's hand,
// removes it, and returns the discarded tile.
func (p *PlayerState) DiscardRandomTileFromHand(seed uint64) *Tile {
	if len(p.Hand) == 0 {
		return nil
	}

	r := rand.New(rand.NewPCG(seed, 0))
	index := r.IntN(len(p.Hand))
	discardedTile := p.Hand[index]

	p.Hand[index] = p.Hand[len(p.Hand)-1]
	p.Hand = p.Hand[:len(p.Hand)-1]

	return discardedTile
}

func (p *PlayerState) AddToDiscard(tile *Tile) {
	if tile == nil {
		return
	}

	p.Discards = append(p.Discards, tile)
}

func (p *PlayerState) GetMatchingInHand(t *Tile) []*Tile {
	res := []*Tile{t}
	for _, h := range p.Hand {
		if h.Suit == t.Suit && h.Rank == t.Rank {
			res = append(res, h)
		}
	}
	return res
}
