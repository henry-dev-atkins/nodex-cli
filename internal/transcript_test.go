package internal

import "testing"

func TestSelectTurnsPrecedence(t *testing.T) {
	tz := &Transcript{
		SessionID: "sess_1",
		Turns: []Turn{
			{ID: "T1", User: "u1", Assistant: "a1"},
			{ID: "T2", User: "u2", Assistant: "a2"},
			{ID: "T3", User: "u3", Assistant: "a3"},
			{ID: "T4", User: "u4", Assistant: "a4"},
		},
	}

	last := SelectTurns(tz, 2, 0, 0)
	if len(last) != 2 || last[0].ID != "T3" || last[1].ID != "T4" {
		t.Fatalf("expected last 2 turns, got %#v", last)
	}

	rng := SelectTurns(tz, 0, 2, 3)
	if len(rng) != 2 || rng[0].ID != "T2" || rng[1].ID != "T3" {
		t.Fatalf("expected range turns T2..T3, got %#v", rng)
	}

	def := SelectTurns(tz, 0, 0, 0)
	if len(def) != 3 || def[0].ID != "T2" || def[2].ID != "T4" {
		t.Fatalf("expected default tail 3 turns, got %#v", def)
	}
}
