# Coding Style

This document defines the coding conventions used in KuFlow.

---

# General Principles

* Prefer readability over cleverness.
* Keep functions small.
* Handle errors explicitly.
* Avoid unnecessary abstractions.
* Write code that is easy to debug.

---

# Go Formatting

Always use:

```bash
gofmt -w .
```

Never manually align code with spaces.

---

# Package Naming

Use lowercase package names.

Good:

```text
proxy
storage
config
logger
```

Bad:

```text
Proxy
DataBase
HTTPServer
```

---

# File Naming

Use snake_case for multi-word files.

Examples:

```text
http_proxy.go
request_logger.go
postgres_storage.go
```

---

# Error Handling

Always check errors immediately.

Good:

```go
rows, err := db.Query(ctx, query)
if err != nil {
    return err
}
```

Bad:

```go
rows, _ := db.Query(ctx, query)
```

---

# Logging

Use structured logs.

Good:

```text
level=info msg="request processed" host=github.com latency_ms=24
```

Avoid:

```text
Request processed successfully!!!
```

---

# Configuration

All configuration should come from environment variables.

Do not hardcode:

* ports
* database credentials
* file paths
* API keys

---

# Database

* All schema changes go through migrations.
* Never modify production tables manually.
* Use descriptive column names.

Examples:

```text
created_at
status_code
response_size
```

---

# Git Commits

Use short imperative messages.

Good:

```text
Add PostgreSQL configuration
Implement request logging
Create initial migration
```

Bad:

```text
fixed stuff
changes
update
```

---

# Comments

Write comments only when they add useful context.

Good:

```go
// Measure upstream response latency before writing the response.
```

Bad:

```go
// Increment i
i++
```
