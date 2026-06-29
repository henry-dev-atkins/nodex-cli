package cmd

import (
	"flag"
	"fmt"

	"nodex-cli/internal"
)

func runSwitch(args []string) error {
	fs := flag.NewFlagSet("switch", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: nodex switch <name>")
	}
	name := fs.Arg(0)

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	if err := internal.SwitchBranch(root, name); err != nil {
		return err
	}
	fmt.Printf("Switched to branch: %s\n", name)
	return nil
}
