# Contributing

Thanks for contributing to Nodex.

## Development Setup

Requirements:
1. Go installed and available on PATH
2. Codex CLI available as `codex` (for runtime workflows)

Build:

```bash
go build ./...
```

Run tests:

```bash
go test ./...
```

Recommended verify pass:

```bash
go test ./...
go vet ./...
go build ./...
```

## Code Style

1. Keep command handlers in `cmd/` thin and orchestration-only.
2. Put behavior and persistence logic in `internal/`.
3. Prefer explicit, blunt error messages.
4. Keep dependencies minimal (stdlib-first).

## Pull Request Expectations

Include:
1. What changed and why
2. Behavior impact for users (commands, flags, storage schema)
3. Test evidence (`go test`, smoke steps, etc.)
4. Documentation updates if behavior changed

## Manual Smoke Checklist

1. `nodex init`
2. `nodex attach-session main <session-id>`
3. `nodex branch go-spike`
4. `nodex attach-session go-spike <session-id>`
5. `nodex merge go-spike --into main --mode summary --last 3`
6. `nodex apply merge-0001 --dry-run`
7. `nodex mark-applied merge-0001 --yes`

## Scope Discipline

Nodex intentionally prefers small, explicit behavior over hidden automation.
When proposing changes:
1. Explain tradeoffs
2. Prefer deterministic workflows
3. Avoid introducing complexity without clear user benefit
