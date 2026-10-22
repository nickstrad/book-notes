#!/usr/bin/env bash
# Lesson 11 — test harness. Runs the chaos suite (test/chaos) against the
# running stack: the gateway's node client under six Toxiproxy faults and a
# stale epoch, the node's blob-store client against four kinds of partner
# misbehaviour (the fake's /admin/mode), and the control breaker's open and
# recovery. With REF, the same tests (from the working tree) run against
# REF's clients.
source "$(dirname "$0")/../lib.sh"

reset_stack
tree=$LF_TREE
if [[ $tree != "$PROJECT_DIR" ]]; then
  mkdir -p "$tree/test"
  rm -rf "$tree/test/chaos" && cp -r "$PROJECT_DIR/test/chaos" "$tree/test/chaos"
fi
log "chaos suite against $( [[ $tree == "$PROJECT_DIR" ]] && echo 'the working tree' || echo "$REF" )'s clients"
out=$RESULTS/11-chaos.json
lab/chaos.sh "$tree" > "$out" 2>&1 || true

echo
echo "  client      misbehaviour     outcome        elapsed"
jq -r 'select(.Action == "output") | .Output' "$out" | grep -o 'MATRIX.*' | sed 's/^MATRIX /  /'
echo
jq -r 'select(.Test != null and (.Action == "pass" or .Action == "fail")) | "  \(.Action | ascii_upcase)  \(.Test)  (\(.Elapsed) s)"' "$out"
jq -r 'select(.Action == "output") | .Output' "$out" | grep -E '^\s+chaos_test.go:[0-9]+:' | grep -v MATRIX | sed 's/^/  /' || true
total=$(jq -r 'select(.Test != null and (.Action == "pass" or .Action == "fail")) | .Test' "$out" | wc -l)
failed=$(jq -r 'select(.Test != null and .Action == "fail") | .Test' "$out" | wc -l)
if ((total == 0)); then jq -r 'select(.Action == "output") | .Output' "$out" | tail -n 20; fi

check chaos_tests_run "$total" "==" 12
check chaos_tests_failed "$failed" "==" 0
verdict
