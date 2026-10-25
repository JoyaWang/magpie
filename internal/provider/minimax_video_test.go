package provider

import (
	"slices"
	"testing"
)

// MiniMax's video models are declared on the preset itself (it has one list
// and no regions), and PlanVideos reads them as a region's.
func TestMiniMaxPresetDeclaresVideoModels(t *testing.T) {
	for _, id := range []string{"minimax", "minimax-cn"} {
		pr := Preset(id)
		if pr == nil {
			t.Fatalf("preset %q missing", id)
		}
		if len(pr.Videos) == 0 {
			t.Fatalf("preset %q declares no video models", id)
		}
		p := Provider{ID: id, Preset: id, Chat: pr.Chat}
		got := modelIDs(p.PlanVideos())
		if !slices.Equal(got, pr.Videos) {
			t.Fatalf("%s PlanVideos = %v, want the preset's %v", id, got, pr.Videos)
		}
		if !slices.Contains(got, "MiniMax-Hailuo-2.3") {
			t.Fatalf("%s video models = %v, want Hailuo 2.3 among them", id, got)
		}
	}
}
