# Nodex

Nodex is a tiny local CLI for branching and merging Codex conversations.

It does not rewrite chat history.  
It creates named branches that point to Codex sessions, then prepares explicit context blocks you can merge back into another branch.

## What It Does

1. Create named branches from the current branch/session metadata.
2. Manage branch lifecycle (`create`, `rename`, `delete`) with local safety guards.
3. Resume Codex on a branch session.
4. Attach explicit session IDs to branches (`attach-session`).
5. Prepare deterministic merge artifacts from source transcript turns.
6. Apply merges as explicit handoff blocks (`--dry-run` supported).
7. Expose the same core operations over MCP (`stdio` JSON-RPC).

## What It Does Not Do

1. No GUI.
2. No DAG/history rewrite.
3. No database or daemon.
4. No fully automatic merge injection workflow in v1.
5. No hosted/remote service.

## Install

```bash
go build -o nodex .
```

## Command Summary

```bash
nodex init
nodex status
nodex branches
nodex branch <name> [--from <branch>]
nodex branch rename <old> <new>
nodex branch delete <name> [--yes] [--force]
nodex switch <name>
nodex chat [branch]
nodex attach-session <branch> <session-id>
nodex merge <src> --into <dest> --mode <summary|decision|verbose> [--message "..."] [--last N] [--from-turn N --to-turn N] [--contains "..."] [--prepare]
nodex apply <merge-id> [--echo-only] [--dry-run] [--mark-applied] [--yes]
nodex mark-applied <merge-id> [--yes]
nodex show <branch>
nodex mcp serve
```

## Quick Start

```bash
nodex init
nodex chat
nodex attach-session main <session-id>
nodex branch go-spike
nodex switch go-spike
nodex chat
nodex attach-session go-spike <session-id>
nodex merge go-spike --into main --mode decision --message "Decide whether to switch from Python to Go." --contains "risk"
nodex apply merge-0001 --dry-run
nodex mark-applied merge-0001 --yes
```

## Branch Lifecycle

```bash
# create from current branch
nodex branch go-spike

# rename branch
nodex branch rename go-spike go-spike-2

# delete branch (prompts by default)
nodex branch delete go-spike-2

# force delete branch and detach its children
nodex branch delete go-spike-2 --force --yes
```

## Session Attachment Workflow

Nodex uses explicit session assignment.

1. Start/resume chat with `nodex chat [branch]`.
2. Identify the Codex session id.
3. Attach it to the branch:

```bash
nodex attach-session <branch> <session-id>
```

Notes:
1. `head_session_id` is always updated.
2. `root_session_id` is set only if empty.

## Merge Selector Examples

```bash
# last N turns
nodex merge spike --into main --mode summary --last 5

# explicit turn range
nodex merge spike --into main --mode verbose --from-turn 3 --to-turn 7

# filter selected turns by substring (case-insensitive)
nodex merge spike --into main --mode decision --contains risk --message "What are the migration risks?"
```

## Apply Safety

```bash
# no side effects: print only + planned actions
nodex apply merge-0004 --dry-run

# explicit apply mark step
nodex mark-applied merge-0004 --yes
```

## MCP Setup (Codex)

```bash
codex mcp add nodex -- nodex mcp serve
```

### MCP Tools

1. `nodex_status`
2. `nodex_branches`
3. `nodex_branch_create`
4. `nodex_branch_switch`
5. `nodex_branch_rename`
6. `nodex_branch_delete`
7. `nodex_branch_attach_session`
8. `nodex_merge_prepare`
9. `nodex_merge_apply`
10. `nodex_mark_applied`

`tools/call` responses include structured JSON content (`type: json`) and a text fallback.

## Data Storage

Nodex stores workspace data in `.nodex/`:

1. `workspace.json`
2. `current_branch`
3. `branches/*.json`
4. `merges/*.json`
5. `cache/transcript-<branch>.json`

## Manual Smoke Test

1. `nodex init`
2. `nodex attach-session main <session-id>`
3. `nodex branch go-spike`
4. `nodex attach-session go-spike <session-id>`
5. `nodex merge go-spike --into main --mode summary --last 3`
6. `nodex apply merge-0001 --dry-run`
7. `nodex mark-applied merge-0001 --yes`
8. `nodex mcp serve` and verify tools are visible to MCP client

## Honest v1 Compromises

1. Automatic session forking is not automated; branch metadata can exist without a session id.
2. Session attachment is explicit: use `nodex attach-session <branch> <session-id>` after starting or resuming chats.
3. `nodex apply` is manual-first: it prints merge text and may open a resumed chat, but paste handoff is still expected.
4. Transcript extraction is best effort across local and cache backends; it can still fail if session data is unavailable.
5. Secret redaction is warning-only and heuristic.

## Quality Checklist

`v1` quality gates and mapped CI/release/docs tasks are tracked in [QUALITY_CHECKLIST_v1.md](plans/QUALITY_CHECKLIST_v1.md).

## Additional Docs

1. [ARCHITECTURE.md](ARCHITECTURE.md)
2. [CONTRIBUTING.md](CONTRIBUTING.md)
3. [SECURITY.md](SECURITY.md)

## License

MIT. See [LICENSE](LICENSE).
