SHELL := /bin/bash
COMPOSE = docker compose -f deploy/compose.yml
.PHONY: up down seed reset verify dwell checks test logs qctl
up:
	$(COMPOSE) up -d --build
	@bash lab/wait.sh
down:
	$(COMPOSE) down -v --remove-orphans
seed:
	bash lab/reset.sh
reset: seed
verify:
	REF='$(REF)' bash lab/verify.sh $(or $(LESSON),00)
dwell:
	bash lab/dwell.sh $(LESSON)
checks:
	go test -tags checks ./test/checks -count=1
test:
	go test ./...
logs:
	$(COMPOSE) logs --tail=100
qctl:
	@go run ./cmd/qctl $(ARGS)
