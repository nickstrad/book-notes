#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "$0")"
failed=0
for n in $(seq -w 1 27); do
  bash "D$n.sh" || failed=$((failed+1))
done
printf 'Proof failures: %d\n' "$failed"
[[ $failed == 0 ]]
