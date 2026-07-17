package domain

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
