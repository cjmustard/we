package service_test

import (
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/we/edit"
	"github.com/df-mc/we/guardrail"
	"github.com/df-mc/we/history"
	"github.com/df-mc/we/service"
)

func TestApplyBrushRejectsLargeBrush(t *testing.T) {
	withTx(t, func(tx *world.Tx) {
		batch := history.NewBatch(true)
		err := service.ApplyBrush(tx, service.BrushActor{}, cube.Pos{0, 0, 0}, service.BrushConfig{
			Type:   "cube",
			Length: 2,
			Width:  1,
			Height: 1,
		}, edit.DefaultSchematicStore(), guardrail.Limits{MaxBrushVolume: 1}, batch)
		if err == nil || !strings.Contains(err.Error(), "brush volume 2 exceeds limit 1") {
			t.Fatalf("ApplyBrush error = %v, want brush limit error", err)
		}
		if batch.Len() != 0 {
			t.Fatalf("batch Len = %d, want 0", batch.Len())
		}
	})
}

func TestBrushAnchorFromSurfaceExtendsShapeOutward(t *testing.T) {
	cfg := service.BrushConfig{Type: service.BrushSphere, Radius: 3, Height: 7}
	tests := []struct {
		name    string
		surface cube.Pos
		face    cube.Face
		want    cube.Pos
	}{
		{name: "up", surface: cube.Pos{10, 65, 10}, face: cube.FaceUp, want: cube.Pos{10, 68, 10}},
		{name: "down", surface: cube.Pos{10, 63, 10}, face: cube.FaceDown, want: cube.Pos{10, 60, 10}},
		{name: "east", surface: cube.Pos{11, 64, 10}, face: cube.FaceEast, want: cube.Pos{14, 64, 10}},
		{name: "west", surface: cube.Pos{9, 64, 10}, face: cube.FaceWest, want: cube.Pos{6, 64, 10}},
		{name: "south", surface: cube.Pos{10, 64, 11}, face: cube.FaceSouth, want: cube.Pos{10, 64, 14}},
		{name: "north", surface: cube.Pos{10, 64, 9}, face: cube.FaceNorth, want: cube.Pos{10, 64, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.BrushAnchorFromSurface(tt.surface, tt.face, cfg); got != tt.want {
				t.Fatalf("BrushAnchorFromSurface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBrushAnchorFromSurfaceLeavesLineAtSurface(t *testing.T) {
	cfg := service.BrushConfig{Type: service.BrushLine, Radius: 3, Height: 7}
	surface := cube.Pos{10, 65, 10}
	if got := service.BrushAnchorFromSurface(surface, cube.FaceUp, cfg); got != surface {
		t.Fatalf("BrushAnchorFromSurface() = %v, want %v", got, surface)
	}
}
