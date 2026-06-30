# iqtoolkit-analyzer

PostgreSQL health checking and performance tuning recommendations CLI (Go).

## Build & test

- `make build` / `make test` / `make lint` / `make cover`
- Integration tests for `internal/dbconn` require `TEST_DATABASE_URL` (CI runs them against postgres:17)

## Hard rules

- **Autovacuum must NEVER be disabled.** Recommendation text for `autovacuum = off` is unconditionally critical — no caveats or escape hatches ("unless you run manual vacuums" etc.).
- Recommendation messages should be firm and opinionated when DBA consensus is clear, and always include the actual numbers so users can judge urgency.

## Changelog

- **Router substring bypass fix** — Fixed security vulnerability in `getTargetFromProxyTable()` in `src/router.ts`. Replaced unanchored `indexOf` check with a three-branch key classifier (path-only, host+path, host-only) that enforces exact host matching. Prevents crafted Host headers from bypassing routing boundaries. Spec: `.kiro/specs/router-substring-bypass-fix/`

## Roadmap (future releases)

- **Configurable rule thresholds** — currently hardcoded in `internal/recommendations/recommendations.go`:
  - dead-tuple ratio (20%, 10k floor), seq_scan (>1000), max_connections (>500), avg duration (>500ms), slow query count (>10), error count (>100)
- **Customizable recommendations** — let users disable rules, adjust severities, and define custom rules to empower them. Likely a `rules.yaml` in `~/.config/iqtoolkit-analyzer/` plus flag overrides.
- Exception: autovacuum-off severity is not user-configurable (see hard rules).
