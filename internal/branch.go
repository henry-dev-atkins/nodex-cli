package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var branchNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type Branch struct {
	Name          string `json:"name"`
	Parent        string `json:"parent"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	RootSessionID string `json:"root_session_id"`
	HeadSessionID string `json:"head_session_id"`
	HeadSummary   string `json:"head_summary"`
	LastMergeID   string `json:"last_merge_id"`
	Notes         string `json:"notes"`
	Status        string `json:"status"`
}

func ValidateBranchName(name string) error {
	if name == "" {
		return errors.New("branch name is required")
	}
	if !branchNameRe.MatchString(name) {
		return fmt.Errorf("invalid branch name: %s (expected [A-Za-z0-9._-]+)", name)
	}
	return nil
}

func branchPath(root, name string) string {
	return filepath.Join(branchesDir(root), name+".json")
}

func LoadBranch(root, name string) (*Branch, error) {
	if err := ValidateBranchName(name); err != nil {
		return nil, err
	}
	var b Branch
	if err := readJSONFile(branchPath(root, name), &b); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("branch does not exist: %s", name)
		}
		return nil, err
	}
	return &b, nil
}

func SaveBranch(root string, b *Branch) error {
	if err := ValidateBranchName(b.Name); err != nil {
		return err
	}
	b.UpdatedAt = nowUTC()
	if b.CreatedAt == "" {
		b.CreatedAt = b.UpdatedAt
	}
	if strings.TrimSpace(b.Status) == "" {
		b.Status = "ready"
	}
	return writeJSONAtomic(branchPath(root, b.Name), b)
}

func ListBranches(root string) ([]*Branch, error) {
	entries, err := os.ReadDir(branchesDir(root))
	if err != nil {
		return nil, err
	}
	out := make([]*Branch, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		b, err := LoadBranch(root, name)
		if err != nil {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func BranchExists(root, name string) bool {
	return fileExists(branchPath(root, name))
}

func CurrentBranch(root string) (*Branch, error) {
	ws, err := LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	return LoadBranch(root, ws.CurrentBranch)
}

func SwitchBranch(root, name string) error {
	if _, err := LoadBranch(root, name); err != nil {
		return err
	}
	return SetCurrentBranch(root, name)
}

func CreateBranch(root, newName, fromName string, codex CodexAdapter) (*Branch, error) {
	if err := ValidateBranchName(newName); err != nil {
		return nil, err
	}
	if BranchExists(root, newName) {
		return nil, fmt.Errorf("branch already exists: %s", newName)
	}

	if fromName == "" {
		ws, err := LoadWorkspace(root)
		if err != nil {
			return nil, err
		}
		fromName = ws.CurrentBranch
	}
	src, err := LoadBranch(root, fromName)
	if err != nil {
		return nil, err
	}

	var newSessionID string
	if src.HeadSessionID != "" && codex != nil {
		if sid, err := codex.Fork(src.HeadSessionID); err == nil {
			newSessionID = sid
		}
	}

	b := &Branch{
		Name:          newName,
		Parent:        fromName,
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: newSessionID,
		HeadSessionID: newSessionID,
		HeadSummary:   "",
		LastMergeID:   "",
		Notes:         "",
		Status:        "ready",
	}
	if err := SaveBranch(root, b); err != nil {
		return nil, err
	}
	return b, nil
}

func AttachSession(root, branchName, sessionID string) (*Branch, error) {
	if err := ValidateSessionID(sessionID); err != nil {
		return nil, err
	}
	b, err := LoadBranch(root, branchName)
	if err != nil {
		return nil, err
	}
	b.HeadSessionID = sessionID
	if strings.TrimSpace(b.RootSessionID) == "" {
		b.RootSessionID = sessionID
	}
	if err := SaveBranch(root, b); err != nil {
		return nil, err
	}
	return b, nil
}

func RenameBranch(root, oldName, newName string) error {
	if err := ValidateBranchName(oldName); err != nil {
		return err
	}
	if err := ValidateBranchName(newName); err != nil {
		return err
	}
	if oldName == newName {
		return nil
	}
	if !BranchExists(root, oldName) {
		return fmt.Errorf("branch does not exist: %s", oldName)
	}
	if BranchExists(root, newName) {
		return fmt.Errorf("branch already exists: %s", newName)
	}

	b, err := LoadBranch(root, oldName)
	if err != nil {
		return err
	}
	b.Name = newName
	if err := SaveBranch(root, b); err != nil {
		return err
	}
	if err := os.Remove(branchPath(root, oldName)); err != nil {
		return err
	}

	items, err := ListBranches(root)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Parent != oldName {
			continue
		}
		item.Parent = newName
		if err := SaveBranch(root, item); err != nil {
			return err
		}
	}

	merges, err := ListMerges(root)
	if err != nil {
		return err
	}
	for _, m := range merges {
		changed := false
		if m.SourceBranch == oldName {
			m.SourceBranch = newName
			changed = true
		}
		if m.DestBranch == oldName {
			m.DestBranch = newName
			changed = true
		}
		if changed {
			if err := SaveMerge(root, m); err != nil {
				return err
			}
		}
	}

	ws, err := LoadWorkspace(root)
	if err != nil {
		return err
	}
	updateCurrentFile := false
	if ws.CurrentBranch == oldName {
		ws.CurrentBranch = newName
		updateCurrentFile = true
	}
	if ws.DefaultBranch == oldName {
		ws.DefaultBranch = newName
	}
	if err := SaveWorkspace(root, ws); err != nil {
		return err
	}
	if updateCurrentFile {
		if err := writeCurrentBranch(root, newName); err != nil {
			return err
		}
	}
	return nil
}

func DeleteBranch(root, name string, force bool) error {
	if err := ValidateBranchName(name); err != nil {
		return err
	}
	ws, err := LoadWorkspace(root)
	if err != nil {
		return err
	}
	if ws.CurrentBranch == name {
		return fmt.Errorf("cannot delete current branch: %s", name)
	}
	if ws.DefaultBranch == name {
		return fmt.Errorf("cannot delete default branch: %s", name)
	}
	if _, err := LoadBranch(root, name); err != nil {
		return err
	}

	items, err := ListBranches(root)
	if err != nil {
		return err
	}
	hasChildren := false
	for _, item := range items {
		if item.Parent == name {
			hasChildren = true
			if !force {
				return fmt.Errorf("branch %s has child branch %s; use --force to detach children", name, item.Name)
			}
		}
	}

	if hasChildren && force {
		for _, item := range items {
			if item.Parent != name {
				continue
			}
			item.Parent = ""
			if err := SaveBranch(root, item); err != nil {
				return err
			}
		}
	}

	return os.Remove(branchPath(root, name))
}
