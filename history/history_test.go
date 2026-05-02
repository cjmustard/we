package history_test

import (
	"testing"
	_ "unsafe"

	mcblock "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/we/history"
	"github.com/df-mc/we/parse"
)

//go:linkname finaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func finaliseBlockRegistry()

func withTx(t *testing.T, f func(tx *world.Tx)) {
	t.Helper()
	finaliseBlockRegistry()
	w := world.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()
	<-w.Exec(f)
}

func TestBrushHistoryIsIsolatedFromMainHistory(t *testing.T) {
	withTx(t, func(tx *world.Tx) {
		pos := cube.Pos{0, 0, 0}
		h := history.NewHistory(10)
		mainBatch := history.NewBatch(false)
		mainBatch.SetBlock(tx, pos, mcblock.Stone{})
		h.Record(mainBatch)
		brushBatch := history.NewBatch(true)
		brushBatch.SetBlock(tx, pos, mcblock.Gold{})
		h.Record(brushBatch)

		if !h.Undo(tx, false) {
			t.Fatal("main undo returned false")
		}
		if !parse.SameBlock(tx.Block(pos), mcblock.Air{}) {
			t.Fatal("main undo did not skip brush stack and restore main before-state")
		}
		if !h.Undo(tx, true) {
			t.Fatal("brush undo returned false")
		}
		if !parse.SameBlock(tx.Block(pos), mcblock.Stone{}) {
			t.Fatal("brush undo did not restore brush before-state")
		}
	})
}

func TestRecordReturnsChangedCountAndSkipsNoOps(t *testing.T) {
	var failure string
	withTx(t, func(tx *world.Tx) {
		h := history.NewHistory(10)
		batch := history.NewBatch(false)
		batch.SetBlock(tx, cube.Pos{0, 0, 0}, mcblock.Stone{})
		batch.SetBlock(tx, cube.Pos{1, 0, 0}, mcblock.Dirt{})
		batch.SetBlock(tx, cube.Pos{2, 0, 0}, mcblock.Air{})

		if got := h.Record(batch); got != 2 {
			failure = "Record() changed count mismatch"
			return
		}
		if got := h.Record(history.NewBatch(false)); got != 0 {
			failure = "Record() stored empty batch"
		}
	})
	if failure != "" {
		t.Fatal(failure)
	}
}
