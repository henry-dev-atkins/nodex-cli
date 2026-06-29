package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MergeArtifact struct {
	ID              string          `json:"id"`
	CreatedAt       string          `json:"created_at"`
	SourceBranch    string          `json:"source_branch"`
	DestBranch      string          `json:"dest_branch"`
	SourceSessionID string          `json:"source_session_id"`
	DestSessionID   string          `json:"dest_session_id"`
	Mode            string          `json:"mode"`
	Message         string          `json:"message"`
	Selection       Selection       `json:"selection"`
	SourceMaterial  SourceMaterial  `json:"source_material"`
	RedactionReport RedactionReport `json:"redaction_report"`
	PreviewText     string          `json:"preview_text"`
	Applied         bool            `json:"applied"`
	AppliedAt       string          `json:"applied_at"`
	ApplyError      string          `json:"apply_error"`
}

type Selection struct {
	Last     int    `json:"last"`
	FromTurn int    `json:"from_turn"`
	ToTurn   int    `json:"to_turn"`
	Contains string `json:"contains"`
}

type SourceMaterial struct {
	Turns []Turn `json:"turns"`
}

func mergePath(root, id string) string {
	return filepath.Join(mergesDir(root), id+".json")
}

func isValidMergeMode(mode string) bool {
	switch mode {
	case "summary", "decision", "verbose":
		return true
	default:
		return false
	}
}

func NextMergeID(root string) string {
	entries, err := os.ReadDir(mergesDir(root))
	if err != nil {
		return "merge-0001"
	}
	maxN := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		n := parseMergeNumber(id)
		if n > maxN {
			maxN = n
		}
	}
	return fmt.Sprintf("merge-%04d", maxN+1)
}

func SaveMerge(root string, m *MergeArtifact) error {
	if strings.TrimSpace(m.ID) == "" {
		m.ID = NextMergeID(root)
	}
	if m.CreatedAt == "" {
		m.CreatedAt = nowUTC()
	}
	return writeJSONAtomic(mergePath(root, m.ID), m)
}

func LoadMerge(root, id string) (*MergeArtifact, error) {
	var m MergeArtifact
	if err := readJSONFile(mergePath(root, id), &m); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("merge does not exist: %s", id)
		}
		return nil, err
	}
	return &m, nil
}

func ListMerges(root string) ([]*MergeArtifact, error) {
	entries, err := os.ReadDir(mergesDir(root))
	if err != nil {
		return nil, err
	}
	out := make([]*MergeArtifact, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		m, err := LoadMerge(root, id)
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return parseMergeNumber(out[i].ID) < parseMergeNumber(out[j].ID)
	})
	return out, nil
}

func PendingMergesCount(root string) (int, error) {
	items, err := ListMerges(root)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, m := range items {
		if !m.Applied {
			n++
		}
	}
	return n, nil
}

func PrepareMerge(root, srcName, destName, mode, message string, last, from, to int, contains string, tx TranscriptStore) (*MergeArtifact, error) {
	if !isValidMergeMode(mode) {
		return nil, fmt.Errorf("invalid mode: %s (expected summary|decision|verbose)", mode)
	}
	if last > 0 && (from > 0 || to > 0) {
		return nil, errors.New("use either --last or --from-turn/--to-turn, not both")
	}
	src, err := LoadBranch(root, srcName)
	if err != nil {
		return nil, err
	}
	dest, err := LoadBranch(root, destName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(src.HeadSessionID) == "" {
		return nil, fmt.Errorf("source branch has no head session id: %s", srcName)
	}
	if strings.TrimSpace(dest.HeadSessionID) == "" {
		return nil, fmt.Errorf("destination branch has no head session id: %s", destName)
	}
	transcript, err := tx.Extract(src.HeadSessionID)
	if err != nil {
		return nil, err
	}
	_ = SaveTranscriptCache(root, srcName, transcript)

	if from > 0 || to > 0 {
		if from <= 0 || to <= 0 {
			return nil, errors.New("both --from-turn and --to-turn must be set together")
		}
		if from > to {
			return nil, errors.New("--from-turn must be <= --to-turn")
		}
		if to > len(transcript.Turns) {
			return nil, fmt.Errorf("--to-turn %d exceeds transcript length %d", to, len(transcript.Turns))
		}
	}
	turns := SelectTurns(transcript, last, from, to)
	turns = filterTurnsByContains(turns, contains)
	if len(turns) == 0 {
		if strings.TrimSpace(contains) != "" {
			return nil, errors.New("no turns match --contains filter")
		}
		return nil, errors.New("no turns selected for merge")
	}

	preview := RenderMergePreview(mode, srcName, destName, message, turns)
	redaction := ScanText(root, preview)

	m := &MergeArtifact{
		ID:              NextMergeID(root),
		CreatedAt:       nowUTC(),
		SourceBranch:    srcName,
		DestBranch:      destName,
		SourceSessionID: src.HeadSessionID,
		DestSessionID:   dest.HeadSessionID,
		Mode:            mode,
		Message:         message,
		Selection:       Selection{Last: last, FromTurn: from, ToTurn: to, Contains: contains},
		SourceMaterial:  SourceMaterial{Turns: turns},
		RedactionReport: redaction,
		PreviewText:     preview,
		Applied:         false,
		AppliedAt:       "",
		ApplyError:      "",
	}
	if err := SaveMerge(root, m); err != nil {
		return nil, err
	}
	src.LastMergeID = m.ID
	_ = SaveBranch(root, src)
	return m, nil
}

func filterTurnsByContains(turns []Turn, contains string) []Turn {
	if strings.TrimSpace(contains) == "" {
		return turns
	}
	needle := strings.ToLower(strings.TrimSpace(contains))
	out := make([]Turn, 0, len(turns))
	for _, t := range turns {
		user := strings.ToLower(t.User)
		assistant := strings.ToLower(t.Assistant)
		if strings.Contains(user, needle) || strings.Contains(assistant, needle) {
			out = append(out, t)
		}
	}
	return out
}

func MarkMergeApplied(root, mergeID string) (*MergeArtifact, error) {
	m, err := LoadMerge(root, mergeID)
	if err != nil {
		return nil, err
	}
	m.Applied = true
	m.AppliedAt = nowUTC()
	m.ApplyError = ""
	if err := SaveMerge(root, m); err != nil {
		return nil, err
	}

	dest, err := LoadBranch(root, m.DestBranch)
	if err == nil {
		dest.LastMergeID = m.ID
		_ = SaveBranch(root, dest)
	}
	return m, nil
}

func MarkMergeApplyError(root, mergeID, msg string) error {
	m, err := LoadMerge(root, mergeID)
	if err != nil {
		return err
	}
	m.ApplyError = msg
	return SaveMerge(root, m)
}
