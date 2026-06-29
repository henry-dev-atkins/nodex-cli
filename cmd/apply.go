package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"nodex-cli/internal"
)

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	echoOnly := fs.Bool("echo-only", false, "print merge block only")
	dryRun := fs.Bool("dry-run", false, "print merge block and planned actions only; no side effects")
	markApplied := fs.Bool("mark-applied", false, "set merge artifact as applied")
	yes := fs.Bool("yes", false, "assume yes for confirmation prompts")
	mergeID, parseArgs := peelLeadingPositional(args)
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if mergeID == "" {
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: nodex apply <merge-id> [--echo-only] [--dry-run] [--mark-applied] [--yes]")
		}
		mergeID = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex apply <merge-id> [--echo-only] [--dry-run] [--mark-applied] [--yes]")
	}
	if *dryRun && *markApplied {
		return fmt.Errorf("--dry-run cannot be combined with --mark-applied; use `nodex mark-applied %s` separately", mergeID)
	}

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	m, err := internal.LoadMerge(root, mergeID)
	if err != nil {
		return err
	}

	fmt.Printf("Applying merge %s into %s\n", m.ID, m.DestBranch)
	fmt.Println("--- BEGIN MERGE BLOCK ---")
	fmt.Println(m.PreviewText)
	fmt.Println("--- END MERGE BLOCK ---")

	if *dryRun {
		fmt.Println("Dry run: no Codex resume and no merge state changes were made.")
	} else if !*echoOnly {
		if strings.TrimSpace(m.DestSessionID) == "" {
			fmt.Println("No destination head session id found; paste manually into destination branch chat.")
		} else {
			ws, err := internal.LoadWorkspace(root)
			if err != nil {
				return err
			}
			codex := &internal.ShellCodexAdapter{Bin: ws.CodexBin}
			if err := codex.Detect(); err != nil {
				return err
			}
			if err := codex.Resume(m.DestSessionID); err != nil {
				_ = internal.MarkMergeApplyError(root, m.ID, err.Error())
				return err
			}
		}
	}

	shouldMark := false
	if *markApplied {
		if *yes {
			shouldMark = true
		} else {
			shouldMark = confirm("Mark this merge as applied? [y/N]: ")
		}
	}
	if shouldMark {
		if _, err := internal.MarkMergeApplied(root, m.ID); err != nil {
			return err
		}
		fmt.Printf("Marked applied: %s\n", m.ID)
	} else {
		fmt.Printf("Left unapplied: %s\n", m.ID)
	}

	fmt.Println("Next:")
	fmt.Printf("1. paste this into Codex on branch %s\n", m.DestBranch)
	fmt.Println("2. continue the conversation")
	fmt.Printf("3. optionally run: nodex mark-applied %s --yes\n", m.ID)
	return nil
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	rd := bufio.NewReader(os.Stdin)
	line, _ := rd.ReadString('\n')
	return parseBoolFlag(line)
}
