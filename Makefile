DATABASE_URL=postgres://proxy:proxy@localhost:5432/proxydb?sslmode=disable

.PHONY: run

run:
	go run ./cmd/kuflow

.PHONY: migrate-up

migrate-up:
	migrate \
	-path migrations \
	-database "$(DATABASE_URL)" \
	up

.PHONY: migrate-down

migrate-down:
	migrate \
	-path migrations \
	-database "$(DATABASE_URL)" \
	down

.PHONY: migrate-force

migrate-force:
	migrate \
	-path migrations \
	-database "$(DATABASE_URL)" \
	force 0