# litefleet: make up / make seed / make verify LESSON=NN [REF=<tag>] / make down
COMPOSE := docker compose -f deploy/compose.yml
LESSON  ?= 00
export REF REF_SERVICES

.PHONY: up down build seed migrate backfill contract rollout godot verify calibrate audit report test test-chaos logs ps

build:
	$(COMPOSE) --profile lab build

up: build
	lab/up.sh

seed:
	$(COMPOSE) run --rm seedgen

migrate:
	lab/migrate.sh

backfill:
	$(COMPOSE) --profile jobs run --rm --no-deps migrate backfill

contract:
	$(COMPOSE) --profile jobs run --rm --no-deps migrate migrate -contract

rollout:
	lab/rollout.sh

godot:
	lab/godot.sh

down:
	$(COMPOSE) --profile lab --profile maintenance down -v --remove-orphans

verify:
	lab/verify.sh $(LESSON)

calibrate:
	lab/calibrate.sh

audit:
	lab/audit.sh

report:
	lab/report.sh $(CSV)

test:
	go test ./...

test-chaos:
	lab/chaos.sh | jq -r 'select(.Action == "output") | .Output' | grep -E 'MATRIX|^(--- |ok|FAIL|PASS)' 

logs:
	$(COMPOSE) logs -f --tail=50

ps:
	$(COMPOSE) ps
