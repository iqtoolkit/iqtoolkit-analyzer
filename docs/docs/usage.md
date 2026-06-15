---
sidebar_position: 3
---

# Usage

## Quick Start

```bash
# Basic analysis — connect to PostgreSQL and analyze logs
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log

# Log-only mode — analyze a log file without a database connection
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.log
```

## Commands

### `analyze` — Analyze logs and configuration

```bash
iqtoolkit-analyzer analyze [flags]
```

**Required flags:**

| Flag | Description |
|------|-------------|
| `--log-file` | Path to PostgreSQL log file |

**Optional flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dsn` | _(none)_ | PostgreSQL connection string. Omit to run in **log-only mode** |
| `--slow-threshold` | `1000` | Slow query threshold in milliseconds |
| `--log-format` | auto-detect | Log format: `stderr`, `csvlog`, or `jsonlog` |
| `--ai-provider` | _(none)_ | AI provider: `openai`, `anthropic`, `gemini`, or `kiro` |
| `--ai-model` | _(provider default)_ | Override the AI model |
| `--format` | `text` | Output format: `text`, `json`, or `markdown` |
| `--output` | _(stdout)_ | Write output to a file instead of stdout |

### `report` — Generate an HTML report

```bash
iqtoolkit-analyzer report \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --output report.html
```

| Flag | Default | Description |
|------|---------|-------------|
| `--dsn` | _(required)_ | PostgreSQL connection string |
| `--output` | `report.html` | Output HTML file path |

The HTML report contains:
- PostgreSQL server version
- All runtime settings from `pg_settings`
- All available and installed extensions with version info

### Global Flags

```bash
iqtoolkit-analyzer --version
iqtoolkit-analyzer --help
iqtoolkit-analyzer analyze --help
```

## Examples

```bash
# Adjust the slow query threshold to 500ms
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --slow-threshold 500

# Parse a CSV-formatted log file
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.csv \
  --log-format csvlog

# Include AI-enhanced recommendations
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider openai

# Output as JSON to a file
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --format json --output analysis.json
```

## Sample Output

**Text format:**

```
=== Summary ===
Total entries:    14832
Error count:      47
Slow queries:     12
Avg duration:     45.2ms
Peak error time:  2024-03-15 14:00:00 +0000 UTC

=== Recommendations ===
[critical][performance] Average query duration (45.2ms) exceeds threshold
[warning][performance] 12 slow queries detected (threshold: 1000ms)
[warning][reliability] 47 errors detected in log file
[info][configuration] shared_buffers is set to default value; consider increasing
```

**JSON format:**

```json
{
  "summary": {
    "total_entries": 14832,
    "error_count": 47,
    "slow_queries": 12,
    "avg_duration": "45.2ms",
    "peak_error_time": "2024-03-15T14:00:00Z"
  },
  "recommendations": [
    {
      "severity": "critical",
      "category": "performance",
      "message": "Average query duration (45.2ms) exceeds threshold"
    }
  ]
}
```
