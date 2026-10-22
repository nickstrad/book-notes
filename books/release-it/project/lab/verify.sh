#!/usr/bin/env bash
# make verify LESSON=NN [REF=<git ref>] [REF_SERVICES="gateway control node"]
#
# Runs lab/verify/NN-*.sh against the stack. With REF, the services named in
# REF_SERVICES (default: gateway control node) are built from REF's tree, and
# so are deploy/haproxy.cfg and the migrations; everything that is harness
# (partner fakes, load generator, scenarios, report SQL, thresholds) comes
# from the working tree. That is how a lesson's script reproduces the flaw on
# the previous tag.
source "$(dirname "$0")/lib.sh"

lesson=${1:?usage: verify.sh NN}
script=$(ls "$LAB_DIR"/verify/"$lesson"-*.sh 2>/dev/null | head -n 1)
[[ -n $script ]] || { echo "no verify script for lesson $lesson" >&2; exit 2; }

log "building images from the working tree"
compose --profile lab build -q >/dev/null 2>&1 || compose --profile lab build

if [[ -n ${REF:-} ]]; then
  repo=$(git -C "$PROJECT_DIR" rev-parse --show-toplevel)
  prefix=$(git -C "$PROJECT_DIR" rev-parse --show-prefix)
  sha=$(git -C "$repo" rev-parse --short "$REF^{commit}")
  dir=/tmp/litefleet-ref/$sha
  if [[ ! -d $dir ]]; then
    mkdir -p "$dir"
    git -C "$repo" archive "$sha" "$prefix" | tar -x -C "$dir" --strip-components="$(tr -cd / <<<"$prefix" | wc -c)"
  fi
  for svc in ${REF_SERVICES:-gateway control node}; do
    log "building $svc from $REF ($sha)"
    docker build -q -f "$dir/deploy/Dockerfile" --build-arg CMD="$svc" -t "litefleet/$svc:ref" "$dir" >/dev/null
    case $svc in
      gateway) export GATEWAY_TAG=ref ;;
      control) export CONTROL_TAG=ref ;;
      node) export NODE_TAG=ref ;;
    esac
  done
  export LF_CONF="$dir/deploy"
  # Migrations, profiles and online migrations follow the tree control comes from.
  [[ " ${REF_SERVICES:-gateway control node} " == *" control "* ]] && export LF_TREE="$dir"
  echo "REF=$REF ($sha): ${REF_SERVICES:-gateway control node} built from REF; harness from the working tree"
fi

exec bash "$script"
