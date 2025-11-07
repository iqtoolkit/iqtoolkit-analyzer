# 🚦 Slow Query Doctor Roadmap

## Current Release: v0.1.5

- PostgreSQL slow query log parsing
- Query normalization and grouping
- Statistical analysis (impact scores, percentiles)
- GPT-4 powered optimization recommendations
- Markdown report generation
- Docker containerization
- Support for different PostgreSQL log formats (plain, CSV, JSON)
- Configuration file support (.slowquerydoctor.yml)
- Multi-line query parsing and special character handling
- Improved error messages and user guidance

---

## v0.1.x - Completed (November 2025)

**Focus:** Building a solid foundation with OpenAI integration

- [x] Add better sample outputs showing diverse query patterns
- [x] Improve recommendation quality (distinguish between missing indexes, query rewrites, schema changes)
- [x] Add more detailed explanations for why queries are slow
- [x] Handle edge cases in log parsing (multi-line queries, special characters)
- [x] Add support for different PostgreSQL log formats (plain, CSV, JSON)
- [x] Improve error messages and user guidance
- [x] Add configuration file support (.slowquerydoctor.yml)
- [x] Release v0.1.5 as stable foundation

---

## v0.2.0 - Privacy & API Support (Q1 2026)

**Focus:** Privacy-focused analysis and API integration

- [ ] Replace OpenAI with Ollama for enhanced privacy
  - Support configurable OpenAI URL/endpoint for enterprise users
  - Support both OpenAI and Ollama backends
- [ ] Add FastAPI web service layer
  - REST API endpoints for query analysis
  - Streaming analysis support
  - File upload handling
  - Background task processing
  - Rate limiting and API key auth
  - Swagger/OpenAPI documentation
- [ ] Enhanced analysis features
  - EXPLAIN plan analysis integration
  - Common anti-pattern detection
  - Query complexity scoring
  - Multi-file log analysis
  - Before/after comparison reports
  - HTML report generation
  - JSON/CSV export options
- [ ] Docker deployment
  - Docker Compose setup
  - Ollama container integration
  - Production configuration
  - GPU support (optional)

---

## v0.3.0 - Self-Learning & Predictive Analysis (Q2 2026)

**Focus:** ML-based intelligence and historical tracking

- [ ] Track query performance over time (historical database)
- [ ] Identify performance regression patterns
- [ ] ML-based anomaly detection for new slow queries
- [ ] Confidence scoring for recommendations
- [ ] Trend analysis (queries getting slower over time)
- [ ] Automatic baseline detection
- [ ] Predictive alerts for queries likely to become slow
- [ ] Learn from user feedback (which recommendations worked)

---

## v0.4.0 - Multi-Database Support (Q3 2026)

**Focus:** Expand beyond PostgreSQL

- [ ] MySQL slow query log support
  - Slow query log parsing
  - Performance schema integration
  - MySQL-specific recommendations
- [ ] SQL Server Extended Events support
  - XEvents log parsing
  - DMV integration
  - SQL Server-specific optimizations
- [ ] Database-agnostic query analysis layer
- [ ] Cross-database performance comparison
- [ ] Multi-format log handling
- [ ] Database-specific index recommendations

> **Note**: Focusing on PostgreSQL (v0.2.0), MySQL, and SQL Server—the most common enterprise databases. Oracle support is not planned as it requires significantly different log formats and dedicated engineering resources.

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

- **Ravi Bhatia:** ML/self-learning system for recommendations → Planned for v0.3.0
- **Uri Dimant:** Better examples showing query tuning (not just index recommendations) → In progress for v0.1.2

---

## Contributing

See issues labeled with `good-first-issue` or `help-wanted` for ways to contribute!

Questions or suggestions? Email: gio@gmartinez.net
---

**Made with ❤️ for Database performance optimization**