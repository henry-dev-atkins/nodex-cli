package internal

import (
	"strings"
	"testing"
)

func TestRenderMergePreviewDecision(t *testing.T) {
	turns := []Turn{
		{ID: "T1", User: "What should we do?", Assistant: "Switch to Go for CLI speed."},
	}
	out := RenderMergePreview("decision", "go-spike", "main", "Decide language", turns)
	if !strings.Contains(out, "Imported branch: go-spike") {
		t.Fatalf("missing source branch in preview: %s", out)
	}
	if !strings.Contains(out, "Mode: decision") {
		t.Fatalf("missing mode in preview: %s", out)
	}
	if !strings.Contains(out, "Question: Decide language") {
		t.Fatalf("missing message in preview: %s", out)
	}
	if !strings.Contains(out, "Use this imported decision context") {
		t.Fatalf("missing destination instruction in preview: %s", out)
	}
}
