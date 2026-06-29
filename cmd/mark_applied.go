package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runMarkApplied(args []string) error {
	fs := flag.NewFlagSet("mark-applied", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "skip confirmation prompt")
	mergeID, parseArgs := peelLeadingPositional(args)
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if mergeID == "" {
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: nodex mark-applied <merge-id> [--yes]")
		}
		mergeID = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex mark-applied <merge-id> [--yes]")
	}

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	m, err := internal.LoadMerge(root, mergeID)
	if err != nil {
		return err
	}
	if m.Applied {
		fmt.Printf("Merge already applied: %s\n", m.ID)
		return nil
	}

	if !*yes && !confirm(fmt.Sprintf("Mark merge %s as applied? [y/N]: ", mergeID)) {
		fmt.Println("Cancelled.")
		return nil
	}
	m, err = internal.MarkMergeApplied(root, mergeID)
	if err != nil {
		return err
	}
	fmt.Printf("Marked applied: %s\n", m.ID)
	return nil
}
