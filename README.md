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

## How It Works

```
PostgreSQL Logs ──> Log Parser ──> Metrics Analyzer ──┐
                                                      ├──> Recommendation Engine ──> Report
PostgreSQL DB ───> Config Reader ─────────────────────┘
```

1. The **log parser** reads PostgreSQL log files and extracts structured entries (timestamp, level, message, duration).
2. The **database connector** queries `pg_settings` to retrieve current configuration values and their sources.
3. The **metrics analyzer** processes parsed log entries and database settings into a summary report.
4. The **recommendation engine** evaluates the report against best-practice thresholds and produces categorized, severity-rated suggestions.

## Project Structure

```
iqtoolkit-analyzer/
├── internal/
│   ├── dbconn/             # Database connection and pg_settings queries
│   ├── logparser/          # PostgreSQL log file parsing
│   ├── metrics/            # Log and config analysis into reports
│   └── recommendations/    # Recommendation generation from metrics
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
```

The tool will output a report containing:
- Summary metrics (total entries, error count, slow queries, average duration)
- Peak error times by hour
- Categorized recommendations with severity levels

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
