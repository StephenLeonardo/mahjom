package domain

import "math/rand/v2"

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
