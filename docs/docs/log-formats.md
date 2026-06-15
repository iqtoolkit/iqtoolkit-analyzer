---
sidebar_position: 5
---

# Log Formats

iqtoolkit-analyzer auto-detects the log format. You can override detection with `--log-format`.

## Supported Formats

| Format | PostgreSQL Setting | Description |
|--------|-------------------|-------------|
| `stderr` | `log_destination = 'stderr'` | Default line-based format |
| `csvlog` | `log_destination = 'csvlog'` | Comma-separated values |
| `jsonlog` | `log_destination = 'jsonlog'` | JSON (one object per line, PG 15+) |

## Examples

```bash
# Auto-detect (default)
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.log

# Explicit stderr format
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.log --log-format stderr

# CSV log format
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.csv --log-format csvlog

# JSON log format (PostgreSQL 15+)
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.json --log-format jsonlog
```

## Extension Requirements

For full data collection, iqtoolkit-analyzer uses these PostgreSQL extensions:

| Extension | Purpose | Required? |
|-----------|---------|-----------|
| `pg_stat_statements` | Top queries by execution time | Recommended |
| `pg_buffercache` | Buffer cache usage analysis | Optional |

If an extension is available but not installed, the tool will prompt:

```
Extension "pg_stat_statements" is available but not installed.
Run: CREATE EXTENSION pg_stat_statements;
```

If unavailable, the tool continues without that data source — no failure.
