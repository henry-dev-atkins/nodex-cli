package internal

import (
	"testing"
)

type fakeTranscriptStore struct {
	tx *Transcript
}

func (f *fakeTranscriptStore) Extract(sessionID string) (*Transcript, error) {
	return f.tx, nil
}

func TestPrepareMergeContainsFilter(t *testing.T) {
	root := t.TempDir()
	if err := InitWorkspace(root, "codex", false); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "src",
		Parent:        "main",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_src",
		HeadSessionID: "sess_src",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	main, err := LoadBranch(root, "main")
	if err != nil {
		t.Fatal(err)
	}
	main.RootSessionID = "sess_main"
	main.HeadSessionID = "sess_main"
	if err := SaveBranch(root, main); err != nil {
		t.Fatal(err)
	}

	tx := &fakeTranscriptStore{
		tx: &Transcript{
			SessionID: "sess_src",
			Turns: []Turn{
				{ID: "T1", User: "General idea", Assistant: "Do architecture cleanup first."},
				{ID: "T2", User: "Any risk here?", Assistant: "Major risk is migration churn."},
			},
		},
	}

	m, err := PrepareMerge(root, "src", "main", "summary", "", 0, 0, 0, "risk", tx)
	if err != nil {
		t.Fatal(err)
	}
	if m.Selection.Contains != "risk" {
		t.Fatalf("expected contains filter persisted, got %q", m.Selection.Contains)
	}
	if len(m.SourceMaterial.Turns) != 1 {
		t.Fatalf("expected single filtered turn, got %d", len(m.SourceMaterial.Turns))
	}
	if m.SourceMaterial.Turns[0].ID != "T2" {
		t.Fatalf("expected filtered turn T2, got %s", m.SourceMaterial.Turns[0].ID)
	}
}

func TestPrepareMergeContainsNoMatch(t *testing.T) {
	root := t.TempDir()
	if err := InitWorkspace(root, "codex", false); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "src",
		Parent:        "main",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_src",
		HeadSessionID: "sess_src",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	main, err := LoadBranch(root, "main")
	if err != nil {
		t.Fatal(err)
	}
	main.RootSessionID = "sess_main"
	main.HeadSessionID = "sess_main"
	if err := SaveBranch(root, main); err != nil {
		t.Fatal(err)
	}
	tx := &fakeTranscriptStore{
		tx: &Transcript{
			SessionID: "sess_src",
			Turns: []Turn{
				{ID: "T1", User: "General", Assistant: "Cleanup"},
			},
		},
	}
	if _, err := PrepareMerge(root, "src", "main", "summary", "", 0, 0, 0, "no-such-text", tx); err == nil {
		t.Fatal("expected contains filter no-match to fail")
	}
}
