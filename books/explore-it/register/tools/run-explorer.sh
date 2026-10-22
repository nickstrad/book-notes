#!/usr/bin/env bash
set -uo pipefail
base=$(python3 -c 'import json; print(json.load(open("/tmp/explore-it-coordinator/current.json"))["base"])')
cd "$base"
date -u +%FT%TZ > run-start.txt
timeout --signal=TERM 3600s claude --model opus --effort high --tools Bash,Read,Write,Edit,Glob,Grep --allowedTools Bash,Read,Write,Edit,Glob,Grep --verbose --output-format stream-json -p < input.md > run.jsonl 2> run.stderr
status=$?
date -u +%FT%TZ > run-end.txt
printf '%s\n' "$status" > exit-status.txt
printf 'Explorer exit=%s; log=%s/run.jsonl\n' "$status" "$base"
exit "$status"
