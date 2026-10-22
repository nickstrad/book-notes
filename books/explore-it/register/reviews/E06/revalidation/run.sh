#!/usr/bin/env bash
set -uo pipefail
cd /tmp/explore-it-revalidate06-claude/source
out=/tmp/explore-it-revalidate06-claude
date -u +%FT%TZ > "$out/start.txt"
timeout --signal=TERM 900s claude --model claude-opus-5-5 --effort high --tools Bash,Read,Write,Edit,Glob,Grep --allowedTools Bash,Read,Write,Edit,Glob,Grep --verbose --output-format stream-json -p < "$out/input.md" > "$out/run.jsonl" 2> "$out/stderr.log"
status=$?
date -u +%FT%TZ > "$out/end.txt"
printf '%s\n' "$status" > "$out/exit-status.txt"
printf 'Claude revalidation exit=%s\n' "$status"
exit "$status"
