package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultBranchName = "main"
	defaultCodexBin   = "codex"
)

type Workspace struct {
	Version       int    `json:"version"`
	CreatedAt     string `json:"created_at"`
	DefaultBranch string `json:"default_branch"`
	CurrentBranch string `json:"current_branch"`
	CodexBin      string `json:"codex_bin"`
}

func FindRepoRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	cur := abs
	for {
		if dirExists(filepath.Join(cur, ".nodex")) || dirExists(filepath.Join(cur, ".git")) {
			return cur, nil
		}
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
	}
	return abs, nil
}

func NodexDir(root string) string {
	return filepath.Join(root, ".nodex")
}

func workspacePath(root string) string {
	return filepath.Join(NodexDir(root), "workspace.json")
}

func currentBranchPath(root string) string {
	return filepath.Join(NodexDir(root), "current_branch")
}

func branchesDir(root string) string {
	return filepath.Join(NodexDir(root), "branches")
}

func mergesDir(root string) string {
	return filepath.Join(NodexDir(root), "merges")
}

func cacheDir(root string) string {
	return filepath.Join(NodexDir(root), "cache")
}

func InitWorkspace(root string, codexBin string, force bool) error {
	nodex := NodexDir(root)
	if dirExists(nodex) {
		if !force {
			return fmt.Errorf("%s already exists (use --force to reinitialize)", nodex)
		}
		if err := os.RemoveAll(nodex); err != nil {
			return err
		}
	}
	if codexBin == "" {
		codexBin = defaultCodexBin
	}
	if err := ensureDir(branchesDir(root)); err != nil {
		return err
	}
	if err := ensureDir(mergesDir(root)); err != nil {
		return err
	}
	if err := ensureDir(cacheDir(root)); err != nil {
		return err
	}

	ws := &Workspace{
		Version:       1,
		CreatedAt:     nowUTC(),
		DefaultBranch: defaultBranchName,
		CurrentBranch: defaultBranchName,
		CodexBin:      codexBin,
	}
	if err := SaveWorkspace(root, ws); err != nil {
		return err
	}
	if err := writeCurrentBranch(root, defaultBranchName); err != nil {
		return err
	}

	mainBranch := &Branch{
		Name:          defaultBranchName,
		Parent:        "",
		CreatedAt:     nowUTC(),
		UpdatedAt:     nowUTC(),
		RootSessionID: "",
		HeadSessionID: "",
		HeadSummary:   "",
		LastMergeID:   "",
		Notes:         "",
		Status:        "ready",
	}
	return SaveBranch(root, mainBranch)
}

func IsInitialized(root string) bool {
	return fileExists(workspacePath(root))
}

func LoadWorkspace(root string) (*Workspace, error) {
	var ws Workspace
	if err := readJSONFile(workspacePath(root), &ws); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("nodex not initialized in this repo")
		}
		return nil, err
	}
	if ws.CodexBin == "" {
		ws.CodexBin = defaultCodexBin
	}
	return &ws, nil
}

func SaveWorkspace(root string, ws *Workspace) error {
	return writeJSONAtomic(workspacePath(root), ws)
}

func writeCurrentBranch(root, name string) error {
	body := []byte(strings.TrimSpace(name) + "\n")
	return writeFileAtomic(currentBranchPath(root), body, 0o644)
}

func SetCurrentBranch(root, name string) error {
	ws, err := LoadWorkspace(root)
	if err != nil {
		return err
	}
	ws.CurrentBranch = name
	if err := SaveWorkspace(root, ws); err != nil {
		return err
	}
	return writeCurrentBranch(root, name)
}
