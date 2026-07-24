# Development Guide

This document describes how to set up and run KuFlow for local development.

---

# Requirements

The following software must be installed:

* Ubuntu (WSL recommended)
* Docker
* Git
* Visual Studio Code
* VS Code Remote - WSL extension

---

# Project Location

The project should be stored inside the Linux filesystem.

Recommended location:

```bash
~/kuflow
```

Avoid developing inside `/mnt/c/...` because Docker and Go tooling work better on the native Linux filesystem.

---

# Opening the Project

Open Ubuntu and run:

```bash
cd ~/kuflow
code .
```

VS Code should show **WSL: Ubuntu** in the bottom-left corner.

---

# Starting PostgreSQL

Run:

```bash
docker-compose up -d
```

Check running containers:

```bash
docker ps
```

You should see:

```text
proxy-postgres
```

---

# Stopping Services

```bash
docker-compose down
```

---

# Connecting to PostgreSQL

```bash
docker exec -it proxy-postgres psql -U proxy -d proxydb
```

Useful commands:

```sql
\\dt
\\d table_name
\\q
```

---

# Project Structure

```text
kuflow/
├── docs/
├── migrations/
├── proxy/
├── database/
├── scripts/
├── docker-compose.yml
└── README.md
```

---

# Workflow

Recommended workflow:

1. Update documentation.
2. Create or update migrations.
3. Implement Go code.
4. Test locally.
5. Commit changes.

---

# First-Time Setup

```bash
cd ~/kuflow

docker-compose up -d

docker ps
```

After this the development environment is ready.
