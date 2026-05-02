package edit

import (
	mcblock "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/we/geo"
	"github.com/df-mc/we/history"
)

// denseBlockEntry is one prepared block write plus the batch index that tracks
// its before/after snapshots. Keeping this tiny data carrier explicit lets the
// fast BuildStructure path preserve the same history semantics as Batch.SetBlock.
type denseBlockEntry struct {
	Pos   cube.Pos
	Index int
	Block world.Block
	Liq   world.Liquid
}

type denseBlockStructure struct {
	d       [3]int
	entries []denseBlockEntry
}

func (s denseBlockStructure) Dimensions() [3]int { return s.d }

func (s denseBlockStructure) At(x, y, z int, _ func(x, y, z int) world.Block) (world.Block, world.Liquid) {
	entry := s.entries[(x*s.d[1]+y)*s.d[2]+z]
	return entry.Block, entry.Liq
}

// writeDenseArea applies a full cuboid through Dragonfly's chunk-batched
// BuildStructure path. It snapshots every position before the structure write
// and refreshes after snapshots afterwards so undo/redo behavior stays identical
// to repeated Batch.SetBlock calls.
func writeDenseArea(tx *world.Tx, area geo.Area, blockAt func(cube.Pos) world.Block, batch *history.Batch) {
	n := int(area.Volume())
	batch.Grow(n)
	entries := make([]denseBlockEntry, 0, n)
	area.Range(func(x, y, z int) {
		pos := cube.Pos{x, y, z}
		block := blockAt(pos)
		var liq world.Liquid
		if block == nil {
			block = mcblock.Air{}
		}
		entries = append(entries, denseBlockEntry{Pos: pos, Index: batch.EnsurePos(tx, pos), Block: block, Liq: liq})
	})
	tx.BuildStructure(area.Min, denseBlockStructure{d: [3]int{area.Dx(), area.Dy(), area.Dz()}, entries: entries})
	for _, entry := range entries {
		batch.SetAfterForIndex(tx, entry.Index, entry.Pos)
	}
}
