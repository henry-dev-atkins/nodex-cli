package cmd

import (
	"fmt"

	"nodex-cli/internal"
)

func runMCP(args []string) error {
	if len(args) != 1 || args[0] != "serve" {
		return fmt.Errorf("usage: nodex mcp serve")
	}
	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	return internal.ServeMCP(root)
}
