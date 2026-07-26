<div align="center">

# KuFlow

### High-performance Reverse Proxy and Telemetry Platform written in Go

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-MIT-green)
![Status](https://img.shields.io/badge/Status-Active%20Development-orange)

Production-inspired reverse proxy built to explore modern Go backend development,
network programming, telemetry collection and distributed systems architecture.

</div>

---

# Vision

KuFlow is an open-source educational project created by two Computer Engineering students.

The project aims to evolve into a modular, production-ready reverse proxy platform focused on:

- performance
- observability
- maintainability
- extensibility

Rather than building a simple HTTP forwarder, KuFlow is designed to resemble the architecture of real backend infrastructure used in production environments.

---

# Goals

This project is intended to gain practical experience in:

- Go Backend Development
- HTTP Reverse Proxy
- Networking
- Distributed Systems
- Clean Architecture
- System Design
- PostgreSQL
- Docker
- Telemetry
- Observability
- API Design
- Concurrent Programming

---

# Current Features

- HTTP server
- Reverse Proxy foundation
- Middleware pipeline
- Request logging
- Request ID generation
- Telemetry architecture
- PostgreSQL integration
- SQL migrations
- Environment configuration (.env)
- Dependency Injection (App container)
- Layered architecture

---

# Technology Stack

| Category | Technology |
|-----------|------------|
| Language | Go |
| HTTP | net/http |
| Database | PostgreSQL 16 |
| Database Driver | pgx/v5 |
| Containers | Docker |
| Version Control | Git + GitHub |
| IDE | VS Code |
| Development OS | Ubuntu (WSL) |

---

# Architecture Overview

```text
                    HTTP Request
                          │
                          ▼
                 Middleware Pipeline
        (Logger • RequestID • Telemetry)
                          │
                          ▼
                     HTTP Handlers
                          │
                          ▼
                        Services
                          │
                          ▼
                     Repositories
                          │
                          ▼
                      PostgreSQL
```

The application follows a layered architecture with dependency injection through a central `App` container.

---

# Project Structure

```text
kuflow/
│
├── cmd/
│   └── kuflow/
│
├── internal/
│   ├── app/
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── proxy/
│   ├── repository/
│   ├── requestid/
│   ├── router/
│   ├── server/
│   └── service/
│
├── docs/
│   ├── api.md
│   ├── architecture.md
│   ├── coding-style.md
│   ├── database.md
│   ├── decisions.md
│   ├── deployment.md
│   ├── development.md
│   ├── networking.md
│   └── roadmap.md
│
├── migrations/
│
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

# Getting Started

## Clone repository

```bash
git clone git@github.com:Rowlyge/kuflow.git

cd kuflow
```

---

## Start PostgreSQL

```bash
docker compose up -d
```

---

## Install dependencies

```bash
go mod tidy
```

---

## Configure environment

```bash
cp .env.example .env
```

Edit the configuration if necessary.

---

## Run the application

```bash
go run ./cmd/kuflow
```

Server:

```
http://localhost:8080
```

Health endpoint:

```
GET /health
```

---

# Development Workflow

Main branches:

- `main` — stable branch
- `develop` — active development

Feature development is performed in separate branches and merged into `develop` using Pull Requests.

---

# Roadmap

The project is planned to include:

- Reverse Proxy Engine
- Multiple upstream servers
- Round Robin load balancing
- Least Connections balancing
- Passive health checks
- Active health checks
- Retry policy
- Rate limiting
- Authentication
- Structured logging
- Metrics
- Prometheus integration
- Configuration hot reload
- Admin API
- Dashboard
- Request analytics

---

# Documentation

Complete project documentation is available in the `docs/` directory.

- Architecture
- Database Design
- API Specification
- Development Guide
- Deployment
- Networking
- Coding Style
- Design Decisions
- Roadmap

---

# Current Status

Current progress:

- Architecture completed
- PostgreSQL integrated
- Middleware pipeline implemented
- Reverse Proxy foundation implemented
- Dependency Injection implemented

### Next milestone

Develop a production-ready Reverse Proxy Engine with intelligent request forwarding and telemetry collection.

---

# Authors

**Michail Sokun**

**Georgiy Kuzin**

---

<div align="center">

Made with Go, PostgreSQL and Docker.

</div>