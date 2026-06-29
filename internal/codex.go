package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CodexAdapter interface {
	Detect() error
	Resume(sessionID string) error
	Start() (string, error)
	Fork(sessionID string) (string, error)
}

type ShellCodexAdapter struct {
	Bin string
}

func (c *ShellCodexAdapter) bin() string {
	if strings.TrimSpace(c.Bin) == "" {
		return defaultCodexBin
	}
	return c.Bin
}

func (c *ShellCodexAdapter) Detect() error {
	_, err := exec.LookPath(c.bin())
	if err != nil {
		return errors.New("codex binary not found on PATH")
	}
	return nil
}

func (c *ShellCodexAdapter) Resume(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("branch has no head session id")
	}
	cmd := exec.Command(c.bin(), "resume", sessionID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *ShellCodexAdapter) Start() (string, error) {
	cmd := exec.Command(c.bin())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	// Session attachment is explicit in this version (use nodex attach-session).
	return "", nil
}

func (c *ShellCodexAdapter) Fork(sessionID string) (string, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("source branch has no head session id")
	}
	return "", fmt.Errorf("automatic fork not supported in v1; create branch metadata only")
}
