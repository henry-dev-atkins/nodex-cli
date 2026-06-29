package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type RPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func ServeMCP(root string) error {
	dec := json.NewDecoder(bufio.NewReader(os.Stdin))
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	for {
		var msg RPCMessage
		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch msg.Method {
		case "initialize":
			_ = enc.Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      msg.ID,
				"result": map[string]any{
					"protocolVersion": "2024-11-05",
					"serverInfo": map[string]any{
						"name":    "nodex",
						"version": "0.1.0",
					},
					"capabilities": map[string]any{
						"tools": map[string]any{},
					},
				},
			})
		case "tools/list":
			_ = enc.Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      msg.ID,
				"result": map[string]any{
					"tools": listTools(),
				},
			})
		case "tools/call":
			out, err := handleToolCall(root, msg.Params)
			if err != nil {
				_ = enc.Encode(rpcError(msg.ID, -32000, err.Error()))
			} else {
				_ = enc.Encode(map[string]any{
					"jsonrpc": "2.0",
					"id":      msg.ID,
					"result": map[string]any{
						"content": []map[string]any{
							{
								"type": "json",
								"json": out,
							},
							{
								"type": "text",
								"text": mustJSON(out),
							},
						},
					},
				})
			}
		default:
			_ = enc.Encode(rpcError(msg.ID, -32601, "method not found"))
		}
	}
}

func listTools() []map[string]any {
	return []map[string]any{
		{
			"name":        "nodex_status",
			"description": "Get current branch, branch list, and pending merges.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "nodex_branches",
			"description": "List branches and their head session ids.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "nodex_branch_create",
			"description": "Create a branch from another branch.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"from": map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		{
			"name":        "nodex_branch_switch",
			"description": "Switch current branch.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		{
			"name":        "nodex_branch_rename",
			"description": "Rename a branch.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"old_name": map[string]any{"type": "string"},
					"new_name": map[string]any{"type": "string"},
				},
				"required": []string{"old_name", "new_name"},
			},
		},
		{
			"name":        "nodex_branch_delete",
			"description": "Delete a branch.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":  map[string]any{"type": "string"},
					"force": map[string]any{"type": "boolean"},
				},
				"required": []string{"name"},
			},
		},
		{
			"name":        "nodex_branch_attach_session",
			"description": "Attach a session id to a branch head.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":       map[string]any{"type": "string"},
					"session_id": map[string]any{"type": "string"},
				},
				"required": []string{"name", "session_id"},
			},
		},
		{
			"name":        "nodex_merge_prepare",
			"description": "Prepare a merge artifact and preview text.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source_branch": map[string]any{"type": "string"},
					"dest_branch":   map[string]any{"type": "string"},
					"mode":          map[string]any{"type": "string"},
					"message":       map[string]any{"type": "string"},
					"last":          map[string]any{"type": "integer"},
					"from_turn":     map[string]any{"type": "integer"},
					"to_turn":       map[string]any{"type": "integer"},
					"contains":      map[string]any{"type": "string"},
				},
				"required": []string{"source_branch", "dest_branch", "mode"},
			},
		},
		{
			"name":        "nodex_merge_apply",
			"description": "Load merge artifact and return handoff preview.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"merge_id": map[string]any{"type": "string"},
				},
				"required": []string{"merge_id"},
			},
		},
		{
			"name":        "nodex_mark_applied",
			"description": "Mark a merge artifact as applied.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"merge_id": map[string]any{"type": "string"},
				},
				"required": []string{"merge_id"},
			},
		},
	}
}

func handleToolCall(root string, raw json.RawMessage) (any, error) {
	var p toolCallParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, errors.New("invalid tools/call params")
	}
	switch p.Name {
	case "nodex_status":
		return MCPStatus(root)
	case "nodex_branches":
		return MCPBranches(root)
	case "nodex_branch_create":
		name := toString(p.Arguments["name"])
		from := toString(p.Arguments["from"])
		return MCPBranchCreate(root, name, from)
	case "nodex_branch_switch":
		name := toString(p.Arguments["name"])
		return MCPBranchSwitch(root, name)
	case "nodex_branch_rename":
		oldName := toString(p.Arguments["old_name"])
		newName := toString(p.Arguments["new_name"])
		return MCPBranchRename(root, oldName, newName)
	case "nodex_branch_delete":
		name := toString(p.Arguments["name"])
		force := toBool(p.Arguments["force"])
		return MCPBranchDelete(root, name, force)
	case "nodex_branch_attach_session":
		name := toString(p.Arguments["name"])
		sessionID := toString(p.Arguments["session_id"])
		return MCPBranchAttachSession(root, name, sessionID)
	case "nodex_merge_prepare":
		src := toString(p.Arguments["source_branch"])
		dest := toString(p.Arguments["dest_branch"])
		mode := toString(p.Arguments["mode"])
		message := toString(p.Arguments["message"])
		last := toInt(p.Arguments["last"])
		fromTurn := toInt(p.Arguments["from_turn"])
		toTurn := toInt(p.Arguments["to_turn"])
		contains := toString(p.Arguments["contains"])
		return MCPMergePrepare(root, src, dest, mode, message, last, fromTurn, toTurn, contains)
	case "nodex_merge_apply":
		mergeID := toString(p.Arguments["merge_id"])
		return MCPMergeApply(root, mergeID)
	case "nodex_mark_applied":
		mergeID := toString(p.Arguments["merge_id"])
		return MCPMarkApplied(root, mergeID)
	default:
		return nil, fmt.Errorf("unknown tool: %s", p.Name)
	}
}

func MCPStatus(root string) (map[string]any, error) {
	ws, err := LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	branches, err := ListBranches(root)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(branches))
	for _, b := range branches {
		names = append(names, b.Name)
	}
	pending, err := PendingMergesCount(root)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"current_branch": ws.CurrentBranch,
		"branches":       names,
		"pending_merges": pending,
	}, nil
}

func MCPBranches(root string) (map[string]any, error) {
	ws, err := LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	branches, err := ListBranches(root)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(branches))
	for _, b := range branches {
		items = append(items, map[string]any{
			"name":            b.Name,
			"parent":          b.Parent,
			"head_session_id": b.HeadSessionID,
		})
	}
	return map[string]any{
		"current_branch": ws.CurrentBranch,
		"items":          items,
	}, nil
}

func MCPBranchCreate(root, name, from string) (map[string]any, error) {
	ws, err := LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	codex := &ShellCodexAdapter{Bin: ws.CodexBin}
	b, err := CreateBranch(root, name, from, codex)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"name":            b.Name,
		"parent":          b.Parent,
		"head_session_id": b.HeadSessionID,
	}, nil
}

func MCPBranchSwitch(root, name string) (map[string]any, error) {
	if err := SwitchBranch(root, name); err != nil {
		return nil, err
	}
	return map[string]any{
		"current_branch": name,
	}, nil
}

func MCPBranchRename(root, oldName, newName string) (map[string]any, error) {
	if err := RenameBranch(root, oldName, newName); err != nil {
		return nil, err
	}
	return map[string]any{
		"old_name": oldName,
		"new_name": newName,
	}, nil
}

func MCPBranchDelete(root, name string, force bool) (map[string]any, error) {
	if err := DeleteBranch(root, name, force); err != nil {
		return nil, err
	}
	return map[string]any{
		"name":    name,
		"deleted": true,
	}, nil
}

func MCPBranchAttachSession(root, name, sessionID string) (map[string]any, error) {
	b, err := AttachSession(root, name, sessionID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"name":            b.Name,
		"head_session_id": b.HeadSessionID,
		"root_session_id": b.RootSessionID,
	}, nil
}

func MCPMergePrepare(root, src, dest, mode, message string, last, fromTurn, toTurn int, contains string) (map[string]any, error) {
	tx := NewDefaultTranscriptStore(root)
	m, err := PrepareMerge(root, src, dest, mode, message, last, fromTurn, toTurn, contains, tx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"merge_id":           m.ID,
		"preview_text":       m.PreviewText,
		"redaction_findings": m.RedactionReport.Findings,
	}, nil
}

func MCPMergeApply(root, mergeID string) (map[string]any, error) {
	m, err := LoadMerge(root, mergeID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"merge_id":     m.ID,
		"applied":      m.Applied,
		"handoff":      "paste_required",
		"preview_text": m.PreviewText,
	}, nil
}

func MCPMarkApplied(root, mergeID string) (map[string]any, error) {
	m, err := MarkMergeApplied(root, mergeID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"merge_id":   m.ID,
		"applied":    m.Applied,
		"applied_at": m.AppliedAt,
	}, nil
}

func rpcError(id any, code int, msg string) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": msg,
		},
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	default:
		return 0
	}
}

func toBool(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || strings.TrimSpace(b) == "1"
	default:
		return false
	}
}
