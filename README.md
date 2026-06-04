# iqtoolkit-analyzer

A CLI tool for PostgreSQL health checking and performance tuning recommendations.

## Overview

iqtoolkit-analyzer connects to a PostgreSQL database, reads its runtime configuration, parses PostgreSQL log files, analyzes performance metrics, and generates actionable tuning recommendations. It acts as an intelligent advisor that identifies performance bottlenecks, configuration issues, and reliability concerns across your PostgreSQL deployment.

Recommendations are categorized by concern (performance, configuration, reliability) and assigned severity levels (critical, warning, info) so you can prioritize the most impactful changes first.

## Features

- **Log Parsing** - Extracts timestamps, log levels, messages, and query durations from PostgreSQL log files. Identifies slow queries by configurable duration thresholds.
- **Metrics Analysis** - Calculates total log entries, error counts, slow query counts, average query duration, and peak error times (by hour).
- **Configuration Review** - Connects to PostgreSQL and inspects runtime settings via `pg_settings`, checking for suboptimal values in parameters like `shared_buffers`, `work_mem`, and `log_min_duration_statement`.
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

iqtoolkit-analyzer uses the [cobra](https://github.com/spf13/cobra) CLI framework. Connect to your PostgreSQL instance and point it at your log files to receive a health report:

```bash
# Analyze PostgreSQL logs and configuration
iqtoolkit-analyzer analyze --dsn "postgres://user:pass@localhost:5432/mydb" --log-file /var/log/postgresql/postgresql.log

# Adjust the slow query threshold (default unit: milliseconds)
iqtoolkit-analyzer analyze --dsn "postgres://user:pass@localhost:5432/mydb" --log-file /var/log/postgresql/postgresql.log --slow-threshold 500

# Generate an HTML report with all settings, extensions, and version
iqtoolkit-analyzer report --dsn "postgres://user:pass@localhost:5432/mydb" --output report.html
```

The `analyze` command outputs:
- Summary metrics (total entries, error count, slow queries, average duration)
- Peak error times by hour
- Categorized recommendations with severity levels

The `report` command generates a self-contained HTML file containing:
- PostgreSQL server version
- All runtime settings from `pg_settings`
- All available and installed extensions with version info

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
