package cmd

import (
	"flag"
	"fmt"
	"strings"

	"nodex-cli/internal"
)

func runBranch(args []string) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "rename":
			return runBranchRename(args[1:])
		case "delete":
			return runBranchDelete(args[1:])
		}
	}

	fs := flag.NewFlagSet("branch", flag.ContinueOnError)
	from := fs.String("from", "", "source branch name (default: current branch)")
	newName, parseArgs := peelLeadingPositional(args)
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if newName == "" {
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: nodex branch <name> [--from <branch>]")
		}
		newName = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex branch <name> [--from <branch>]")
	}

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	ws, err := internal.LoadWorkspace(root)
	if err != nil {
		return err
	}
	codex := &internal.ShellCodexAdapter{Bin: ws.CodexBin}
	b, err := internal.CreateBranch(root, newName, *from, codex)
	if err != nil {
		return err
	}

	fmt.Printf("Created branch: %s\n", b.Name)
	fmt.Printf("Parent: %s\n", b.Parent)
	fmt.Printf("Head session: %s\n", defaultText(b.HeadSessionID, "<none>"))
	if b.HeadSessionID == "" {
		fmt.Println("Note: automatic fork unavailable in v1; run `nodex chat " + b.Name + "` to start a session.")
	}
	return nil
}

func runBranchRename(args []string) error {
	fs := flag.NewFlagSet("branch rename", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: nodex branch rename <old> <new>")
	}
	oldName := fs.Arg(0)
	newName := fs.Arg(1)

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	if err := internal.RenameBranch(root, oldName, newName); err != nil {
		return err
	}
	fmt.Printf("Renamed branch: %s -> %s\n", oldName, newName)
	return nil
}

func runBranchDelete(args []string) error {
	fs := flag.NewFlagSet("branch delete", flag.ContinueOnError)
	force := fs.Bool("force", false, "detach child branches and delete anyway")
	yes := fs.Bool("yes", false, "skip confirmation prompt")
	name, parseArgs := peelLeadingPositional(args)
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if name == "" {
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: nodex branch delete <name> [--yes] [--force]")
		}
		name = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex branch delete <name> [--yes] [--force]")
	}

	if !*yes {
		prompt := fmt.Sprintf("Delete branch %s? [y/N]: ", name)
		if *force {
			prompt = fmt.Sprintf("Force-delete branch %s (children detached)? [y/N]: ", name)
		}
		if !confirm(prompt) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	if err := internal.DeleteBranch(root, name, *force); err != nil {
		return err
	}
	fmt.Printf("Deleted branch: %s\n", name)
	return nil
}
