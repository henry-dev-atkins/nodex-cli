package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func readJSONFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty json file")
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("invalid json in %s: %w", path, err)
	}
	return nil
}

func writeJSONAtomic(path string, in any) error {
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o644)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parseMergeNumber(id string) int {
	if !strings.HasPrefix(id, "merge-") {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimPrefix(id, "merge-"))
	if err != nil {
		return -1
	}
	return n
}

var sessionIDValidRe = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

func ValidateSessionID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("session id is required")
	}
	if !sessionIDValidRe.MatchString(id) {
		return fmt.Errorf("invalid session id: %s", id)
	}
	return nil
}
