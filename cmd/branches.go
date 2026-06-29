package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runBranches(args []string) error {
	fs := flag.NewFlagSet("branches", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex branches")
	}
	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	ws, err := internal.LoadWorkspace(root)
	if err != nil {
		return err
	}
	items, err := internal.ListBranches(root)
	if err != nil {
		return err
	}
	for _, b := range items {
		marker := " "
		if b.Name == ws.CurrentBranch {
			marker = "*"
		}
		if b.Parent == "" {
			fmt.Printf("%s %-10s %s\n", marker, b.Name, defaultText(b.HeadSessionID, "<none>"))
		} else {
			fmt.Printf("%s %-10s %s parent=%s\n", marker, b.Name, defaultText(b.HeadSessionID, "<none>"), b.Parent)
		}
	}
	return nil
}
