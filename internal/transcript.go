package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Transcript struct {
	SessionID string `json:"session_id"`
	Turns     []Turn `json:"turns"`
}

type Turn struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	Assistant string `json:"assistant"`
}

type CachedTranscript struct {
	Branch      string `json:"branch"`
	SessionID   string `json:"session_id"`
	GeneratedAt string `json:"generated_at"`
	Turns       []Turn `json:"turns"`
}

type TranscriptStore interface {
	Extract(sessionID string) (*Transcript, error)
}

type TranscriptBackend interface {
	Name() string
	Extract(sessionID string) (*Transcript, error)
}

type LocalTranscriptStore struct {
	Root string
}

func (l *LocalTranscriptStore) Name() string {
	return "local"
}

type CachedTranscriptStore struct {
	Root string
}

func (c *CachedTranscriptStore) Name() string {
	return "cache"
}

type MultiTranscriptStore struct {
	Backends []TranscriptBackend
}

func NewDefaultTranscriptStore(root string) TranscriptStore {
	return &MultiTranscriptStore{
		Backends: []TranscriptBackend{
			&LocalTranscriptStore{Root: root},
			&CachedTranscriptStore{Root: root},
		},
	}
}

func (m *MultiTranscriptStore) Extract(sessionID string) (*Transcript, error) {
	if len(m.Backends) == 0 {
		return nil, errors.New("no transcript backends configured")
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("source branch has no head session id")
	}
	failures := make([]string, 0, len(m.Backends))
	for _, backend := range m.Backends {
		tx, err := backend.Extract(sessionID)
		if err == nil {
			return tx, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", backend.Name(), err))
	}
	return nil, fmt.Errorf("transcript extraction failed for session %s; backends tried: %s", sessionID, strings.Join(failures, "; "))
}

func (l *LocalTranscriptStore) Extract(sessionID string) (*Transcript, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("source branch has no head session id")
	}
	path, err := l.findSessionFile(sessionID)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read transcript file: %w", err)
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, errors.New("invalid JSON in transcript file")
	}
	turns := normalizeTurns(doc)
	if len(turns) == 0 {
		return nil, errors.New("no user/assistant exchanges found")
	}
	return &Transcript{
		SessionID: sessionID,
		Turns:     turns,
	}, nil
}

func (l *LocalTranscriptStore) findSessionFile(sessionID string) (string, error) {
	roots := make([]string, 0, 8)
	if l.Root != "" {
		roots = append(roots,
			filepath.Join(l.Root, ".codex", "sessions"),
			filepath.Join(l.Root, ".codex"),
		)
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots,
			filepath.Join(home, ".codex", "sessions"),
			filepath.Join(home, ".config", "codex", "sessions"),
			filepath.Join(home, "AppData", "Roaming", "Codex", "sessions"),
			filepath.Join(home, "AppData", "Local", "Codex", "sessions"),
		)
	}
	if p := os.Getenv("CODEX_SESSIONS_DIR"); strings.TrimSpace(p) != "" {
		roots = append([]string{p}, roots...)
	}
	seen := map[string]struct{}{}
	for _, root := range roots {
		root = filepath.Clean(root)
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		if !dirExists(root) {
			continue
		}
		hit, err := findFileContaining(root, sessionID)
		if err == nil && hit != "" {
			return hit, nil
		}
	}
	return "", errors.New("session file not found in known local codex paths")
}

func (c *CachedTranscriptStore) Extract(sessionID string) (*Transcript, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("source branch has no head session id")
	}
	entries, err := os.ReadDir(cacheDir(c.Root))
	if err != nil {
		return nil, errors.New("cache directory unavailable")
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "transcript-") || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		var cached CachedTranscript
		path := filepath.Join(cacheDir(c.Root), e.Name())
		if err := readJSONFile(path, &cached); err != nil {
			continue
		}
		if cached.SessionID != sessionID {
			continue
		}
		if len(cached.Turns) == 0 {
			continue
		}
		return &Transcript{
			SessionID: sessionID,
			Turns:     cached.Turns,
		}, nil
	}
	return nil, errors.New("no matching transcript cache entry")
}

func findFileContaining(root, needle string) (string, error) {
	var hit string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".jsonl") {
			return nil
		}
		if strings.Contains(d.Name(), needle) {
			hit = path
			return errors.New("found")
		}
		// fallback: small content scan if session id is not in filename.
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if len(data) > 2*1024*1024 {
			return nil
		}
		if strings.Contains(string(data), needle) {
			hit = path
			return errors.New("found")
		}
		return nil
	})
	if err != nil && err.Error() == "found" {
		return hit, nil
	}
	if hit != "" {
		return hit, nil
	}
	return "", err
}

func normalizeTurns(doc any) []Turn {
	if turns := turnsFromExplicitPairs(doc); len(turns) > 0 {
		return assignTurnIDs(turns)
	}
	msgs := collectRoleMessages(doc)
	if len(msgs) == 0 {
		return nil
	}
	turns := make([]Turn, 0, len(msgs)/2+1)
	pendingUser := ""
	for _, m := range msgs {
		switch m.Role {
		case "user":
			if pendingUser != "" {
				turns = append(turns, Turn{User: pendingUser, Assistant: ""})
			}
			pendingUser = m.Text
		case "assistant":
			turns = append(turns, Turn{User: pendingUser, Assistant: m.Text})
			pendingUser = ""
		}
	}
	if pendingUser != "" {
		turns = append(turns, Turn{User: pendingUser, Assistant: ""})
	}
	filtered := make([]Turn, 0, len(turns))
	for _, t := range turns {
		if strings.TrimSpace(t.User) == "" && strings.TrimSpace(t.Assistant) == "" {
			continue
		}
		filtered = append(filtered, t)
	}
	return assignTurnIDs(filtered)
}

func turnsFromExplicitPairs(v any) []Turn {
	root, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := root["turns"]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]Turn, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		user := firstNonEmpty(
			stringField(m, "user"),
			stringField(m, "prompt"),
			stringField(m, "question"),
		)
		assistant := firstNonEmpty(
			stringField(m, "assistant"),
			stringField(m, "response"),
			stringField(m, "answer"),
		)
		if strings.TrimSpace(user) == "" && strings.TrimSpace(assistant) == "" {
			continue
		}
		out = append(out, Turn{User: user, Assistant: assistant})
	}
	return out
}

type roleMessage struct {
	Role string
	Text string
}

func collectRoleMessages(v any) []roleMessage {
	out := []roleMessage{}
	collectRoleMessagesRec(v, &out)
	return out
}

func collectRoleMessagesRec(v any, out *[]roleMessage) {
	switch x := v.(type) {
	case map[string]any:
		role := strings.ToLower(strings.TrimSpace(stringField(x, "role")))
		if role == "user" || role == "assistant" {
			text := extractMessageText(x)
			if strings.TrimSpace(text) != "" {
				*out = append(*out, roleMessage{Role: role, Text: text})
			}
		}
		for _, child := range x {
			collectRoleMessagesRec(child, out)
		}
	case []any:
		for _, child := range x {
			collectRoleMessagesRec(child, out)
		}
	}
}

func extractMessageText(m map[string]any) string {
	if s := stringField(m, "text"); s != "" {
		return s
	}
	if s := stringField(m, "message"); s != "" {
		return s
	}
	if raw, ok := m["content"]; ok {
		switch v := raw.(type) {
		case string:
			return v
		case []any:
			parts := make([]string, 0, len(v))
			for _, p := range v {
				switch pv := p.(type) {
				case string:
					if strings.TrimSpace(pv) != "" {
						parts = append(parts, pv)
					}
				case map[string]any:
					if t := firstNonEmpty(stringField(pv, "text"), stringField(pv, "content")); t != "" {
						parts = append(parts, t)
					}
				}
			}
			return strings.Join(parts, "\n")
		case map[string]any:
			return firstNonEmpty(stringField(v, "text"), stringField(v, "content"))
		}
	}
	return ""
}

func stringField(m map[string]any, key string) string {
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func assignTurnIDs(turns []Turn) []Turn {
	for i := range turns {
		turns[i].ID = fmt.Sprintf("T%d", i+1)
	}
	return turns
}

func SelectTurns(t *Transcript, last, from, to int) []Turn {
	if t == nil || len(t.Turns) == 0 {
		return nil
	}
	if last > 0 {
		if last >= len(t.Turns) {
			return append([]Turn(nil), t.Turns...)
		}
		return append([]Turn(nil), t.Turns[len(t.Turns)-last:]...)
	}
	if from > 0 && to >= from && to <= len(t.Turns) {
		return append([]Turn(nil), t.Turns[from-1:to]...)
	}
	if len(t.Turns) <= 3 {
		return append([]Turn(nil), t.Turns...)
	}
	return append([]Turn(nil), t.Turns[len(t.Turns)-3:]...)
}

func SaveTranscriptCache(root, branch string, t *Transcript) error {
	if t == nil {
		return nil
	}
	cache := CachedTranscript{
		Branch:      branch,
		SessionID:   t.SessionID,
		GeneratedAt: nowUTC(),
		Turns:       t.Turns,
	}
	path := filepath.Join(cacheDir(root), fmt.Sprintf("transcript-%s.json", branch))
	return writeJSONAtomic(path, cache)
}
