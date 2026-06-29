# Nodex v1 Quality Checklist

This checklist turns quality principles into concrete execution tasks.

Tags:
- `[CI]` automation and enforcement in pipelines
- `[Release]` packaging/versioning/distribution
- `[Docs]` user and contributor documentation

## 1) Architecture and Maintainability

Definition of done:
- Command handlers stay thin (parse input + call internal package + print result).
- Domain logic remains in `internal/*`.
- File formats stay stable and explicitly documented.

Tasks:
- [ ] `[CI]` Add `go vet ./...` and `go test ./...` gate on every PR.
- [ ] `[CI]` Add a lightweight check that no command package imports implementation-only helpers from other command files.
- [ ] `[Docs]` Add `ARCHITECTURE.md` describing package boundaries and data flow (`cmd -> internal -> storage/adapters`).
- [ ] `[Docs]` Document `.nodex` file schema and compatibility expectations in one place.

## 2) CLI UX Contract

Definition of done:
- Errors are actionable and blunt.
- Exit status is predictable (`0` success, non-zero failure).
- Help text and examples match real behavior.

Tasks:
- [ ] `[CI]` Add table-driven command tests for required usage and error paths (missing args, unknown branch, uninitialized workspace).
- [ ] `[CI]` Add smoke assertions for expected stdout fragments on core commands (`init`, `status`, `merge`, `apply`).
- [ ] `[Docs]` Add a "CLI contracts" section: stdout/stderr expectations and exit code policy.
- [ ] `[Docs]` Add an examples section covering both success paths and common failures.

## 3) Test Strategy (Confidence Without Bloat)

Definition of done:
- Pure logic is unit tested.
- Command behavior has minimal integration coverage.
- One reproducible temp-workspace smoke run exists.

Tasks:
- [ ] `[CI]` Keep unit tests for branch validation, merge rendering, turn selection, merge ID generation green.
- [ ] `[CI]` Add a scripted smoke test using a temp workspace and synthetic transcript fixture.
- [ ] `[CI]` Run tests on Windows and Linux (same Go version matrix).
- [ ] `[Docs]` Keep a short manual smoke checklist in README for local verification.

## 4) Compatibility and Versioning Discipline

Definition of done:
- Versioning and changelog process are explicit.
- Breaking CLI/storage changes are intentional and announced.
- Deprecated behavior has a clear migration window.

Tasks:
- [ ] `[Release]` Adopt SemVer and tag releases (`vX.Y.Z`).
- [ ] `[Release]` Add `CHANGELOG.md` with Keep a Changelog format.
- [ ] `[Release]` Define compatibility policy for `.nodex/workspace.json` version field.
- [ ] `[Docs]` Add deprecation policy (minimum warning period before removals).

## 5) Trust, Security, and Supply Chain

Definition of done:
- Users can verify what they run.
- Security reporting path is documented.
- Known limitations are visible, not hidden.

Tasks:
- [ ] `[Release]` Produce checksums (`SHA256SUMS`) for binaries.
- [ ] `[Release]` Sign release artifacts (for example with `cosign` or GPG).
- [ ] `[Docs]` Add `SECURITY.md` with disclosure and response expectations.
- [ ] `[Docs]` Add support matrix (OS + tested Go versions) and explicit non-goals.
- [ ] `[CI]` Add dependency/vulnerability scan step (`govulncheck`).

## 6) Contributor Velocity and Onboarding

Definition of done:
- New contributors can run checks quickly.
- Expectations for PRs/issues are explicit.
- Project decisions are easy to understand.

Tasks:
- [ ] `[Docs]` Add `CONTRIBUTING.md` (setup, test commands, PR expectations).
- [ ] `[Docs]` Add GitHub issue templates for bug reports and feature requests.
- [ ] `[Docs]` Add PR template requiring: behavior change summary, test evidence, docs impact.
- [ ] `[CI]` Add a single `verify` target/script used locally and in CI.
- [ ] `[Docs]` Add "invariants and tradeoffs" notes (what is intentionally hacky in v1 and why).

## 7) Minimum v1 Gate

Ship `v1.0.0` only after these are complete:
- [ ] `[CI]` PR gate runs `go test`, `go vet`, and `go build` on Windows + Linux.
- [ ] `[CI]` Temp-workspace smoke script passes.
- [ ] `[Release]` SemVer tag + changelog entry created.
- [ ] `[Release]` Checksums published with binaries.
- [ ] `[Docs]` `README`, `ARCHITECTURE.md`, `CONTRIBUTING.md`, and `SECURITY.md` present.
- [ ] `[Docs]` CLI usage/examples and known limitations are accurate for the release.
