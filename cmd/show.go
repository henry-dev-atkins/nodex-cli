package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: nodex show <branch>")
	}
	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	b, err := internal.LoadBranch(root, fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Printf("Branch: %s\n", b.Name)
	fmt.Printf("Parent: %s\n", defaultText(b.Parent, "<none>"))
	fmt.Printf("Root session: %s\n", defaultText(b.RootSessionID, "<none>"))
	fmt.Printf("Head session: %s\n", defaultText(b.HeadSessionID, "<none>"))
	fmt.Printf("Last merge: %s\n", defaultText(b.LastMergeID, "<none>"))
	fmt.Printf("Last summary: %s\n", defaultText(b.HeadSummary, "<none>"))
	return nil
}
