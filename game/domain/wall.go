package domain

import "math/rand/v2"

type IWall interface {
	Shuffle(seed uint64)
	Draw() (*Tile, bool)
	Remaining() int64
}

// DrawPile is the remaining wall; normal draws and post-kang draws both come from this pile
// (engine may take from front vs back; order is equivalent for randomness).
type Wall struct {
	DrawPile []*Tile
}

func (w *Wall) Shuffle(seed uint64) {
	source := rand.NewPCG(seed, 0)
	r := rand.New(source)
	r.Shuffle(len(w.DrawPile), func(i, j int) {
		w.DrawPile[i], w.DrawPile[j] = w.DrawPile[j], w.DrawPile[i]
	})
}

// Draw removes and returns the next tile from the front of the wall.
// It returns false when the wall is empty.
func (w *Wall) Draw() (*Tile, bool) {
	if len(w.DrawPile) == 0 {
		return nil, false
	}

	tile := w.DrawPile[0]
	w.DrawPile[0] = nil
	w.DrawPile = w.DrawPile[1:]
	return tile, true
}

// Remaining returns the number of tiles left in the wall.
func (w *Wall) Remaining() int64 {
	return int64(len(w.DrawPile))
}
