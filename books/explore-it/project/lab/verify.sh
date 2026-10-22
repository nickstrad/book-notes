#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
lesson=${1:-00}
if [[ -n ${REF:-} && ${QUEUELET_REF_ACTIVE:-0} != 1 ]]; then
  root=$(git rev-parse --show-toplevel)
  scratch=$(mktemp -d /tmp/queuelet-ref.XXXXXX)
  trap 'docker compose -f "$scratch/books/explore-it/project/deploy/compose.yml" down -v >/dev/null 2>&1; rm -rf "$scratch"' EXIT
  git -C "$root" archive "$REF" books/explore-it/project | tar -x -C "$scratch"
  export COMPOSE_PROJECT_NAME=queuelet-ref QUEUELET_API_PORT=18210 QUEUELET_HOOK_PORT=18211 QUEUELET_PG_PORT=15443 QUEUELET_IMAGE_TAG=ref
  export QUEUELET_URL=http://localhost:18210 QUEUELET_HOOK_URL=http://localhost:18211 QUEUELET_REF_ACTIVE=1
  target="$scratch/books/explore-it/project"
  docker compose -f "$target/deploy/compose.yml" up -d --build
  bash lab/wait.sh
  # Run current instruments against the historical binaries.
  cp -R lab/. "$target/lab/"
  mkdir -p "$target/test/checks"
  cp -R test/checks/. "$target/test/checks/"
  (cd "$target"; REF="$REF" bash lab/verify.sh "$lesson")
  exit $?
fi
if [[ $lesson == 00 ]]; then
  bash lab/reset.sh
  go test ./...
  go test -tags checks ./test/checks -run L00 -count=1
  echo 'BASELINE HOLDS: enqueue→lease→ack→webhook; unit and happy-path checks pass'
  exit 0
fi
# Each fixer adds an ids file and immutable black-box repro scripts with its lesson.
idsfile="lab/repro/lesson-$lesson.ids"
[[ -f $idsfile ]] || { echo "No reproductions registered for lesson $lesson" >&2; exit 2; }
bash lab/reset.sh
present=0; absent=0
while read -r id; do
  [[ -n $id ]] || continue
  result=$(bash "lab/repro/$id.sh")
  printf '%s\n' "$result"
  if [[ $result == PRESENT:* ]]; then ((present+=1)); elif [[ $result == ABSENT:* ]]; then ((absent+=1)); else exit 2; fi
done < "$idsfile"
if [[ -n ${REF:-} ]]; then
  [[ $absent == 0 ]] || { echo 'Historical reproduction incomplete' >&2; exit 1; }
  echo "DEFECTS REPRODUCED: $present present"
else
  [[ $present == 0 ]] || exit 1
  go test -tags checks ./test/checks -run "L$lesson" -count=1
  echo "FIXED AND CHECKED: $absent absent; regression checks pass"
fi
