# Database Design

## Purpose

The database stores telemetry collected by KuFlow during HTTP request processing.

The primary goal is not to store user data, but to collect operational metadata for monitoring, analytics, and future system improvements.

---

# Design Goals

The database should:

- Store request metadata.
- Support analytical queries.
- Be easy to extend.
- Avoid unnecessary data duplication.
- Be managed through SQL migrations.

---

# Core Entity

The primary entity in the MVP is an HTTP request.

Each processed request generates one telemetry record.

---

# Database Engine

PostgreSQL was selected because it provides:

- Excellent SQL support
- Strong indexing capabilities
- Rich data types
- High performance
- Wide adoption in Data Engineering

---

# Initial Database Schema

The first version of the database contains one main table:

```text
requests
```

Future versions may introduce:

```text
clients

hosts

users

sessions

metrics
```

---

# Request Entity

| Field | Type | Required | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | ✅ | Unique request identifier |
| method | VARCHAR(10) | ✅ | HTTP method |
| host | VARCHAR(255) | ✅ | Destination host |
| path | TEXT | ✅ | Request path |
| status_code | SMALLINT | ✅ | HTTP response code |
| request_size | INTEGER | ✅ | Request size (bytes) |
| response_size | INTEGER | ✅ | Response size (bytes) |
| latency_ms | INTEGER | ✅ | Processing time |
| client_ip | INET | Optional | Client IP address |
| created_at | TIMESTAMP | ✅ | Request timestamp |

---

# Indexing Strategy

The following fields should be indexed:

- host
- status_code
- created_at

These indexes improve analytical query performance.

---

# Naming Convention

Tables

- snake_case
- plural form

Example:

```text
requests
users
sessions
```

Columns

- snake_case
- descriptive names

Example:

```text
created_at

status_code

request_size
```

---

# Migration Strategy

Database schema changes are managed through SQL migration files.

Example:

```text
001_init.sql

002_add_indexes.sql

003_add_metrics.sql
```

Each migration must be:

- Incremental
- Reproducible
- Version controlled

---

# Future Database Evolution

Version 1

```text
requests
```

↓

Version 2

```text
requests
clients
hosts
```

↓

Version 3

```text
requests
clients
hosts
metrics
```

↓

Version 4

```text
requests
clients
hosts
metrics
dashboards
```