# KuFlow

> High-performance Reverse Proxy and HTTP Telemetry Platform written in Go.

## Overview

KuFlow is an educational open-source project developed by two Computer Engineering students.

The goal of the project is to build a production-like reverse proxy while learning real backend and infrastructure engineering practices, including:

* Backend Development
* Networking
* Reverse Proxy Architecture
* Load Balancing
* System Design
* PostgreSQL
* Database Migrations
* Docker
* Go

Unlike a simple reverse proxy, KuFlow is designed to become a complete HTTP traffic collection and analysis platform.

---

## Features

* Reverse HTTP Proxy
* Multiple Upstream Servers
* Round Robin Load Balancing
* Automatic Health Checker
* HTTP Telemetry Collection
* PostgreSQL Telemetry Storage
* Versioned Database Migrations
* Dockerized Infrastructure
* Clean Project Architecture

---

## Technology Stack

| Layer           | Technology            |
| --------------- | --------------------- |
| Language        | Go                    |
| HTTP            | net/http              |
| Reverse Proxy   | httputil.ReverseProxy |
| Database        | PostgreSQL            |
| Migrations      | golang-migrate        |
| Containers      | Docker                |
| Version Control | Git                   |
| IDE             | VS Code               |
| OS              | Ubuntu (WSL)          |

---

## Project Structure

```text
kuflow/
│
├── cmd/
│   ├── kuflow/
│   └── upstream/
│
├── docs/
│   ├── architecture.md
│   ├── database.md
│   └── roadmap.md
│
├── internal/
│   ├── app/
│   ├── balancer/
│   ├── clientip/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── health/
│   ├── middleware/
│   ├── model/
│   ├── proxy/
│   ├── repository/
│   ├── service/
│   └── upstream/
│
├── migrations/
│
├── docker-compose.yml
├── Makefile
├── .env.example
├── .gitignore
├── LICENSE
└── README.md
```

---

## Implemented Components

* Proxy Engine
* Upstream Manager
* Round Robin Balancer
* Automatic Health Checker
* Reverse Proxy Middleware
* HTTP Telemetry Middleware
* PostgreSQL Repository
* Versioned Database Migrations

---

## Development Principles

* Keep architecture simple.
* Prefer composition over complexity.
* Write readable and maintainable code.
* Keep business logic independent from infrastructure.
* Use version-controlled database migrations.
* Build production-like components instead of educational shortcuts.

---

## Current Status

Current implementation includes:

* Reverse Proxy Engine
* Multiple Upstream Support
* Round Robin Load Balancer
* Automatic Health Checker
* HTTP Telemetry Collection
* PostgreSQL Persistence
* Database Migrations using golang-migrate

### Next Milestones

* Configurable Load Balancing Algorithms
* Graceful Shutdown
* Advanced Telemetry
* Analytics Pipeline

---

## Authors

* Michail Sokun
* Georgiy Kuzin
