package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runChat(args []string) error {
	fs := flag.NewFlagSet("chat", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("usage: nodex chat [branch]")
	}
	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	ws, err := internal.LoadWorkspace(root)
	if err != nil {
		return err
	}
	branchName := ws.CurrentBranch
	if fs.NArg() == 1 {
		branchName = fs.Arg(0)
	}

	b, err := internal.LoadBranch(root, branchName)
	if err != nil {
		return err
	}
	codex := &internal.ShellCodexAdapter{Bin: ws.CodexBin}
	if err := codex.Detect(); err != nil {
		return err
	}

	if b.HeadSessionID == "" {
		fmt.Printf("Starting Codex for branch: %s\n", b.Name)
		if _, err := codex.Start(); err != nil {
			return err
		}
		fmt.Println("Session capture is explicit in this build; branch head remains unchanged.")
		fmt.Printf("Next: nodex attach-session %s <session-id>\n", b.Name)
		return nil
	}

	fmt.Printf("Resuming branch %s (%s)\n", b.Name, b.HeadSessionID)
	return codex.Resume(b.HeadSessionID)
}
