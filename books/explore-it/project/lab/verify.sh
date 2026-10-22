#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
lesson=${1:-00}
if [[ -n ${REF:-} && ${QUEUELET_REF_ACTIVE:-0} != 1 ]]; then
  root=$(git rev-parse --show-toplevel)
  scratch=$(mktemp -d /tmp/queuelet-ref.XXXXXX)
  original_project=$PWD
  target="$scratch/books/explore-it/project"
  cleanup_ref() {
    status=$?
    trap - EXIT
    set +e
    if ((status != 0)); then
      mkdir -p "$original_project/results"
      diagnostics=$(mktemp -d "$original_project/results/ref-failure.XXXXXX")
      if [[ -n $diagnostics ]]; then
        printf 'REF=%s\nlesson=%s\nexit_status=%s\n' "$REF" "$lesson" "$status" > "$diagnostics/run.txt"
        docker compose -f "$target/deploy/compose.yml" ps -a > "$diagnostics/compose-ps.log" 2>&1
        docker compose -f "$target/deploy/compose.yml" logs --no-color > "$diagnostics/compose.log" 2>&1
        printf 'REF verification failed (exit %s); diagnostics: %s\n' "$status" "$diagnostics" >&2
      else
        echo 'Could not create REF diagnostics directory; dumping Compose state to stderr' >&2
        docker compose -f "$target/deploy/compose.yml" ps -a >&2
        docker compose -f "$target/deploy/compose.yml" logs --no-color >&2
      fi
    fi
    docker compose -f "$target/deploy/compose.yml" down -v >/dev/null 2>&1
    rm -rf "$scratch"
    exit "$status"
  }
  trap cleanup_ref EXIT
  git -C "$root" archive "$REF" books/explore-it/project | tar -x -C "$scratch"
  export COMPOSE_PROJECT_NAME=queuelet-ref QUEUELET_API_PORT=18210 QUEUELET_HOOK_PORT=18211 QUEUELET_PG_PORT=15443 QUEUELET_IMAGE_TAG=ref
  export QUEUELET_URL=http://localhost:18210 QUEUELET_HOOK_URL=http://localhost:18211 QUEUELET_REF_ACTIVE=1
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
mapfile -t ids < "$idsfile"
expected=${#ids[@]}
((expected > 0)) || { echo "PROBE ERROR: empty reproduction list" >&2; exit 2; }
declare -A seen=()
for id in "${ids[@]}"; do
  [[ $id =~ ^([DU][0-9]{2}|[AT][0-9]+)$ ]] || { echo "PROBE ERROR: invalid reproduction ID $id" >&2; exit 2; }
  [[ ! -v seen[$id] ]] || { echo "PROBE ERROR: duplicate reproduction ID $id" >&2; exit 2; }
  seen[$id]=1
  result=$(bash "lab/repro/$id.sh" </dev/null)
  printf '%s\n' "$result"
  if [[ $result == "PRESENT: $id — "* ]]; then ((present+=1)); elif [[ $result == "ABSENT: $id — "* ]]; then ((absent+=1)); else echo "PROBE ERROR: invalid verdict for $id" >&2; exit 2; fi
done
[[ $((present+absent)) == "$expected" ]] || exit 2
controls=()
[[ $lesson == 07 ]] && controls=(D18 D19 D20)
[[ $lesson == 08 ]] && controls=(D21 D22 D23 D24)
# The missing-guard group (fixed in E01), now taken from the transition table.
[[ $lesson == 09 ]] && controls=(D12 D14 D18 D19 D20 D27)
if ((${#controls[@]})); then
  for control in "${controls[@]}"; do
    result=$(bash "lab/repro/$control.sh" </dev/null)
    printf '%s\n' "$result"
    [[ $result == "ABSENT: $control — "* ]] || { echo "PROBE ERROR: early-fixed control $control failed" >&2; exit 2; }
  done
fi
if [[ -n ${REF:-} ]]; then
  if [[ ${QUEUELET_DWELL:-0} == 1 ]]; then
    echo "DWELL: $present present; $absent absent"
    exit 0
  fi
  [[ $absent == 0 ]] || { echo 'Historical reproduction incomplete' >&2; exit 1; }
  echo "DEFECTS REPRODUCED: $present present"
else
  [[ $present == 0 ]] || { echo "VERIFICATION FAILED: $present present; $absent absent" >&2; exit 1; }
  go test -tags checks ./test/checks -run "L$lesson" -count=1
  echo "FIXED AND CHECKED: $absent absent; regression checks pass"
fi
