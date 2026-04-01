.PHONY: migrate-up migrate-down infra-up infra-down run

migrate-up:
	scripts/migrate-up.sh

migrate-down:
	scripts/migrate-down.sh

infra-up:
	docker compose up -d

infra-down:
	docker compose down

run:
	scripts/run.sh
