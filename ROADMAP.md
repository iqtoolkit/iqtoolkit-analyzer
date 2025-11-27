# 🚦 Slow Query Doctor Roadmap

## Current Release: v0.1.5 ✅ SHIPPED

- Database slow query log parsing (PostgreSQL focus)
- Query normalization and grouping
- Statistical analysis (impact scores, percentiles)
- AI-powered optimization recommendations (OpenAI only)
- Markdown report generation
- Docker containerization
- Anti-pattern detection
- Multi-format log support (plain, CSV, JSON)

**AI Provider Limitation**: v0.1.x supports **OpenAI GPT models only** (requires OPENAI_API_KEY). For sensitive database logs, consider waiting for v0.2.0 with local Ollama support and configurable AI providers.

---

## v0.1.6 - Final v0.1.x Feature Release (November 2025) 🔒 FEATURE FREEZE

**Focus:** Complete documentation and prepare for v0.2.0

**🚨 IMPORTANT**: This is the **FINAL v0.1.x release with new features**. Any future v0.1.x updates (v0.1.7+) will be **bug fixes only**. All new features go to v0.2.0+.

- [x] Add comprehensive ARCHITECTURE.md documentation
- [x] Update all references from "PostgreSQL-specific" to "database log analyzer"
- [x] Clear roadmap timeline and scope boundaries
- [x] Project discipline guidelines (.gitmessage)
- [x] Prepare codebase for multi-database expansion in v0.4.0
- [x] Plan configurable AI provider architecture (Ollama default, OpenAI optional)
- [x] Design extensible AI provider system for future models (Claude, Gemini, etc.)
- [x] Create placeholder sample log directories for MySQL and SQL Server
- [x] Add comprehensive AI provider extensibility guide in ARCHITECTURE.md
- [x] Establish release tagging strategy (v0.1.6-final-feature)

**Post-v0.1.6**: Only critical bug fixes allowed in v0.1.x branch. Feature development moves to v0.2.0.

---

## v0.2.2 - MongoDB Support + AI Provider Flexibility (Nov 2025) ✅ STABLE

**Focus:** MongoDB slow query analysis and configurable AI providers

**🎉 SHIPPED FEATURES**
- ✅ **MongoDB slow query analyzer** with profiler integration
- ✅ **MongoDB query pattern recognition** and normalization
- ✅ **Enhanced configuration system** (expanded YAML configuration with MongoDB support)
- ✅ **HTML report generation** (interactive dashboards with MongoDB-specific insights)
- ✅ **Multi-format reporting** (JSON, Markdown, HTML)
- ✅ **CLI integration** with database-specific subcommands
- ✅ **Configurable AI providers** (Ollama + OpenAI) - **SHIPPED EARLY** 🚀
- ✅ **Privacy-first AI** with local Ollama support (enterprise-ready)
- ✅ **Flexible model configuration** (custom hosts, multiple models)

**Complete Feature List:**
- [x] **MongoDB slow query analyzer** (complete with profiler integration)
- [x] **Enhanced configuration system** (expanded .iqtoolkit-analyzer.yml options)
- [x] **HTML report generation** (interactive dashboards)
- [x] **Multi-file analysis** (batch processing)
- [x] **Ollama integration** (arctic-text2sql-r1:7b default model)
- [x] **OpenAI integration** (gpt-4o-mini default model)
- [x] **Configurable LLM providers** via LLMConfig
- [x] **Custom Ollama host support** for self-hosted deployments

---

## v0.2.3 - PostgreSQL EXPLAIN + MongoDB Enhancements (Q1 2026) 🎯 NEXT

**Focus:** Advanced query plan analysis and MongoDB optimization

**PostgreSQL EXPLAIN Analyzer (RFC-002):**
- [ ] **EXPLAIN plan parser** (JSON/YAML formats)
- [ ] **Execution metrics extraction** (performance, I/O, estimation accuracy)
- [ ] **Anti-pattern detection** (seq scans, inefficient joins, poor filters)
- [ ] **Index recommendations** from EXPLAIN analysis
- [ ] **Unified reporting** (integrate EXPLAIN insights into existing reports)
- [ ] **LLM-powered insights** for complex execution plans
- [ ] **PostgreSQL 9.6-16 support** with version-specific handling

**MongoDB Enhancements:**
- [ ] **MongoDB aggregation pipeline optimization** recommendations
- [ ] **MongoDB indexing strategy** analysis and suggestions
- [ ] **Query complexity scoring and classification**
- [ ] **Enhanced anti-pattern detection** for MongoDB-specific issues

**Developer Experience:**
- [ ] **FastAPI backend** for programmatic access (optional REST API)
- [ ] **Improved error messages** and troubleshooting guides
- [ ] **Performance benchmarks** and optimization documentation

---

## v0.3.0 - MySQL & SQL Server Support (Q2 2026) 📋 PLANNED

**Focus:** Expand to traditional SQL databases (moved up from v0.4.0)

**MySQL Support:**
- [ ] **MySQL slow query log parser** (standard and JSON formats)
- [ ] **MySQL-specific anti-patterns** (covering indexes, table scans, temp tables)
- [ ] **InnoDB-specific optimizations** (buffer pool, transaction isolation)
- [ ] **MySQL EXPLAIN analysis** (including EXPLAIN FORMAT=JSON)

**SQL Server Support:**
- [ ] **Extended Events parser** (query performance tracking)
- [ ] **Query Store integration** (if available)
- [ ] **SQL Server execution plan analysis** (XML format)
- [ ] **SQL Server-specific recommendations** (indexes, statistics, query hints)

**Unified Features:**
- [ ] **Cross-database performance comparison** (PostgreSQL + MongoDB + MySQL + SQL Server)
- [ ] **Database-agnostic query analysis engine** (shared anti-pattern detection)
- [ ] **Unified configuration** for multiple database types
- [ ] **Comparison reports** (identify slowest queries across all databases)

---

## v0.4.0 - Self-Learning & ML Intelligence (Q3 2026) 📋 PLANNED

**Focus:** ML-based intelligence and historical tracking

**Historical Analysis:**
- [ ] **Track query performance over time** (SQLite/PostgreSQL storage)
- [ ] **Trend analysis** (queries getting slower/faster over time)
- [ ] **Automatic baseline detection** (establish normal performance)
- [ ] **Performance regression detection** (alert on degradation)

**Machine Learning:**
- [ ] **ML-based anomaly detection** for new slow queries
- [ ] **Identify performance regression patterns** (recurring issues)
- [ ] **Confidence scoring** for recommendations (based on historical success)
- [ ] **Predictive alerts** for queries likely to become slow
- [ ] **Learn from user feedback** (which recommendations worked)

**Integration:**
- [ ] **CI/CD integration** (fail builds on performance regression)
- [ ] **Prometheus/Grafana integration** (real-time monitoring)
- [ ] **Alerting system** (Slack/Teams/email notifications)

---

## v1.0.0 - Production Ready (Q4 2026)

**Focus:** Enterprise-grade stability and features

- [ ] Web UI for easier analysis
- [ ] API for programmatic access
- [ ] Authentication and multi-user support
- [ ] Scheduled analysis and alerting
- [ ] Integration with monitoring tools (Prometheus, Grafana)
- [ ] Performance regression CI/CD integration
- [ ] Comprehensive test coverage
- [ ] Enterprise support options

---

## Future Ideas (Backlog)

- Real-time query monitoring integration
- Automated optimization application (with approval workflow)
- Query workload simulator
- Cost estimation for cloud databases
- Integration with query plan visualizers
- Mobile app for on-call DBAs
- Slack/Teams notification integration
- AI-powered query rewriting suggestions

---

## Community Requests

Track feature requests from users here:

- **Ravi Bhatia:** ML/self-learning system for recommendations → BACKLOG (v0.3.0, Q2 2026)
- **Uri Dimant:** Query rewrites (not just indexes) → IMPLEMENTED ✅

## Version Timeline Summary

| Version | Timeline | Status | Key Features |
|---------|----------|--------|--------------|
| v0.1.5 | ✅ SHIPPED | Mature | PostgreSQL analyzer with OpenAI only |
| v0.1.6 | ✅ SHIPPED | Mature | Final v0.1.x feature release, documentation |
| v0.2.2 | ✅ **STABLE** | **Current** | **MongoDB + Ollama/OpenAI providers** 🚀 |
| v0.2.3 | Q1 2026 | 🎯 Next | **PostgreSQL EXPLAIN analyzer + MongoDB enhancements** |
| ~~v0.2.4~~ | ~~Q1 2026~~ | ❌ **Cancelled** | Features merged into v0.2.2 and v0.2.3 |
| v0.3.0 | Q2 2026 | 📋 Planned | **MySQL + SQL Server support** (moved up) |
| v0.4.0 | Q3 2026 | 📋 Planned | **ML/self-learning, anomaly detection** |
| v1.0.0 | Q4 2026 | 📋 Planned | Web UI, enterprise features |

**🎉 Ahead of Schedule:** v0.2.4's AI provider flexibility shipped **3 months early** in v0.2.2!

---

## Release & Versioning (short)

See the published release process and versioning guide in the documentation: `docs/release-process.md`.
Key points:
- `VERSION` at repo root is the single source of truth.
- Use the provided GitHub Action `/.github/workflows/propagate-version.yml` to propagate and tag releases.
- Automate changelog generation (Release Drafter or conventional-changelog) and add security scans to CI for release gating.


## Contributing

See issues labeled with `good-first-issue` or `help-wanted` for ways to contribute!

Questions or suggestions? Email: gio@gmartinez.net
---

**Made with ❤️ for Database performance optimization**