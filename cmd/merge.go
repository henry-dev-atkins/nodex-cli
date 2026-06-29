package cmd

import (
	"flag"
	"fmt"
	"strings"

	"nodex-cli/internal"
)

func runMerge(args []string) error {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	dest := fs.String("into", "", "destination branch")
	mode := fs.String("mode", "", "summary|decision|verbose")
	message := fs.String("message", "", "decision question or merge note")
	last := fs.Int("last", 0, "include last N turns from source transcript")
	fromTurn := fs.Int("from-turn", 0, "start turn index (1-based)")
	toTurn := fs.Int("to-turn", 0, "end turn index (1-based)")
	contains := fs.String("contains", "", "include only turns containing this text")
	prepareOnly := fs.Bool("prepare", false, "prepare artifact only")
	src, parseArgs := peelLeadingPositional(args)
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if src == "" {
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: nodex merge <src> --into <dest> --mode <summary|decision|verbose> [--message ...] [--last N] [--from-turn N --to-turn N] [--contains ...] [--prepare]")
		}
		src = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return fmt.Errorf("usage: nodex merge <src> --into <dest> --mode <summary|decision|verbose> [--message ...] [--last N] [--from-turn N --to-turn N] [--contains ...] [--prepare]")
	}
	if *dest == "" || *mode == "" {
		return fmt.Errorf("usage: nodex merge <src> --into <dest> --mode <summary|decision|verbose> [--message ...] [--last N] [--from-turn N --to-turn N] [--contains ...] [--prepare]")
	}

	root, err := requireInitializedRoot()
	if err != nil {
		return err
	}
	tx := internal.NewDefaultTranscriptStore(root)
	m, err := internal.PrepareMerge(root, src, *dest, *mode, *message, *last, *fromTurn, *toTurn, *contains, tx)
	if err != nil {
		return err
	}

	fmt.Printf("Prepared merge: %s\n", m.ID)
	fmt.Printf("Source: %s\n", m.SourceBranch)
	fmt.Printf("Destination: %s\n", m.DestBranch)
	fmt.Printf("Mode: %s\n\n", m.Mode)
	fmt.Println(m.PreviewText)
	if len(m.RedactionReport.Findings) > 0 {
		fmt.Printf("\nWarning: redaction findings=%d\n", len(m.RedactionReport.Findings))
		for _, f := range m.RedactionReport.Findings {
			fmt.Printf("- %s %s (%s)\n", f.Type, f.ValuePreview, f.Reason)
		}
	}
	if *prepareOnly {
		fmt.Printf("\nRun: nodex apply %s\n", m.ID)
		return nil
	}
	fmt.Printf("\nRun: nodex apply %s\n", m.ID)
	return nil
}

func peelLeadingPositional(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}
	if strings.HasPrefix(args[0], "-") {
		return "", args
	}
	return args[0], args[1:]
}
