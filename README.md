# iqtoolkit-analyzer

A CLI tool for PostgreSQL health checking and performance tuning recommendations.

## Overview

iqtoolkit-analyzer connects to a PostgreSQL database, reads its runtime configuration, parses PostgreSQL log files, analyzes performance metrics, and generates actionable tuning recommendations. It acts as an intelligent advisor that identifies performance bottlenecks, configuration issues, and reliability concerns across your PostgreSQL deployment.

Recommendations are categorized by concern (performance, configuration, reliability) and assigned severity levels (critical, warning, info) so you can prioritize the most impactful changes first.

## Features

- **Log Parsing** - Extracts timestamps, log levels, messages, and query durations from PostgreSQL log files. Supports `stderr` (default), `csvlog`, and `jsonlog` formats with auto-detection. Identifies slow queries by configurable duration thresholds.
- **Metrics Analysis** - Calculates total log entries, error counts, slow query counts, average query duration, and peak error times (by hour).
- **Configuration Review** - Connects to PostgreSQL and inspects runtime settings via `pg_settings`, checking for suboptimal values in parameters like `shared_buffers`, `work_mem`, and `log_min_duration_statement`.
- **Extended Data Collection** - Queries `pg_stat_statements` (top queries by time), `pg_stat_user_tables` (sequential scans, dead tuples), `pg_stat_user_indexes` (unused indexes), and `pg_buffercache` (buffer cache usage). Automatically checks if required extensions are installed and prompts you to create them if not.
- **Actionable Recommendations** - Generates prioritized suggestions based on collected metrics, flagging high average query durations, excessive slow queries, elevated error counts, and misconfigured parameters.
- **AI-Enhanced Analysis** - Optionally sends metrics to an AI provider (OpenAI, Anthropic, Gemini, or Kiro/Amazon Bedrock) for deeper tuning recommendations beyond rule-based checks.
- **HTML Report** - Generates a self-contained HTML report with all `pg_settings`, installed and available extensions (with versions), and the PostgreSQL server version.

## How It Works

```
PostgreSQL Logs ──> Log Parser ──> Metrics Analyzer ──┐
                                                      ├──> Recommendation Engine ──> Report
PostgreSQL DB ───> Config Reader ─────────────────────┘
                                                      │
                                                      └──> AI Context Builder ──> AI Provider ──> Enhanced Recommendations
```

1. The **log parser** reads PostgreSQL log files and extracts structured entries (timestamp, level, message, duration).
2. The **database connector** queries `pg_settings` to retrieve current configuration values and their sources.
3. The **metrics analyzer** processes parsed log entries and database settings into a summary report.
4. The **recommendation engine** evaluates the report against best-practice thresholds and produces categorized, severity-rated suggestions.
5. The **AI context builder** formats metrics and settings into a structured prompt, then sends it to a configured AI provider (OpenAI, Anthropic, Gemini, or Kiro/Bedrock) for enhanced tuning recommendations.

## AI Provider Configuration

iqtoolkit-analyzer can use AI providers (OpenAI, Anthropic, Gemini, or Kiro/Amazon Bedrock) to generate enhanced tuning recommendations from your PostgreSQL metrics.

### Environment Variables

Set the API key for your chosen provider:

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Google Gemini
export GEMINI_API_KEY="AI..."

# Kiro (Amazon Bedrock) — uses standard AWS credentials (env vars, profile, or IMDS)
export AWS_REGION="us-east-1"
```

### Config File

Alternatively, create `~/.config/iqtoolkit-analyzer/config.json`:

```json
{
  "openai_api_key": "sk-...",
  "anthropic_api_key": "sk-ant-...",
  "gemini_api_key": "AI...",
  "aws_region": "us-east-1"
}
```

Environment variables take precedence over the config file.

## Project Structure

```
iqtoolkit-analyzer/
├── internal/
│   ├── ai/                 # AI provider clients and prompt building
│   ├── dbconn/             # Database connection and pg_settings queries
│   ├── logparser/          # PostgreSQL log file parsing
│   ├── metrics/            # Log and config analysis into reports
│   ├── recommendations/    # Recommendation generation from metrics
│   └── report/             # HTML report generation
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Prerequisites

- Go 1.22 or later
- Access to a PostgreSQL instance (for configuration analysis)
- PostgreSQL log files (for log-based analysis)

## Installation

```bash
go install github.com/iqtoolkit/iqtoolkit-analyzer@latest
```

Or build from source:

```bash
git clone https://github.com/iqtoolkit/iqtoolkit-analyzer.git
cd iqtoolkit-analyzer
go build ./...
```

## Usage

### Quick Start

```bash
# Basic analysis — connect to PostgreSQL and analyze logs
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log

# Log-only mode — analyze a log file without a database connection
iqtoolkit-analyzer analyze --log-file /var/log/postgresql/postgresql.log
```

### Commands

#### `analyze` — Analyze logs and configuration

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
| `--dsn` | _(none)_ | PostgreSQL connection string (`postgres://user:pass@host:port/db`). Omit to run in **log-only mode** — settings and runtime stats are skipped |
| `--slow-threshold` | `1000` | Slow query threshold in milliseconds |
| `--log-format` | auto-detect | Log format: `stderr`, `csvlog`, or `jsonlog` |
| `--ai-provider` | _(none)_ | AI provider: `openai`, `anthropic`, `gemini`, or `kiro` |
| `--ai-model` | _(provider default)_ | Override the AI model (e.g., `gpt-4o`, `claude-sonnet-4-5`). Can also be set via `IQTOOLKIT_AI_MODEL` env var |
| `--format` | `text` | Output format: `text`, `json`, or `markdown` |
| `--output` | _(stdout)_ | Write output to a file instead of stdout |

**Examples:**

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

# Parse a JSON-formatted log file
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.json \
  --log-format jsonlog

# Include AI-enhanced recommendations (requires API key configured)
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider openai

# Use a specific AI model
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider anthropic --ai-model claude-sonnet-4-20250514

# Output as JSON to a file
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --format json --output analysis.json

# Output as Markdown
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --format markdown --output analysis.md
```

**Sample output (text):**

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

=== AI-Enhanced Recommendations ===
Based on your metrics, here are prioritized tuning suggestions:
1. Increase shared_buffers from 128MB to at least 1GB (25% of RAM)...
```

**Sample output (JSON):**

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
  ],
  "ai_recommendations": "..."
}
```

#### `report` — Generate an HTML report

```bash
iqtoolkit-analyzer report [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--dsn` | _(required)_ | PostgreSQL connection string |
| `--output` | `report.html` | Output HTML file path |

```bash
iqtoolkit-analyzer report \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --output report.html
```

The HTML report contains:
- PostgreSQL server version
- All runtime settings from `pg_settings`
- All available and installed extensions with version info

#### Global Flags

```bash
# Show version
iqtoolkit-analyzer --version

# Show help
iqtoolkit-analyzer --help
iqtoolkit-analyzer analyze --help
```

### Log Format Support

iqtoolkit-analyzer auto-detects the log format. You can override detection with `--log-format`:

| Format | PostgreSQL Setting | Description |
|--------|-------------------|-------------|
| `stderr` | `log_destination = 'stderr'` | Default line-based format |
| `csvlog` | `log_destination = 'csvlog'` | Comma-separated values |
| `jsonlog` | `log_destination = 'jsonlog'` | JSON (one object per line, PG 15+) |

### Extension Requirements

For full data collection, iqtoolkit-analyzer uses these PostgreSQL extensions:

| Extension | Purpose | Required? |
|-----------|---------|-----------|
| `pg_stat_statements` | Top queries by execution time | Recommended |
| `pg_buffercache` | Buffer cache usage analysis | Optional |

If an extension is available but not installed, the tool will prompt:

```
Extension "pg_stat_statements" is available but not installed. Run: CREATE EXTENSION pg_stat_statements;
```

If unavailable, the tool continues without that data source — no failure.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
