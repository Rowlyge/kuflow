# Architecture Decisions

This document records important technical decisions made during the development of KuFlow.

---

# ADR-001 — Use Go

## Decision

Use Go as the primary programming language.

## Reason

* Excellent networking support.
* High performance.
* Simple deployment.
* Great concurrency model.
* Strong standard library.

## Consequences

* Easy static binary deployment.
* Low memory usage.
* Fast startup time.

---

# ADR-002 — Use PostgreSQL

## Decision

Use PostgreSQL as the primary database.

## Reason

* Excellent SQL support.
* Rich data types.
* Strong indexing capabilities.
* Widely used in Data Engineering.
* Great Go ecosystem (`pgx`).

## Consequences

* Easy analytical queries.
* Reliable long-term storage.

---

# ADR-003 — Use Docker Compose

## Decision

Use Docker Compose for local development.

## Reason

* Reproducible environment.
* Easy onboarding.
* Simple infrastructure management.

## Consequences

* Developers do not need local PostgreSQL installation.
* Same environment for all contributors.

---

# ADR-004 — Store Telemetry Instead of Payloads

## Decision

Store request metadata, not request/response bodies.

## Reason

* Smaller storage requirements.
* Better privacy.
* Simpler schema.
* Sufficient for performance analysis.

## Stored Data

* method
* host
* path
* status code
* request size
* response size
* latency
* timestamp

---

# ADR-005 — Use Linux Filesystem for Development

## Decision

Develop inside the WSL Linux filesystem.

## Reason

* Better Docker performance.
* Better Go tooling compatibility.
* Fewer path issues.

## Recommended Path

```bash
~/kuflow
```
