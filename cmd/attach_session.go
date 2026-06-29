package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runAttachSession(args []string) error {
	fs := flag.NewFlagSet("attach-session", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: nodex attach-session <branch> <session-id>")
	}
	branchName := fs.Arg(0)
	sessionID := fs.Arg(1)

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	b, err := internal.AttachSession(root, branchName, sessionID)
	if err != nil {
		return err
	}
	fmt.Printf("Attached session %s to branch %s\n", b.HeadSessionID, b.Name)
	return nil
}
