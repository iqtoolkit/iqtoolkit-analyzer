---
sidebar_position: 1
slug: /
---

# Introduction

**iqtoolkit-analyzer** is a CLI tool for PostgreSQL health checking and performance tuning recommendations.

It connects to a PostgreSQL database, reads its runtime configuration, parses PostgreSQL log files, analyzes performance metrics, and generates actionable tuning recommendations. It acts as an intelligent advisor that identifies performance bottlenecks, configuration issues, and reliability concerns across your PostgreSQL deployment.

Recommendations are categorized by concern (performance, configuration, reliability) and assigned severity levels (critical, warning, info) so you can prioritize the most impactful changes first.

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
5. The **AI context builder** formats metrics and settings into a structured prompt, then sends it to a configured AI provider for enhanced tuning recommendations.

## Features

- **Log Parsing** — Extracts timestamps, log levels, messages, and query durations from PostgreSQL log files. Supports `stderr`, `csvlog`, and `jsonlog` formats with auto-detection.
- **Metrics Analysis** — Calculates total log entries, error counts, slow query counts, average query duration, and peak error times.
- **Configuration Review** — Inspects runtime settings via `pg_settings`, checking for suboptimal values.
- **Extended Data Collection** — Queries `pg_stat_statements`, `pg_stat_user_tables`, `pg_stat_user_indexes`, and `pg_buffercache`.
- **Actionable Recommendations** — Generates prioritized suggestions based on collected metrics.
- **AI-Enhanced Analysis** — Optionally uses OpenAI, Anthropic, Gemini, or Kiro/Amazon Bedrock for deeper tuning recommendations.
- **HTML Report** — Generates a self-contained HTML report with all settings and extension info.
