#!/usr/bin/env bash
# make test-chaos — run the chaos test suite (test/chaos, build tag `chaos`)
# inside the Compose network against the running stack. With a directory
# argument, run that tree's clients instead (the verify uses it for REF).
# Prints `go test -json` lines.
source "$(dirname "$0")/lib.sh"
tree=${1:-$PROJECT_DIR}
docker run --rm --network litefleet_default \
  -v "$tree":/src -w /src \
  -v "$(go env GOMODCACHE 2>/dev/null || echo /root/go/pkg/mod)":/go/pkg/mod \
  -v litefleet-gocache:/root/.cache/go-build \
  -v "$RESULTS":/results:ro \
  -e CGO_ENABLED=0 golang:1.26-alpine \
  go test -tags chaos -count=1 -json ./test/chaos/
