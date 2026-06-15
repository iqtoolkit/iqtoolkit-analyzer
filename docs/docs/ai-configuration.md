---
sidebar_position: 4
---

# AI Configuration

iqtoolkit-analyzer can use AI providers to generate enhanced tuning recommendations from your PostgreSQL metrics.

## Supported Providers

| Provider | Flag Value | Model Examples |
|----------|-----------|----------------|
| OpenAI | `openai` | `gpt-4o`, `gpt-4o-mini` |
| Anthropic | `anthropic` | `claude-sonnet-4-5`, `claude-sonnet-4-20250514` |
| Google Gemini | `gemini` | `gemini-pro` |
| Kiro (Amazon Bedrock) | `kiro` | Uses AWS credentials |

## Environment Variables

Set the API key for your chosen provider:

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Google Gemini
export GEMINI_API_KEY="AI..."

# Kiro (Amazon Bedrock) — uses standard AWS credentials
export AWS_REGION="us-east-1"
```

## Config File

Alternatively, create `~/.config/iqtoolkit-analyzer/config.json`:

```json
{
  "openai_api_key": "sk-...",
  "anthropic_api_key": "sk-ant-...",
  "gemini_api_key": "AI...",
  "aws_region": "us-east-1"
}
```

:::info
Environment variables take precedence over the config file.
:::

## Usage

```bash
# Use OpenAI for enhanced recommendations
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider openai

# Use a specific model
iqtoolkit-analyzer analyze \
  --dsn "postgres://user:pass@localhost:5432/mydb" \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider anthropic --ai-model claude-sonnet-4-20250514

# Set model via environment variable
export IQTOOLKIT_AI_MODEL="gpt-4o"
iqtoolkit-analyzer analyze \
  --log-file /var/log/postgresql/postgresql.log \
  --ai-provider openai
```

## Sample AI Output

```
=== AI-Enhanced Recommendations ===
Based on your metrics, here are prioritized tuning suggestions:
1. Increase shared_buffers from 128MB to at least 1GB (25% of RAM)...
2. Set effective_cache_size to 75% of total RAM...
3. Consider enabling pg_stat_statements for query-level insights...
```
