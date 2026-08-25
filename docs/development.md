# KuFlow Deployment Guide

## Requirements

- Docker
- Docker Compose

## 1. Clone repository

git clone ...
cd kuflow

## 2. Configure environment

cp .env.production .env

Edit:

- DATABASE_URL
- POSTGRES_PASSWORD
- PROXY_UPSTREAMS

## 3. Start services

docker compose up -d --build

## 4. Verify containers

docker compose ps

Expected:
- postgres healthy
- kuflow running

## 5. View logs

docker logs -f kuflow

Expected:
- Database migrations applied
- Connected to PostgreSQL
- HTTP server started

## 6. Health check

curl http://SERVER_IP:8080/health

## 7. Create API key

curl -X POST ...

## 8. List API keys

curl http://SERVER_IP:8080/admin/api-keys

## 9. Stop services

docker compose down