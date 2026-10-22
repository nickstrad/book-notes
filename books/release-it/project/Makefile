# litefleet: make up / make seed / make verify LESSON=NN [REF=<tag>] / make down
COMPOSE := docker compose -f deploy/compose.yml
LESSON  ?= 00
export REF REF_SERVICES

.PHONY: up down build seed verify calibrate audit report test logs ps

build:
	$(COMPOSE) --profile lab build

up: build
	lab/up.sh

seed:
	$(COMPOSE) run --rm seedgen

down:
	$(COMPOSE) --profile lab down -v --remove-orphans

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

logs:
	$(COMPOSE) logs -f --tail=50

ps:
	$(COMPOSE) ps
