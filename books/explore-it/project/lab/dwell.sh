#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
lesson=${1:?lesson required}
printf 'tag\tverdict\n'
for tag in explore-it/v0-inherited $(git tag -l 'explore-it/lesson-*' | sort); do
  number=${tag##*-}
  if [[ $tag == *lesson-* ]] && ((10#$number > 10#$lesson)); then continue; fi
  status=0
  report=$(QUEUELET_DWELL=1 REF="$tag" bash lab/verify.sh "$lesson" 2>&1) || status=$?
  if ((status != 0)); then printf '%s\tPROBE ERROR (exit %s)\n%s\n' "$tag" "$status" "$report"; exit "$status"; fi
  printf '%s\t%s\n' "$tag" "$(printf '%s\n' "$report" | sed -n '/^PRESENT:/p; /^ABSENT:/p; /^DEFECTS REPRODUCED:/p; /^FIXED AND CHECKED:/p')"
done
