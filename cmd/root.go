package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"nodex-cli/internal"
)

func Execute(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "status":
		return runStatus(args[1:])
	case "branches":
		return runBranches(args[1:])
	case "branch":
		return runBranch(args[1:])
	case "switch":
		return runSwitch(args[1:])
	case "chat":
		return runChat(args[1:])
	case "merge":
		return runMerge(args[1:])
	case "apply":
		return runApply(args[1:])
	case "mark-applied":
		return runMarkApplied(args[1:])
	case "attach-session":
		return runAttachSession(args[1:])
	case "show":
		return runShow(args[1:])
	case "mcp":
		return runMCP(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() {
	fmt.Println(`nodex - local branch-and-merge manager for Codex conversations

Usage:
  nodex init [--force]
  nodex status
  nodex branches
  nodex branch <name> [--from <branch>]
  nodex branch rename <old> <new>
  nodex branch delete <name> [--yes] [--force]
  nodex switch <name>
  nodex chat [branch]
  nodex attach-session <branch> <session-id>
  nodex merge <src> --into <dest> --mode <summary|decision|verbose> [--message "..."] [--last N] [--from-turn N --to-turn N] [--contains "..."] [--prepare]
  nodex apply <merge-id> [--echo-only] [--dry-run] [--mark-applied] [--yes]
  nodex mark-applied <merge-id> [--yes]
  nodex show <branch>
  nodex mcp serve`)
}

func findRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return internal.FindRepoRoot(cwd)
}

func requireInitializedRoot() (string, error) {
	root, err := findRoot()
	if err != nil {
		return "", err
	}
	if !internal.IsInitialized(root) {
		return "", errors.New("nodex not initialized in this repo")
	}
	return root, nil
}

func parseBoolFlag(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
