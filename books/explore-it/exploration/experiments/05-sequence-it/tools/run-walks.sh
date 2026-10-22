#!/usr/bin/env bash
# Background runner: BIAS=1 walks for the given seeds, 200 steps each; writes results/e05/walk-b<seed>.txt
source "$(dirname "$0")/e05env.sh"
cd "$(dirname "$0")/.."
for s in "$@"; do BIAS=1 bash walk.sh "$s" 200 > "$RES/walk-b$s.txt" 2>&1; done
echo done > "$RES/walk-b.done"
