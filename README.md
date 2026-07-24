# KuFlow

> High-performance HTTP Proxy and Telemetry Platform written in Go.

## Overview

KuFlow is an educational open-source project developed by two Computer Engineering students.

The main purpose of the project is to build a production-like HTTP proxy server while learning:

- Backend Development
- Networking
- Data Engineering
- System Design
- Docker
- PostgreSQL
- Go

Instead of only forwarding HTTP requests, KuFlow also collects telemetry for further analysis.

---

## Features

- HTTP request forwarding
- Telemetry collection
- PostgreSQL storage
- Dockerized infrastructure
- SQL migrations
- Clean project architecture

---

## Technology Stack

| Layer | Technology |
|--------|------------|
| Language | Go |
| Database | PostgreSQL |
| Containers | Docker |
| Version Control | Git |
| IDE | VS Code |
| OS | Ubuntu (WSL) |

---

## Project Structure

```text
kuflow/
│
├── docs/
│   ├── architecture.md
│   ├── database.md
│   └── roadmap.md
│
├── migrations/
│
├── proxy/
│
├── database/
│
├── scripts/
│
├── docker-compose.yml
├── .env.example
├── .gitignore
├── LICENSE
└── README.md
```

---

## Development Principles

- Keep architecture simple.
- Write readable code.
- Document important decisions.
- Use version-controlled database migrations.
- Prefer explicit solutions over hidden magic.

---

## Current Status

Project is currently in the architecture and database design stage.

The next milestone is implementing the first HTTP proxy service.

---

## Authors

- Michail Sokun
- Georgiy Kuzin