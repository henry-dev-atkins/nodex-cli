package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex status")
	}
	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	ws, err := internal.LoadWorkspace(root)
	if err != nil {
		return err
	}
	cur, err := internal.LoadBranch(root, ws.CurrentBranch)
	if err != nil {
		return err
	}
	branches, err := internal.ListBranches(root)
	if err != nil {
		return err
	}
	pending, err := internal.PendingMergesCount(root)
	if err != nil {
		return err
	}
	fmt.Printf("Current branch: %s\n", ws.CurrentBranch)
	fmt.Printf("Head session: %s\n", defaultText(cur.HeadSessionID, "<none>"))
	fmt.Printf("Branches: %d\n", len(branches))
	fmt.Printf("Pending merges: %d\n", pending)
	return nil
}

func defaultText(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
