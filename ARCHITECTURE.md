# Architecture

This document describes Nodex internals at a practical level for maintainers.

## High-Level Shape

Nodex is a single Go binary with:
1. Thin command entrypoints in `cmd/`
2. Core logic in `internal/`
3. File-based state in `.nodex/`
4. Optional MCP server mode over stdio JSON-RPC

Flow:
1. `main.go` -> `cmd.Execute(...)`
2. command handler parses flags
3. handler calls internal functions
4. internal layer reads/writes `.nodex/*` JSON files

## Package Boundaries

`cmd/`
1. CLI argument parsing and usage text
2. User-facing output formatting
3. No storage schema logic

`internal/workspace.go`
1. Workspace init/load/save
2. Root discovery (`FindRepoRoot`)
3. Current branch pointer sync

`internal/branch.go`
1. Branch CRUD and validation
2. Branch lifecycle operations (`rename`, `delete`, `attach-session`)
3. Branch reference updates

`internal/merge.go`
1. Merge artifact schema and persistence
2. Merge selection pipeline (`last` / range / contains filter)
3. Apply state marking

`internal/transcript.go`
1. Transcript backend abstraction
2. Default multi-backend orchestration
3. Local extraction + cache fallback

`internal/codex.go`
1. Shell adapter for Codex process interaction
2. Detect, start, resume, fork (best-effort/unsupported path)

`internal/mcp.go`
1. JSON-RPC loop for `initialize`, `tools/list`, `tools/call`
2. Tool dispatch to same internal operations used by CLI

## Storage Model

All state is local under `.nodex/`:
1. `workspace.json`
2. `current_branch`
3. `branches/<name>.json`
4. `merges/merge-####.json`
5. `cache/transcript-<branch>.json`

Core rule:
1. A branch points to one `head_session_id`.
2. Merges are explicit artifacts and are never auto-deleted.

## Session Model

Nodex uses explicit session assignment:
1. Start/resume chat with `nodex chat`
2. Attach session with `nodex attach-session <branch> <session-id>`

This avoids hidden state changes and keeps branch/session mapping auditable.

## Merge Selection Pipeline

Order:
1. Base selection (`--last`, or `--from-turn`/`--to-turn`, or default tail)
2. Optional text filter (`--contains`, case-insensitive)
3. Deterministic template rendering (`summary`, `decision`, `verbose`)
4. Redaction scan (warning only in current version)

## MCP Response Shape

`tools/call` success returns:
1. `content` item with `type: "json"` and structured output object
2. `content` text fallback with JSON string for compatibility

## Invariants

1. Branch names must match `[A-Za-z0-9._-]+`.
2. `workspace.current_branch` and `.nodex/current_branch` stay synchronized.
3. Merge IDs are monotonic (`merge-0001`, `merge-0002`, ...).
4. `applied=true` is only set by explicit mark flow.
