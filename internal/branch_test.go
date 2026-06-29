package internal

import "testing"

func TestValidateBranchName(t *testing.T) {
	valid := []string{"main", "go-spike", "feat.v1", "branch_01"}
	for _, name := range valid {
		if err := ValidateBranchName(name); err != nil {
			t.Fatalf("expected valid branch name %q, got error: %v", name, err)
		}
	}

	invalid := []string{"", "bad name", "slash/name", "name?"}
	for _, name := range invalid {
		if err := ValidateBranchName(name); err == nil {
			t.Fatalf("expected invalid branch name %q to fail", name)
		}
	}
}
