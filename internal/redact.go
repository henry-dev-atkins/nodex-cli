package internal

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Finding struct {
	Type         string `json:"type"`
	ValuePreview string `json:"value_preview"`
	Reason       string `json:"reason"`
}

type RedactionReport struct {
	Findings []Finding `json:"findings"`
}

var secretRegexes = []struct {
	re     *regexp.Regexp
	reason string
}{
	{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), "aws access key-like token"},
	{regexp.MustCompile(`\bsk-[A-Za-z0-9]{20,}\b`), "openai-style key-like token"},
	{regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`), "github token-like string"},
	{regexp.MustCompile(`\b(?:xoxb|xoxp)-[A-Za-z0-9-]{20,}\b`), "slack token-like string"},
}

var tokenLikeRe = regexp.MustCompile(`[A-Za-z0-9_\-]{20,}`)

func ScanText(root, text string) RedactionReport {
	report := RedactionReport{Findings: []Finding{}}
	seen := map[string]struct{}{}
	add := func(t, value, reason string) {
		key := t + "|" + value
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		report.Findings = append(report.Findings, Finding{
			Type:         t,
			ValuePreview: previewSecret(value),
			Reason:       reason,
		})
	}

	for _, rx := range secretRegexes {
		matches := rx.re.FindAllString(text, -1)
		for _, m := range matches {
			add("regex", m, rx.reason)
		}
	}

	for _, tok := range tokenLikeRe.FindAllString(text, -1) {
		if shannon(tok) >= 3.7 {
			add("entropy", tok, "high entropy token-like string")
		}
	}

	for _, deny := range loadDenylist(root) {
		if strings.Contains(text, deny) {
			add("denylist", deny, "matches .nodex/redact.txt")
		}
	}

	return report
}

func loadDenylist(root string) []string {
	path := filepath.Join(NodexDir(root), "redact.txt")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	out := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func previewSecret(s string) string {
	if utf8.RuneCountInString(s) <= 8 {
		return s
	}
	runes := []rune(s)
	return string(runes[:3]) + "..." + string(runes[len(runes)-2:])
}

func shannon(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	counts := map[rune]float64{}
	total := 0.0
	for _, r := range s {
		counts[r]++
		total++
	}
	ent := 0.0
	for _, c := range counts {
		p := c / total
		ent -= p * math.Log2(p)
	}
	return ent
}
