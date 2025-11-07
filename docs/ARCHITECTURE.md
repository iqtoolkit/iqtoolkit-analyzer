# Architecture: v0.2.0 - Privacy-First Database Query Analyzer

> **Database Support Timeline**:
> | Version | Timeline | Features |
> |---------|----------|----------|
> | v0.1.5 | ✅ SHIPPED | PostgreSQL slow query analyzer |
> | v0.2.0 | Nov 2025 - Q1 2026 | EXPLAIN plans, anti-patterns, HTML reports, multi-file analysis, **Ollama integration** |
> | v0.3.0 | Q2 2026 | ML/self-learning, historical tracking, anomaly detection |
> | v0.4.0 | Q3 2026 | **MySQL and SQL Server support** |
> | v1.0.0 | Q4 2026 | Web UI, enterprise features, API |
>
> The architecture is deliberately designed to be database-agnostic, with clear extension points for supporting multiple database engines. Current PostgreSQL focus allows rapid development and validation of core features.
>
> **Note**: We're focusing on PostgreSQL, MySQL, and SQL Server—the most common enterprise databases. Oracle support is not planned as it would require significantly different log formats and dedicated engineering resources.

## Design Principles

### 1. Privacy First: Ollama (Local) Replaces OpenAI (Public)

PostgreSQL logs contain sensitive business data:
- Customer information
- Proprietary query patterns
- Database schema details
- Performance metrics tied to business operations

**v0.1.x Problem**: Used OpenAI's public API → data leaves customer infrastructure → compliance violation

**v0.2.0 Solution**: Ollama for local execution
- ✅ Local execution (customer's infrastructure)
- ✅ No data transmission outside network
- ✅ Enterprise compliance (HIPAA, SOC 2, GDPR ready)
- ✅ Cost-effective at scale (no per-API-call charges)
- ✅ Deterministic results (no dependency on external service availability)

### 2. Enterprise Ready

v0.2.0 features designed for production deployment:

| Feature | Purpose | Enterprise Impact |
|---------|---------|-------------------|
| EXPLAIN Plan Analysis | Understand query execution | Compliance with performance audits |
| Anti-pattern Detection | Catch common mistakes | Reduce bad deployments |
| HTML Reports | Executive-friendly output | Audit trails & documentation |
| Multi-file Analysis | Batch process logs | Handle real production workloads |
| FastAPI Backend | REST API integration | Integrate with existing tools |

### 3. Incremental Shipping

**Strict discipline**: Complete v0.2 before starting v0.3

```
v0.1.5 (SHIPPED)
    ↓
v0.2.0 (Nov 2025 - Q1 2026)
├── EXPLAIN plans
├── Anti-patterns
├── HTML reports
├── Multi-file analysis
├── Ollama integration (CORE)
└── FastAPI backend
    ↓
v0.3.0 (Q2 2026 - NOT YET)
├── ML/self-learning
├── Historical tracking
├── Anomaly detection
└── [DO NOT START UNTIL v0.2 COMPLETE]
```

**Rule**: Don't chase v0.3 features while v0.2 isn't done. ML is 6+ months away.

## Why This Matters for Your Career

You're not just building a tool—you're solving enterprise data compliance. This positioning is critical for evolving from DBA to **AI-enhanced database architect**.

### Career Positioning

1. **DBA Today**: Manages databases, optimizes queries
2. **AI-Enhanced Architect Tomorrow**: Understands how AI can solve enterprise problems responsibly

**v0.2.0 demonstrates**:
- Understanding of enterprise constraints (data privacy)
- Ability to make architectural decisions based on business requirements
- Building for compliance and extensibility
- Database-agnostic design principles

This version establishes the foundation for a professional, multi-database solution while delivering immediate value for PostgreSQL users. The architecture is deliberately modular to support additional database engines without significant refactoring.

**Why PostgreSQL First?**
1. Establish core features with a well-documented database
2. Validate privacy-first architecture
3. Build robust parsing and analysis patterns
4. Create extensible interfaces for future database support

This is what separates hobbyist projects from production systems.

## Technical Architecture

### Core Components

**1. Database Engine Abstraction** (`app/core/db_engines/`)
- Abstract interface for database engines
- PostgreSQL implementation (v0.2.0)
- Extension points for MySQL (v0.4.0)
- Extension points for SQL Server (v0.4.1)

**2. Ollama Client** (`app/core/ollama_client.py`)
- Local LLM inference
- Streaming responses
- Health monitoring
- Timeout handling
- Database-specific prompt templates

**2. Query Analysis Service** (`app/services/query_analyzer.py`)
- Parse LLM responses into structured data
- Extract indexes, rewrites, anti-patterns
- Generate recommendations

**3. Log Parser Service** (`app/services/log_parser.py`)
- Extract slow queries from PostgreSQL logs
- Filter by threshold
- Sort by impact

**4. Report Generator** (`app/services/report_generator.py`)
- HTML report generation
- Syntax highlighting
- Executive summaries

**5. FastAPI Backend** (`app/routers/`)
- `/analyze/query` - Single query analysis
- `/analyze/log` - Batch log analysis
- `/analyze/log/upload` - File upload
- `/analyze/report/html` - HTML report generation
- `/health` - System health check

## Deployment Models

### Model 1: On-Premises (Enterprise Standard)

```
Customer Infrastructure
├── Ollama (local LLM)
├── FastAPI Backend
├── PostgreSQL (to analyze)
└── PostgreSQL Logs (to process)

Benefits:
- Zero data egress
- Full compliance
- Deterministic
- Customer controls versioning
```

### Model 2: Cloud VPC (Hybrid)

```
Customer VPC
├── Ollama (in private VPC)
├── FastAPI Backend
└── PostgreSQL

Your Infrastructure: Monitoring dashboard only
```

### Model 3: SaaS (Future, Post v1.0)

```
Only after:
- v1.0 shipped (enterprise features complete)
- Customer requests multi-tenant
- Compliance review done
```

## Contributors & Feedback Loop

### Uri
**Focus**: Query rewrite suggestions
- Feedback: Show actual SQL rewrites, not just indexes
- Status: Implemented in v0.2.0 `/analyze/query`
- Impact: Makes tool immediately actionable

### Ravi
**Focus**: ML/self-learning ideas
- Suggestion: Historical tracking + anomaly detection
- Timeline: v0.3.0 (Q2 2026), NOT v0.2.0
- Reason: Requires v0.2 baseline + production data first

### Weekly Discipline
- Track feedback in GitHub Issues tagged `v0.2.0`
- Add contributions to ROADMAP.md
- Document why feedback was included/deferred
- Post LinkedIn updates linking to milestones

## Success Metrics for v0.2.0

### Shipped (By Q1 2026)
- [ ] EXPLAIN plan analysis working
- [ ] Anti-patterns detected accurately
- [ ] HTML reports generated cleanly
- [ ] Multi-file analysis handles 100+ queries
- [ ] Ollama integration stable (no OpenAI fallback)
- [ ] FastAPI backend production-ready

### Adoption
- [ ] First customer using on-premises
- [ ] GitHub stars: 100+ (conservative estimate)
- [ ] LinkedIn engagement: 500+ followers
- [ ] Case study: "Why we switched from X to this tool"

## What v0.2.0 Is NOT

- ❌ Not a SaaS product (yet)
- ❌ Not MySQL/SQL Server support (that's v0.4)
- ❌ Not web UI (that's v1.0)
- ❌ Not ML anomaly detection (that's v0.3)
- ❌ Not a replacement for APM tools (complements them)

## Next Steps

1. **This Week (Nov 7-14)**
   - Add CONTEXT.md to GitHub
   - Fix Uri's feedback on query rewrites
   - Test FastAPI backend with 2-3 log patterns
   - Respond to LinkedIn comments with version roadmap
   - Tag v0.1.2 when stable

2. **This Month (Nov 15-30)**
   - Develop EXPLAIN plan parser
   - Build anti-pattern detector
   - Start HTML report generator

3. **December-Q1**
   - Complete v0.2.0 features
   - Security audit (Ollama + FastAPI)
   - Performance testing (100+ concurrent requests)
   - First customer pilot

## References

- FastAPI Backend: See `fastapi-v0.2` branch
- Ollama Integration: `app/core/ollama_client.py`
- Feature Roadmap: `ROADMAP.md`