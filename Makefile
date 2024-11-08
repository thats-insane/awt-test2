include .envrc

.PHONY: run/api
run/api:
	@echo 'Running products/reviews API...'
	@go run ./cmd/api -port=3000 -env=development -limiter-burst=5 -limiter-rps=2 -limiter-enabled=true -db-dsn=${PRODUCTSREVIEWS_DB_DSN}

.PHONY: db/psql
db/psql:
	psql ${PRODUCTSREVIEWS_DB_DSN}

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path=./migrations -database ${PRODUCTSREVIEWS_DB_DSN} up