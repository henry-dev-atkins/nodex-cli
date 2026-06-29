package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "reinitialize .nodex if it already exists")
	codexBin := fs.String("codex-bin", "codex", "codex binary name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex init [--force]")
	}
	root, err := findRoot()
	if err != nil {
		return err
	}
	codex := &internal.ShellCodexAdapter{Bin: *codexBin}
	if err := codex.Detect(); err != nil {
		return err
	}
	if err := internal.InitWorkspace(root, *codexBin, *force); err != nil {
		return err
	}
	fmt.Println("Initialized Nodex workspace.")
	fmt.Println("Current branch: main")
	fmt.Println("Run: nodex chat")
	return nil
}
