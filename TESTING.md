# Testing Guide — iqtoolkit-analyzer

Step-by-step instructions for testing every part of the tool, from unit tests to full end-to-end runs against a live PostgreSQL instance.

## Prerequisites

- Go 1.26+ (`go version`)
- Docker (for the local PostgreSQL used in integration and E2E tests)
- `make` (optional but convenient)
- `golangci-lint` (optional, for lint checks): `brew install golangci-lint`

---

## 1. Build

```bash
make build
# or: go build -o iqtoolkit-analyzer .

./iqtoolkit-analyzer --version
# expect: a version string like v1.0.0-5-gabc1234 (abc1234, 2026-06-10) or "dev"

./iqtoolkit-analyzer --help
./iqtoolkit-analyzer analyze --help
./iqtoolkit-analyzer report --help
```

## 2. Unit tests

```bash
make test
# or: go test -race ./...
```

Expected: all packages pass. `internal/dbconn` integration tests are **skipped** unless `TEST_DATABASE_URL` is set (see §3).

Run a single package or test:

```bash
go test -race ./internal/recommendations/
go test -race -run TestIsRetryable ./internal/ai/
```

Coverage report:

```bash
make cover
# or:
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # opens browser
```

## 3. Integration tests (live PostgreSQL)

Start a disposable PostgreSQL:

```bash
docker run -d --name iqa-pg \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:17
```

Wait for it to be ready:

```bash
docker exec iqa-pg pg_isready -U postgres
# expect: accepting connections
```

Run the integration tests:

```bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres \
  go test -race -v ./internal/dbconn/
```

Expected: 8 tests run (Version, Settings, Extensions, CheckExtension, StatUserTables, StatUserIndexes, QueryTimeout) — none skipped.

## 4. End-to-end: log-only mode (no database)

Create a sample log in **stderr** format:

```bash
cat > /tmp/sample-stderr.log <<'EOF'
2026-06-10 10:00:01 UTC [101] LOG:  database system is ready to accept connections
2026-06-10 10:00:05 UTC [102] ERROR:  relation "users" does not exist
2026-06-10 10:00:06 UTC [102] ERROR:  relation "users" does not exist
2026-06-10 10:00:10 UTC [103] LOG:  duration: 2500.123 ms  statement: SELECT * FROM orders
2026-06-10 10:00:15 UTC [104] LOG:  duration: 4100.500 ms  statement: SELECT * FROM order_items oi JOIN orders o ON o.id = oi.order_id
2026-06-10 10:00:20 UTC [105] LOG:  duration: 150.000 ms  statement: SELECT 1
EOF
```

Run in log-only mode (no `--dsn`):

```bash
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log
```

**Verify:**
- stderr shows: `No --dsn provided: running in log-only mode...`
- Summary: 6 total entries, 2 errors, 2 slow queries (≥1000ms default threshold)
- "Slowest Queries" section lists the 4100ms query **first** (sorted by duration desc)
- A performance recommendation appears (avg duration > 500ms)

Test the `--quiet` flag:

```bash
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log -q 2>/dev/null
# the log-only notice must NOT appear; report still prints to stdout
```

Test the slow threshold:

```bash
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --slow-threshold 100
# now 3 slow queries (150ms query included)
```

## 5. Output formats

```bash
# JSON — validate with jq
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --format json -q | jq .
# verify keys: .summary, .slow_queries (array, sorted), .recommendations

# Markdown
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --format markdown -q

# Write to file
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --format json --output /tmp/report.json
cat /tmp/report.json | jq .summary
```

## 6. Log format auto-detection (csvlog and jsonlog)

**csvlog** sample:

```bash
cat > /tmp/sample.csv <<'EOF'
2026-06-10 10:00:05.000 UTC,postgres,mydb,102,"127.0.0.1:50000",665f0001.66,1,SELECT,2026-06-10 09:59:00 UTC,4/22,0,ERROR,42P01,"relation ""users"" does not exist",,,,,,,,psql,client backend,,0
2026-06-10 10:00:10.000 UTC,postgres,mydb,103,"127.0.0.1:50001",665f0002.67,1,SELECT,2026-06-10 09:59:00 UTC,5/23,0,LOG,00000,"duration: 2500.123 ms  statement: SELECT * FROM orders",,,,,,,,psql,client backend,,0
EOF

./iqtoolkit-analyzer analyze --log-file /tmp/sample.csv -q
# expect: 2 entries, 1 error, 1 slow query (auto-detected as csvlog)
```

**jsonlog** sample:

```bash
cat > /tmp/sample.json <<'EOF'
{"timestamp":"2026-06-10 10:00:05.000 UTC","error_severity":"ERROR","message":"relation \"users\" does not exist"}
{"timestamp":"2026-06-10 10:00:10.000 UTC","error_severity":"LOG","message":"duration: 2500.123 ms  statement: SELECT * FROM orders"}
EOF

./iqtoolkit-analyzer analyze --log-file /tmp/sample.json -q
# expect: 2 entries, 1 error, 1 slow query (auto-detected as jsonlog)
```

Force a format explicitly:

```bash
./iqtoolkit-analyzer analyze --log-file /tmp/sample.csv --log-format csvlog -q
```

## 7. End-to-end: full analysis with database

Using the Docker PostgreSQL from §3, create some activity first:

```bash
docker exec -i iqa-pg psql -U postgres <<'EOF'
CREATE TABLE big (id int, val text);
INSERT INTO big SELECT g, md5(g::text) FROM generate_series(1, 100000) g;
UPDATE big SET val = 'x' WHERE id % 3 = 0;          -- create dead tuples
SELECT count(*) FROM big WHERE val = 'nope';        -- force seq scans
SELECT count(*) FROM big WHERE val = 'nope';
EOF
```

Run the full analysis:

```bash
./iqtoolkit-analyzer analyze \
  --dsn "postgres://postgres:postgres@localhost:5432/postgres" \
  --log-file /tmp/sample-stderr.log
```

**Verify:**
- stderr suggests installing `pg_stat_statements` and `pg_buffercache` (not installed by default)
- Configuration recommendations appear (shared_buffers at default, etc.)
- No autovacuum warning (it's on by default)

Test the autovacuum critical path (then revert!):

```bash
docker exec -i iqa-pg psql -U postgres -c "ALTER SYSTEM SET autovacuum = off; SELECT pg_reload_conf();"
./iqtoolkit-analyzer analyze --dsn "postgres://postgres:postgres@localhost:5432/postgres" --log-file /tmp/sample-stderr.log -q | grep -i autovacuum
# expect: [critical][maintenance] autovacuum is DISABLED...
docker exec -i iqa-pg psql -U postgres -c "ALTER SYSTEM RESET autovacuum; SELECT pg_reload_conf();"
```

## 8. HTML report command

```bash
./iqtoolkit-analyzer report \
  --dsn "postgres://postgres:postgres@localhost:5432/postgres" \
  --output /tmp/report.html

open /tmp/report.html   # macOS
```

**Verify in the browser:**
- PostgreSQL version box
- Recommendations section with color-coded severity badges (if any fire)
- Settings table (full pg_settings)
- Extensions table — installed ones in green, others with a dash

## 9. AI-enhanced analysis (optional, requires API key)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
./iqtoolkit-analyzer analyze \
  --log-file /tmp/sample-stderr.log \
  --ai-provider anthropic
# expect: an "AI-Enhanced Recommendations" section at the end
```

Other providers: `--ai-provider openai` (`OPENAI_API_KEY`), `gemini` (`GEMINI_API_KEY`), `kiro` (AWS credentials + `AWS_REGION`).

Test the model override:

```bash
./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log \
  --ai-provider anthropic --ai-model claude-haiku-4-5
# or via env:
IQTOOLKIT_AI_MODEL=claude-haiku-4-5 ./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --ai-provider anthropic
```

Negative test (bad key — must fail fast, no retry storm):

```bash
ANTHROPIC_API_KEY=bad ./iqtoolkit-analyzer analyze --log-file /tmp/sample-stderr.log --ai-provider anthropic
# expect: "AI analysis failed: ai: anthropic returned 401..." within seconds
```

## 10. Error handling & edge cases

```bash
# Missing required flag
./iqtoolkit-analyzer analyze
# expect: required flag(s) "log-file" not set

# Nonexistent log file
./iqtoolkit-analyzer analyze --log-file /nope.log
# expect: opening log file: ... no such file or directory

# Empty log file
touch /tmp/empty.log && ./iqtoolkit-analyzer analyze --log-file /tmp/empty.log -q
# expect: clean run, 0 entries, "No recommendations"

# Bad DSN — must fail within ~10s (connect timeout)
time ./iqtoolkit-analyzer analyze --dsn "postgres://nope:nope@10.255.255.1:5432/x" --log-file /tmp/sample-stderr.log

# Ctrl-C during a run must exit cleanly (signal handling)
```

## 11. Lint & vulnerability scan

```bash
make lint          # golangci-lint run
go vet ./...
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 12. CI verification

After pushing, check GitHub Actions:

- **test** job — unit + integration tests against the postgres:17 service container, coverage artifact uploaded
- **lint** job — golangci-lint
- **vulncheck** job — govulncheck

```bash
gh run list --limit 3
gh run view --log-failed   # if anything is red
```

## 13. Release build (local snapshot)

```bash
go install github.com/goreleaser/goreleaser/v2@latest
goreleaser release --snapshot --clean
ls dist/
# expect: archives for linux/darwin/windows × amd64/arm64
dist/iqtoolkit-analyzer_darwin_arm64_*/iqtoolkit-analyzer --version  # match your platform
```

## Cleanup

```bash
docker rm -f iqa-pg
rm -f /tmp/sample-stderr.log /tmp/sample.csv /tmp/sample.json /tmp/empty.log /tmp/report.json /tmp/report.html
rm -rf dist/ coverage.out
```
