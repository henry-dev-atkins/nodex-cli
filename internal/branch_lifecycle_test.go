package internal

import "testing"

func TestAttachSessionSetsHeadAndRoot(t *testing.T) {
	root := t.TempDir()
	if err := InitWorkspace(root, "codex", false); err != nil {
		t.Fatal(err)
	}
	b := &Branch{
		Name:          "spike",
		Parent:        "main",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "",
		HeadSessionID: "",
		Status:        "ready",
	}
	if err := SaveBranch(root, b); err != nil {
		t.Fatal(err)
	}
	out, err := AttachSession(root, "spike", "sess_spike_1")
	if err != nil {
		t.Fatal(err)
	}
	if out.HeadSessionID != "sess_spike_1" {
		t.Fatalf("expected head session to be set, got %q", out.HeadSessionID)
	}
	if out.RootSessionID != "sess_spike_1" {
		t.Fatalf("expected root session to be set when empty, got %q", out.RootSessionID)
	}
}

func TestRenameBranchUpdatesReferences(t *testing.T) {
	root := t.TempDir()
	if err := InitWorkspace(root, "codex", false); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "feature",
		Parent:        "main",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_feature",
		HeadSessionID: "sess_feature",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "child",
		Parent:        "feature",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_child",
		HeadSessionID: "sess_child",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	ws, err := LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	ws.CurrentBranch = "feature"
	ws.DefaultBranch = "feature"
	if err := SaveWorkspace(root, ws); err != nil {
		t.Fatal(err)
	}
	if err := writeCurrentBranch(root, "feature"); err != nil {
		t.Fatal(err)
	}
	if err := SaveMerge(root, &MergeArtifact{
		ID:              "merge-0001",
		CreatedAt:       nowUTC(),
		SourceBranch:    "feature",
		DestBranch:      "main",
		SourceSessionID: "sess_feature",
		DestSessionID:   "sess_main",
		Mode:            "summary",
		Selection:       Selection{},
		SourceMaterial:  SourceMaterial{},
		RedactionReport: RedactionReport{},
		PreviewText:     "",
	}); err != nil {
		t.Fatal(err)
	}

	if err := RenameBranch(root, "feature", "renamed"); err != nil {
		t.Fatal(err)
	}
	if BranchExists(root, "feature") {
		t.Fatal("old branch should not exist after rename")
	}
	if !BranchExists(root, "renamed") {
		t.Fatal("new branch should exist after rename")
	}
	child, err := LoadBranch(root, "child")
	if err != nil {
		t.Fatal(err)
	}
	if child.Parent != "renamed" {
		t.Fatalf("expected child parent updated to renamed, got %q", child.Parent)
	}
	ws, err = LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if ws.CurrentBranch != "renamed" || ws.DefaultBranch != "renamed" {
		t.Fatalf("workspace references not updated, got current=%q default=%q", ws.CurrentBranch, ws.DefaultBranch)
	}
	m, err := LoadMerge(root, "merge-0001")
	if err != nil {
		t.Fatal(err)
	}
	if m.SourceBranch != "renamed" {
		t.Fatalf("expected merge source updated to renamed, got %q", m.SourceBranch)
	}
}

func TestDeleteBranchGuardsAndForceDetach(t *testing.T) {
	root := t.TempDir()
	if err := InitWorkspace(root, "codex", false); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "to-delete",
		Parent:        "main",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_del",
		HeadSessionID: "sess_del",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	if err := SaveBranch(root, &Branch{
		Name:          "child",
		Parent:        "to-delete",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "sess_child",
		HeadSessionID: "sess_child",
		Status:        "ready",
	}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBranch(root, "main", false); err == nil {
		t.Fatal("expected delete current/default branch to fail")
	}
	if err := DeleteBranch(root, "to-delete", false); err == nil {
		t.Fatal("expected delete with child branch to fail without force")
	}
	if err := DeleteBranch(root, "to-delete", true); err != nil {
		t.Fatal(err)
	}
	if BranchExists(root, "to-delete") {
		t.Fatal("branch should be deleted")
	}
	child, err := LoadBranch(root, "child")
	if err != nil {
		t.Fatal(err)
	}
	if child.Parent != "" {
		t.Fatalf("expected child parent detached on force delete, got %q", child.Parent)
	}
}
