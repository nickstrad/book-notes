#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
exec python3 probe08.py "$(basename "$0" .sh)"
