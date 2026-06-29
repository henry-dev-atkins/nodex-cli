# Security Policy

## Reporting a Vulnerability

Please report suspected vulnerabilities privately first.

Include:
1. A clear description of the issue
2. Reproduction steps
3. Impact assessment
4. Any suggested mitigations

Until a dedicated security contact is configured, open a private channel with the maintainer and do not publish exploit details publicly.

## Response Expectations

Target response timeline:
1. Initial acknowledgement: within 7 days
2. Triage update: within 14 days
3. Fix/release timing: depends on severity and complexity

## Security Scope Notes

Nodex is a local CLI workflow tool. Important characteristics:
1. It writes local JSON files under `.nodex/`.
2. It shells out to local `codex` binary.
3. Redaction checks are heuristic warnings, not guaranteed secret removal.
4. Merge application can involve manual copy/paste handoff.

## Supported Versions

Security fixes are currently targeted to the latest commit on the main development line.
