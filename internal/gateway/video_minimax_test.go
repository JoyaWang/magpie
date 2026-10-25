package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// MiniMax takes the duration in whole seconds and a resolution tier, not a
// size; a mode or an aspect ratio is turned away (2013), so neither is sent.
func TestMiniMaxVideoBody(t *testing.T) {
	body, err := miniMaxVideoBody("MiniMax-Hailuo-2.3", filming{Prompt: "a panda", Seconds: "6", Size: "1280x720"})
	if err != nil {
		t.Fatal(err)
	}
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req["duration"] != float64(6) {
		t.Fatalf("duration = %v, want 6 (a whole number of seconds)", req["duration"])
	}
	if req["resolution"] != "768P" {
		t.Fatalf("resolution = %v, want 768P for a 720p size", req["resolution"])
	}
	for _, unwanted := range []string{"mode", "aspect_ratio", "size", "seconds"} {
		if _, ok := req[unwanted]; ok {
			t.Fatalf("%q was sent: %s", unwanted, body)
		}
	}
	if req["model"] != "MiniMax-Hailuo-2.3" || req["prompt"] != "a panda" {
		t.Fatalf("model/prompt lost: %s", body)
	}
}

func TestMiniMaxVideoBodyRejectsBadSeconds(t *testing.T) {
	if _, err := miniMaxVideoBody("MiniMax-Hailuo-2.3", filming{Prompt: "x", Seconds: "2.5"}); err == nil {
		t.Fatal("a fractional duration must be refused, as every other videos request does")
	}
}

// MiniMax reports failure inside a 200: the envelope, not the status line, is
// its answer. 2056 is out of quota, 2069 a model it no longer serves.
func TestMiniMaxEnvelopeError(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"out of quota", `{"task_id":"","base_resp":{"status_code":2056,"status_msg":"已达到 Token Plan 用量上限"}}`, "2056"},
		{"model retired", `{"choices":null,"base_resp":{"status_code":2069,"status_msg":"The model \"MiniMax-M2\" is no longer available."}}`, "2069"},
		{"accepted", `{"task_id":"t1","base_resp":{"status_code":0}}`, ""},
		{"no envelope", `{"task_id":"t1"}`, ""},
		{"not json", `<html>oops</html>`, ""},
	}
	for _, tc := range cases {
		err := miniMaxEnvelopeError([]byte(tc.body))
		if tc.want == "" {
			if err != nil {
				t.Fatalf("%s: unexpected error %v", tc.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: err = %v, want it to name %s", tc.name, err, tc.want)
		}
	}
}

// MiniMax's words for a video read as magpie's: Success is done, Fail failed,
// and everything before them is still running.
func TestMiniMaxStateMapping(t *testing.T) {
	for raw, want := range map[string]string{
		"Success": "done", "success": "done", "succeeded": "done",
		"Fail": "failed", "failed": "failed",
		"Processing": "running", "Preparing": "running", "Queueing": "running", "": "running",
	} {
		st := videoState{Status: raw}
		miniMaxState(&st)
		if st.Status != want {
			t.Errorf("miniMaxState(%q) = %q, want %q", raw, st.Status, want)
		}
	}
}

// A MiniMax provider is one of the vendors magpie itself asks for videos, so
// naming one of its models is not "can't make videos".
func TestMiniMaxOnlyFilmsItsOwnModels(t *testing.T) {
	for _, id := range []string{"minimax", "minimax-cn"} {
		if !filmsMiniMax(provider.Provider{ID: id, Preset: id}) {
			t.Fatalf("preset %q must be recognized as making videos itself", id)
		}
	}
	for _, id := range []string{"grok", "openai", "volcengine"} {
		if filmsMiniMax(provider.Provider{ID: id, Preset: id}) {
			t.Fatalf("preset %q is not MiniMax", id)
		}
	}
}
