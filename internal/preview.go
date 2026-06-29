package internal

import (
	"strings"
)

func RenderMergePreview(mode, src, dest, message string, turns []Turn) string {
	switch mode {
	case "decision":
		return renderDecision(src, dest, message, turns)
	case "verbose":
		return renderVerbose(src, dest, message, turns)
	default:
		return renderSummary(src, dest, turns)
	}
}

func renderSummary(src, dest string, turns []Turn) string {
	var b strings.Builder
	b.WriteString("Imported branch: " + src + "\n")
	b.WriteString("Mode: summary\n\n")
	b.WriteString("Included material:\n")
	for _, t := range turns {
		if strings.TrimSpace(t.User) != "" {
			b.WriteString("- " + oneLine(t.User) + "\n")
		}
		if strings.TrimSpace(t.Assistant) != "" {
			b.WriteString("- " + oneLine(t.Assistant) + "\n")
		}
	}
	b.WriteString("\nUse this imported context when answering the next step in " + dest + ".\n")
	return b.String()
}

func renderDecision(src, dest, message string, turns []Turn) string {
	var b strings.Builder
	b.WriteString("Imported branch: " + src + "\n")
	b.WriteString("Mode: decision\n")
	if strings.TrimSpace(message) != "" {
		b.WriteString("Question: " + message + "\n")
	}
	b.WriteString("\nBranch findings:\n")
	for _, t := range turns {
		if strings.TrimSpace(t.User) != "" {
			b.WriteString("- " + oneLine(t.User) + "\n")
		}
		if strings.TrimSpace(t.Assistant) != "" {
			b.WriteString("- " + oneLine(t.Assistant) + "\n")
		}
	}
	b.WriteString("\nRecommendation:\n")
	if len(turns) > 0 && strings.TrimSpace(turns[len(turns)-1].Assistant) != "" {
		b.WriteString(oneLine(turns[len(turns)-1].Assistant) + "\n")
	} else {
		b.WriteString("Review the imported branch findings and make a decision.\n")
	}
	b.WriteString("\nUse this imported decision context when answering the next step in " + dest + ".\n")
	return b.String()
}

func renderVerbose(src, _, message string, turns []Turn) string {
	var b strings.Builder
	b.WriteString("Imported branch: " + src + "\n")
	b.WriteString("Mode: verbose\n")
	if strings.TrimSpace(message) != "" {
		b.WriteString("Request: " + message + "\n")
	}
	b.WriteString("\nKey prompts and answers:\n")
	for _, t := range turns {
		b.WriteString("\n[" + t.ID + "] User: " + t.User + "\n")
		b.WriteString("[" + t.ID + "] Assistant: " + t.Assistant + "\n")
	}
	b.WriteString("\nImportant details:\n")
	for _, t := range turns {
		if strings.TrimSpace(t.Assistant) != "" {
			b.WriteString("- " + oneLine(t.Assistant) + "\n")
		}
	}
	b.WriteString("\nUse this imported context as reference, not as hidden history.\n")
	return b.String()
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if len(s) > 240 {
		return s[:237] + "..."
	}
	return s
}
